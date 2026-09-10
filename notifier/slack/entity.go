package slack

import "encoding/json"

// SlackMessage represents a message to be sent to Slack.
// It includes fields for the channel, text, username, icon URL, thread timestamp,
//
// - Channel: The Slack channel to send the message to (e.g., "#general").
// - Text: The main text of the message.
// - Username: The name of the user sending the message.
// - IconURL: A URL to an image to use as the icon for the message.
// - ThreadTS: The timestamp of the thread to reply to (if applicable).
// - Attachments: A list of attachments to include with the message.
type SlackMessage struct {
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	ThreadTS    string       `json:"thread_ts,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// ToString converts the SlackMessage struct into its JSON string representation.
// It returns the JSON string and an error if the marshaling process fails.
func (sm *SlackMessage) ToString() (string, error) {
	data, err := json.Marshal(sm)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Attachment represents a Slack message attachment.
// Attachments are used to add more context or information to a Slack message.
//
// Fields:
// - Color: A hex color code (e.g., "#36a64f") or a Slack named color (e.g., "good", "warning", "danger").
// - Title: The title of the attachment, displayed in bold.
// - TitleLink: A URL that turns the title into a clickable link.
// - Text: The main text of the attachment, supporting Markdown-style formatting.
// - Fields: A list of fields to display in a table-like format within the attachment.
// - Footer: A small footer text displayed at the bottom of the attachment.
// - AuthorName: The name of the author displayed at the top of the attachment.
// - AuthorIcon: A URL to an icon image displayed next to the author's name.
// - ImageURL: A URL to an image displayed within the attachment.
// - Timestamp: A stringified Unix timestamp to display as a date/time in the attachment footer.
type Attachment struct {
	Color      string      `json:"color,omitempty"`
	Title      string      `json:"title,omitempty"`
	TitleLink  string      `json:"title_link,omitempty"`
	Text       string      `json:"text,omitempty"`
	Fields     []Field     `json:"fields,omitempty"`
	Footer     string      `json:"footer,omitempty"`
	AuthorName string      `json:"author_name,omitempty"`
	AuthorIcon string      `json:"author_icon,omitempty"`
	ImageURL   string      `json:"image_url,omitempty"`
	Timestamp  json.Number `json:"ts,omitempty"` // Use stringified timestamp
}

// Field represents a Slack message field, which is a key-value pair
// displayed in a Slack message attachment. The Title is the name
// of the field, Value is the content of the field, and Short indicates
// whether the field should be displayed in a compact format.
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}
