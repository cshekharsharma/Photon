package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cshekharsharma/photon/notifier/entity"
	"github.com/cshekharsharma/photon/utils/rest"
	"github.com/slack-go/slack"
)

type SlackClient interface {
	PostMessage(channel string, options ...slack.MsgOption) (string, string, error)
}

type SlackNotifier struct {
	request   *entity.Request
	client    SlackClient
	defaultCh string
}

var postSlackWebhook = http.Post

// NewSlackNotifier creates a SlackNotifier with API token-based client.
func NewSlackNotifier(request *entity.Request) *SlackNotifier {
	return &SlackNotifier{
		request:   request,
		client:    slack.New(request.Token),
		defaultCh: request.DefaultChannel,
	}
}

// newSlackNotifierWithClient is useful for mocking Slack client in unit tests.
func newSlackNotifierWithClient(client SlackClient, request *entity.Request) *SlackNotifier {
	return &SlackNotifier{
		request:   request,
		client:    client,
		defaultCh: request.DefaultChannel,
	}
}

// Send sends message to Slack via either API or webhook.
// // The message must be of type *SlackMessage.
func (s *SlackNotifier) Send(message entity.Message) error {
	if _, ok := message.(*SlackMessage); !ok {
		return fmt.Errorf("invalid message type, expecting *SlackMessage, got %T", message)
	}
	slackMessage := message.(*SlackMessage)

	if s.request.UseWebhook {
		return s.sendViaWebhook(slackMessage)
	}
	return s.sendViaToken(slackMessage)
}

// --- TOKEN BASED IMPLEMENTATION ---
func (s *SlackNotifier) sendViaToken(message *SlackMessage) error {
	channel := message.Channel
	if channel == "" {
		channel = s.defaultCh
	}

	attachments := make([]slack.Attachment, len(message.Attachments))
	for i, a := range message.Attachments {
		fields := make([]slack.AttachmentField, len(a.Fields))
		for j, f := range a.Fields {
			fields[j] = slack.AttachmentField{
				Title: f.Title,
				Value: f.Value,
				Short: f.Short,
			}
		}

		attachments[i] = slack.Attachment{
			Color:      a.Color,
			Title:      a.Title,
			TitleLink:  a.TitleLink,
			Text:       a.Text,
			Footer:     a.Footer,
			AuthorName: a.AuthorName,
			AuthorIcon: a.AuthorIcon,
			ImageURL:   a.ImageURL,
			Fields:     fields,
		}

		if a.Timestamp.String() != "" {
			attachments[i].Ts = a.Timestamp
		}
	}

	_, _, err := s.client.PostMessage(
		channel,
		slack.MsgOptionText(message.Text, false),
		slack.MsgOptionAttachments(attachments...),
		slack.MsgOptionTS(message.ThreadTS),
	)

	if err != nil {
		return fmt.Errorf("sending message via Slack API failed: %w", err)
	}
	return nil
}

// --- WEBHOOK BASED IMPLEMENTATION ---
func (s *SlackNotifier) sendViaWebhook(message *SlackMessage) error {
	if s.request.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required when UseWebhook is true")
	}

	payload := map[string]interface{}{
		"text":     message.Text,
		"username": message.Username,
		"icon_url": message.IconURL,
	}

	if len(message.Attachments) > 0 {
		payload["attachments"] = message.Attachments
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack webhook payload: %w", err)
	}

	resp, err := postSlackWebhook(s.request.WebhookURL, rest.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send Slack webhook request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("slack webhook returned non-success status: %s", resp.Status)
	}

	return nil
}
