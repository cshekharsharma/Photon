package meter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cshekharsharma/photon/telemetry"
	"go.opentelemetry.io/otel/metric"
)

var (
	counterMap   sync.Map // map[string]metric.Int64Counter — safely caches counters by name
	histogramMap sync.Map // map[string]metric.Float64Histogram — safely caches histograms by name
	gaugeMap     sync.Map // map[string]*atomic.Int64 — stores values for manually updated gauges

	telemetryGet = telemetry.Get

	testFloat64CallbacksHook func([]metric.Float64Callback)
	testInt64CallbacksHook   func([]metric.Int64Callback)

	RecordCounterFunc         = RecordCounter
	RecordHistogramFunc       = RecordHistogram
	IncrementErrorCounterFunc = IncrementErrorCounter
	IncrementGaugeFunc        = IncrementGauge
	DecrementGaugeFunc        = DecrementGauge
	RecordLatencyFunc         = RecordLatency
	ObservableGaugeFunc       = InitObservableGauge
	ObservableGaugeValueFunc  = SetGaugeValue
	BatchCountersFunc         = AddBatchCounters
)

// RecordCounter increments a named counter metric by a specified value.
// Automatically initializes and caches the counter instrument.
func RecordCounter(name string, value int64, attrMap map[string]interface{}) {
	meter := telemetryGet().Meter

	val, ok := counterMap.Load(name)
	if !ok {
		counter, err := meter.Int64Counter(name)
		if err != nil {
			return
		}
		actual, _ := counterMap.LoadOrStore(name, counter)
		val = actual
	}

	val.(metric.Int64Counter).Add(
		context.Background(),
		value,
		metric.WithAttributes(telemetry.ConvertToAttributes(attrMap)...),
	)
}

// AddBatchCounters allows batch incrementing of multiple counters by name,
// sharing the same attribute set for all.
func AddBatchCounters(namedCounts map[string]int64, attrMap map[string]interface{}) {
	meter := telemetryGet().Meter
	attrs := metric.WithAttributes(telemetry.ConvertToAttributes(attrMap)...)

	for name, value := range namedCounts {
		val, ok := counterMap.Load(name)
		if !ok {
			counter, err := meter.Int64Counter(name)
			if err != nil {
				continue // skip this metric on failure
			}
			actual, _ := counterMap.LoadOrStore(name, counter)
			val = actual
		}
		val.(metric.Int64Counter).Add(context.Background(), value, attrs)
	}
}

// IncrementErrorCounter increments a named counter for errors.
// Automatically attaches error type and message to attributes.
func IncrementErrorCounter(name string, err error, attrMap map[string]interface{}) {
	if err == nil {
		return
	}
	attrMap["error.type"] = fmt.Sprintf("%T", err)
	attrMap["error.message"] = err.Error()

	RecordCounter(name, 1, attrMap)
}

// RecordHistogram records a float64 value to a histogram metric.
// Useful for tracking things like latencies, sizes, durations, etc.
func RecordHistogram(name string, value float64, attrMap map[string]interface{}) {
	meter := telemetryGet().Meter

	val, ok := histogramMap.Load(name)
	if !ok {
		hist, err := meter.Float64Histogram(name)
		if err != nil {
			return
		}
		actual, _ := histogramMap.LoadOrStore(name, hist)
		val = actual
	}

	val.(metric.Float64Histogram).Record(
		context.Background(),
		value,
		metric.WithAttributes(telemetry.ConvertToAttributes(attrMap)...),
	)
}

// RecordLatency is a convenience function to record latency (in milliseconds)
// since the provided start time using a histogram.
func RecordLatency(name string, start time.Time, attrMap map[string]interface{}) {
	duration := time.Since(start).Seconds() * 1000 // convert to ms
	RecordHistogram(name, duration, attrMap)
}

// InitObservableGauge registers a gauge that observes values via a callback function.
// The callback is called automatically during OTEL collection/scrape.
func InitObservableGauge(name string, callback func(context.Context) float64) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("metric name is required")
	}
	meter := telemetryGet().Meter

	cb := func(ctx context.Context, observer metric.Float64Observer) error {
		value := callback(ctx)
		observer.Observe(value)
		return nil
	}

	if testFloat64CallbacksHook != nil {
		cfg := metric.NewFloat64ObservableGaugeConfig(metric.WithFloat64Callback(cb))
		testFloat64CallbacksHook(cfg.Callbacks())
	}

	_, err := meter.Float64ObservableGauge(
		name,
		metric.WithFloat64Callback(cb),
	)

	return err
}

// SetGaugeValue manually sets the value for a named gauge metric.
// It stores the value internally and registers an observable gauge if needed.
func SetGaugeValue(name string, value int64, attrMap map[string]interface{}) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("metric name is required")
	}
	val, ok := gaugeMap.Load(name)
	if !ok {
		var atomicVal atomic.Int64
		atomicVal.Store(value)

		actual, _ := gaugeMap.LoadOrStore(name, &atomicVal)
		return registerObservableGauge(name, actual.(*atomic.Int64), attrMap)
	}

	val.(*atomic.Int64).Store(value)
	return nil
}

// registerObservableGauge registers an Int64ObservableGauge for a manually tracked gauge.
// It uses an atomic value internally and reports it via the callback.
func registerObservableGauge(name string, ref *atomic.Int64, attrMap map[string]interface{}) error {
	meter := telemetryGet().Meter

	cb := func(ctx context.Context, obs metric.Int64Observer) error {
		obs.Observe(ref.Load(), metric.WithAttributes(telemetry.ConvertToAttributes(attrMap)...))
		return nil
	}

	if testInt64CallbacksHook != nil {
		cfg := metric.NewInt64ObservableGaugeConfig(metric.WithInt64Callback(cb))
		testInt64CallbacksHook(cfg.Callbacks())
	}

	_, err := meter.Int64ObservableGauge(
		name,
		metric.WithInt64Callback(cb),
	)

	if err != nil {
		return err
	}

	return nil
}

// IncrementGauge adds +1 to a named gauge.
func IncrementGauge(name string, attrMap map[string]interface{}) error {
	val, ok := gaugeMap.Load(name)
	if !ok {
		var counter atomic.Int64
		counter.Store(1)
		actual, _ := gaugeMap.LoadOrStore(name, &counter)
		return registerObservableGauge(name, actual.(*atomic.Int64), attrMap)
	}

	val.(*atomic.Int64).Add(1)
	return nil
}

// DecrementGauge subtracts -1 from a named gauge.
func DecrementGauge(name string) {
	val, ok := gaugeMap.Load(name)
	if ok {
		val.(*atomic.Int64).Add(-1)
	}
}
