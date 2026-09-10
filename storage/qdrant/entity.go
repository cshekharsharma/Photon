package qdrant

import (
	"encoding/json"
	"strconv"
	"time"
)

// ConnectionConfig holds the configuration parameters required to establish a connection
// to a Qdrant server.
type ConnectionConfig struct {
	BaseURL string // e.g. "http://localhost:6333"
	APIKey  string // optional

	Timeout         time.Duration // per-request hard timeout (http.Client)
	DialTimeout     time.Duration // TCP dial timeout
	IdleConnTimeout time.Duration // idle conn timeout

	// Retry
	RetryMaxAttempts int           // total attempts incl. first (default 3)
	RetryBaseDelay   time.Duration // default 150ms
	RetryMaxDelay    time.Duration // default 2s
	RetryJitter      float64       // 0..1 (default 0.2)
}

const (
	PointIDTypeString int = 1
	PointIDTypeInt    int = 2
)

// PointID supports string or integer IDs.
type PointID struct {
	kind int // 1 string, 2 int
	s    string
	i    int64
}

func PointIDString(v string) PointID {
	return PointID{
		kind: PointIDTypeString,
		s:    v,
	}
}

func PointIDInt(v int64) PointID {
	return PointID{
		kind: PointIDTypeInt,
		i:    v,
	}
}

func (id PointID) MarshalJSON() ([]byte, error) {
	switch id.kind {
	case PointIDTypeString:
		return json.Marshal(id.s)
	case PointIDTypeInt:
		return json.Marshal(id.i)
	default:
		return nil, ErrInvalidPointID
	}
}

func (id *PointID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*id = PointIDString(s)
		return nil
	}

	c := data[0]
	if (c >= '0' && c <= '9') || c == '-' {
		i, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return ErrInvalidPointID
		}
		*id = PointIDInt(i)
		return nil
	}

	return ErrInvalidPointID
}

type Point struct {
	ID      PointID        `json:"id"`
	Vector  []float32      `json:"vector,omitempty"`  // single-vector baseline
	Payload map[string]any `json:"payload,omitempty"` // Qdrant payload
}

type ScoredPoint struct {
	ID      PointID        `json:"id"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload,omitempty"`
	Vector  []float32      `json:"vector,omitempty"`
}

type CollectionInfo struct {
	Status string `json:"status"` // can extend this struct if needed
}

// ----------- Request structs -------------

type UpsertPointsRequest struct {
	Collection string
	Points     []Point
	Wait       bool
}

type DeletePointsRequest struct {
	Collection string
	IDs        []PointID
	Wait       bool
}

type SearchRequest struct {
	Collection     string
	Vector         []float32
	Limit          int
	Offset         int
	WithPayload    bool
	WithVector     bool
	ScoreThreshold *float32

	// Filter is raw Qdrant filter JSON. Keep flexible initially.
	Filter json.RawMessage
}

type CreateCollectionRequest struct {
	Collection string
	Config     CollectionConfig
}
