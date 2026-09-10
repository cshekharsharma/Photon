package contracts

import (
	"fmt"
	"strings"
)

type Channel string
type Priority string
type TemplateMode string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"

	PriorityHigh Priority = "high"
	PriorityLow  Priority = "low"

	TemplateModeResolve  TemplateMode = "resolve"
	TemplateModeOverride TemplateMode = "override"
)

type Recipient struct {
	Email     string `json:"email,omitempty"`
	PhoneE164 string `json:"phone_e164,omitempty"`
}

type Scope struct {
	AppScope     string `json:"app_scope,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	RestaurantID string `json:"restaurant_id,omitempty"`
	LocationID   string `json:"location_id,omitempty"`
	Locale       string `json:"locale,omitempty"`
}

type TemplateReference struct {
	Mode             TemplateMode `json:"mode"`
	Channel          string       `json:"channel,omitempty"`
	Purpose          string       `json:"purpose,omitempty"`
	SubjectTemplate  string       `json:"subject_template,omitempty"`
	TextTemplate     string       `json:"text_template,omitempty"`
	HTMLTemplate     string       `json:"html_template,omitempty"`
	FromAddress      string       `json:"from_address,omitempty"`
	ReplyToAddresses []string     `json:"reply_to_addresses,omitempty"`
	SenderName       string       `json:"sender_name,omitempty"`
}

type NotificationMessage struct {
	MessageID      string                 `json:"message_id"`
	IdempotencyKey string                 `json:"idempotency_key"`
	Priority       Priority               `json:"priority"`
	Channel        Channel                `json:"channel"`
	Purpose        string                 `json:"purpose"`
	ProviderHint   string                 `json:"provider_hint,omitempty"`
	Recipient      Recipient              `json:"recipient"`
	Scope          Scope                  `json:"scope"`
	Template       TemplateReference      `json:"template"`
	Variables      map[string]string      `json:"variables"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	SourceService  string                 `json:"source_service,omitempty"`
}

func (m *NotificationMessage) Validate() error {
	if m == nil {
		return fmt.Errorf("notification message is required")
	}
	if strings.TrimSpace(m.MessageID) == "" {
		return fmt.Errorf("message_id is required")
	}
	if strings.TrimSpace(m.IdempotencyKey) == "" {
		return fmt.Errorf("idempotency_key is required")
	}
	if strings.TrimSpace(m.Purpose) == "" {
		return fmt.Errorf("purpose is required")
	}
	switch m.Channel {
	case ChannelEmail, ChannelSMS:
	default:
		return fmt.Errorf("unsupported channel: %s", m.Channel)
	}
	switch m.Priority {
	case PriorityHigh, PriorityLow:
	default:
		return fmt.Errorf("unsupported priority: %s", m.Priority)
	}
	return nil
}
