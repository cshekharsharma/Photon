package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/cshekharsharma/photon/cloud/entity/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSNSClient struct {
	mock.Mock
}

func (m *MockSNSClient) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sns.PublishOutput), args.Error(1)
}

func (m *MockSNSClient) GetTopicAttributes(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sns.GetTopicAttributesOutput), args.Error(1)
}

func TestPublish(t *testing.T) {
	ctx := context.Background()
	mockSNS := new(MockSNSClient)
	ps := PublishSubscribe{SNSClient: mockSNS}

	input := &pubsub.PublishInput{
		TopicId: "valid-topic",
		Message: "Test message",
	}
	output := &sns.PublishOutput{
		MessageId: aws.String("12345"),
	}

	mockSNS.On("Publish", ctx, mock.AnythingOfType("*sns.PublishInput"), mock.Anything).Return(output, nil)

	result, err := ps.Publish(ctx, input)
	mockSNS.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "12345", result.MessageId)
}

func TestPublish_WithOptionalFields(t *testing.T) {
	ctx := context.Background()
	mockSNS := new(MockSNSClient)
	ps := PublishSubscribe{SNSClient: mockSNS}

	input := &pubsub.PublishInput{
		Message:          "Test message",
		MessageGroupId:   "group",
		MessageDedupeId:  "dedupe",
		PhoneNumber:      "+1234567890",
		Subject:          "subject",
		DeviceEndpointId: "device-arn",
	}
	output := &sns.PublishOutput{
		MessageId: aws.String("999"),
	}

	mockSNS.On("Publish", ctx, mock.AnythingOfType("*sns.PublishInput"), mock.Anything).Return(output, nil)

	result, err := ps.Publish(ctx, input)
	mockSNS.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, "999", result.MessageId)
}

func TestPublish_WithMessageAttributes(t *testing.T) {
	ctx := context.Background()
	mockSNS := new(MockSNSClient)
	ps := PublishSubscribe{SNSClient: mockSNS}

	input := &pubsub.PublishInput{
		TopicId: "valid-topic",
		Message: "Test message",
		MessageAttributes: map[string]string{
			"eventType":        "test.completed",
			"cshekharsharmaId": "ABC123",
		},
	}
	output := &sns.PublishOutput{MessageId: aws.String("with-attrs")}

	mockSNS.On("Publish", ctx, mock.MatchedBy(func(in *sns.PublishInput) bool {
		return in.MessageAttributes["eventType"].StringValue != nil &&
			*in.MessageAttributes["eventType"].StringValue == "test.completed" &&
			in.MessageAttributes["cshekharsharmaId"].StringValue != nil &&
			*in.MessageAttributes["cshekharsharmaId"].StringValue == "ABC123"
	}), mock.Anything).Return(output, nil)

	result, err := ps.Publish(ctx, input)
	mockSNS.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, "with-attrs", result.MessageId)
}

func TestPublish_SkipsEmptyMessageAttributes(t *testing.T) {
	ctx := context.Background()
	mockSNS := new(MockSNSClient)
	ps := PublishSubscribe{SNSClient: mockSNS}

	input := &pubsub.PublishInput{
		TopicId: "valid-topic",
		Message: "Test message",
		MessageAttributes: map[string]string{
			"eventType": "test.completed",
			"":          "empty-key",
			"emptyVal":  "",
		},
	}
	output := &sns.PublishOutput{MessageId: aws.String("with-filtered-attrs")}

	mockSNS.On("Publish", ctx, mock.MatchedBy(func(in *sns.PublishInput) bool {
		_, hasEmptyKey := in.MessageAttributes[""]
		_, hasEmptyValue := in.MessageAttributes["emptyVal"]
		return len(in.MessageAttributes) == 1 &&
			!hasEmptyKey &&
			!hasEmptyValue &&
			*in.MessageAttributes["eventType"].StringValue == "test.completed"
	}), mock.Anything).Return(output, nil)

	result, err := ps.Publish(ctx, input)
	mockSNS.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, "with-filtered-attrs", result.MessageId)
}

func TestPublish_Error(t *testing.T) {
	ctx := context.Background()
	mockSNS := new(MockSNSClient)
	ps := PublishSubscribe{SNSClient: mockSNS}

	mockSNS.On("Publish", ctx, mock.AnythingOfType("*sns.PublishInput"), mock.Anything).Return((*sns.PublishOutput)(nil), errors.New("publish error"))

	_, err := ps.Publish(ctx, &pubsub.PublishInput{Message: "x"})
	assert.Error(t, err)
}

func TestGetTopicAttributes(t *testing.T) {
	ctx := context.Background()
	mockSNS := new(MockSNSClient)
	ps := PublishSubscribe{SNSClient: mockSNS}

	input := &pubsub.GetTopicAttributesInput{
		TopicId: "valid-topic",
	}
	output := &sns.GetTopicAttributesOutput{
		Attributes: map[string]string{"DisplayName": "Test Topic"},
	}

	mockSNS.On("GetTopicAttributes", ctx, mock.AnythingOfType("*sns.GetTopicAttributesInput"), mock.Anything).Return(output, nil)

	result, err := ps.GetTopicAttributes(ctx, input)
	mockSNS.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Topic", result.Attributes["DisplayName"])
}

func TestGetTopicAttributes_Error(t *testing.T) {
	ctx := context.Background()
	mockSNS := new(MockSNSClient)
	ps := PublishSubscribe{SNSClient: mockSNS}

	mockSNS.On("GetTopicAttributes", ctx, mock.AnythingOfType("*sns.GetTopicAttributesInput"), mock.Anything).Return((*sns.GetTopicAttributesOutput)(nil), errors.New("attr error"))

	_, err := ps.GetTopicAttributes(ctx, &pubsub.GetTopicAttributesInput{TopicId: "x"})
	assert.Error(t, err)
}
