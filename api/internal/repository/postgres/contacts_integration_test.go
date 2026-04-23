//go:build integration

package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestContactRepo_Create(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewContactRepo(pool)

	email := "ada@omnir.test"
	tests := []struct {
		name    string
		input   *domain.Contact
		wantErr bool
	}{
		{
			name: "creates contact with all fields",
			input: &domain.Contact{
				FirstName: "Ada",
				LastName:  "Lovelace",
				Email:     &email,
				OwnerID:   ownerID,
				Stage:     domain.ContactStageProspect,
			},
		},
		{
			name: "creates contact with minimal fields",
			input: &domain.Contact{
				FirstName: "Grace",
				LastName:  "Hopper",
				OwnerID:   ownerID,
				Stage:     domain.ContactStageLead,
			},
		},
		{
			name: "defaults empty stage to lead",
			input: &domain.Contact{
				FirstName: "Linus",
				LastName:  "Torvalds",
				OwnerID:   ownerID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.Create(ctx, tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, got.ID)
			assert.Equal(t, defaultOrgID, got.OrgID)
			assert.Equal(t, tt.input.FirstName, got.FirstName)
			assert.Equal(t, tt.input.LastName, got.LastName)
			assert.Equal(t, tt.input.OwnerID, got.OwnerID)
			if tt.input.Stage == "" {
				assert.Equal(t, domain.ContactStageLead, got.Stage)
			}
			assert.Nil(t, got.DeletedAt)
		})
	}
}

func TestContactRepo_GetByID(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewContactRepo(pool)

	email := "ada@omnir.test"
	created, err := repo.Create(ctx, &domain.Contact{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     &email,
		OwnerID:   ownerID,
		Stage:     domain.ContactStageProspect,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "returns existing contact",
			id:   created.ID,
		},
		{
			name:    "returns ErrNotFound for unknown id",
			id:      uuid.New(),
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, created.ID, got.ID)
			assert.Equal(t, "Ada", got.FirstName)
			assert.Equal(t, "Lovelace", got.LastName)
		})
	}
}

func TestContactRepo_Update(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewContactRepo(pool)

	email := "ada@omnir.test"
	created, err := repo.Create(ctx, &domain.Contact{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     &email,
		OwnerID:   ownerID,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)

	newFirst := "Grace"
	newStage := domain.ContactStageCustomer

	tests := []struct {
		name    string
		id      uuid.UUID
		patch   domain.ContactPatch
		wantErr error
		check   func(t *testing.T, got *domain.Contact)
	}{
		{
			name: "updates first_name and stage",
			id:   created.ID,
			patch: domain.ContactPatch{
				FirstName: &newFirst,
				Stage:     &newStage,
			},
			check: func(t *testing.T, got *domain.Contact) {
				assert.Equal(t, "Grace", got.FirstName)
				assert.Equal(t, domain.ContactStageCustomer, got.Stage)
				assert.Equal(t, "Lovelace", got.LastName) // unchanged
			},
		},
		{
			name:    "returns ErrNotFound for unknown id",
			id:      uuid.New(),
			patch:   domain.ContactPatch{FirstName: &newFirst},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.Update(ctx, tt.id, tt.patch)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}

func TestContactRepo_Delete(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewContactRepo(pool)

	created, err := repo.Create(ctx, &domain.Contact{
		FirstName: "Ada",
		LastName:  "Lovelace",
		OwnerID:   ownerID,
		Stage:     domain.ContactStageLead,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "soft-deletes existing contact",
			id:   created.ID,
		},
		{
			name:    "returns ErrNotFound for already-deleted id",
			id:      created.ID, // second call on same id
			wantErr: domain.ErrNotFound,
		},
		{
			name:    "returns ErrNotFound for unknown id",
			id:      uuid.New(),
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			// Verify soft-deleted: GetByID should return not found.
			_, err = repo.GetByID(ctx, tt.id)
			require.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func TestContactRepo_List(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewContactRepo(pool)

	// Seed 3 contacts.
	email1, email2 := "ada@omnir.test", "grace@omnir.test"
	stage1 := domain.ContactStageProspect
	c1, err := repo.Create(ctx, &domain.Contact{FirstName: "Ada", LastName: "Lovelace", Email: &email1, OwnerID: ownerID, Stage: stage1})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Contact{FirstName: "Grace", LastName: "Hopper", Email: &email2, OwnerID: ownerID, Stage: domain.ContactStageLead})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Contact{FirstName: "Margaret", LastName: "Hamilton", OwnerID: ownerID, Stage: domain.ContactStageCustomer})
	require.NoError(t, err)

	tests := []struct {
		name      string
		filter    domain.ContactFilter
		wantCount int
		wantTotal int
	}{
		{
			name:      "lists all contacts",
			filter:    domain.ContactFilter{},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "filters by stage",
			filter:    domain.ContactFilter{Stage: &stage1},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "filters by owner_id",
			filter:    domain.ContactFilter{OwnerID: &ownerID},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "filters by owner_id with no match",
			filter:    domain.ContactFilter{OwnerID: func() *uuid.UUID { id := uuid.New(); return &id }()},
			wantCount: 0,
			wantTotal: 0,
		},
		{
			name:      "searches by name (q)",
			filter:    domain.ContactFilter{Q: "ada"},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "paginates: page 1 limit 2",
			filter:    domain.ContactFilter{Page: 1, Limit: 2},
			wantCount: 2,
			wantTotal: 3,
		},
		{
			name:      "paginates: page 2 limit 2",
			filter:    domain.ContactFilter{Page: 2, Limit: 2},
			wantCount: 1,
			wantTotal: 3,
		},
	}

	_ = c1 // used indirectly via stage filter

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contacts, total, err := repo.List(ctx, tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, contacts, tt.wantCount)
		})
	}
}
