package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

// OrgRepo is the PostgreSQL implementation of repository.OrgRepository.
type OrgRepo struct {
	db *pgxpool.Pool
}

func NewOrgRepo(db *pgxpool.Pool) *OrgRepo {
	return &OrgRepo{db: db}
}

// Create inserts a new organization row. The caller must supply a unique slug.
func (r *OrgRepo) Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	org.CreatedAt = time.Now().UTC()

	err := r.db.QueryRow(ctx, `
		INSERT INTO orgs (id, name, slug, plan, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, slug, plan, created_at
	`, org.ID, org.Name, org.Slug, org.Plan, org.CreatedAt).
		Scan(&org.ID, &org.Name, &org.Slug, &org.Plan, &org.CreatedAt)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// GetByID returns the org with the given ID, or domain.ErrNotFound.
func (r *OrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	var org domain.Organization
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, plan, created_at
		FROM orgs WHERE id = $1
	`, id).Scan(&org.ID, &org.Name, &org.Slug, &org.Plan, &org.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetBySlug returns the org with the given slug, or domain.ErrNotFound.
func (r *OrgRepo) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	var org domain.Organization
	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug, plan, created_at
		FROM orgs WHERE slug = $1
	`, slug).Scan(&org.ID, &org.Name, &org.Slug, &org.Plan, &org.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// SlugExists reports whether the slug is already taken.
func (r *OrgRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM orgs WHERE slug = $1)`, slug,
	).Scan(&exists)
	return exists, err
}
