package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/omnir/crm-api/internal/domain"
)

// SeedUser returns a minimal valid User for use in tests.
func SeedUser() *domain.User {
	return &domain.User{
		ID:        uuid.New(),
		Email:     "test@omnir.test",
		Name:      "Test User",
		Role:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SeedContact returns a minimal valid Contact for use in tests.
func SeedContact(ownerID uuid.UUID) *domain.Contact {
	email := "ada@omnir.test"
	return &domain.Contact{
		ID:        uuid.New(),
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     &email,
		OwnerID:   ownerID,
		Stage:     domain.ContactStageProspect,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SeedAccount returns a minimal valid Account for use in tests.
func SeedAccount(ownerID uuid.UUID) *domain.Account {
	domain_ := "acme.example.com"
	return &domain.Account{
		ID:        uuid.New(),
		Name:      "Acme Corp",
		Domain:    &domain_,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SeedDeal returns a minimal valid Deal for use in tests.
func SeedDeal(ownerID, pipelineID uuid.UUID) *domain.Deal {
	return &domain.Deal{
		ID:         uuid.New(),
		Title:      "New Enterprise Deal",
		ValueCents: 500000,
		Currency:   "USD",
		Stage:      domain.DealStageQualified,
		OwnerID:    ownerID,
		PipelineID: pipelineID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
