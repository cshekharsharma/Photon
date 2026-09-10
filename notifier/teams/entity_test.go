package teams

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeamsMessage_ToString_Empty(t *testing.T) {
	msg := TeamsMessage{}
	jsonStr, err := msg.ToString()
	assert.NoError(t, err)
	assert.Equal(t, `{"@type":"","summary":""}`, jsonStr)
}

func TestTeamsMessage_ToString_MarshalError(t *testing.T) {
	orig := marshalTeamsMessage
	defer func() { marshalTeamsMessage = orig }()

	marshalTeamsMessage = func(v interface{}) ([]byte, error) {
		return nil, errors.New("marshal failed")
	}

	msg := TeamsMessage{}
	_, err := msg.ToString()
	assert.EqualError(t, err, "marshal failed")
}
