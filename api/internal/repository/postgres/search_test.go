package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
)

func TestAddSearchVisibilityWhereAddsOwnerPredicateForPrivateModules(t *testing.T) {
	userID := uuid.New()
	ctx := domain.WithAccessContext(context.Background(), &domain.AccessContext{
		UserID: userID,
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleContacts: {Mode: domain.SharingDefaultPrivate},
		},
	})
	where := "c.deleted_at IS NULL AND c.org_id = $1"
	args := []any{domain.DefaultOrgID, "acme"}

	addSearchVisibilityWhere(ctx, &where, &args, domain.ACLModuleContacts, domain.SharingAccessRead, "c.owner_id = %s")

	require.Contains(t, where, "AND c.owner_id = $3")
	require.Equal(t, []any{domain.DefaultOrgID, "acme", userID}, args)
}

func TestAddSearchVisibilityWhereAddsTicketAssigneeOrSubmitterPredicate(t *testing.T) {
	userID := uuid.New()
	ctx := domain.WithAccessContext(context.Background(), &domain.AccessContext{
		UserID: userID,
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleTickets: {Mode: domain.SharingDefaultPrivate},
		},
	})
	where := "t.deleted_at IS NULL AND t.org_id = $1"
	args := []any{domain.DefaultOrgID, "printer"}

	addSearchVisibilityWhere(ctx, &where, &args, domain.ACLModuleTickets, domain.SharingAccessRead, "(t.assignee_id = %s OR t.submitted_by_user_id = %s)")

	require.Contains(t, where, "AND (t.assignee_id = $3 OR t.submitted_by_user_id = $3)")
	require.Equal(t, []any{domain.DefaultOrgID, "printer", userID}, args)
}

func TestAddSearchVisibilityWhereSkipsPublicReadModules(t *testing.T) {
	ctx := domain.WithAccessContext(context.Background(), &domain.AccessContext{
		UserID: uuid.New(),
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleAccounts: {Mode: domain.SharingDefaultPublicRO},
		},
	})
	where := "a.deleted_at IS NULL AND a.org_id = $1"
	args := []any{domain.DefaultOrgID, "acme"}

	addSearchVisibilityWhere(ctx, &where, &args, domain.ACLModuleAccounts, domain.SharingAccessRead, "a.owner_id = %s")

	require.Equal(t, "a.deleted_at IS NULL AND a.org_id = $1", where)
	require.Equal(t, []any{domain.DefaultOrgID, "acme"}, args)
}
