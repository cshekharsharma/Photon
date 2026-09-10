// PublishSubscribeInterface is a common interface that has to be implemented
// by all cloud vendor specific implementions (ie AWS, Azure, GCP).
// This interface provides a cloud agnostic approach to the workflow, where consumers
// of publish-subscribe cloud workflow do not have to know the exact cloud vendor, can work
// with the inteface signature itself.
package contract

import (
	"context"

	"github.com/cshekharsharma/photon/cloud/entity/pubsub"
)

// PublishSubscribeInterface defines methods for a cloud notification service like AWS SNS.
type PublishSubscribeInterface interface {

	// Publish a message to a specified topic
	Publish(ctx context.Context, input *pubsub.PublishInput) (*pubsub.PublishResult, error)

	// Get topic attributes
	GetTopicAttributes(ctx context.Context, input *pubsub.GetTopicAttributesInput) (*pubsub.GetTopicAttributesResult, error)
}
