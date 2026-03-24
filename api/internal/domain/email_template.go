package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EmailTemplate struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *EmailTemplate) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if t.Subject == "" {
		return fmt.Errorf("%w: subject is required", ErrValidation)
	}
	if t.Body == "" {
		return fmt.Errorf("%w: body is required", ErrValidation)
	}
	return nil
}

type EmailTemplatePatch struct {
	Name    *string `json:"name,omitempty"`
	Subject *string `json:"subject,omitempty"`
	Body    *string `json:"body,omitempty"`
}

type EmailTemplateFilter struct {
	OrgID uuid.UUID
	Q     string
	Page  int
	Limit int
}
