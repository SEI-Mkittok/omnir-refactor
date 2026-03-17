//go:build integration

package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountRepo_Create(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewAccountRepo(pool)

	accountDomain := "acme.example.com"
	industry := "Technology"
	size := domain.AccountSize11_50

	tests := []struct {
		name    string
		input   *domain.Account
		wantErr bool
	}{
		{
			name: "creates account with all fields",
			input: &domain.Account{
				Name:     "Acme Corp",
				Domain:   &accountDomain,
				Industry: &industry,
				Size:     &size,
				OwnerID:  ownerID,
			},
		},
		{
			name: "creates account with minimal fields",
			input: &domain.Account{
				Name:    "Minimal Corp",
				OwnerID: ownerID,
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
			assert.Equal(t, tt.input.Name, got.Name)
			assert.Equal(t, tt.input.OwnerID, got.OwnerID)
			assert.Nil(t, got.DeletedAt)
		})
	}
}

func TestAccountRepo_GetByID(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewAccountRepo(pool)

	created, err := repo.Create(ctx, &domain.Account{
		Name:    "Acme Corp",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "returns existing account",
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
			assert.Equal(t, "Acme Corp", got.Name)
		})
	}
}

func TestAccountRepo_Update(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewAccountRepo(pool)

	created, err := repo.Create(ctx, &domain.Account{
		Name:    "Acme Corp",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	newName := "Acme International"
	newIndustry := "Finance"

	tests := []struct {
		name    string
		id      uuid.UUID
		patch   domain.AccountPatch
		wantErr error
		check   func(t *testing.T, got *domain.Account)
	}{
		{
			name: "updates name and industry",
			id:   created.ID,
			patch: domain.AccountPatch{
				Name:     &newName,
				Industry: &newIndustry,
			},
			check: func(t *testing.T, got *domain.Account) {
				assert.Equal(t, "Acme International", got.Name)
				require.NotNil(t, got.Industry)
				assert.Equal(t, "Finance", *got.Industry)
			},
		},
		{
			name:    "returns ErrNotFound for unknown id",
			id:      uuid.New(),
			patch:   domain.AccountPatch{Name: &newName},
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

func TestAccountRepo_Delete(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewAccountRepo(pool)

	created, err := repo.Create(ctx, &domain.Account{
		Name:    "Acme Corp",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "soft-deletes existing account",
			id:   created.ID,
		},
		{
			name:    "returns ErrNotFound for already-deleted id",
			id:      created.ID,
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
			_, err = repo.GetByID(ctx, tt.id)
			require.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func TestAccountRepo_List(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewAccountRepo(pool)

	industry1 := "Technology"
	industry2 := "Finance"
	size1 := domain.AccountSize11_50

	_, err := repo.Create(ctx, &domain.Account{Name: "Acme Corp", Industry: &industry1, Size: &size1, OwnerID: ownerID})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Account{Name: "Beta Finance", Industry: &industry2, OwnerID: ownerID})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Account{Name: "Gamma Tech", Industry: &industry1, OwnerID: ownerID})
	require.NoError(t, err)

	tests := []struct {
		name      string
		filter    domain.AccountFilter
		wantCount int
		wantTotal int
	}{
		{
			name:      "lists all accounts",
			filter:    domain.AccountFilter{},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "filters by industry",
			filter:    domain.AccountFilter{Industry: &industry1},
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name:      "filters by size",
			filter:    domain.AccountFilter{Size: &size1},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "filters by owner_id",
			filter:    domain.AccountFilter{OwnerID: &ownerID},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "searches by name (q)",
			filter:    domain.AccountFilter{Q: "acme"},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "paginates: page 1 limit 2",
			filter:    domain.AccountFilter{Page: 1, Limit: 2},
			wantCount: 2,
			wantTotal: 3,
		},
		{
			name:      "paginates: page 2 limit 2",
			filter:    domain.AccountFilter{Page: 2, Limit: 2},
			wantCount: 1,
			wantTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, total, err := repo.List(ctx, tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, accounts, tt.wantCount)
		})
	}
}
