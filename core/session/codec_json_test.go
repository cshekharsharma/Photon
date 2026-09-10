package session

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONCodec_RoundTrip(t *testing.T) {
	codec := jsonCodec{}
	deadline := time.Now().Add(time.Hour).UTC()
	values := map[string]interface{}{"a": "b"}

	b, err := codec.Encode(deadline, values)
	assert.NoError(t, err)

	decodedDeadline, decodedValues, err := codec.Decode(b)
	assert.NoError(t, err)
	assert.Equal(t, deadline.Unix(), decodedDeadline.Unix())
	assert.Equal(t, values, decodedValues)
}

func TestJSONCodec_DecodeEmpty(t *testing.T) {
	codec := jsonCodec{}
	deadline, values, err := codec.Decode(nil)
	assert.NoError(t, err)
	assert.True(t, deadline.IsZero())
	assert.Empty(t, values)
}

func TestJSONCodec_DecodeInvalidJSON(t *testing.T) {
	codec := jsonCodec{}
	_, _, err := codec.Decode([]byte("nope"))
	assert.Error(t, err)
}

func TestJSONCodec_DecodeInvalidPayload(t *testing.T) {
	codec := jsonCodec{}
	_, _, err := codec.Decode([]byte(`{"deadline":0,"values":{}}`))
	assert.ErrorIs(t, err, errInvalidPayload)
}

func TestJSONCodec_DecodeNilValues(t *testing.T) {
	codec := jsonCodec{}
	deadline := time.Now().Add(time.Hour).UTC().Unix()
	payload := []byte(`{"deadline":`)
	payload = append(payload, []byte(fmt.Sprintf("%d", deadline))...)
	payload = append(payload, []byte(`}`)...)

	decodedDeadline, decodedValues, err := codec.Decode(payload)
	assert.NoError(t, err)
	assert.Equal(t, deadline, decodedDeadline.Unix())
	assert.Empty(t, decodedValues)
}
