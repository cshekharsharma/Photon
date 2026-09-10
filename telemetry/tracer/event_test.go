package tracer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEventOptions_FullValues(t *testing.T) {
	attr := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	timestamp := time.Now()

	opts := NewEventOptions("event_name", attr, true, timestamp)

	assert.Equal(t, "event_name", opts.Name)
	assert.Equal(t, attr, opts.Attributes)
	assert.Equal(t, true, opts.WithStackTrace)
	assert.Equal(t, timestamp, opts.WithTimestamp)
}

func TestNewEventOptions_EmptyAttributes(t *testing.T) {
	timestamp := time.Now()

	opts := NewEventOptions("event_empty", nil, false, timestamp)

	assert.Equal(t, "event_empty", opts.Name)
	assert.Nil(t, opts.Attributes)
	assert.False(t, opts.WithStackTrace)
	assert.Equal(t, timestamp, opts.WithTimestamp)
}

func TestNewEventOptions_EmptyName(t *testing.T) {
	opts := NewEventOptions("", nil, false, time.Time{})

	assert.Equal(t, "", opts.Name)
	assert.Nil(t, opts.Attributes)
	assert.False(t, opts.WithStackTrace)
	assert.True(t, opts.WithTimestamp.IsZero())
}
