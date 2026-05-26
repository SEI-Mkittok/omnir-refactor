package handler

import (
	"testing"

	"github.com/omnir/crm-api/internal/domain"
)

func TestPreviewWebformMapsRequiredFields(t *testing.T) {
	form := &domain.Webform{
		TargetModule: domain.WebformTargetLead,
		Fields: []domain.WebformField{
			{Key: "email", Label: "Email", Required: true, TargetField: "email"},
			{Key: "company", Label: "Company", TargetField: "company"},
		},
	}

	preview := previewWebform(form, map[string]interface{}{"email": "ada@example.com", "company": "Analytical Engines"})

	if len(preview.MissingFields) != 0 {
		t.Fatalf("expected no missing fields, got %v", preview.MissingFields)
	}
	if preview.MappedFields["email"] != "ada@example.com" {
		t.Fatalf("expected mapped email, got %v", preview.MappedFields["email"])
	}
}

func TestMailConditionMatchesSubjectContains(t *testing.T) {
	msg := &domain.EmailInboxMessage{Subject: "Urgent billing issue"}
	condition := domain.MailConverterCondition{Field: "subject", Operator: "contains", Value: "billing"}

	if !mailConditionMatches(condition, msg) {
		t.Fatal("expected subject condition to match")
	}
}
