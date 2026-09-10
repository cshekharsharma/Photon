package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/cshekharsharma/photon/cloud/entity/pubsub"
	"github.com/cshekharsharma/photon/utils/types"
)

// awsSnsClientInterface defines an interface for interacting with AWS Simple Notification Service (SNS).
// It abstracts the SNS operations to facilitate easier testing and decoupling from the actual AWS SDK client implementations.
// The interface includes methods to publish messages to SNS topics and to retrieve attributes of SNS topics.
type awsSnsClientInterface interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	GetTopicAttributes(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error)
}

// PublishSubscribe is a struct that encapsulates an AWS SNS client to provide simplified access to SNS functionalities.
// It uses an implementation of the awsSnsClientInterface to perform operations, allowing for flexibility in testing and
// operations management.
//
// Attributes:
// - SNSClient: An instance of awsSnsClientInterface which is used to interact with AWS SNS.
type PublishSubscribe struct {
	SNSClient awsSnsClientInterface
}

// Publish attempts to send a message to an SNS topic or directly as an SMS.
// It takes input of type *pubsub.PublishInput which includes various message attributes
// such as the topic ID, message content, group ID, deduplication ID, phone number,
// subject, and a device endpoint ID. It wraps these attributes into an SNS PublishInput
// and sends the message using the SNS client.
//
// Parameters:
// - input: Contains all necessary information for publishing the message, including:
//   - TopicId: The ARN of the SNS topic to which the message will be sent.
//   - Message: The content of the message to be published.
//   - MessageGroupId: (Optional) The ID of the message group for FIFO topics.
//     Messages with the same MessageGroupId are delivered in order. This field is required
//     for FIFO topics.
//   - MessageDedupeId: (Optional) The ID for deduplication of messages in FIFO topics.
//     If two messages with the same MessageDedupeId are sent within a 5-minute interval,
//     only one will be delivered. This field is also required for FIFO topics.
//   - PhoneNumber: (Optional) The phone number to send an SMS message to, in E.164 format.
//   - Subject: (Optional) The subject line for email notifications sent via SNS.
//   - DeviceEndpointId: (Optional) The endpoint ID for targeting a specific device.
//
// Returns:
// - *pubsub.PublishResult: Contains the message ID assigned by AWS SNS if the publish operation is successful.
// - error: An error object that indicates an issue during the publish operation, such as problems with input validation
//   or network issues.
//
// Conditions:
// - If both TopicId and PhoneNumber are specified, an error will be returned.
// - If both TopicId and DeviceEndpointId are specified, an error will be returned.
// - If both PhoneNumber and DeviceEndpointId are specified, an error will be returned.
// - Only one of TopicId, PhoneNumber, or DeviceEndpointId can be specified at a time.
// - MessageGroupId and MessageDedupeId must be specified together if using a FIFO topic.
//
// Usage:
// - This method is ideal for applications requiring reliable delivery of notifications, alerts,
//   or other short messages to SNS topics or directly to SMS recipients.

func (ps *PublishSubscribe) Publish(ctx context.Context, input *pubsub.PublishInput) (*pubsub.PublishResult, error) {
	var snsPublishInput = sns.PublishInput{
		Message: aws.String(input.Message),
	}

	if !types.IsEmpty(input.TopicId) {
		snsPublishInput.TargetArn = aws.String(input.TopicId)
	}
	if !types.IsEmpty(input.MessageGroupId) {
		snsPublishInput.MessageGroupId = aws.String(input.MessageGroupId)
	}
	if !types.IsEmpty(input.MessageDedupeId) {
		snsPublishInput.MessageDeduplicationId = aws.String(input.MessageDedupeId)
	}
	if !types.IsEmpty(input.PhoneNumber) {
		snsPublishInput.PhoneNumber = aws.String(input.PhoneNumber)
	}
	if !types.IsEmpty(input.Subject) {
		snsPublishInput.Subject = aws.String(input.Subject)
	}
	if !types.IsEmpty(input.DeviceEndpointId) {
		snsPublishInput.TargetArn = aws.String(input.DeviceEndpointId)
	}
	if len(input.MessageAttributes) > 0 {
		snsPublishInput.MessageAttributes = make(map[string]snstypes.MessageAttributeValue, len(input.MessageAttributes))
		for key, value := range input.MessageAttributes {
			if types.IsEmpty(key) || types.IsEmpty(value) {
				continue
			}
			snsPublishInput.MessageAttributes[key] = snstypes.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(value),
			}
		}
	}

	awsResult, err := ps.SNSClient.Publish(ctx, &snsPublishInput)

	if err != nil {
		return nil, err
	}

	result := new(pubsub.PublishResult)
	result.MessageId = *awsResult.MessageId

	return result, nil
}

// GetTopicAttributes fetches the attributes of a specific SNS topic. It requires an input of type *pubsub.GetTopicAttributesInput,
// which should include the topic ID. This method wraps the topic ID into an SNS GetTopicAttributesInput and retrieves
// the attributes using the SNS client.
//
// Parameters:
// - input: Specifies the topic ID for which attributes are being retrieved.
//
// Returns:
// - *pubsub.GetTopicAttributesResult: Contains a map of the topic attributes if the operation is successful.
// - error: An error object that may occur during the attribute retrieval process, such as if the topic does not exist or network issues.
//
// Usage:
// - Useful for managing and monitoring SNS topics, especially when adjusting topic configurations or auditing topic usage and permissions.
func (ps *PublishSubscribe) GetTopicAttributes(ctx context.Context, input *pubsub.GetTopicAttributesInput) (*pubsub.GetTopicAttributesResult, error) {
	awsResult, err := ps.SNSClient.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: aws.String(input.TopicId),
	})

	if err != nil {
		return nil, err
	}

	result := new(pubsub.GetTopicAttributesResult)
	result.Attributes = awsResult.Attributes

	return result, nil
}
