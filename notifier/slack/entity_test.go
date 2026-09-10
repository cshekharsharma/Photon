package slack

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlackMessage_ToString_Empty(t *testing.T) {
	msg := SlackMessage{}
	jsonStr, err := msg.ToString()
	assert.NoError(t, err)
	assert.Equal(t, `{}`, jsonStr)
}

func TestSlackMessage_ToString_MarshalError(t *testing.T) {
	msg := SlackMessage{
		Attachments: []Attachment{
			{Timestamp: json.Number("not-a-number")},
		},
	}

	_, err := msg.ToString()
	assert.Error(t, err)
}
