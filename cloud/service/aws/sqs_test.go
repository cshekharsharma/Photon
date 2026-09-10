package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSQSClient struct {
	mock.Mock
	sqs.Client
}

func (m *MockSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.ChangeMessageVisibilityOutput), args.Error(1)
}

func (m *MockSQSClient) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.GetQueueAttributesOutput), args.Error(1)
}

func TestSendMessage(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}
	messageID := "msg-123"
	md5Body := "md5-body"

	mockSQS.On("SendMessage", mock.Anything, mock.Anything).Return(&sqs.SendMessageOutput{
		MessageId:        aws.String(messageID),
		MD5OfMessageBody: aws.String(md5Body),
	}, nil)

	result, err := mq.SendMessage(context.Background(), "url", "Hello, World!", "group-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, messageID, result.MessageId)
	assert.Equal(t, md5Body, result.MessageBodyMd5)

	mockSQS.AssertExpectations(t)
}

func TestSendMessage_FifoDedupFallback(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}

	oldCrypto := getCryptoSafeRandomString
	oldRandom := getRandomString
	getCryptoSafeRandomString = func(uint8) (string, error) { return "", errors.New("fail") }
	getRandomString = func(int64) string { return "fallback" }
	defer func() {
		getCryptoSafeRandomString = oldCrypto
		getRandomString = oldRandom
	}()

	mockSQS.On("SendMessage", mock.Anything, mock.MatchedBy(func(in *sqs.SendMessageInput) bool {
		return in.MessageGroupId != nil && in.MessageDeduplicationId != nil &&
			*in.MessageDeduplicationId == "dedup-fallback"
	})).Return(&sqs.SendMessageOutput{
		MessageId:        aws.String("id"),
		MD5OfMessageBody: aws.String("md5"),
	}, nil)

	_, err := mq.SendMessage(context.Background(), "queue.fifo", "body", "group-1")
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

func TestSendMessageWithDedupe_UsesExplicitDedupeID(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}

	mockSQS.On("SendMessage", mock.Anything, mock.MatchedBy(func(in *sqs.SendMessageInput) bool {
		return in.MessageGroupId != nil && *in.MessageGroupId == "group-1" &&
			in.MessageDeduplicationId != nil && *in.MessageDeduplicationId == "idem-1"
	})).Return(&sqs.SendMessageOutput{
		MessageId:        aws.String("id"),
		MD5OfMessageBody: aws.String("md5"),
	}, nil)

	_, err := mq.SendMessageWithDedupe(context.Background(), "queue.fifo", "body", "group-1", "idem-1")
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

func TestSendMessage_Error(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}

	mockSQS.On("SendMessage", mock.Anything, mock.Anything).Return((*sqs.SendMessageOutput)(nil), errors.New("send error"))

	_, err := mq.SendMessage(context.Background(), "url", "body", "")
	assert.Error(t, err)
}

func TestReceiveMessages(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}

	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything).Return(&sqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				MessageId:     aws.String("msg-001"),
				Body:          aws.String("Message Body"),
				MD5OfBody:     aws.String("md5-body"),
				ReceiptHandle: aws.String("handle-001"),
			},
		},
	}, nil)

	results, err := mq.ReceiveMessages(context.Background(), "url", 1)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "msg-001", results[0].MessageId)

	mockSQS.AssertExpectations(t)
}

func TestReceiveMessages_EmptyAndError(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}

	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything).Return(&sqs.ReceiveMessageOutput{
		Messages: []types.Message{},
	}, nil).Once()

	results, err := mq.ReceiveMessages(context.Background(), "url", 1)
	assert.NoError(t, err)
	assert.Len(t, results, 0)

	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything).Return((*sqs.ReceiveMessageOutput)(nil), errors.New("recv error")).Once()
	_, err = mq.ReceiveMessages(context.Background(), "url", 1)
	assert.Error(t, err)
}

func TestDeleteMessageSuccess(t *testing.T) {
	ctx := context.TODO()
	mockSQS := new(MockSQSClient)
	messageQueue := MessageQueue{SqsClient: mockSQS}
	queueUrl := "http://example.com/queue"
	receiptHandle := "receipt-handle-123"

	mockSQS.On("DeleteMessage", ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: aws.String(receiptHandle),
	}).Return(&sqs.DeleteMessageOutput{}, nil)

	result, err := messageQueue.DeleteMessage(ctx, queueUrl, receiptHandle)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockSQS.AssertExpectations(t)
}

func TestDeleteMessage_Error(t *testing.T) {
	mockSQS := new(MockSQSClient)
	messageQueue := MessageQueue{SqsClient: mockSQS}
	mockSQS.On("DeleteMessage", mock.Anything, mock.Anything).Return((*sqs.DeleteMessageOutput)(nil), errors.New("delete error"))

	_, err := messageQueue.DeleteMessage(context.Background(), "url", "rh")
	assert.Error(t, err)
}

func TestChangeMessageVisibility(t *testing.T) {
	ctx := context.TODO()
	mockSQS := new(MockSQSClient)
	messageQueue := MessageQueue{SqsClient: mockSQS}
	queueUrl := "http://example.com/queue"
	receiptHandle := "receipt-handle-123"

	mockSQS.On("ChangeMessageVisibility", ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(queueUrl),
		ReceiptHandle:     aws.String(receiptHandle),
		VisibilityTimeout: 90,
	}).Return(&sqs.ChangeMessageVisibilityOutput{}, nil)

	result, err := messageQueue.ChangeMessageVisibility(ctx, queueUrl, receiptHandle, 90)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockSQS.AssertExpectations(t)
}

func TestChangeMessageVisibility_Error(t *testing.T) {
	mockSQS := new(MockSQSClient)
	messageQueue := MessageQueue{SqsClient: mockSQS}
	mockSQS.On("ChangeMessageVisibility", mock.Anything, mock.Anything).Return((*sqs.ChangeMessageVisibilityOutput)(nil), errors.New("visibility error"))

	_, err := messageQueue.ChangeMessageVisibility(context.Background(), "url", "rh", 90)
	assert.Error(t, err)
}

func TestGetQueueAttributes(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}

	mockSQS.On("GetQueueAttributes", mock.Anything, mock.Anything).Return(&sqs.GetQueueAttributesOutput{
		Attributes: map[string]string{"ApproximateNumberOfMessages": "10"},
	}, nil)

	result, err := mq.GetQueueAttributes(context.Background(), "url")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "10", result.Attributes["ApproximateNumberOfMessages"])

	mockSQS.AssertExpectations(t)
}

func TestGetQueueAttributes_Error(t *testing.T) {
	mockSQS := new(MockSQSClient)
	mq := MessageQueue{SqsClient: mockSQS}
	mockSQS.On("GetQueueAttributes", mock.Anything, mock.Anything).Return((*sqs.GetQueueAttributesOutput)(nil), errors.New("attr error"))

	_, err := mq.GetQueueAttributes(context.Background(), "url")
	assert.Error(t, err)
}
