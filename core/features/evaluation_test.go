package features

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockSetConfig(cfg *FeatureConfig) {
	mutex.Lock()
	defer mutex.Unlock()
	configCache = cfg
	areFeaturesInitialized = true
}

func TestGetFeatureValue_ValidEvaluation(t *testing.T) {
	resetInitState()

	now := time.Now().UTC()
	cfg := &FeatureConfig{
		Attributes: &Attributes{
			Canary:       true,
			Platforms:    []string{"web"},
			Environments: []string{"prod"},
			Regions:      []string{"us"},
		},
		Features: map[string]*Feature{
			"featureA": {
				Enabled:      true,
				Datatype:     "STRING",
				DefaultValue: "default",
				Variants: []*Variant{
					{Name: "A", Weight: 100, Label: "v1"},
				},
				Rollouts: []*Rollout{
					{Platform: "web", Environment: "prod", Region: "us", Percentage: 100},
				},
				Rules: &Rules{
					Conjunction: "AND",
					Conditions: []*Condition{
						{Key: "country", Operator: "EQUALS", Value: "us"},
					},
				},
				Schedule: &Schedule{
					Start: now.Add(-time.Hour).Format(time.RFC3339),
					End:   now.Add(time.Hour).Format(time.RFC3339),
				},
			},
		},
	}

	configCache = cfg
	areFeaturesInitialized = true

	val, err := GetFeatureValue("featureA", EvaluationOptions{
		Platform:    "web",
		Environment: "prod",
		Region:      "us",
		RuleContext: map[string]interface{}{"country": "us"},
	})

	require.NoError(t, err)
	require.Equal(t, "A", val)
}

func TestGetFeatureValue_GranularCases(t *testing.T) {
	now := time.Now().UTC()
	start := now.Add(-1 * time.Hour).Format(time.RFC3339)
	end := now.Add(1 * time.Hour).Format(time.RFC3339)

	cfg := &FeatureConfig{
		Features: map[string]*Feature{
			"featureX": {
				Name:         "featureX",
				Enabled:      true,
				Datatype:     DatatypeString,
				DefaultValue: "default",
				Schedule: &Schedule{
					Start: start,
					End:   end,
				},
				Variants: []*Variant{
					{Name: "A", Weight: 100, Label: "v1"},
				},
				Rollouts: []*Rollout{
					{
						Platform:    "web",
						Environment: "prod",
						Region:      "us",
						Percentage:  100,
					},
				},
				Rules: &Rules{
					Conjunction: ConjunctionAnd,
					Conditions: []*Condition{
						{Key: "userTier", Operator: OperatorEquals, Value: "premium"},
					},
				},
			},
		},
	}

	// Override global config cache
	configCache = cfg
	areFeaturesInitialized = true

	tests := []struct {
		name       string
		opts       EvaluationOptions
		want       interface{}
		expectFail bool
	}{
		{
			name: "All match - should return variant",
			opts: EvaluationOptions{
				Platform:    "web",
				Environment: "prod",
				Region:      "us",
				RuleContext: map[string]interface{}{"userTier": "premium"},
			},
			want: "A",
		},
		{
			name: "Platform mismatch - should return default",
			opts: EvaluationOptions{
				Platform:    "mobile",
				Environment: "prod",
				Region:      "us",
				RuleContext: map[string]interface{}{"userTier": "premium"},
			},
			want: "default",
		},
		{
			name: "Region mismatch - should return default",
			opts: EvaluationOptions{
				Platform:    "web",
				Environment: "prod",
				Region:      "eu",
				RuleContext: map[string]interface{}{"userTier": "premium"},
			},
			want: "default",
		},
		{
			name: "Rule fails - should return default",
			opts: EvaluationOptions{
				Platform:    "web",
				Environment: "prod",
				Region:      "us",
				RuleContext: map[string]interface{}{"userTier": "basic"},
			},
			want: "default",
		},
		{
			name: "Schedule expired - should return default",
			opts: EvaluationOptions{
				Platform:    "web",
				Environment: "prod",
				Region:      "us",
				RuleContext: map[string]interface{}{"userTier": "premium"},
			},
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Schedule expired - should return default" {
				cfg.Features["featureX"].Schedule = &Schedule{
					Start: now.Add(-2 * time.Hour).Format(time.RFC3339),
					End:   now.Add(-1 * time.Hour).Format(time.RFC3339),
				}
			} else {
				cfg.Features["featureX"].Schedule = &Schedule{
					Start: start,
					End:   end,
				}
			}

			got, _ := GetFeatureValue("featureX", tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetFeatureValue_FeatureConfigNotInitialized(t *testing.T) {
	resetInitState()

	_, err := GetFeatureValue("someFeature", EvaluationOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "feature config not initialized")
}

func TestGetFeatureValue_FeatureNotFound(t *testing.T) {
	cfg := &FeatureConfig{
		Features: map[string]*Feature{
			"disabledFeature": {
				Enabled:      false,
				DefaultValue: "off",
			},
		},
	}
	mockSetConfig(cfg)

	_, err := GetFeatureValue("nonExistentFeature", EvaluationOptions{
		Platform:    "web",
		Environment: "production",
		Region:      "us",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "feature 'nonExistentFeature' not found")
}

func TestGetFeatureValue_FeatureDisabled(t *testing.T) {
	cfg := &FeatureConfig{
		Features: map[string]*Feature{
			"disabledFeature": {
				Enabled:      false,
				DefaultValue: "off",
			},
		},
	}
	mockSetConfig(cfg)

	result, err := GetFeatureValue("disabledFeature", EvaluationOptions{})
	assert.NoError(t, err)
	assert.Equal(t, "off", result)
}

func Test_isFeatureActive(t *testing.T) {
	assert.True(t, isFeatureActive(&Feature{Enabled: true}))
	assert.False(t, isFeatureActive(&Feature{Enabled: false}))
}

func Test_isScheduleValid(t *testing.T) {
	assert.True(t, isScheduleValid(nil)) // nil schedule should be valid

	now := time.Now().UTC()
	past := now.Add(-2 * time.Hour).Format(time.RFC3339)
	future := now.Add(2 * time.Hour).Format(time.RFC3339)
	tooEarly := now.Add(2 * time.Hour).Format(time.RFC3339)
	tooLate := now.Add(-2 * time.Hour).Format(time.RFC3339)

	assert.True(t, isScheduleValid(&Schedule{Start: past, End: future}))
	assert.False(t, isScheduleValid(&Schedule{Start: tooEarly, End: future}))
	assert.False(t, isScheduleValid(&Schedule{Start: past, End: tooLate}))
	assert.False(t, isScheduleValid(&Schedule{Start: "not-a-date", End: future}))
	assert.False(t, isScheduleValid(&Schedule{Start: past, End: "invalid-date"}))
}

func Test_isRolloutEligible(t *testing.T) {
	assert.True(t, isRolloutEligible(nil, EvaluationOptions{})) // nil rollouts should be valid

	rollouts := []*Rollout{
		{Platform: "web", Environment: "prod", Region: "us", Percentage: 100},
	}
	opts := EvaluationOptions{Platform: "web", Environment: "prod", Region: "us"}
	assert.True(t, isRolloutEligible(rollouts, opts))

	rollouts[0].Percentage = 0
	assert.False(t, isRolloutEligible(rollouts, opts))

	opts = EvaluationOptions{Platform: "android", Environment: "prod", Region: "us"}
	assert.False(t, isRolloutEligible(rollouts, opts))

	opts = EvaluationOptions{Platform: "web", Environment: "qa", Region: "us"}
	assert.False(t, isRolloutEligible(rollouts, opts))

	opts = EvaluationOptions{Platform: "web", Environment: "prod", Region: "in"}
	assert.False(t, isRolloutEligible(rollouts, opts))
}

func Test_evaluateConditions_AND(t *testing.T) {
	conds := []*Condition{
		{Key: "country", Operator: OperatorEquals, Value: "IN"},
		{Key: "age", Operator: OperatorGreaterThanOrEqual, Value: 18},
	}
	ruleCtx := map[string]interface{}{"country": "IN", "age": 25}
	assert.True(t, evaluateConditions(conds, ConjunctionAnd, ruleCtx))

	ruleCtx = map[string]interface{}{"country": "IN", "age": 15}
	assert.False(t, evaluateConditions(conds, ConjunctionAnd, ruleCtx))
}

func Test_evaluateConditions_OR(t *testing.T) {
	conds := []*Condition{
		{Key: "country", Operator: OperatorEquals, Value: "IN"},
		{Key: "age", Operator: OperatorGreaterThanOrEqual, Value: 18},
	}
	ruleCtx := map[string]interface{}{"country": "US", "age": 25}
	assert.True(t, evaluateConditions(conds, ConjunctionOr, ruleCtx))

	ruleCtx = map[string]interface{}{"country": "US", "age": 16}
	assert.False(t, evaluateConditions(conds, ConjunctionOr, ruleCtx))
}

func Test_evaluateConditions_Invalid(t *testing.T) {
	conds := []*Condition{
		{Key: "country", Operator: OperatorEquals, Value: "IN"},
		{Key: "age", Operator: OperatorGreaterThanOrEqual, Value: 18},
	}

	ruleCtx := map[string]interface{}{"country": "US", "age": 16}
	assert.False(t, evaluateConditions(conds, "invalid", ruleCtx))
}

func Test_matchCondition(t *testing.T) {
	assert.True(t, matchCondition(&Condition{Operator: OperatorEquals, Value: "x"}, "x"))
	assert.False(t, matchCondition(&Condition{Operator: OperatorEquals, Value: "x"}, "y"))

	assert.True(t, matchCondition(&Condition{Operator: OperatorNotEquals, Value: "x"}, "y"))
	assert.False(t, matchCondition(&Condition{Operator: OperatorNotEquals, Value: "x"}, "x"))

	assert.True(t, matchCondition(&Condition{Operator: OperatorIn, Value: []interface{}{"a", "b"}}, "a"))
	assert.False(t, matchCondition(&Condition{Operator: OperatorIn, Value: []interface{}{"a", "b"}}, "c"))
	assert.False(t, matchCondition(&Condition{Operator: OperatorIn, Value: "s"}, 1))

	assert.True(t, matchCondition(&Condition{Operator: OperatorNotIn, Value: []interface{}{"a", "b"}}, "c"))
	assert.False(t, matchCondition(&Condition{Operator: OperatorNotIn, Value: []interface{}{"a", "b"}}, "a"))
	assert.False(t, matchCondition(&Condition{Operator: OperatorNotIn, Value: "s"}, 1))

	assert.True(t, matchCondition(&Condition{Operator: OperatorGreaterThan, Value: 5}, 6))
	assert.False(t, matchCondition(&Condition{Operator: OperatorGreaterThan, Value: 5}, 3))

	assert.True(t, matchCondition(&Condition{Operator: OperatorLessThan, Value: 5}, 3))
	assert.False(t, matchCondition(&Condition{Operator: OperatorLessThan, Value: 5}, 6))

	assert.True(t, matchCondition(&Condition{Operator: OperatorGreaterThanOrEqual, Value: 5}, 5))
	assert.False(t, matchCondition(&Condition{Operator: OperatorGreaterThanOrEqual, Value: 5}, 3))

	assert.True(t, matchCondition(&Condition{Operator: OperatorLessThanOrEqual, Value: 5}, 5))
	assert.False(t, matchCondition(&Condition{Operator: OperatorLessThanOrEqual, Value: 5}, 6))

	assert.False(t, matchCondition(&Condition{Operator: "UNKNOWN"}, 1))
}

func Test_selectVariant(t *testing.T) {
	feature := &Feature{
		DefaultValue: "none",
		Variants: []*Variant{
			{Name: "a", Weight: 50, Label: "v1"},
			{Name: "b", Weight: 50, Label: "v2"},
		},
	}

	// Should return either a or b
	firstSelection := selectVariant(feature, "my-random-key")
	assert.Contains(t, []string{"a", "b"}, firstSelection)

	secondSelection := selectVariant(feature, "my-random-key")
	assert.Equal(t, firstSelection, secondSelection)

	selected := selectVariant(feature, "")
	assert.Contains(t, []string{"a", "b"}, selected)

	feature2 := &Feature{
		DefaultValue: "default",
		Variants:     []*Variant{}, // no variants
	}

	val := selectVariant(feature2, "")
	assert.Equal(t, "default", val)
}

func Test_toFloat64(t *testing.T) {
	val, ok := toFloat64(5)
	assert.True(t, ok)
	assert.Equal(t, 5.0, val)

	val, ok = toFloat64(int64(10))
	assert.True(t, ok)
	assert.Equal(t, 10.0, val)

	val, ok = toFloat64(2.5)
	assert.True(t, ok)
	assert.Equal(t, 2.5, val)

	val, ok = toFloat64("str")
	assert.False(t, ok)
	assert.Equal(t, 0.0, val)

	diff := compareFloats("invalid", 5)
	assert.Equal(t, 0.0, diff)

	diff = compareFloats(5, "invalid")
	assert.Equal(t, 0.0, diff)

	diff = compareFloats("a", "b")
	assert.Equal(t, 0.0, diff)
}

// -------- Benchmark tests -------- //
func BenchmarkGetFeatureValue(b *testing.B) {
	cfg := &FeatureConfig{
		Attributes: &Attributes{
			Canary:       true,
			Platforms:    []string{"web"},
			Environments: []string{"prod"},
			Regions:      []string{"us"},
		},
		Features: map[string]*Feature{
			"featureX": {
				Name:         "featureX",
				Enabled:      true,
				Datatype:     DatatypeString,
				DefaultValue: "default",
				Schedule: &Schedule{
					Start: time.Now().Add(-time.Hour).Format(time.RFC3339),
					End:   time.Now().Add(time.Hour).Format(time.RFC3339),
				},
				Variants: []*Variant{
					{Name: "A", Weight: 30},
					{Name: "B", Weight: 30},
					{Name: "C", Weight: 40},
				},
				Rollouts: []*Rollout{
					{Platform: "web", Environment: "prod", Region: "us", Percentage: 100},
				},
				Rules: &Rules{
					Conjunction: ConjunctionAnd,
					Conditions: []*Condition{
						{Key: "userTier", Operator: OperatorEquals, Value: "premium"},
					},
				},
			},
		},
	}

	mockSetConfig(cfg)

	evalOpts := EvaluationOptions{
		Platform:            "web",
		Environment:         "prod",
		Region:              "us",
		RuleContext:         map[string]interface{}{"userTier": "premium"},
		EvaluationBucketKey: "user-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := GetFeatureValue("featureX", evalOpts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSelectVariant(b *testing.B) {
	feature := &Feature{
		DefaultValue: "none",
		Variants: []*Variant{
			{Name: "A", Weight: 20, Label: "v1"},
			{Name: "B", Weight: 20, Label: "v2"},
			{Name: "C", Weight: 20, Label: "v3"},
			{Name: "D", Weight: 20, Label: "v4"},
			{Name: "E", Weight: 20, Label: "v5"},
		},
	}

	bucketKey := "user-id-12345"

	b.Run("WithBucketKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = selectVariant(feature, bucketKey)
		}
	})

	b.Run("WithoutBucketKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = selectVariant(feature, "")
		}
	})
}
