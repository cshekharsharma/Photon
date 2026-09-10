package teams

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cshekharsharma/photon/notifier/entity"
	"github.com/cshekharsharma/photon/utils/rest"
)

var (
	marshalTeamsPayload = json.Marshal
	postTeamsWebhook    = http.Post
)

// TeamsNotifier sends messages to Microsoft Teams via Incoming Webhook.
// It implements the Notifier interface.
type TeamsNotifier struct {
	request *entity.Request
}

// NewTeamsNotifier creates a new TeamsNotifier with the provided request.
func NewTeamsNotifier(request *entity.Request) *TeamsNotifier {
	return &TeamsNotifier{request: request}
}

// Send sends a Teams message via webhook. It supports facts and actions.
// The message must be of type *TeamsMessage.
func (t *TeamsNotifier) Send(message entity.Message) error {
	if t.request == nil || t.request.WebhookURL == "" {
		return errors.New("webhook URL is required for Teams")
	}

	tMessage, ok := message.(*TeamsMessage)
	if !ok {
		return fmt.Errorf("invalid message type: expected *TeamsMessage, got %T", message)
	}

	card := buildCardStructure(tMessage)

	body, err := marshalTeamsPayload(card)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams payload: %w", err)
	}

	resp, err := postTeamsWebhook(t.request.WebhookURL, rest.ContentTypeJSON, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send Teams request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("teams webhook returned non-success status: %s", resp.Status)
	}

	return nil
}

func buildCardStructure(msg *TeamsMessage) map[string]interface{} {
	card := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"title":      msg.Title,
		"summary":    msg.Summary,
		"text":       msg.Text,
		"themeColor": msg.ThemeColor,
	}

	// Sections with facts
	if len(msg.Sections) > 0 {
		var sections []map[string]interface{}
		for _, section := range msg.Sections {
			facts := make([]map[string]string, 0)
			for _, f := range section.Facts {
				facts = append(facts, map[string]string{
					"name":  f.Name,
					"value": f.Value,
				})
			}

			sectionData := map[string]interface{}{
				"facts": facts,
			}

			if section.ActivityTitle != "" {
				sectionData["activityTitle"] = section.ActivityTitle
			}
			if section.ActivitySubtitle != "" {
				sectionData["activitySubtitle"] = section.ActivitySubtitle
			}
			if section.ActivityImage != "" {
				sectionData["activityImage"] = section.ActivityImage
			}

			sections = append(sections, sectionData)
		}
		card["sections"] = sections
	}

	// Potential actions
	if len(msg.PotentialAction) > 0 {
		var actions []map[string]interface{}
		for _, a := range msg.PotentialAction {
			action := map[string]interface{}{
				"@type":   a.Type,
				"name":    a.Name,
				"targets": make([]map[string]string, 0),
			}
			for _, target := range a.Targets {
				action["targets"] = append(action["targets"].([]map[string]string), map[string]string{
					"os":  target.OS,
					"uri": target.URI,
				})
			}
			actions = append(actions, action)
		}
		card["potentialAction"] = actions
	}

	return card
}
