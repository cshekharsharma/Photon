package contract

import (
	"context"

	"github.com/cshekharsharma/photon/cloud/entity/messagequeue"
)

// MessageQueueInterface interface is a common interface that has to be implemented
// by all cloud vendor specific implementions (ie AWS, Azure, GCP).
// This interface provides a cloud agnostic approach to the workflow, where consumers
// of message queue cloud workflow do not have to know the exact cloud vendor, can work
// with the inteface signature itself.
type MessageQueueInterface interface {

	// Send single message to cloud message queue
	SendMessage(ctx context.Context, queueUrl string, messageBody string, messageGroupId string) (*messagequeue.SendMessageResult, error)

	// Receive multiple messages from cloud message queue
	ReceiveMessages(ctx context.Context, queueUrl string, maxItems int32) ([]*messagequeue.ReceiveMessageResult, error)

	// Delete single message from cloud message queue
	DeleteMessage(ctx context.Context, queueUrl string, receiptHandle string) (*messagequeue.DeleteMessageResult, error)

	// Change visibility timeout for a received message
	ChangeMessageVisibility(ctx context.Context, queueUrl string, receiptHandle string, visibilityTimeoutSeconds int32) (*messagequeue.ChangeMessageVisibilityResult, error)

	// Get message queue attributes
	GetQueueAttributes(ctx context.Context, queueUrl string) (*messagequeue.GetQueueAttributesResult, error)
}
