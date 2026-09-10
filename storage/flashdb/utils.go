package flashdb

import (
	"maps"
	"math"

	"github.com/cshekharsharma/photon/utils/system"
)

var getInMemorySizeInBytesHook = system.GetInMemorySizeInBytes

// clone returns a shallow copy of the provided map. If the input map is nil,
// it returns nil. The function copies all key-value pairs from the original
// map to a new map, but does not perform deep copying of the values.
func clone(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	copy := make(map[string]any, len(m))
	maps.Copy(copy, m)

	return copy
}

// addInt adds the given delta to the value cur, which can be of type int, int32, int64, or float64.
// It returns the result and a boolean indicating success. For int/int32 we do
// pre-addition boundary checks to avoid silent wraparound on overflow.
func addInt(cur any, delta int64) (any, bool) {
	switch x := cur.(type) {
	case int:
		// Pre-check overflow/underflow against int bounds on this arch.
		if delta > 0 && int64(x) > int64(math.MaxInt)-delta {
			return nil, false
		}
		if delta < 0 && int64(x) < int64(math.MinInt)-delta {
			return nil, false
		}
		return int(int64(x) + delta), true

	case int32:
		if delta > 0 && int64(x) > int64(math.MaxInt32)-delta {
			return nil, false
		}
		if delta < 0 && int64(x) < int64(math.MinInt32)-delta {
			return nil, false
		}
		return int32(int64(x) + delta), true

	case int64:
		// int64 add is exact in Go; we still detect overflow via bounds if desired,
		// but since both operands are int64-compatible, we won’t overflow here.
		return x + delta, true

	case float64:
		return x + float64(delta), true

	default:
		return nil, false
	}
}

// computeSize returns the byte size for storing (key, value), including
// key bytes and an estimated per-record overhead.
func computeSize(key string, value map[string]any) int64 {
	const recordOverhead = int64(96) // struct + LRU node + map entry (approx)
	vSize := safeDeepSize(value)
	return int64(len(key)) + recordOverhead + vSize
}

// safeDeepSize wraps the user-provided reflect-based estimator with panic
// protection and a conservative fallback.
func safeDeepSize(v any) (size int64) {
	defer func() {
		if r := recover(); r != nil {
			size = 0
		}
	}()

	size, err := getInMemorySizeInBytesHook(v)
	if err != nil {
		return 0
	}

	if size < 0 {
		return 0
	}

	return size
}
