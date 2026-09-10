package teams

import "encoding/json"

var marshalTeamsMessage = json.Marshal

type TeamsMessage struct {
	Type            string        `json:"@type"`                     // Always "MessageCard"
	Summary         string        `json:"summary"`                   // Required: Plain-text summary
	ThemeColor      string        `json:"themeColor,omitempty"`      // Optional: Color for the card (hexadecimal)
	Title           string        `json:"title,omitempty"`           // Title for the card
	Text            string        `json:"text,omitempty"`            // Main content (supports basic markdown)
	Sections        []Section     `json:"sections,omitempty"`        // Array of sections
	PotentialAction []TeamsAction `json:"potentialAction,omitempty"` // Array of actions
}

func (tm *TeamsMessage) ToString() (string, error) {
	data, err := marshalTeamsMessage(tm)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type Section struct {
	ActivityTitle    string `json:"activityTitle,omitempty"`
	ActivitySubtitle string `json:"activitySubtitle,omitempty"`
	ActivityImage    string `json:"activityImage,omitempty"`
	Facts            []Fact `json:"facts,omitempty"`
	Text             string `json:"text,omitempty"` // Additional markdown content
}

type Fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TeamsAction struct {
	Type    string   `json:"@type"`
	Name    string   `json:"name"`
	Targets []Target `json:"targets"`
}

type Target struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}
