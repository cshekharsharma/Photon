package entity

// Message represents an interface that defines the behavior for message entities.
// Any type implementing this interface must provide a method to convert the message
// into its string representation.
//
// The ToString method is expected to return a string representation of the message
// along with an error if the conversion fails. This allows for flexible handling
// of message types while ensuring proper error management during the conversion process.
type Message interface {
	ToString() (string, error)
}
