package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

const userCols = `id, org_id, email, name, role, avatar_url, created_at, updated_at, deleted_at`

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

// HasAdminUser returns true if at least one non-deleted admin user exists across all orgs.
func (r *UserRepo) HasAdminUser(ctx context.Context) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE deleted_at IS NULL AND role = 'admin')`).Scan(&exists)
	return exists, err
}

// Create inserts a new user with the given bcrypt password hash.
func (r *UserRepo) Create(ctx context.Context, u *domain.User, passwordHash string) (*domain.User, error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		u.OrgID = orgID
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

// GetByID returns a user by ID, scoped to the org in context.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	q := `SELECT ` + userCols + ` FROM users WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	return scanUser(r.db.QueryRow(ctx, q, args...))
}

// Update applies a partial patch to a user.
func (r *UserRepo) Update(ctx context.Context, id uuid.UUID, patch domain.UserPatch) (*domain.User, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.Name != nil {
		addArg("name", *patch.Name)
	}
	if patch.Email != nil {
		addArg("email", *patch.Email)
	}
	if patch.Role != nil {
		addArg("role", *patch.Role)
	}

	whereClause := fmt.Sprintf(`id=$%d AND deleted_at IS NULL`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, userCols,
	)
	return scanUser(r.db.QueryRow(ctx, query, args...))
}

// Delete soft-deletes a user.
func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE users SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	result, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// List returns users matching the filter along with the total count.
func (r *UserRepo) List(ctx context.Context, f domain.UserFilter) ([]*domain.User, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	where := []string{"deleted_at IS NULL"}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = f.OrgID
	}
	if orgID != uuid.Nil {
		addWhere("org_id", orgID)
	}

	if f.Role != nil {
		addWhere("role", *f.Role)
	}
	if f.Q != "" {
		where = append(where, fmt.Sprintf(
			`(name ILIKE $%d OR email ILIKE $%d)`, i, i+1,
		))
		args = append(args, "%"+f.Q+"%", "%"+f.Q+"%")
		i += 2
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE `+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortCol := "created_at"
	allowedSorts := map[string]bool{
		"created_at": true, "updated_at": true,
		"name": true, "email": true,
	}
	if allowedSorts[f.Sort] {
		sortCol = f.Sort
	}
	order := "DESC"
	if strings.ToUpper(f.Order) == "ASC" {
		order = "ASC"
	}

	rows, err := r.db.Query(ctx,
		fmt.Sprintf(
			`SELECT %s FROM users WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
			userCols, whereClause, sortCol, order, i, i+1,
		),
		append(args, f.Limit, offset)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Role,
			&u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	return users, total, rows.Err()
}
