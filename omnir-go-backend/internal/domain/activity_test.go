package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/omnir/crm-api/internal/domain"
)

func TestActivity_Validate(t *testing.T) {
	ownerID := uuid.New()

	tests := []struct {
		name    string
		input   domain.Activity
		wantErr bool
	}{
		{
			name: "valid activity — minimum fields",
			input: domain.Activity{
				Type:    domain.ActivityTypeCall,
				Subject: "Follow-up call",
				OwnerID: ownerID,
			},
			wantErr: false,
		},
		{
			name: "valid activity — all types",
			input: domain.Activity{
				Type:    domain.ActivityTypeEmail,
				Subject: "Send proposal",
				OwnerID: ownerID,
			},
			wantErr: false,
		},
		{
			name: "missing subject",
			input: domain.Activity{
				Type:    domain.ActivityTypeTask,
				OwnerID: ownerID,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			input: domain.Activity{
				Type:    domain.ActivityType("webinar"),
				Subject: "Something",
				OwnerID: ownerID,
			},
			wantErr: true,
		},
		{
			name: "missing owner_id",
			input: domain.Activity{
				Type:    domain.ActivityTypeMeeting,
				Subject: "Kick-off meeting",
			},
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

func TestActivityType_IsValid(t *testing.T) {
	valid := []domain.ActivityType{
		domain.ActivityTypeCall,
		domain.ActivityTypeEmail,
		domain.ActivityTypeMeeting,
		domain.ActivityTypeTask,
	}
	for _, at := range valid {
		if !at.IsValid() {
			t.Errorf("expected activity type %q to be valid", at)
		}
	}

	if domain.ActivityType("bogus").IsValid() {
		t.Error("expected activity type 'bogus' to be invalid")
	}
}
