package features

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFeatureConfig_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		config   *FeatureConfig
		wantErr  bool
		errorSub string
	}{
		{
			name: "valid config",
			config: &FeatureConfig{
				Attributes: &Attributes{
					Environments: []string{"production"},
					Regions:      []string{"us"},
					Platforms:    []string{"web"},
				},
				Features: map[string]*Feature{
					"featureA": {
						Enabled:      true,
						Datatype:     DatatypeString,
						DefaultValue: "val",
						Variants: []*Variant{
							{Name: "a", Weight: 50, Label: "v1"},
							{Name: "b", Weight: 50, Label: "v2"},
						},
						Rules: &Rules{
							Conjunction: ConjunctionAnd,
							Conditions: []*Condition{
								{Key: "user", Operator: OperatorEquals, Value: "abc"},
							},
						},
						Rollouts: []*Rollout{
							{Platform: "web", Environment: "production", Region: "us", Percentage: 50},
						},
						Schedule: &Schedule{Start: "2024-01-01T00:00:00Z", End: "2025-01-01T00:00:00Z"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid datatype",
			config: &FeatureConfig{
				Features: map[string]*Feature{
					"f": {Datatype: "unknown"},
				},
			},
			wantErr:  true,
			errorSub: "invalid datatype",
		},
		{
			name: "variant weight not 100",
			config: &FeatureConfig{
				Features: map[string]*Feature{
					"f": {
						Datatype: DatatypeString,
						Variants: []*Variant{
							{Weight: 30},
							{Weight: 30},
						},
					},
				},
			},
			wantErr:  true,
			errorSub: "variant weight must sum to 100",
		},
		{
			name: "schedule invalid",
			config: &FeatureConfig{
				Features: map[string]*Feature{
					"f": {
						Datatype: DatatypeString,
						Schedule: &Schedule{Start: "2025-01-01T00:00:00Z", End: "2024-01-01T00:00:00Z"},
					},
				},
			},
			wantErr:  true,
			errorSub: "start time must be before",
		},
		{
			name: "rollout env not allowed",
			config: &FeatureConfig{
				Attributes: &Attributes{Environments: []string{"prod"}, Regions: []string{"us"}},
				Features: map[string]*Feature{
					"f": {
						Datatype: DatatypeString,
						Rollouts: []*Rollout{{Environment: "qa", Region: "us", Percentage: 50}},
					},
				},
			},
			wantErr:  true,
			errorSub: "not in allowed environments",
		},
		{
			name: "condition key empty",
			config: &FeatureConfig{
				Features: map[string]*Feature{
					"f": {
						Datatype: DatatypeString,
						Rules: &Rules{
							Conjunction: ConjunctionAnd,
							Conditions:  []*Condition{{Key: "", Operator: OperatorEquals, Value: "val"}},
						},
					},
				},
			},
			wantErr:  true,
			errorSub: "key cannot be empty",
		},
		{
			name: "condition value nil",
			config: &FeatureConfig{
				Features: map[string]*Feature{
					"f": {
						Datatype: DatatypeString,
						Rules: &Rules{
							Conjunction: ConjunctionAnd,
							Conditions:  []*Condition{{Key: "u", Operator: OperatorEquals}},
						},
					},
				},
			},
			wantErr:  true,
			errorSub: "value cannot be nil",
		},
		{
			name: "operator not in list",
			config: &FeatureConfig{
				Features: map[string]*Feature{
					"f": {
						Datatype: DatatypeString,
						Rules: &Rules{
							Conjunction: ConjunctionAnd,
							Conditions:  []*Condition{{Key: "u", Operator: "BAD_OP", Value: "val"}},
						},
					},
				},
			},
			wantErr:  true,
			errorSub: "invalid operator",
		},
		{
			name: "operator IN but not array",
			config: &FeatureConfig{
				Features: map[string]*Feature{
					"f": {
						Datatype: DatatypeString,
						Rules: &Rules{
							Conjunction: ConjunctionAnd,
							Conditions:  []*Condition{{Key: "u", Operator: OperatorIn, Value: "val"}},
						},
					},
				},
			},
			wantErr:  true,
			errorSub: "must have an array value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, _ := json.Marshal(tc.config)
			schema := []byte(`{"type": "object"}`) // minimal schema to bypass

			err := ValidateFeatureConfig(cfg, schema)
			if tc.wantErr {
				if err == nil || (tc.errorSub != "" && !contains(err.Error(), tc.errorSub)) {
					t.Errorf("expected error containing %q, got: %v", tc.errorSub, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected valid config, got: %v", err)
				}
			}
		})
	}
}

func TestValidateFeatureConfig_UnmarshalError(t *testing.T) {
	schema := []byte(`{"type":"object"}`)
	// `attributes` must be an object in FeatureConfig; array forces json.Unmarshal error branch.
	config := []byte(`{"attributes":[],"features":{}}`)

	err := ValidateFeatureConfig(config, schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal feature config")
}

func contains(s, substr string) bool {
	return substr == "" || (s != "" && substr != "" && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	return len(substr) <= len(s) && (len(substr) == 0 || len(s) > 0 && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || stringContains(s[1:], substr))))
}

func TestValidateRules_EdgeCases(t *testing.T) {
	t.Run("MissingConjunctionWithConditions", func(t *testing.T) {
		rules := &Rules{
			Conjunction: "",
			Conditions: []*Condition{
				{
					Key:      "region",
					Operator: OperatorEquals,
					Value:    "US",
				},
			},
		}
		err := validateRules(DatatypeString, rules)
		assert.ErrorContains(t, err, "conjunction cannot be empty")
	})

	t.Run("InvalidConjunctionValue", func(t *testing.T) {
		rules := &Rules{
			Conjunction: "XOR",
			Conditions:  []*Condition{},
		}
		err := validateRules(DatatypeString, rules)
		assert.ErrorContains(t, err, "conjunction must be one of")
	})
}

func Test_validateRollout_EdgeCases(t *testing.T) {
	t.Run("InvalidPlatform_NotInAllowedList-1", func(t *testing.T) {
		rollouts := []*Rollout{
			{
				Platform:    "desktop", // Not in allowed list
				Environment: "production",
				Region:      "us",
				Percentage:  50,
			},
		}
		attrs := &Attributes{
			Platforms:    []string{"web", "mobile"}, // "desktop" not allowed
			Environments: []string{"production"},
			Regions:      []string{"us", "eu"},
		}

		err := validateRollout(rollouts, attrs)
		assert.ErrorContains(t, err, "rollout platform 'desktop' not in allowed platforms")
	})

	t.Run("InvalidPlatform_NotInAllowedList-2", func(t *testing.T) {
		rollouts := []*Rollout{
			{
				Platform:    "web", // Not in allowed list
				Environment: "production",
				Region:      "us",
				Percentage:  50,
			},
		}
		attrs := &Attributes{
			Platforms:    []string{"web", "mobile"}, // "desktop" not allowed
			Environments: []string{"production"},
			Regions:      []string{"au", "eu"},
		}

		err := validateRollout(rollouts, attrs)
		assert.ErrorContains(t, err, "rollout region 'us' not in allowed regions")
	})

	t.Run("InvalidPercentage_OutOfBounds", func(t *testing.T) {
		rollouts := []*Rollout{
			{
				Environment: "production",
				Region:      "us",
				Percentage:  150, // invalid percentage
			},
		}
		attrs := &Attributes{
			Environments: []string{"production"},
			Regions:      []string{"us"},
		}

		err := validateRollout(rollouts, attrs)
		assert.ErrorContains(t, err, "rollout percentage must be between 0 and 100")
	})
}

func TestValidateSchedule_InvalidTimeFormat(t *testing.T) {
	t.Run("InvalidStartTime", func(t *testing.T) {
		schedule := &Schedule{
			Start: "bad-start-time",       // Invalid format
			End:   "2024-08-31T23:59:59Z", // Valid format
		}
		err := validateSchedule(schedule)
		assert.ErrorContains(t, err, "invalid start time")
	})

	t.Run("InvalidEndTime", func(t *testing.T) {
		schedule := &Schedule{
			Start: "2024-08-01T00:00:00Z", // Valid format
			End:   "bad-end-time",         // Invalid format
		}
		err := validateSchedule(schedule)
		assert.ErrorContains(t, err, "invalid end time")
	})
}

func Test_validateSchema_ErrorHandling(t *testing.T) {
	t.Run("SchemaValidationError_dueToInvalidSchemaFormat", func(t *testing.T) {
		badSchema := []byte(`{ this is not valid JSON }`)
		validDoc := []byte(`{}`) // dummy

		err := validateSchema(validDoc, badSchema)
		assert.ErrorContains(t, err, "schema validation error")
	})

	t.Run("SchemaValidationFails_dueToInvalidDocument", func(t *testing.T) {
		// Minimal valid schema that expects "foo" to be a string
		schema := []byte(`{
			"type": "object",
			"properties": {
				"foo": { "type": "string" }
			},
			"required": ["foo"]
		}`)

		invalidDoc := []byte(`{ "foo": 123 }`) // wrong type for "foo"

		err := validateSchema(invalidDoc, schema)
		assert.ErrorContains(t, err, "schema validation failed")
	})
}

func Test_validateVariants_EdgeCases(t *testing.T) {
	t.Run("VariantWeightNegative", func(t *testing.T) {
		variants := []*Variant{
			{Name: "A", Weight: -5},
		}
		err := validateVariants(variants)
		assert.ErrorContains(t, err, "variant weight must be non-negative")
	})

	t.Run("VariantWeightAbove100", func(t *testing.T) {
		variants := []*Variant{
			{Name: "A", Weight: 105},
		}
		err := validateVariants(variants)
		assert.ErrorContains(t, err, "variant weight must be less than or equal to 100")
	})
}

func TestValidateFeatureConfig_InvalidSchema(t *testing.T) {
	invalidJSON := []byte(`{ invalid json `)
	schemaJSON := []byte(`{ "type": "object" }`) // minimal schema

	err := ValidateFeatureConfig(invalidJSON, schemaJSON)
	assert.ErrorContains(t, err, "schema validation error")
}

func TestApplyCustomValidations_EmptyFeatures(t *testing.T) {
	config := &FeatureConfig{
		Version:  "1.0",
		Features: map[string]*Feature{}, // empty features map
	}

	err := applyContentValidations(config)
	assert.NoError(t, err)
}

func TestParseTime_EmptyString(t *testing.T) {
	tm, err := parseTime("")
	assert.NoError(t, err)
	assert.Nil(t, tm)
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
		expectError    bool
	}{
		{
			input:          "2025-07-21T10:00:00Z",
			expectedOutput: "2025-07-21 10:00:00 +0000 UTC",
			expectError:    false,
		},
		{
			input:          "",
			expectedOutput: "",
			expectError:    false,
		},
		{
			input:       "2025-07-21 10:00:00", // missing T and Z
			expectError: true,
		},
		{
			input:       "not-a-date",
			expectError: true,
		},
	}

	for _, test := range tests {
		result, err := parseTime(test.input)

		if test.expectError {

			if err == nil {
				t.Errorf("expected error for input %q, got nil", test.input)
			}
		} else {

			if err != nil {
				t.Errorf("unexpected error for input %q: %v", test.input, err)
			} else if result == nil && test.input != "" {
				t.Errorf("expected non-nil result for input %q", test.input)
			} else if result != nil && result.String() != test.expectedOutput {
				t.Errorf("unexpected result for input %q: got %v, want %v", test.input, result.String(), test.expectedOutput)
			}
		}
	}
}
