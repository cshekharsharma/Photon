package slack

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cshekharsharma/photon/notifier/entity"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---
type MockSlackClient struct {
	mock.Mock
}

func (m *MockSlackClient) PostMessage(channel string, options ...slack.MsgOption) (string, string, error) {
	args := m.Called(channel, options)
	return args.String(0), args.String(1), args.Error(2)
}

// --- Tests ---

func TestNewSlackNotifier(t *testing.T) {
	req := &entity.Request{
		Token: "dummy-token",
	}
	n := NewSlackNotifier(req)
	assert.NotNil(t, n)
	assert.Equal(t, req.Token, n.request.Token)
	assert.Equal(t, req.DefaultChannel, n.request.DefaultChannel)
	assert.NotNil(t, n.client)
	assert.Equal(t, req.DefaultChannel, n.defaultCh)
}

func TestSlackNotifier_Send_Token_Success(t *testing.T) {
	mockClient := new(MockSlackClient)
	req := &entity.Request{
		Token:          "dummy-token",
		DefaultChannel: "#general",
	}
	n := newSlackNotifierWithClient(mockClient, req)

	msg := &SlackMessage{
		Text:    "Hello World",
		Channel: "#test",
		Attachments: []Attachment{
			{
				Title: "Test",
				Fields: []Field{
					{Title: "Env", Value: "Prod", Short: true},
				},
				Timestamp: json.Number("1234567890"),
			},
		},
	}

	mockClient.On("PostMessage", "#test", mock.Anything).Return("ts1", "ch1", nil)

	err := n.Send(msg)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSlackNotifier_Send_Token_FallbackToDefaultChannel(t *testing.T) {
	mockClient := new(MockSlackClient)
	req := &entity.Request{
		Token:          "dummy-token",
		DefaultChannel: "#fallback",
	}
	n := newSlackNotifierWithClient(mockClient, req)

	msg := &SlackMessage{Text: "No channel set"}

	mockClient.On("PostMessage", "#fallback", mock.Anything).Return("ts2", "ch2", nil)

	err := n.Send(msg)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSlackNotifier_Send_Token_Error(t *testing.T) {
	mockClient := new(MockSlackClient)
	n := newSlackNotifierWithClient(mockClient, &entity.Request{Token: "t", DefaultChannel: "#fallback"})

	msg := &SlackMessage{Channel: "#fail", Text: "Failing"}
	mockClient.On("PostMessage", "#fail", mock.Anything).Return("", "", errors.New("mock failure"))

	err := n.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock failure")
}

func TestSlackNotifier_Send_InvalidMessageType(t *testing.T) {
	n := newSlackNotifierWithClient(new(MockSlackClient), &entity.Request{})
	err := n.Send(nil) // invalid
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message type")
}

func TestSlackNotifier_Send_Webhook_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))

	defer ts.Close()

	req := &entity.Request{UseWebhook: true, WebhookURL: ts.URL}
	n := &SlackNotifier{request: req}

	msg := &SlackMessage{
		Text:    "via webhook",
		Channel: "#test",
		Attachments: []Attachment{
			{
				Title: "Test",
				Fields: []Field{
					{Title: "Env", Value: "Prod", Short: true},
				},
				Timestamp: json.Number("1234567890"),
			},
		},
	}

	err := n.Send(msg)
	assert.NoError(t, err)
}

func TestSlackNotifier_Send_Webhook_HTTPFailure(t *testing.T) {
	req := &entity.Request{UseWebhook: true, WebhookURL: "http://127.0.0.1:1"} // unreachable
	n := &SlackNotifier{request: req}
	msg := &SlackMessage{Text: "fail"}
	err := n.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send Slack webhook request")
}

func TestSlackNotifier_Send_Webhook_NonSuccessStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	req := &entity.Request{UseWebhook: true, WebhookURL: ts.URL}
	n := &SlackNotifier{request: req}
	msg := &SlackMessage{Text: "bad status"}
	err := n.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-success status")
}

func TestSlackNotifier_Send_WebhookURLMissing(t *testing.T) {
	request := &entity.Request{
		UseWebhook: true,
		WebhookURL: "", // Missing webhook URL
	}
	notifier := NewSlackNotifier(request)

	err := notifier.Send(&SlackMessage{
		Text: "This should fail due to missing webhook URL",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook URL is required")
}

func TestSlackNotifier_Send_Webhook_MarshalError(t *testing.T) {
	req := &entity.Request{UseWebhook: true, WebhookURL: "http://example.com"}
	n := &SlackNotifier{request: req}

	msg := &SlackMessage{
		Text: "bad payload",
		Attachments: []Attachment{
			{Timestamp: json.Number("not-a-number")},
		},
	}

	err := n.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal Slack webhook payload")
}

func TestSlackNotifier_Send_Webhook_CloseError(t *testing.T) {
	orig := postSlackWebhook
	defer func() { postSlackWebhook = orig }()

	postSlackWebhook = func(string, string, io.Reader) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       closeErrorBody{Reader: strings.NewReader("")},
		}, nil
	}

	n := &SlackNotifier{request: &entity.Request{
		UseWebhook: true,
		WebhookURL: "http://example.com",
	}}

	err := n.Send(&SlackMessage{Text: "close error"})
	assert.NoError(t, err)
}

type closeErrorBody struct {
	*strings.Reader
}

func (b closeErrorBody) Close() error {
	return errors.New("close failed")
}
