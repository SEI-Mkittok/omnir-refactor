package domain

import "testing"

func TestWebformValidateRejectsDuplicateFieldKeys(t *testing.T) {
	form := &Webform{
		Name:         "Contact us",
		TargetModule: WebformTargetLead,
		Fields: []WebformField{
			{Key: "email", Label: "Email", Type: "email"},
			{Key: "email", Label: "Work email", Type: "email"},
		},
	}

	if err := form.Validate(); err == nil {
		t.Fatal("expected duplicate field validation error")
	}
}

func TestMailConverterRuleValidateRequiresConditionsAndActions(t *testing.T) {
	rule := &MailConverterRule{Name: "Support mail", Status: MailConverterRuleActive}

	if err := rule.Validate(); err == nil {
		t.Fatal("expected validation error for missing conditions and actions")
	}
}

func TestValidateAutomationConfigAcceptsTicketCreated(t *testing.T) {
	trigger := AutomationTrigger{Type: TriggerTicketCreated, Config: map[string]interface{}{}}
	actions := []AutomationAction{{Type: ActionCreateActivity, Config: map[string]interface{}{"subject": "Review ticket"}}}

	if err := ValidateAutomationConfig(&trigger, nil, actions); err != nil {
		t.Fatalf("expected ticket_created automation config to be valid: %v", err)
	}
}
