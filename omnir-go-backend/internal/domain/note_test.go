package domain_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

func TestNote_Validate(t *testing.T) {
	validAuthor := uuid.New()
	validEntity := uuid.New()

	tests := []struct {
		name    string
		input   domain.Note
		wantErr bool
	}{
		{
			name: "valid note",
			input: domain.Note{
				Content:    "Some note",
				EntityType: domain.NoteEntityContact,
				EntityID:   validEntity,
				AuthorID:   validAuthor,
			},
			wantErr: false,
		},
		{
			name: "missing content",
			input: domain.Note{
				EntityType: domain.NoteEntityContact,
				EntityID:   validEntity,
				AuthorID:   validAuthor,
			},
			wantErr: true,
		},
		{
			name: "invalid entity type",
			input: domain.Note{
				Content:    "Note",
				EntityType: "invoice",
				EntityID:   validEntity,
				AuthorID:   validAuthor,
			},
			wantErr: true,
		},
		{
			name: "nil entity id",
			input: domain.Note{
				Content:    "Note",
				EntityType: domain.NoteEntityDeal,
				EntityID:   uuid.Nil,
				AuthorID:   validAuthor,
			},
			wantErr: true,
		},
		{
			name: "nil author id",
			input: domain.Note{
				Content:    "Note",
				EntityType: domain.NoteEntityAccount,
				EntityID:   validEntity,
				AuthorID:   uuid.Nil,
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

func TestNoteEntityType_IsValid(t *testing.T) {
	valid := []domain.NoteEntityType{
		domain.NoteEntityContact,
		domain.NoteEntityAccount,
		domain.NoteEntityDeal,
	}
	for _, et := range valid {
		if !et.IsValid() {
			t.Errorf("expected entity_type %q to be valid", et)
		}
	}

	if domain.NoteEntityType("invoice").IsValid() {
		t.Error("expected entity_type 'invoice' to be invalid")
	}
}
