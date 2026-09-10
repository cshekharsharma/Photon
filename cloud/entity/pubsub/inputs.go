// Package pubsub provides structures for publishing messages to topics and managing topic attributes in a messaging system.
package pubsub

// PublishInput contains the parameters required for publishing a message to a specific topic.
// It includes various identifiers that can control message routing and delivery properties,
// such as deduplication and targeting specific devices or phone numbers.
type PublishInput struct {
	Message           string            // Message is the content of the message to be published.
	TopicId           string            // TopicId identifies the topic to which the message will be published.
	MessageGroupId    string            // MessageGroupId is an identifier used for message grouping in the target topic (useful in ordered delivery scenarios).
	MessageDedupeId   string            // MessageDedupeId is an identifier for message deduplication. This prevents publishing duplicate messages within a certain time frame.
	MessageAttributes map[string]string // MessageAttributes are string attributes used for SNS subscription filtering and routing.
	DeviceEndpointId  string            // DeviceEndpointId identifies a specific device endpoint for targeted message delivery.
	PhoneNumber       string            // PhoneNumber is used when the message needs to be sent to a specific phone number.
	Subject           string            // Subject is a brief description or summary of the message content.
}

// GetTopicAttributesInput contains the parameters for retrieving attributes of a specific topic.
// This structure is typically used in administrative tasks such as monitoring and configuring topics.
type GetTopicAttributesInput struct {
	TopicId string // TopicId identifies the topic for which attributes are being retrieved.
}
