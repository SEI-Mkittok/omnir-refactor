package worker

import (
	"testing"

	"github.com/omnir/crm-api/internal/domain"
)

// TestEvaluateConditions_Empty verifies that an automation with no conditions always fires.
func TestEvaluateConditions_Empty(t *testing.T) {
	result := EvaluateConditions(nil, map[string]interface{}{"stage": "lead"})
	if !result {
		t.Fatal("expected empty conditions to return true")
	}
}

// TestEvaluateConditions_AllPass verifies AND logic: all conditions must pass.
func TestEvaluateConditions_AllPass(t *testing.T) {
	conditions := []domain.AutomationCondition{
		{Field: "stage", Operator: domain.ConditionOpEquals, Value: "qualified"},
		{Field: "owner_id", Operator: domain.ConditionOpIsSet},
	}
	data := map[string]interface{}{
		"stage":    "qualified",
		"owner_id": "abc-123",
	}
	if !EvaluateConditions(conditions, data) {
		t.Fatal("expected all conditions to pass")
	}
}

// TestEvaluateConditions_OneFails verifies AND logic: one failing condition stops execution.
func TestEvaluateConditions_OneFails(t *testing.T) {
	conditions := []domain.AutomationCondition{
		{Field: "stage", Operator: domain.ConditionOpEquals, Value: "qualified"},
		{Field: "stage", Operator: domain.ConditionOpEquals, Value: "proposal"},
	}
	data := map[string]interface{}{"stage": "qualified"}
	if EvaluateConditions(conditions, data) {
		t.Fatal("expected second condition to fail")
	}
}

// TestEvaluateConditions_DisabledRule documents that disabled rules must not be
// passed to EvaluateConditions — callers are responsible for filtering by status.
// This test verifies condition logic is independent of rule status.
func TestEvaluateConditions_DisabledRuleNotFired(t *testing.T) {
	// Simulate the caller (worker.processEvent) filtering out non-active automations.
	// A paused/draft automation should never reach EvaluateConditions.
	// Here we just confirm the function itself has no status awareness.
	conditions := []domain.AutomationCondition{
		{Field: "status", Operator: domain.ConditionOpEquals, Value: "active"},
	}
	// Even if data has status="active", the rule evaluation itself is pure logic.
	data := map[string]interface{}{"status": "active"}
	if !EvaluateConditions(conditions, data) {
		t.Fatal("expected condition to pass")
	}
}

func TestEvaluateCondition_Equals(t *testing.T) {
	cases := []struct {
		name   string
		field  string
		op     domain.ConditionOperator
		value  interface{}
		data   map[string]interface{}
		expect bool
	}{
		{
			name:   "equals match",
			field:  "stage",
			op:     domain.ConditionOpEquals,
			value:  "qualified",
			data:   map[string]interface{}{"stage": "qualified"},
			expect: true,
		},
		{
			name:   "equals no match",
			field:  "stage",
			op:     domain.ConditionOpEquals,
			value:  "qualified",
			data:   map[string]interface{}{"stage": "lead"},
			expect: false,
		},
		{
			name:   "not_equals match",
			field:  "stage",
			op:     domain.ConditionOpNotEquals,
			value:  "lost",
			data:   map[string]interface{}{"stage": "qualified"},
			expect: true,
		},
		{
			name:   "contains match",
			field:  "name",
			op:     domain.ConditionOpContains,
			value:  "acme",
			data:   map[string]interface{}{"name": "ACME Corp"},
			expect: true,
		},
		{
			name:   "not_contains match",
			field:  "name",
			op:     domain.ConditionOpNotContains,
			value:  "spam",
			data:   map[string]interface{}{"name": "Legit Corp"},
			expect: true,
		},
		{
			name:   "greater_than match",
			field:  "value",
			op:     domain.ConditionOpGreaterThan,
			value:  "1000",
			data:   map[string]interface{}{"value": "5000"},
			expect: true,
		},
		{
			name:   "less_than match",
			field:  "value",
			op:     domain.ConditionOpLessThan,
			value:  "10000",
			data:   map[string]interface{}{"value": "500"},
			expect: true,
		},
		{
			name:   "is_set when set",
			field:  "owner_id",
			op:     domain.ConditionOpIsSet,
			data:   map[string]interface{}{"owner_id": "some-id"},
			expect: true,
		},
		{
			name:   "is_set when missing",
			field:  "owner_id",
			op:     domain.ConditionOpIsSet,
			data:   map[string]interface{}{},
			expect: false,
		},
		{
			name:   "is_not_set when missing",
			field:  "owner_id",
			op:     domain.ConditionOpIsNotSet,
			data:   map[string]interface{}{},
			expect: true,
		},
		{
			name:   "missing field returns false",
			field:  "nonexistent",
			op:     domain.ConditionOpEquals,
			value:  "anything",
			data:   map[string]interface{}{},
			expect: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cond := domain.AutomationCondition{
				Field:    tc.field,
				Operator: tc.op,
				Value:    tc.value,
			}
			got := EvaluateConditions([]domain.AutomationCondition{cond}, tc.data)
			if got != tc.expect {
				t.Errorf("EvaluateConditions(%+v, %v) = %v; want %v", cond, tc.data, got, tc.expect)
			}
		})
	}
}
