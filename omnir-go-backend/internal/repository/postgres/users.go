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

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// CountAll returns the total number of non-deleted users (across all orgs).
func (r *UserRepo) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	return count, err
}

// Create inserts a new user with the given bcrypt password hash.
func (r *UserRepo) Create(ctx context.Context, u *domain.User, passwordHash string) (*domain.User, error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO users (id, org_id, email, name, role, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, org_id, email, name, role, avatar_url, created_at, updated_at, deleted_at
	`, u.ID, u.OrgID, u.Email, u.Name, u.Role, passwordHash, u.CreatedAt, u.UpdatedAt)

	return scanUser(row)
}

// FindByEmail returns the user and bcrypt password hash for the given email.
// Returns nil user (not error) when not found.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	var u domain.User
	var passwordHash string
	err := r.db.QueryRow(ctx, `
		SELECT id, org_id, email, name, role, avatar_url, created_at, updated_at, deleted_at, password_hash
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(
		&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role,
		&u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &passwordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}
	return &u, passwordHash, nil
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role,
		&u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
