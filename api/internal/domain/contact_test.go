package domain_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

func strPtr(s string) *string { return &s }

func TestContact_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.Contact
		wantErr bool
	}{
		{
			name:    "valid contact minimum fields",
			input:   domain.Contact{FirstName: "Ada", LastName: "Lovelace", OwnerID: uuid.New()},
			wantErr: false,
		},
		{
			name:    "valid contact all fields",
			input:   domain.Contact{FirstName: "Ada", LastName: "Lovelace", OwnerID: uuid.New(), Email: strPtr("ada@example.com")},
			wantErr: false,
		},
		{
			name:    "missing first name",
			input:   domain.Contact{Email: strPtr("ada@example.com")},
			wantErr: true,
		},
		{
			name:    "invalid email format",
			input:   domain.Contact{FirstName: "Ada", LastName: "Lovelace", OwnerID: uuid.New(), Email: strPtr("not-an-email")},
			wantErr: true,
		},
		{
			name:    "empty email is allowed",
			input:   domain.Contact{FirstName: "Ada", LastName: "Lovelace", OwnerID: uuid.New(), Email: strPtr("")},
			wantErr: false,
		},
		{
			name:    "invalid stage",
			input:   domain.Contact{FirstName: "Ada", LastName: "Lovelace", OwnerID: uuid.New(), Stage: "unknown_stage"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContactStage_IsValid(t *testing.T) {
	valid := []domain.ContactStage{
		domain.ContactStageLead,
		domain.ContactStageProspect,
		domain.ContactStageCustomer,
		domain.ContactStageChurned,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected stage %q to be valid", s)
		}
	}

	if domain.ContactStage("bogus").IsValid() {
		t.Error("expected stage 'bogus' to be invalid")
	}
}
