package session

import (
	"encoding/json"
	"errors"
	"time"
)

var errInvalidPayload = errors.New("invalid session payload")

type jsonPayload struct {
	Deadline int64                  `json:"deadline"`
	Values   map[string]interface{} `json:"values"`
}

type jsonCodec struct{}

func (jsonCodec) Encode(deadline time.Time, values map[string]interface{}) ([]byte, error) {
	payload := jsonPayload{
		Deadline: deadline.Unix(),
		Values:   values,
	}
	return json.Marshal(payload)
}

func (jsonCodec) Decode(b []byte) (time.Time, map[string]interface{}, error) {
	if len(b) == 0 {
		return time.Time{}, map[string]interface{}{}, nil
	}

	var payload jsonPayload
	if err := json.Unmarshal(b, &payload); err != nil {
		return time.Time{}, nil, err
	}
	if payload.Deadline == 0 {
		return time.Time{}, nil, errInvalidPayload
	}
	if payload.Values == nil {
		payload.Values = map[string]interface{}{}
	}

	return time.Unix(payload.Deadline, 0).UTC(), payload.Values, nil
}
