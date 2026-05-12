package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// SearchRepo implements repository.SearchRepository using PostgreSQL full-text search.
type SearchRepo struct {
	db *pgxpool.Pool
}

func NewSearchRepo(db *pgxpool.Pool) *SearchRepo {
	return &SearchRepo{db: db}
}

// Search runs to_tsvector @@ plainto_tsquery queries across tickets, contacts,
// accounts, and deals, all scoped to the org in context.
// Returns results grouped by entity type to match the frontend SearchResult shape.
func (r *SearchRepo) Search(ctx context.Context, filter domain.SearchFilter) (*domain.SearchGroupedResult, error) {
	orgID, hasOrg := domain.OrgIDFromContext(ctx)
	if !hasOrg || orgID == uuid.Nil {
		return &domain.SearchGroupedResult{
			Contacts: []domain.SearchContact{},
			Accounts: []domain.SearchAccount{},
			Deals:    []domain.SearchDeal{},
			Tickets:  []domain.SearchTicket{},
		}, nil
	}

	out := &domain.SearchGroupedResult{
		Contacts: []domain.SearchContact{},
		Accounts: []domain.SearchAccount{},
		Deals:    []domain.SearchDeal{},
		Tickets:  []domain.SearchTicket{},
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if (filter.EntityType == "" || filter.EntityType == domain.SearchEntityContacts) && filter.ContactID == nil {
		if err := r.searchContacts(ctx, orgID, filter, out); err != nil {
			return nil, fmt.Errorf("contacts search: %w", err)
		}
	}
	if filter.EntityType == "" || filter.EntityType == domain.SearchEntityAccounts {
		if err := r.searchAccounts(ctx, orgID, filter, out); err != nil {
			return nil, fmt.Errorf("accounts search: %w", err)
		}
	}
	if filter.EntityType == "" || filter.EntityType == domain.SearchEntityDeals {
		if err := r.searchDeals(ctx, orgID, filter, out); err != nil {
			return nil, fmt.Errorf("deals search: %w", err)
		}
	}
	if filter.EntityType == "" || filter.EntityType == domain.SearchEntityTickets {
		if err := r.searchTickets(ctx, orgID, filter, out); err != nil {
			return nil, fmt.Errorf("tickets search: %w", err)
		}
	}

	return out, nil
}

func relatedAccount(id, name sql.NullString) *domain.SearchRelationship {
	if !id.Valid || !name.Valid || id.String == "" {
		return nil
	}
	return &domain.SearchRelationship{EntityType: "account", ID: id.String, Name: name.String}
}

func relatedContact(id, name sql.NullString) *domain.SearchRelationship {
	if !id.Valid || !name.Valid || id.String == "" {
		return nil
	}
	return &domain.SearchRelationship{EntityType: "contact", ID: id.String, Name: name.String}
}

func relationshipFilterClause(column string, argIndex int, relationshipType string) (string, []any) {
	if strings.TrimSpace(relationshipType) == "" {
		return "", nil
	}
	return fmt.Sprintf(" AND %s = $%d", column, argIndex), []any{relationshipType}
}

func addSearchVisibilityWhere(ctx context.Context, where *string, args *[]any, module domain.ACLModule, accessLevel domain.SharingAccessLevel, ownerExprs ...string) {
	predicate, predicateArgs := accessVisibilityPredicate(ctx, len(*args)+1, module, accessLevel, ownerExprs...)
	if predicate == "" {
		return
	}
	*where += " AND " + predicate
	*args = append(*args, predicateArgs...)
}

func (r *SearchRepo) searchContacts(ctx context.Context, orgID uuid.UUID, filter domain.SearchFilter, out *domain.SearchGroupedResult) error {
	args := []any{orgID, filter.Query}
	accountArgIndex := 0
	where := `
		c.deleted_at IS NULL
		AND c.org_id = $1
		AND to_tsvector('english',
			c.first_name || ' ' || c.last_name || ' ' ||
			coalesce(c.email, '') || ' ' || coalesce(c.phone, ''))
		  @@ plainto_tsquery('english', $2)`
	if filter.AccountID != nil {
		args = append(args, *filter.AccountID)
		accountArgIndex = len(args)
		where += fmt.Sprintf(`
		AND (
			c.account_id = $%d
			OR EXISTS (
				SELECT 1 FROM account_contacts ac
				WHERE ac.contact_id = c.id AND ac.account_id = $%d
			)
		)`, len(args), len(args))
	}
	if accountArgIndex > 0 {
		if clause, extra := relationshipFilterClause(`
		CASE
			WHEN c.account_id = ra.id THEN 'primary'
			WHEN rac.is_primary THEN 'primary'
			WHEN rac.account_id IS NOT NULL THEN 'linked'
			ELSE ''
		END`, len(args)+1, filter.RelationshipType); clause != "" {
			where += clause
			args = append(args, extra...)
		}
	}
	addSearchVisibilityWhere(ctx, &where, &args, domain.ACLModuleContacts, domain.SharingAccessRead, "c.owner_id")
	args = append(args, filter.Limit)
	accountContactJoin := "LEFT JOIN account_contacts rac ON false"
	if accountArgIndex > 0 {
		accountContactJoin = fmt.Sprintf("LEFT JOIN account_contacts rac ON rac.contact_id = c.id AND rac.account_id = $%d", accountArgIndex)
	}

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT c.id, c.first_name, c.last_name, coalesce(c.email, ''), c.stage,
		       ra.id::text, ra.name,
		       CASE
		         WHEN c.account_id = ra.id THEN 'primary'
		         WHEN rac.is_primary THEN 'primary'
		         WHEN rac.account_id IS NOT NULL THEN 'linked'
		         ELSE ''
		       END
		FROM contacts c
		LEFT JOIN accounts ra ON ra.id = c.account_id AND ra.deleted_at IS NULL
		%s
		WHERE %s
		ORDER BY c.updated_at DESC
		LIMIT $%d`, accountContactJoin, where, len(args)), args...)
	if err != nil {
		return fmt.Errorf("contacts search: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.SearchContact
		var accountID, accountName sql.NullString
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Stage, &accountID, &accountName, &c.RelationshipType); err != nil {
			return err
		}
		c.RelatedAccount = relatedAccount(accountID, accountName)
		out.Contacts = append(out.Contacts, c)
	}
	return rows.Err()
}

func (r *SearchRepo) searchAccounts(ctx context.Context, orgID uuid.UUID, filter domain.SearchFilter, out *domain.SearchGroupedResult) error {
	args := []any{orgID, filter.Query}
	where := `
		a.deleted_at IS NULL
		AND a.org_id = $1
		AND to_tsvector('english',
			a.name || ' ' || coalesce(a.domain, '') || ' ' || coalesce(a.industry, ''))
		  @@ plainto_tsquery('english', $2)`
	joins := ""
	selectRelated := "NULL::text, NULL::text, NULL::text, NULL::text, ''"
	if filter.AccountID != nil {
		args = append(args, *filter.AccountID)
		joins += fmt.Sprintf(`
		LEFT JOIN account_relationships ar
		  ON ar.org_id = a.org_id
		 AND ar.deleted_at IS NULL
		 AND ((ar.parent_account_id = $%d AND ar.child_account_id = a.id)
		   OR (ar.child_account_id = $%d AND ar.parent_account_id = a.id))`, len(args), len(args))
		where += " AND ar.id IS NOT NULL"
		selectRelated = fmt.Sprintf("$%d::text, scope_account.name, NULL::text, NULL::text, ar.relationship_type::text", len(args)+1)
		args = append(args, *filter.AccountID)
		joins += fmt.Sprintf(" LEFT JOIN accounts scope_account ON scope_account.id = $%d", len(args))
		if filter.RelationshipType != "" {
			args = append(args, filter.RelationshipType)
			where += fmt.Sprintf(" AND ar.relationship_type::text = $%d", len(args))
		}
	}
	if filter.ContactID != nil {
		args = append(args, *filter.ContactID)
		joins += fmt.Sprintf(" LEFT JOIN contacts rc ON rc.id = $%d AND rc.deleted_at IS NULL", len(args))
		where += fmt.Sprintf(`
		AND (
			a.id = rc.account_id
			OR EXISTS (
				SELECT 1 FROM account_contacts ac
				WHERE ac.account_id = a.id AND ac.contact_id = $%d
			)
		)`, len(args))
		selectRelated = "NULL::text, NULL::text, rc.id::text, trim(rc.first_name || ' ' || rc.last_name), 'linked'"
	}
	addSearchVisibilityWhere(ctx, &where, &args, domain.ACLModuleAccounts, domain.SharingAccessRead, "a.owner_id")
	args = append(args, filter.Limit)

	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT a.id, a.name, a.domain, a.industry, %s
		FROM accounts a
		%s
		WHERE %s
		ORDER BY a.updated_at DESC
		LIMIT $%d`, selectRelated, joins, where, len(args)), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var a domain.SearchAccount
		var accountID, accountName, contactID, contactName sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &a.Domain, &a.Industry, &accountID, &accountName, &contactID, &contactName, &a.RelationshipType); err != nil {
			return err
		}
		a.RelatedAccount = relatedAccount(accountID, accountName)
		a.RelatedContact = relatedContact(contactID, contactName)
		out.Accounts = append(out.Accounts, a)
	}
	return rows.Err()
}

func (r *SearchRepo) searchDeals(ctx context.Context, orgID uuid.UUID, filter domain.SearchFilter, out *domain.SearchGroupedResult) error {
	args := []any{orgID, filter.Query}
	where := "d.deleted_at IS NULL AND d.org_id = $1 AND to_tsvector('english', d.title) @@ plainto_tsquery('english', $2)"
	joins := "LEFT JOIN accounts ra ON ra.id = d.account_id LEFT JOIN contacts rc ON rc.id = d.contact_id LEFT JOIN deal_contacts dc ON dc.deal_id = d.id AND dc.contact_id = d.contact_id"
	if filter.AccountID != nil {
		args = append(args, *filter.AccountID)
		where += fmt.Sprintf(" AND d.account_id = $%d", len(args))
	}
	if filter.ContactID != nil {
		args = append(args, *filter.ContactID)
		joins += fmt.Sprintf(" LEFT JOIN deal_contacts scoped_dc ON scoped_dc.deal_id = d.id AND scoped_dc.contact_id = $%d", len(args))
		where += fmt.Sprintf(" AND (d.contact_id = $%d OR scoped_dc.contact_id IS NOT NULL)", len(args))
		if filter.RelationshipType != "" {
			args = append(args, filter.RelationshipType)
			where += fmt.Sprintf(" AND coalesce(nullif(scoped_dc.role, ''), 'linked') = $%d", len(args))
		}
	}
	addSearchVisibilityWhere(ctx, &where, &args, domain.ACLModuleDeals, domain.SharingAccessRead, "d.owner_id")
	args = append(args, filter.Limit)
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT d.id, d.title, d.stage, d.value_cents,
		       ra.id::text, ra.name, rc.id::text, trim(rc.first_name || ' ' || rc.last_name),
		       CASE WHEN dc.role IS NOT NULL AND dc.role <> '' THEN dc.role WHEN d.contact_id IS NOT NULL THEN 'primary' WHEN d.account_id IS NOT NULL THEN 'linked' ELSE '' END
		FROM deals d
		%s
		WHERE %s
		ORDER BY d.updated_at DESC
		LIMIT $%d`, joins, where, len(args)), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var d domain.SearchDeal
		var valueCents int64
		var accountID, accountName, contactID, contactName sql.NullString
		if err := rows.Scan(&d.ID, &d.Title, &d.Stage, &valueCents, &accountID, &accountName, &contactID, &contactName, &d.RelationshipType); err != nil {
			return err
		}
		d.Value = float64(valueCents) / 100.0
		d.RelatedAccount = relatedAccount(accountID, accountName)
		d.RelatedContact = relatedContact(contactID, contactName)
		out.Deals = append(out.Deals, d)
	}
	return rows.Err()
}

func (r *SearchRepo) searchTickets(ctx context.Context, orgID uuid.UUID, filter domain.SearchFilter, out *domain.SearchGroupedResult) error {
	args := []any{orgID, filter.Query}
	where := `
		  t.deleted_at IS NULL
		  AND t.org_id = $1
		  AND (
		      to_tsvector('english', t.subject || ' ' || coalesce(t.description, '')) @@ plainto_tsquery('english', $2)
		    OR EXISTS (
		        SELECT 1 FROM ticket_comments tc
		        WHERE tc.ticket_id = t.id
		          AND tc.deleted_at IS NULL
		          AND to_tsvector('english', tc.body) @@ plainto_tsquery('english', $2)
		    )
		  )`
	if filter.AccountID != nil {
		args = append(args, *filter.AccountID)
		where += fmt.Sprintf(" AND t.account_id = $%d", len(args))
	}
	if filter.ContactID != nil {
		args = append(args, *filter.ContactID)
		where += fmt.Sprintf(" AND t.contact_id = $%d", len(args))
	}
	addSearchVisibilityWhere(ctx, &where, &args, domain.ACLModuleTickets, domain.SharingAccessRead, "t.assignee_id", "t.submitted_by_user_id")
	args = append(args, filter.Limit)
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT t.id, t.subject, t.status, t.priority,
		       ra.id::text, ra.name, rc.id::text, trim(rc.first_name || ' ' || rc.last_name),
		       CASE WHEN t.contact_id IS NOT NULL THEN 'requester' WHEN t.account_id IS NOT NULL THEN 'linked' ELSE '' END
		FROM tickets t
		LEFT JOIN accounts ra ON ra.id = t.account_id
		LEFT JOIN contacts rc ON rc.id = t.contact_id
		WHERE %s
		ORDER BY t.updated_at DESC
		LIMIT $%d`, where, len(args)), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.SearchTicket
		var accountID, accountName, contactID, contactName sql.NullString
		if err := rows.Scan(&t.ID, &t.Subject, &t.Status, &t.Priority, &accountID, &accountName, &contactID, &contactName, &t.RelationshipType); err != nil {
			return err
		}
		t.RelatedAccount = relatedAccount(accountID, accountName)
		t.RelatedContact = relatedContact(contactID, contactName)
		out.Tickets = append(out.Tickets, t)
	}
	return rows.Err()
}
