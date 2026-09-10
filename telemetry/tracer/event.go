package tracer

import "time"

// EventOptions holds the options for creating an event in tracer span.
// It includes the event name, attributes, and options for stack trace and timestamp.
type EventOptions struct {
	Name           string
	Attributes     map[string]interface{}
	WithStackTrace bool
	WithTimestamp  time.Time
}

// NewEventOptions creates a new EventOptions instance with the provided name, attributes, and options.
func NewEventOptions(
	name string,
	attrib map[string]interface{},
	withStackTrace bool,
	withTimeStamp time.Time,
) *EventOptions {

	return &EventOptions{
		Name:           name,
		Attributes:     attrib,
		WithStackTrace: withStackTrace,
		WithTimestamp:  withTimeStamp,
	}
}
