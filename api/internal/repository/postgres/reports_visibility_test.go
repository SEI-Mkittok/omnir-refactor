package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

func TestAppendReportReadVisibilityAddsPrivateOwnerPredicate(t *testing.T) {
	userID := uuid.New()
	ctx := domain.WithAccessContext(context.Background(), &domain.AccessContext{
		UserID: userID,
		OrgID:  domain.DefaultOrgID,
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleDeals: {Mode: domain.SharingDefaultPrivate},
		},
	})

	query := "SELECT COUNT(*) FROM deals d WHERE d.org_id = $1 AND d.created_at >= $2"
	args := []interface{}{domain.DefaultOrgID, "2026-01-01"}

	appendReportReadVisibility(ctx, &query, &args, domain.ACLModuleDeals, "d.owner_id")

	if !strings.Contains(query, "d.owner_id = $3") {
		t.Fatalf("expected owner visibility predicate to use next placeholder, got query: %s", query)
	}
	if len(args) != 3 {
		t.Fatalf("expected one visibility arg, got %d", len(args))
	}
	if got, ok := args[2].(uuid.UUID); !ok || got != userID {
		t.Fatalf("expected visibility arg to be user id %s, got %#v", userID, args[2])
	}
}

func TestAppendReportReadVisibilitySkipsAllRecordAccess(t *testing.T) {
	ctx := domain.WithAccessContext(context.Background(), &domain.AccessContext{
		UserID: uuid.New(),
		OrgID:  domain.DefaultOrgID,
		Sharing: map[domain.ACLModule]domain.ACLSharingAccess{
			domain.ACLModuleContacts: {Mode: domain.SharingDefaultPublicRO},
		},
	})

	query := "SELECT COUNT(*) FROM contacts c WHERE c.org_id = $1"
	args := []interface{}{domain.DefaultOrgID}

	appendReportReadVisibility(ctx, &query, &args, domain.ACLModuleContacts, "c.owner_id")

	if strings.Contains(query, "c.owner_id") {
		t.Fatalf("expected no owner predicate for public read access, got query: %s", query)
	}
	if len(args) != 1 {
		t.Fatalf("expected args unchanged, got %d", len(args))
	}
}
