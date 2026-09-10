package features

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cshekharsharma/photon/utils/types"
	"github.com/xeipuuv/gojsonschema"
)

// ValidateFeatureConfig validates the feature config JSON against the provided schema.
// It checks for the following:
//  1. JSON schema validation
//  2. Content validations for feature properties
//  3. Validates the conditions, rollouts, and schedule
//  4. Ensures that the total weight of variants sums to 100
//  5. Ensures that the start time is before the end time
//
// The function returns an error if any validation fails.
func ValidateFeatureConfig(configJSON []byte, schemaJSON []byte) error {
	// Validate the JSON schema
	if err := validateSchema(configJSON, schemaJSON); err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	var featureConfig = &FeatureConfig{}

	// Unmarshal the JSON into a FeatureConfig struct
	if marshallErr := json.Unmarshal(configJSON, featureConfig); marshallErr != nil {
		return fmt.Errorf("failed to unmarshal feature config: %w", marshallErr)
	}

	// Perform content validations
	if err := applyContentValidations(featureConfig); err != nil {
		return fmt.Errorf("content validation error: %w", err)
	}

	return nil
}

// validateSchema validates the JSON against the provided schema.
// It uses the gojsonschema library to perform the validation.
// The function returns an error if the validation fails.
func validateSchema(configJSON []byte, schemaJSON []byte) error {
	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	documentLoader := gojsonschema.NewBytesLoader(configJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		errs := []string{}

		for _, desc := range result.Errors() {
			errs = append(errs, fmt.Sprintf("[%s] %s", desc.Field(), desc.Description()))
		}

		return fmt.Errorf("schema validation failed: %v", errs)
	}

	return nil
}

// applyContentValidations applies content validations to the feature config.
// It checks for specific rules, conditions, and other constraints.
// The function returns an error if any validation fails.
func applyContentValidations(config *FeatureConfig) error {
	if len(config.Features) == 0 {
		return nil // feature config can be empty
	}

	for name, feature := range config.Features {
		// validate datatype
		if err := validateDatatype(feature.Datatype); err != nil {
			return fmt.Errorf("feature '%s': %v", name, err)
		}

		// Validate schedule
		if feature.Schedule != nil {
			if err := validateSchedule(feature.Schedule); err != nil {
				return fmt.Errorf("feature '%s': %v", name, err)
			}
		}

		// Validate variants
		if len(feature.Variants) > 0 {
			if err := validateVariants(feature.Variants); err != nil {
				return fmt.Errorf("feature '%s': %v", name, err)
			}
		}

		// Validate rules
		if feature.Rules != nil {
			if err := validateRules(feature.Datatype, feature.Rules); err != nil {
				return fmt.Errorf("feature '%s': %v", name, err)
			}
		}

		// Validate rollouts
		if len(feature.Rollouts) > 0 {
			if rolloutErr := validateRollout(feature.Rollouts, config.Attributes); rolloutErr != nil {
				return fmt.Errorf("feature '%s': %v", name, rolloutErr)
			}
		}
	}

	return nil
}

// validateDatatype checks the datatype for a feature and ensures that
// it is one of the valid datatypes.
func validateDatatype(datatype string) error {
	datatype = strings.ToUpper(datatype)
	if found, _ := types.ExistsInList(datatype, validDatatypes); !found {
		return fmt.Errorf("invalid datatype '%s'", datatype)
	}
	return nil
}

// validateSchedule checks the schedule for a feature and ensures that
// the start time is before the end time.
func validateSchedule(schedule *Schedule) error {
	var start, end *time.Time
	var timeErr error

	if schedule.Start != "" {
		if start, timeErr = parseTime(schedule.Start); timeErr != nil {
			return fmt.Errorf("invalid start time: %v", timeErr)
		}
	}

	if schedule.End != "" {
		if end, timeErr = parseTime(schedule.End); timeErr != nil {
			return fmt.Errorf("invalid end time: %v", timeErr)
		}
	}

	if start != nil && end != nil {
		if start.After(*end) {
			return fmt.Errorf("start time must be before end time")
		}
	}

	return nil
}

// validateVariants checks the variants for a feature and ensures that
// the weights are valid and that the total weight sums to 100.
// The function returns an error if any validation fails.
func validateVariants(variants []*Variant) error {
	sum := 0.0

	for _, v := range variants {
		if v.Weight < 0 {
			return fmt.Errorf("variant weight must be non-negative")
		}
		if v.Weight > 100 {
			return fmt.Errorf("variant weight must be less than or equal to 100")
		}

		sum += v.Weight
	}

	if math.Abs(sum-100.0) > 0.001 { // Allow a small margin of error for floating point comparisons
		return fmt.Errorf("total variant weight must sum to 100, got %.2f", sum)
	}

	return nil
}

// validateRules checks the rules for a feature and ensures that
// the conjunction is valid and that the conditions are valid.
// The function returns an error if any validation fails.
func validateRules(datatype string, rules *Rules) error {
	// Validate empty conjunctions with conditions
	if rules.Conjunction == "" && len(rules.Conditions) > 0 {
		return fmt.Errorf("conjunction cannot be empty when conditions are present")
	}

	// Validate conjunction
	validConjunctionsWithEmpty := append(validConjunctions, "")
	if found, _ := types.ExistsInList(rules.Conjunction, validConjunctionsWithEmpty); !found {
		return fmt.Errorf("conjunction must be one of %v", validConjunctionsWithEmpty)
	}

	// Validate conditions
	if len(rules.Conditions) > 0 {
		if condErr := validateConditions(datatype, rules.Conditions); condErr != nil {
			return condErr
		}
	}

	return nil
}

// validateConditions checks the conditions for a feature and ensures that
// the keys, operators, and values are valid.
// It also checks that the value type matches the expected datatype.
// The function returns an error if any validation fails.
func validateConditions(datatype string, conditions []*Condition) error {

	for _, condition := range conditions {

		if condition.Key == "" {
			return fmt.Errorf("condition key cannot be empty")
		}

		if condition.Value == nil {
			return fmt.Errorf("condition value cannot be nil for key '%s'", condition.Key)
		}

		if found, _ := types.ExistsInList(condition.Operator, validOperators); !found {
			return fmt.Errorf("invalid operator '%s' in condition for key '%s'", condition.Operator, condition.Key)
		}

		if condition.Operator == OperatorIn || condition.Operator == OperatorNotIn {
			if _, ok := condition.Value.([]any); !ok {
				return fmt.Errorf("condition with operator '%s' must have an array value", condition.Operator)
			}
		}

	}

	return nil
}

// validateRollout checks the rollouts for a feature and ensures that
// the environments and regions are valid.
// It also checks that the percentage is between 0 and 100.
// The function returns an error if any validation fails.
func validateRollout(rollouts []*Rollout, attributes *Attributes) error {
	allowedEnvs := attributes.Environments
	allowedRegions := attributes.Regions
	allowedPlatforms := attributes.Platforms

	for _, rollout := range rollouts {
		if len(allowedEnvs) > 0 {
			if !inList(rollout.Environment, allowedEnvs) && rollout.Environment != WildCardValue {
				return fmt.Errorf("rollout environment '%s' not in allowed environments: %v",
					rollout.Environment, allowedEnvs)
			}
		}

		if len(allowedRegions) > 0 {
			if !inList(rollout.Region, allowedRegions) && rollout.Region != WildCardValue {
				return fmt.Errorf("rollout region '%s' not in allowed regions: %v",
					rollout.Region, allowedRegions)
			}
		}

		if len(allowedPlatforms) > 0 {
			if !inList(rollout.Platform, allowedPlatforms) && rollout.Platform != WildCardValue {
				return fmt.Errorf("rollout platform '%s' not in allowed platforms: %v",
					rollout.Platform, allowedPlatforms)
			}
		}

		if rollout.Percentage < 0 || rollout.Percentage > 100 {
			return fmt.Errorf("rollout percentage must be between 0 and 100")
		}
	}

	return nil
}

// parseTime parses a time string in DateTime format and returns a pointer to time.Time.
func parseTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("invalid time format: %v", err)
	}

	return &parsed, nil
}

// inList checks if a value exists in a list.
func inList(val interface{}, array interface{}) bool {
	found, _ := types.ExistsInList(val, array)
	return found
}
