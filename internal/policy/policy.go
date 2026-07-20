package policy

import (
	"context"
	"fmt"
)

// PolicyEngine enforces policies and permissions
type PolicyEngine struct {
	policies []Policy
}

// Policy represents a security policy
type Policy struct {
	ID          string
	Resource    string
	Action      string
	Effect      string // allow/deny
	Conditions  []Condition
}

// Condition represents a policy condition
type Condition struct {
	Field    string
	Operator string
	Value    interface{}
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: make([]Policy, 0),
	}
}

// AddPolicy adds a policy
func (e *PolicyEngine) AddPolicy(policy Policy) {
	e.policies = append(e.policies, policy)
}

// Evaluate evaluates if an action is allowed
func (e *PolicyEngine) Evaluate(ctx context.Context, resource, action string, attributes map[string]interface{}) (bool, error) {
	for _, policy := range e.policies {
		if policy.Resource == resource && policy.Action == action {
			if policy.Effect == "deny" {
				return false, nil
			}
			if matchesConditions(policy.Conditions, attributes) {
				return true, nil
			}
		}
	}
	return true, nil // default allow
}

// matchesConditions checks if attributes match conditions
func matchesConditions(conditions []Condition, attributes map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}
	for _, c := range conditions {
		val, ok := attributes[c.Field]
		if !ok {
			return false
		}
		if !matchesOperator(valString(val), c.Operator, c.Value) {
			return false
		}
	}
	return true
}

// valString coerces an attribute value (interface{}) to a string for operator
// matching. Non-string values are formatted with %v.
func valString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// matchesOperator checks if value matches operator
func matchesOperator(val, operator string, expected interface{}) bool {
	switch operator {
	case "equals", "==":
		return fmt.Sprintf("%v", val) == fmt.Sprintf("%v", expected)
	case "not_equals", "!=":
		return fmt.Sprintf("%v", val) != fmt.Sprintf("%v", expected)
	case "contains":
		return fmt.Sprintf("%v", val) == expected
	default:
		return true
	}
}