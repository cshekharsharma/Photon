// Package pubsub provides structures for managing messages and topics within a publish-subscribe messaging system.
package pubsub

// PublishResult contains the result of a message publish operation to a topic.
// It primarily includes the unique identifier of the message once it has been successfully published.
type PublishResult struct {
	MessageId string // MessageId is the unique identifier assigned to the successfully published message.
}

// GetTopicAttributesResult holds the results of a request to retrieve attributes for a specific topic.
// It contains a map of attribute names to their corresponding values, detailing various properties of the topic.
type GetTopicAttributesResult struct {
	Attributes map[string]string // Attributes is a map where each key-value pair represents the name and value of an attribute of the topic.
}
