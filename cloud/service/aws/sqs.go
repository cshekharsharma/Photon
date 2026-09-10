package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/cshekharsharma/photon/cloud/entity/messagequeue"
	typeutils "github.com/cshekharsharma/photon/utils/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// FifoQueueSuffix is the suffix used for FIFO (First-In-First-Out) queues in SQS.
const FifoQueueSuffix = ".fifo"

// awsSQSClientInterface defines the interface for interacting with AWS Simple Queue Service (SQS).
// This interface abstracts the queue messaging operations to enable easier testing and decoupling from the AWS SDK.
// It provides methods for sending, receiving, and deleting messages, as well as retrieving queue attributes.
type awsSQSClientInterface interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error)
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

// MessageQueue represents a message queue implementation specifically tailored
// for AWS's Simple Queue Service (SQS). It provides methods to interact with
// the SQS service, such as sending, receiving, deleting messages and fetching queue attributes.
type MessageQueue struct {
	SqsClient awsSQSClientInterface
}

var (
	getCryptoSafeRandomString = typeutils.GetCryptoSafeRandomString
	getRandomString           = typeutils.GetRandomString
)

// SendMessage sends a single message to an SQS queue identified by its URL.
//
// Parameters:
//   - queueUrl: The URL of the target SQS queue.
//   - messageBody: The content of the message to be sent.
//   - messageGroupId: A unique identifier for the message group. Relevant for FIFO (First-In-First-Out) queues.
//
// Returns:
//   - A pointer to a SendMessageResult structure containing the details of the sent message.
//   - An error object if any issues are encountered during the sending process. Otherwise, nil.
func (mq MessageQueue) SendMessage(ctx context.Context, queueUrl string, messageBody string, messageGroupId string) (*messagequeue.SendMessageResult, error) {
	dedupeID := ""
	if strings.HasSuffix(queueUrl, FifoQueueSuffix) {
		randomstr, err := getCryptoSafeRandomString(32)
		if err != nil {
			randomstr = getRandomString(32)
		}
		dedupeID = fmt.Sprintf("dedup-%s", randomstr)
	}
	return mq.sendMessage(ctx, queueUrl, messageBody, messageGroupId, dedupeID)
}

// SendMessageWithDedupe sends a single message using an explicit FIFO
// deduplication id. It is used by higher-level publishers that already own an
// idempotency key.
func (mq MessageQueue) SendMessageWithDedupe(ctx context.Context, queueUrl string, messageBody string, messageGroupId string, dedupeID string) (*messagequeue.SendMessageResult, error) {
	return mq.sendMessage(ctx, queueUrl, messageBody, messageGroupId, dedupeID)
}

func (mq MessageQueue) sendMessage(ctx context.Context, queueUrl string, messageBody string, messageGroupId string, dedupeID string) (*messagequeue.SendMessageResult, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueUrl),
		MessageBody: aws.String(messageBody),
	}

	if strings.HasSuffix(queueUrl, FifoQueueSuffix) {
		input.MessageGroupId = aws.String(messageGroupId)
		input.MessageDeduplicationId = aws.String(dedupeID)
	}

	awsResult, err := mq.SqsClient.SendMessage(ctx, input)

	if err != nil {
		return nil, err
	}

	result := new(messagequeue.SendMessageResult)

	result.MessageId = *awsResult.MessageId
	result.MessageBodyMd5 = *awsResult.MD5OfMessageBody

	return result, nil
}

// ReceiveMessages fetches messages from an SQS queue identified by its URL. It retrieves messages up to
// the specified maxItems limit.
//
// Parameters:
//   - queueUrl: The URL of the SQS queue from which messages are to be received.
//   - maxItems: The maximum number of messages to retrieve in a single call.
//
// Returns:
//   - A slice of pointers to ReceiveMessageResult structures, each containing details of a received message.
//   - An error object if issues are encountered during the retrieval. Otherwise, nil.
func (mq MessageQueue) ReceiveMessages(ctx context.Context, queueUrl string, maxItems int32) ([]*messagequeue.ReceiveMessageResult, error) {
	awsResults, err := mq.SqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: maxItems,
		WaitTimeSeconds:     20, // Enable long polling to reduce empty receives and CPU spin.
	})

	if err != nil {
		return nil, err
	}

	var results = []*messagequeue.ReceiveMessageResult{}

	if len(awsResults.Messages) > 0 {
		for i := 0; i < len(awsResults.Messages); i++ {
			result := new(messagequeue.ReceiveMessageResult)

			result.MessageId = *awsResults.Messages[i].MessageId
			result.MessageBody = *awsResults.Messages[i].Body
			result.MessageBodyMd5 = *awsResults.Messages[i].MD5OfBody
			result.ReceiptHandle = *awsResults.Messages[i].ReceiptHandle

			results = append(results, result)
		}
	}

	return results, nil
}

// DeleteMessage removes a message from an SQS queue identified by its URL.
//
// Parameters:
//   - queueUrl: The URL of the SQS queue from which the message should be deleted.
//   - receiptHandle: A unique identifier associated with the act of receiving the message.
//
// Returns:
//   - A pointer to a DeleteMessageResult structure containing details related to the deletion operation.
//   - An error object if any issues are encountered during the deletion process. Otherwise, nil.
func (mq MessageQueue) DeleteMessage(ctx context.Context, queueUrl string, receiptHandle string) (*messagequeue.DeleteMessageResult, error) {
	awsResult, err := mq.SqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: aws.String(receiptHandle),
	})

	if err != nil {
		return nil, err
	}

	result := new(messagequeue.DeleteMessageResult)
	result.ResultMetadata = awsResult.ResultMetadata

	return result, nil
}

// ChangeMessageVisibility updates the visibility timeout for a received SQS message.
func (mq MessageQueue) ChangeMessageVisibility(ctx context.Context, queueUrl string, receiptHandle string, visibilityTimeoutSeconds int32) (*messagequeue.ChangeMessageVisibilityResult, error) {
	awsResult, err := mq.SqsClient.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(queueUrl),
		ReceiptHandle:     aws.String(receiptHandle),
		VisibilityTimeout: visibilityTimeoutSeconds,
	})

	if err != nil {
		return nil, err
	}

	result := new(messagequeue.ChangeMessageVisibilityResult)
	result.ResultMetadata = awsResult.ResultMetadata

	return result, nil
}

// GetQueueAttributes fetches the attributes of an SQS queue identified by its URL. It retrieves
// all available attributes for the given queue.
//
// Parameters:
//   - queueUrl: The URL of the SQS queue whose attributes are to be fetched.
//
// Returns:
//   - A pointer to a GetQueueAttributesResult structure containing the retrieved queue attributes.
//   - An error object if any issues are encountered during the retrieval process. Otherwise, nil.
func (mq MessageQueue) GetQueueAttributes(ctx context.Context, queueUrl string) (*messagequeue.GetQueueAttributesResult, error) {
	awsResult, err := mq.SqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueUrl),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})

	if err != nil {
		return nil, err
	}

	result := new(messagequeue.GetQueueAttributesResult)

	result.ResultMetadata = awsResult.ResultMetadata
	result.Attributes = awsResult.Attributes

	return result, nil
}
