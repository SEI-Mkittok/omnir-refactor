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
func (r *SearchRepo) Search(ctx context.Context, q string, limit int) ([]domain.SearchResultItem, int, error) {
	orgID, hasOrg := domain.OrgIDFromContext(ctx)
	if !hasOrg || orgID == uuid.Nil {
		return []domain.SearchResultItem{}, 0, nil
	}

	var results []domain.SearchResultItem

	if err := r.searchTickets(ctx, orgID, q, limit, &results); err != nil {
		return nil, 0, fmt.Errorf("tickets search: %w", err)
	}
	if err := r.searchContacts(ctx, orgID, q, limit, &results); err != nil {
		return nil, 0, fmt.Errorf("contacts search: %w", err)
	}
	if err := r.searchAccounts(ctx, orgID, q, limit, &results); err != nil {
		return nil, 0, fmt.Errorf("accounts search: %w", err)
	}
	if err := r.searchDeals(ctx, orgID, q, limit, &results); err != nil {
		return nil, 0, fmt.Errorf("deals search: %w", err)
	}

	return results, len(results), nil
}

func (r *SearchRepo) searchTickets(ctx context.Context, orgID uuid.UUID, q string, limit int, out *[]domain.SearchResultItem) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, subject, coalesce(description, '')
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
		var id uuid.UUID
		var subject, description string
		if err := rows.Scan(&id, &subject, &description); err != nil {
			return err
		}
		excerpt := description
		if len(excerpt) > 120 {
			excerpt = excerpt[:120] + "…"
		}
		*out = append(*out, domain.SearchResultItem{
			Type:    "ticket",
			ID:      id.String(),
			Title:   subject,
			Excerpt: excerpt,
			URL:     "/tickets/" + id.String(),
		})
	}
	return rows.Err()
}

func (r *SearchRepo) searchContacts(ctx context.Context, orgID uuid.UUID, q string, limit int, out *[]domain.SearchResultItem) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, first_name, last_name, coalesce(email, '')
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
		var id uuid.UUID
		var firstName, lastName, email string
		if err := rows.Scan(&id, &firstName, &lastName, &email); err != nil {
			return err
		}
		*out = append(*out, domain.SearchResultItem{
			Type:    "contact",
			ID:      id.String(),
			Title:   firstName + " " + lastName,
			Excerpt: email,
			URL:     "/contacts/" + id.String(),
		})
	}
	return rows.Err()
}

func (r *SearchRepo) searchAccounts(ctx context.Context, orgID uuid.UUID, q string, limit int, out *[]domain.SearchResultItem) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, coalesce(domain, '')
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
		var id uuid.UUID
		var name, accountDomain string
		if err := rows.Scan(&id, &name, &accountDomain); err != nil {
			return err
		}
		*out = append(*out, domain.SearchResultItem{
			Type:    "account",
			ID:      id.String(),
			Title:   name,
			Excerpt: accountDomain,
			URL:     "/accounts/" + id.String(),
		})
	}
	return rows.Err()
}

func (r *SearchRepo) searchDeals(ctx context.Context, orgID uuid.UUID, q string, limit int, out *[]domain.SearchResultItem) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, stage
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
		var id uuid.UUID
		var title, stage string
		if err := rows.Scan(&id, &title, &stage); err != nil {
			return err
		}
		*out = append(*out, domain.SearchResultItem{
			Type:    "deal",
			ID:      id.String(),
			Title:   title,
			Excerpt: stage,
			URL:     "/deals/" + id.String(),
		})
	}
	return rows.Err()
}
