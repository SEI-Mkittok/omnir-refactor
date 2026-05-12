package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

func TestAddActivityParentVisibilityWhereAddsParentPredicates(t *testing.T) {
	userID := uuid.New()
	ctx := domain.WithAccessContext(context.Background(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleContacts: {Mode: domain.SharingDefaultPrivate},
			domain.ACLModuleAccounts: {Mode: domain.SharingDefaultPrivate},
			domain.ACLModuleDeals:    {Mode: domain.SharingDefaultPrivate},
		},
	})
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1

	addActivityParentVisibilityWhere(ctx, &where, &args, &idx, domain.SharingAccessRead, "activities")
	joined := strings.Join(where, " AND ")

	for _, want := range []string{
		"activities.contact_id IS NULL OR EXISTS",
		"FROM contacts p",
		"activities.account_id IS NULL OR EXISTS",
		"FROM accounts p",
		"activities.deal_id IS NULL OR EXISTS",
		"FROM deals p",
		"p.owner_id = $1",
		"p.owner_id = $2",
		"p.owner_id = $3",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected activity parent visibility SQL to contain %q, got: %s", want, joined)
		}
	}
	if len(args) != 3 {
		t.Fatalf("expected one owner arg per parent module, got %d", len(args))
	}
	for i, arg := range args {
		if got, ok := arg.(uuid.UUID); !ok || got != userID {
			t.Fatalf("expected arg %d to be user id %s, got %#v", i, userID, arg)
		}
	}
}

func TestAddActivityParentVisibilityWhereSkipsWithoutAccessContext(t *testing.T) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1

	addActivityParentVisibilityWhere(context.Background(), &where, &args, &idx, domain.SharingAccessRead, "activities")

	if len(where) != 1 {
		t.Fatalf("expected where unchanged without access context, got %#v", where)
	}
	if len(args) != 0 || idx != 1 {
		t.Fatalf("expected args and idx unchanged, got len(args)=%d idx=%d", len(args), idx)
	}
}
