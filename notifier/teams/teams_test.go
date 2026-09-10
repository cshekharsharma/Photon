package teams

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cshekharsharma/photon/notifier/entity"
	"github.com/stretchr/testify/assert"
)

// Mock message implementing entity.Message
type mockTeamsMessage struct{}

func (m *mockTeamsMessage) ToString() (string, error) {
	return "", nil
}

func TestTeamsNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), `"@type":"MessageCard"`)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req := &entity.Request{
		WebhookURL: server.URL,
	}
	notifier := NewTeamsNotifier(req)

	msg := &TeamsMessage{
		Title:      "Deployment Update",
		Summary:    "New build deployed",
		Text:       "The service was successfully deployed.",
		ThemeColor: "0076D7",
		Sections: []Section{{
			Facts:            []Fact{{Name: "Status", Value: "Success"}},
			ActivityTitle:    "Deployment Details",
			ActivitySubtitle: "Build #1234",
			ActivityImage:    "https://example.com/image.png",
		}},
		PotentialAction: []TeamsAction{{
			Type: "OpenUri",
			Name: "View Logs",
			Targets: []Target{{
				OS:  "default",
				URI: "https://example.com/logs",
			}},
		}},
	}

	err := notifier.Send(msg)
	assert.NoError(t, err)
}

func TestTeamsNotifier_Send_InvalidMessageType(t *testing.T) {
	req := &entity.Request{WebhookURL: "http://example.com"}
	notifier := NewTeamsNotifier(req)

	err := notifier.Send(&mockTeamsMessage{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message type")
}

func TestTeamsNotifier_Send_HTTPRequestError(t *testing.T) {
	req := &entity.Request{WebhookURL: "http://localhost:9999"} // assume nothing is running
	notifier := NewTeamsNotifier(req)

	msg := &TeamsMessage{
		Title:   "Test",
		Summary: "Testing",
		Text:    "This should cause HTTP error",
	}

	err := notifier.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send Teams request")
}

func TestTeamsNotifier_Send_NonSuccessHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	req := &entity.Request{WebhookURL: server.URL}
	notifier := NewTeamsNotifier(req)

	msg := &TeamsMessage{
		Title:   "Forbidden Test",
		Summary: "This should fail",
		Text:    "Simulating 403 response",
	}

	err := notifier.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "teams webhook returned non-success status: 403 Forbidden")
}

func TestTeamsNotifier_Send_NoWebhook(t *testing.T) {
	req := &entity.Request{WebhookURL: ""}
	notifier := NewTeamsNotifier(req)

	msg := &TeamsMessage{Text: "Should fail"}
	err := notifier.Send(msg)
	assert.EqualError(t, err, "webhook URL is required for Teams")
}

func TestTeamsNotifier_Send_MarshalError(t *testing.T) {
	orig := marshalTeamsPayload
	defer func() { marshalTeamsPayload = orig }()

	marshalTeamsPayload = func(v interface{}) ([]byte, error) {
		return nil, errors.New("payload marshal error")
	}

	req := &entity.Request{WebhookURL: "http://example.com"}
	notifier := NewTeamsNotifier(req)

	msg := &TeamsMessage{
		Title:   "Test",
		Summary: "Summary",
		Text:    "Body",
	}

	err := notifier.Send(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal Teams payload")
}

func TestTeamsNotifier_Send_CloseError(t *testing.T) {
	orig := postTeamsWebhook
	defer func() { postTeamsWebhook = orig }()

	postTeamsWebhook = func(string, string, io.Reader) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       closeErrorBody{Reader: strings.NewReader("")},
		}, nil
	}

	notifier := NewTeamsNotifier(&entity.Request{WebhookURL: "http://example.com"})
	err := notifier.Send(&TeamsMessage{Text: "close error"})
	assert.NoError(t, err)
}

type closeErrorBody struct {
	*strings.Reader
}

func (b closeErrorBody) Close() error {
	return errors.New("close failed")
}
