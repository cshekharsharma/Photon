package qdrant

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestVectorsConfigMarshalJSON_Errors(t *testing.T) {
	v := VectorsConfig{
		Single: &VectorParams{Size: 1, Distance: DistanceCosine},
		Named:  map[string]VectorParams{"a": {Size: 1, Distance: DistanceDot}},
	}
	if _, err := v.MarshalJSON(); err == nil {
		t.Fatalf("expected error when both single and named vectors are set")
	}

	v = VectorsConfig{}
	if _, err := v.MarshalJSON(); err == nil {
		t.Fatalf("expected error when vectors config is empty")
	}
}

func TestVectorsConfigMarshalJSON_SingleAndNamed(t *testing.T) {
	v := VectorsConfig{Single: &VectorParams{Size: 3, Distance: DistanceCosine}}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal single: %v", err)
	}
	var single VectorParams
	if err := json.Unmarshal(b, &single); err != nil {
		t.Fatalf("unmarshal single: %v", err)
	}
	if single.Size != 3 || single.Distance != DistanceCosine {
		t.Fatalf("unexpected single vector params: %+v", single)
	}

	v = VectorsConfig{Named: map[string]VectorParams{"text": {Size: 5, Distance: DistanceDot}}}
	b, err = json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal named: %v", err)
	}
	var named map[string]VectorParams
	if err := json.Unmarshal(b, &named); err != nil {
		t.Fatalf("unmarshal named: %v", err)
	}
	vp := named["text"]
	if vp.Size != 5 || vp.Distance != DistanceDot {
		t.Fatalf("unexpected named vector params: %+v", vp)
	}
}

func TestApplyDefaultsAndToJSON(t *testing.T) {
	cfg := CollectionConfig{Vectors: VectorsConfig{Single: &VectorParams{Size: 4}}}
	if err := cfg.ApplyDefaults(DefaultCollectionDefaults); err != nil {
		t.Fatalf("apply defaults: %v", err)
	}
	if cfg.Vectors.Single.Distance != DistanceCosine {
		t.Fatalf("expected default distance to be applied")
	}
	if cfg.HNSW == nil || cfg.Optimizers == nil {
		t.Fatalf("expected defaults for HNSW and Optimizers")
	}

	cfg = CollectionConfig{Vectors: VectorsConfig{Named: map[string]VectorParams{"a": {Size: 2}}}}
	if err := cfg.ApplyDefaults(DefaultCollectionDefaults); err != nil {
		t.Fatalf("apply defaults named: %v", err)
	}
	if cfg.Vectors.Named["a"].Distance != DistanceCosine {
		t.Fatalf("expected default distance for named vector")
	}

	cfg = CollectionConfig{Vectors: VectorsConfig{Single: &VectorParams{Size: 1, Distance: DistanceDot}}, Extras: map[string]any{"foo": "bar", "vectors": "bad"}}
	m, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if m["foo"] != "bar" {
		t.Fatalf("expected extras to be merged")
	}
	if _, ok := m["vectors"].(map[string]any); !ok {
		t.Fatalf("expected typed vectors to win over extras")
	}

	cfg = CollectionConfig{Vectors: VectorsConfig{Single: &VectorParams{Size: 1}, Named: map[string]VectorParams{"x": {Size: 1}}}}
	if _, err := cfg.ToJSON(); err == nil {
		t.Fatalf("expected ToJSON error when both single and named vectors are set")
	}
}

func TestApplyDefaults_Errors(t *testing.T) {
	cfg := CollectionConfig{}
	if err := cfg.ApplyDefaults(DefaultCollectionDefaults); err == nil {
		t.Fatalf("expected error for missing vectors config")
	}

	cfg = CollectionConfig{Vectors: VectorsConfig{Single: &VectorParams{Size: 0}}}
	if err := cfg.ApplyDefaults(DefaultCollectionDefaults); err == nil {
		t.Fatalf("expected error for invalid single vector size")
	}

	cfg = CollectionConfig{Vectors: VectorsConfig{Named: map[string]VectorParams{"bad": {Size: 0}}}}
	if err := cfg.ApplyDefaults(DefaultCollectionDefaults); err == nil {
		t.Fatalf("expected error for invalid named vector size")
	}
}

func TestApplyDefaults_DoesNotOverrideProvidedConfigs(t *testing.T) {
	hnsw := &HNSWConfig{M: 48, EfConstruct: 256}
	optim := &OptimizersConfig{DefaultSegmentNumber: 8}
	cfg := CollectionConfig{
		Vectors: VectorsConfig{
			Single: &VectorParams{Size: 8, Distance: DistanceDot},
		},
		HNSW:       hnsw,
		Optimizers: optim,
	}

	if err := cfg.ApplyDefaults(DefaultCollectionDefaults); err != nil {
		t.Fatalf("apply defaults: %v", err)
	}
	if cfg.HNSW != hnsw || cfg.Optimizers != optim {
		t.Fatalf("expected provided pointers to be preserved")
	}
	if cfg.Vectors.Single.Distance != DistanceDot {
		t.Fatalf("expected provided distance to be preserved")
	}
}

func TestToJSON_NoExtras(t *testing.T) {
	cfg := CollectionConfig{
		Vectors: VectorsConfig{
			Single: &VectorParams{Size: 4, Distance: DistanceCosine},
		},
	}
	m, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if _, ok := m["vectors"]; !ok {
		t.Fatalf("expected vectors in serialized output")
	}
}

func TestToJSON_UnmarshalError(t *testing.T) {
	orig := unmarshalCollectionJSONHook
	defer func() { unmarshalCollectionJSONHook = orig }()
	unmarshalCollectionJSONHook = func(data []byte, v interface{}) error {
		return errors.New("forced unmarshal error")
	}

	cfg := CollectionConfig{
		Vectors: VectorsConfig{
			Single: &VectorParams{Size: 4, Distance: DistanceCosine},
		},
	}
	if _, err := cfg.ToJSON(); err == nil {
		t.Fatalf("expected unmarshal error")
	}
}
