package postgres

import (
	"context"
	"fmt"

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
func (r *SearchRepo) Search(ctx context.Context, q string, limit int) (*domain.SearchGroupedResult, error) {
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

	if err := r.searchContacts(ctx, orgID, q, limit, out); err != nil {
		return nil, fmt.Errorf("contacts search: %w", err)
	}
	if err := r.searchAccounts(ctx, orgID, q, limit, out); err != nil {
		return nil, fmt.Errorf("accounts search: %w", err)
	}
	if err := r.searchDeals(ctx, orgID, q, limit, out); err != nil {
		return nil, fmt.Errorf("deals search: %w", err)
	}
	if err := r.searchTickets(ctx, orgID, q, limit, out); err != nil {
		return nil, fmt.Errorf("tickets search: %w", err)
	}

	return out, nil
}

func (r *SearchRepo) searchContacts(ctx context.Context, orgID uuid.UUID, q string, limit int, out *domain.SearchGroupedResult) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, first_name, last_name, coalesce(email, ''), stage
		FROM contacts
		WHERE deleted_at IS NULL
		  AND org_id = $1
		  AND to_tsvector('english',
		      first_name || ' ' || last_name || ' ' ||
		      coalesce(email, '') || ' ' || coalesce(phone, ''))
		    @@ plainto_tsquery('english', $2)
		ORDER BY updated_at DESC
		LIMIT $3`,
		orgID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.SearchContact
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.Stage); err != nil {
			return err
		}
		out.Contacts = append(out.Contacts, c)
	}
	return rows.Err()
}

func (r *SearchRepo) searchAccounts(ctx context.Context, orgID uuid.UUID, q string, limit int, out *domain.SearchGroupedResult) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, domain, industry
		FROM accounts
		WHERE deleted_at IS NULL
		  AND org_id = $1
		  AND to_tsvector('english',
		      name || ' ' || coalesce(domain, '') || ' ' || coalesce(industry, ''))
		    @@ plainto_tsquery('english', $2)
		ORDER BY updated_at DESC
		LIMIT $3`,
		orgID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var a domain.SearchAccount
		if err := rows.Scan(&a.ID, &a.Name, &a.Domain, &a.Industry); err != nil {
			return err
		}
		out.Accounts = append(out.Accounts, a)
	}
	return rows.Err()
}

func (r *SearchRepo) searchDeals(ctx context.Context, orgID uuid.UUID, q string, limit int, out *domain.SearchGroupedResult) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, stage, value_cents
		FROM deals
		WHERE deleted_at IS NULL
		  AND org_id = $1
		  AND to_tsvector('english', title) @@ plainto_tsquery('english', $2)
		ORDER BY updated_at DESC
		LIMIT $3`,
		orgID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var d domain.SearchDeal
		var valueCents int64
		if err := rows.Scan(&d.ID, &d.Title, &d.Stage, &valueCents); err != nil {
			return err
		}
		d.Value = float64(valueCents) / 100.0
		out.Deals = append(out.Deals, d)
	}
	return rows.Err()
}

func (r *SearchRepo) searchTickets(ctx context.Context, orgID uuid.UUID, q string, limit int, out *domain.SearchGroupedResult) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, subject, status, priority
		FROM tickets
		WHERE deleted_at IS NULL
		  AND org_id = $1
		  AND (
		      to_tsvector('english', subject || ' ' || coalesce(description, '')) @@ plainto_tsquery('english', $2)
		    OR EXISTS (
		        SELECT 1 FROM ticket_comments tc
		        WHERE tc.ticket_id = tickets.id
		          AND tc.deleted_at IS NULL
		          AND to_tsvector('english', tc.body) @@ plainto_tsquery('english', $2)
		    )
		  )
		ORDER BY updated_at DESC
		LIMIT $3`,
		orgID, q, limit,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.SearchTicket
		if err := rows.Scan(&t.ID, &t.Subject, &t.Status, &t.Priority); err != nil {
			return err
		}
		out.Tickets = append(out.Tickets, t)
	}
	return rows.Err()
}
