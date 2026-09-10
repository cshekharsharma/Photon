// Package messagequeue provides structures and tools for interacting with a message queue system.
package messagequeue

// SendMessageResult holds the results of a message send operation.
// It includes the unique identifier for the message and an MD5 hash of the message body
// to verify the integrity of the message data.
type SendMessageResult struct {
	MessageId      string // Unique identifier of the sent message.
	MessageBodyMd5 string // MD5 hash of the message body.
}

// ReceiveMessageResult represents the result of a message retrieval operation from the message queue.
// It includes details such as the message ID, the actual message body, a hash of the message body
// for data integrity verification, and a receipt handle that is used to delete the message.
type ReceiveMessageResult struct {
	MessageId      string // Unique identifier of the received message.
	MessageBody    string // The actual body of the message.
	MessageBodyMd5 string // MD5 hash of the message body.
	ReceiptHandle  string // Receipt handle of the message, used for further operations like deletion.
}

// DeleteMessageResult encapsulates the result of a message deletion operation.
// It contains metadata related to the result of the deletion process which might
// include status codes, timestamps, or other relevant information, depending on the implementation.
type DeleteMessageResult struct {
	ResultMetadata interface{} // Metadata associated with the deletion operation.
}

// ChangeMessageVisibilityResult encapsulates the result of changing a message visibility timeout.
type ChangeMessageVisibilityResult struct {
	ResultMetadata interface{} // Metadata associated with the visibility change operation.
}

// GetQueueAttributesResult holds the results of an operation to fetch attributes of a message queue.
// It contains a map of attributes describing various aspects of the queue such as size, capacity,
// and other operational parameters.
type GetQueueAttributesResult struct {
	ResultMetadata interface{}       // Metadata associated with fetching the queue attributes.
	Attributes     map[string]string // A map of attribute names and their corresponding values.
}
