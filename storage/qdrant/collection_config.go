package qdrant

import (
	"encoding/json"
	"errors"
)

var unmarshalCollectionJSONHook = json.Unmarshal

//
// ---------- Enums & constants ----------
//

type Distance string

const (
	DistanceCosine    Distance = "Cosine"
	DistanceDot       Distance = "Dot"
	DistanceEuclid    Distance = "Euclid"
	DistanceManhattan Distance = "Manhattan"
)

//
// ---------- Public API ----------
//

// CollectionConfig represents the most commonly used Qdrant collection settings.
// It is opinionated but flexible.
//
// Rule:
// - Typed fields for the common 80%
// - Extras for forward compatibility / advanced configs
type CollectionConfig struct {
	Vectors VectorsConfig `json:"vectors"`

	// Optional tuning
	HNSW       *HNSWConfig       `json:"hnsw_config,omitempty"`
	Optimizers *OptimizersConfig `json:"optimizers_config,omitempty"`

	// Distributed / storage knobs
	ShardNumber       int   `json:"shard_number,omitempty"`
	ReplicationFactor int   `json:"replication_factor,omitempty"`
	OnDiskPayload     *bool `json:"on_disk_payload,omitempty"`

	// Escape hatch for unsupported / future fields
	Extras map[string]any `json:"-"`
}

//
// ---------- Vectors ----------
//

// VectorsConfig supports either:
// - single unnamed vector
// - named vectors
type VectorsConfig struct {
	Single *VectorParams           `json:"-"`
	Named  map[string]VectorParams `json:"-"`
}

type VectorParams struct {
	Size     int      `json:"size"`
	Distance Distance `json:"distance"`
}

func (v VectorsConfig) MarshalJSON() ([]byte, error) {
	if v.Single != nil && len(v.Named) > 0 {
		return nil, errors.New("qdrant: cannot specify both single and named vectors")
	}
	if v.Single == nil && len(v.Named) == 0 {
		return nil, errors.New("qdrant: vectors config is required")
	}

	if v.Single != nil {
		return json.Marshal(v.Single)
	}
	return json.Marshal(v.Named)
}

//
// ---------- Index / optimizer configs ----------
//

type HNSWConfig struct {
	M           int `json:"m,omitempty"`
	EfConstruct int `json:"ef_construct,omitempty"`
}

type OptimizersConfig struct {
	DefaultSegmentNumber int `json:"default_segment_number,omitempty"`
}

//
// ---------- Defaults ----------
//

type Defaults struct {
	Distance Distance
	HNSW     HNSWConfig
	Optim    OptimizersConfig
}

var DefaultCollectionDefaults = Defaults{
	Distance: DistanceCosine,
	HNSW:     HNSWConfig{M: 16, EfConstruct: 128},
	Optim:    OptimizersConfig{DefaultSegmentNumber: 2},
}

//
// ---------- Defaulting & serialization ----------
//

// ApplyDefaults fills missing values without overriding user intent.
func (c *CollectionConfig) ApplyDefaults(d Defaults) error {
	// vectors: mandatory
	if c.Vectors.Single == nil && len(c.Vectors.Named) == 0 {
		return errors.New("qdrant: vectors config is required")
	}

	// single vector defaults
	if c.Vectors.Single != nil {
		if c.Vectors.Single.Size <= 0 {
			return errors.New("qdrant: vector size must be > 0")
		}
		if c.Vectors.Single.Distance == "" {
			c.Vectors.Single.Distance = d.Distance
		}
	}

	// named vectors defaults
	for name, vp := range c.Vectors.Named {
		if vp.Size <= 0 {
			return errors.New("qdrant: named vector '" + name + "' size must be > 0")
		}
		if vp.Distance == "" {
			vp.Distance = d.Distance
			c.Vectors.Named[name] = vp
		}
	}

	// framework defaults (only if user didn't specify)
	if c.HNSW == nil {
		tmp := d.HNSW
		c.HNSW = &tmp
	}
	if c.Optimizers == nil {
		tmp := d.Optim
		c.Optimizers = &tmp
	}

	return nil
}

// ToJSON builds the final request body for Qdrant,
// merging typed config + Extras safely.
func (c CollectionConfig) ToJSON() (map[string]any, error) {
	cc := c // copy

	if err := cc.ApplyDefaults(DefaultCollectionDefaults); err != nil {
		return nil, err
	}

	b, err := json.Marshal(cc)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := unmarshalCollectionJSONHook(b, &m); err != nil {
		return nil, err
	}

	// merge Extras (typed config wins)
	for k, v := range cc.Extras {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}

	return m, nil
}
