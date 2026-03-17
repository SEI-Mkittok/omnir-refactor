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

func TestDealRepo_Create(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewDealRepo(pool)

	tests := []struct {
		name    string
		input   *domain.Deal
		wantErr bool
	}{
		{
			name: "creates deal with all required fields",
			input: &domain.Deal{
				Title:      "Enterprise Deal",
				ValueCents: 500000,
				Currency:   "USD",
				Stage:      domain.DealStageQualified,
				OwnerID:    ownerID,
				PipelineID: defaultPipelineID,
			},
		},
		{
			name: "creates deal with default currency",
			input: &domain.Deal{
				Title:      "Minimal Deal",
				Stage:      domain.DealStageLead,
				OwnerID:    ownerID,
				PipelineID: defaultPipelineID,
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
			assert.Equal(t, tt.input.Title, got.Title)
			assert.Equal(t, tt.input.OwnerID, got.OwnerID)
			assert.Equal(t, tt.input.PipelineID, got.PipelineID)
			assert.Nil(t, got.DeletedAt)
			if tt.input.Currency == "" {
				assert.Equal(t, "USD", got.Currency)
			}
		})
	}
}

func TestDealRepo_GetByID(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewDealRepo(pool)

	created, err := repo.Create(ctx, &domain.Deal{
		Title:      "Enterprise Deal",
		ValueCents: 500000,
		Stage:      domain.DealStageQualified,
		OwnerID:    ownerID,
		PipelineID: defaultPipelineID,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "returns existing deal",
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
			assert.Equal(t, "Enterprise Deal", got.Title)
			assert.Equal(t, int64(500000), got.ValueCents)
		})
	}
}

func TestDealRepo_Update(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewDealRepo(pool)

	created, err := repo.Create(ctx, &domain.Deal{
		Title:      "Enterprise Deal",
		ValueCents: 500000,
		Stage:      domain.DealStageQualified,
		OwnerID:    ownerID,
		PipelineID: defaultPipelineID,
	})
	require.NoError(t, err)

	newTitle := "Updated Deal"
	newValue := int64(1000000)
	newStage := domain.DealStageProposal

	tests := []struct {
		name    string
		id      uuid.UUID
		patch   domain.DealPatch
		wantErr error
		check   func(t *testing.T, got *domain.Deal)
	}{
		{
			name: "updates title, value, and stage",
			id:   created.ID,
			patch: domain.DealPatch{
				Title:      &newTitle,
				ValueCents: &newValue,
				Stage:      &newStage,
			},
			check: func(t *testing.T, got *domain.Deal) {
				assert.Equal(t, "Updated Deal", got.Title)
				assert.Equal(t, int64(1000000), got.ValueCents)
				assert.Equal(t, domain.DealStageProposal, got.Stage)
			},
		},
		{
			name:    "returns ErrNotFound for unknown id",
			id:      uuid.New(),
			patch:   domain.DealPatch{Title: &newTitle},
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

func TestDealRepo_Delete(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewDealRepo(pool)

	created, err := repo.Create(ctx, &domain.Deal{
		Title:      "Enterprise Deal",
		Stage:      domain.DealStageQualified,
		OwnerID:    ownerID,
		PipelineID: defaultPipelineID,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "soft-deletes existing deal",
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

func TestDealRepo_List(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	repo := postgres.NewDealRepo(pool)

	stage1 := domain.DealStageQualified
	stage2 := domain.DealStageLead

	_, err := repo.Create(ctx, &domain.Deal{Title: "Enterprise Deal", ValueCents: 500000, Stage: stage1, OwnerID: ownerID, PipelineID: defaultPipelineID})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Deal{Title: "SMB Deal", ValueCents: 10000, Stage: stage2, OwnerID: ownerID, PipelineID: defaultPipelineID})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Deal{Title: "Startup Deal", ValueCents: 5000, Stage: stage2, OwnerID: ownerID, PipelineID: defaultPipelineID})
	require.NoError(t, err)

	tests := []struct {
		name      string
		filter    domain.DealFilter
		wantCount int
		wantTotal int
	}{
		{
			name:      "lists all deals",
			filter:    domain.DealFilter{},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "filters by stage qualified",
			filter:    domain.DealFilter{Stage: &stage1},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "filters by stage lead",
			filter:    domain.DealFilter{Stage: &stage2},
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name:      "filters by owner_id",
			filter:    domain.DealFilter{OwnerID: &ownerID},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "filters by pipeline_id",
			filter:    domain.DealFilter{PipelineID: &defaultPipelineID},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:      "searches by title (q)",
			filter:    domain.DealFilter{Q: "enterprise"},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "paginates: page 1 limit 2",
			filter:    domain.DealFilter{Page: 1, Limit: 2},
			wantCount: 2,
			wantTotal: 3,
		},
		{
			name:      "paginates: page 2 limit 2",
			filter:    domain.DealFilter{Page: 2, Limit: 2},
			wantCount: 1,
			wantTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deals, total, err := repo.List(ctx, tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Len(t, deals, tt.wantCount)
		})
	}
}

func TestDealRepo_AddContact_ListContacts(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)
	dealRepo := postgres.NewDealRepo(pool)
	contactRepo := postgres.NewContactRepo(pool)

	deal, err := dealRepo.Create(ctx, &domain.Deal{
		Title:      "Deal with contacts",
		Stage:      domain.DealStageQualified,
		OwnerID:    ownerID,
		PipelineID: defaultPipelineID,
	})
	require.NoError(t, err)

	c1, err := contactRepo.Create(ctx, &domain.Contact{FirstName: "Ada", LastName: "Lovelace", OwnerID: ownerID, Stage: domain.ContactStageLead})
	require.NoError(t, err)
	c2, err := contactRepo.Create(ctx, &domain.Contact{FirstName: "Grace", LastName: "Hopper", OwnerID: ownerID, Stage: domain.ContactStageLead})
	require.NoError(t, err)

	// Add both contacts.
	require.NoError(t, dealRepo.AddContact(ctx, deal.ID, c1.ID, "primary"))
	require.NoError(t, dealRepo.AddContact(ctx, deal.ID, c2.ID, "stakeholder"))

	t.Run("lists linked contacts", func(t *testing.T) {
		contacts, err := dealRepo.ListContacts(ctx, deal.ID)
		require.NoError(t, err)
		assert.Len(t, contacts, 2)
	})

	t.Run("GetByID populates contacts", func(t *testing.T) {
		got, err := dealRepo.GetByID(ctx, deal.ID)
		require.NoError(t, err)
		assert.Len(t, got.Contacts, 2)
	})

	t.Run("upsert updates role on conflict", func(t *testing.T) {
		require.NoError(t, dealRepo.AddContact(ctx, deal.ID, c1.ID, "champion"))
		contacts, err := dealRepo.ListContacts(ctx, deal.ID)
		require.NoError(t, err)
		assert.Len(t, contacts, 2) // still 2, role updated
	})
}
