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
	if !strings.Contains(query, "FROM crm_sharing_rules sr") {
		t.Fatalf("expected advanced sharing predicate, got query: %s", query)
	}
	if len(args) != 7 {
		t.Fatalf("expected owner and advanced sharing args, got %d", len(args))
	}
	if got, ok := args[2].(uuid.UUID); !ok || got != userID {
		t.Fatalf("expected visibility arg to be user id %s, got %#v", userID, args[2])
	}
}

func TestScopedReportOrgID(t *testing.T) {
	baseOrg := uuid.New()
	overrideOrg := uuid.New()
	ctx := domain.WithOrgID(context.Background(), baseOrg)

	got, err := scopedReportOrgID(ctx, domain.ReportFilter{})
	if err != nil {
		t.Fatalf("expected org id from context, got error: %v", err)
	}
	if got != baseOrg {
		t.Fatalf("expected context org %s, got %s", baseOrg, got)
	}

	got, err = scopedReportOrgID(ctx, domain.ReportFilter{OrgID: &overrideOrg})
	if err != nil {
		t.Fatalf("expected override org id, got error: %v", err)
	}
	if got != overrideOrg {
		t.Fatalf("expected override org %s, got %s", overrideOrg, got)
	}

	if _, err := scopedReportOrgID(context.Background(), domain.ReportFilter{}); err != domain.ErrNotFound {
		t.Fatalf("expected missing org context to fail with ErrNotFound, got %v", err)
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
