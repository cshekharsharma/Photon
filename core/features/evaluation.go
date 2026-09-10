package features

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"time"
)

// GetFeatureValue determines the evaluated value for a given feature key based on the provided evaluation options.
//
// It performs the following checks in order:
//  1. Ensures feature config is initialized and feature exists
//  2. Checks if the feature is enabled
//  3. Verifies the current time is within the feature’s schedule (if defined)
//  4. Checks whether the feature applies to the provided rollout context (platform, environment, region)
//  5. Evaluates all rules (if defined) against the rule context
//
// If all checks pass, a weighted variant is selected and returned.
// If any check fails, the default value for the feature is returned.
func GetFeatureValue(featureKey string, opts EvaluationOptions) (interface{}, error) {
	cfg, err := GetFeatureConfigStore()

	if err != nil {
		return nil, fmt.Errorf("feature config not initialized: %w", err)
	}

	feature, ok := cfg.Features[featureKey]
	if !ok {
		return nil, fmt.Errorf("feature '%s' not found in loaded feature config", featureKey)
	}

	if !isFeatureActive(feature) {
		return feature.DefaultValue, nil
	}

	if !isScheduleValid(feature.Schedule) {
		return feature.DefaultValue, nil
	}

	if !isRolloutEligible(feature.Rollouts, opts) {
		return feature.DefaultValue, nil
	}

	if feature.Rules != nil && len(feature.Rules.Conditions) > 0 {
		if !evaluateConditions(feature.Rules.Conditions, feature.Rules.Conjunction, opts.RuleContext) {
			return feature.DefaultValue, nil
		}
	}

	return selectVariant(feature, opts.EvaluationBucketKey), nil
}

// isFeatureActive returns true if the feature is marked as enabled.
func isFeatureActive(feature *Feature) bool {
	return feature.Enabled
}

// isScheduleValid checks whether the current time falls within the feature’s defined schedule window.
// If no schedule is defined, it returns true.
func isScheduleValid(schedule *Schedule) bool {
	if schedule == nil {
		return true
	}

	now := time.Now().UTC()

	start, err1 := parseTime(schedule.Start)
	if err1 != nil {
		return false
	}

	end, err2 := parseTime(schedule.End)
	if err2 != nil {
		return false
	}

	if (start != nil && now.Before(*start)) || (end != nil && now.After(*end)) {
		return false
	}

	return true
}

// isRolloutEligible determines if the feature should be rolled out based on platform, environment, and region.
// It supports wildcard values ("*") and uses a random percentage match to simulate rollout gating.
func isRolloutEligible(rollouts []*Rollout, opts EvaluationOptions) bool {
	if len(rollouts) == 0 {
		return true
	}

	for _, r := range rollouts {
		isPlatformEligible := opts.Platform == r.Platform || r.Platform == WildCardValue
		isEnvironmentEligible := opts.Environment == r.Environment || r.Environment == WildCardValue
		isRegionEligible := opts.Region == r.Region || r.Region == WildCardValue

		if isPlatformEligible && isEnvironmentEligible && isRegionEligible {
			return rand.Float64()*100 <= r.Percentage
		}
	}

	return false
}

// selectVariant performs weighted random selection among the feature's variants.
// If weights are not correctly configured or selection fails, the default value is returned.
// The selection is based on the provided bucket key or a random float if no key is given.
// The bucket key is hashed to ensure consistent results across evaluations.
func selectVariant(feature *Feature, bucketKey string) interface{} {
	var roll float64

	if bucketKey != "" {
		h := fnv.New64a()
		_, _ = h.Write([]byte(bucketKey))
		roll = float64(h.Sum64()%10000) / 100.0
	} else {
		roll = rand.Float64() * 100
	}

	sum := 0.0
	for _, v := range feature.Variants {
		sum += v.Weight
		if roll <= sum {
			return v.Name
		}
	}

	return feature.DefaultValue
}

// evaluateConditions evaluates all conditions in the feature's rules block using the specified conjunction (AND/OR).
// The rule context must provide matching keys for all conditions.
// Returns true if the conditions are satisfied as per the conjunction logic.
func evaluateConditions(conditions []*Condition, conjunction string, ruleCtx map[string]interface{}) bool {
	switch strings.ToUpper(conjunction) {
	case ConjunctionOr:
		for _, condition := range conditions {
			if matchCondition(condition, ruleCtx[condition.Key]) {
				return true
			}
		}
		return false

	case ConjunctionAnd:
		for _, condition := range conditions {
			if !matchCondition(condition, ruleCtx[condition.Key]) {
				return false
			}
		}
		return true

	default:
		return false
	}
}

// matchCondition compares a condition’s expected value against the actual value provided in the rule context.
// It supports standard operators such as equality, inequality, range, and list membership.
func matchCondition(c *Condition, val interface{}) bool {
	switch c.Operator {
	case OperatorEquals:
		return val == c.Value

	case OperatorNotEquals:
		return val != c.Value

	case OperatorIn:
		list, ok := c.Value.([]interface{})
		if !ok {
			return false
		}
		for _, item := range list {
			if item == val {
				return true
			}
		}
		return false

	case OperatorNotIn:
		list, ok := c.Value.([]interface{})
		if !ok {
			return false
		}
		for _, item := range list {
			if item == val {
				return false
			}
		}
		return true

	case OperatorGreaterThan:
		return compareFloats(val, c.Value) > 0

	case OperatorLessThan:
		return compareFloats(val, c.Value) < 0

	case OperatorGreaterThanOrEqual:
		return compareFloats(val, c.Value) >= 0

	case OperatorLessThanOrEqual:
		return compareFloats(val, c.Value) <= 0

	default:
		return false
	}
}

// compareFloats converts and compares two interface values as float64.
// Returns the difference a - b. Returns 0 if either value is not numeric.
func compareFloats(a, b interface{}) float64 {
	fa, okA := toFloat64(a)
	fb, okB := toFloat64(b)

	if !okA || !okB {
		return 0
	}

	return fa - fb
}

// toFloat64 attempts to safely convert an interface{} to float64.
// Supports int, int64, and float64 types. Returns false if conversion fails.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true

	case int64:
		return float64(val), true

	case float64:
		return val, true

	default:
		return 0, false
	}
}
