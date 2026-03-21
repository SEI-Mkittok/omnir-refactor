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

// CalendarConnectionRepo is the Postgres implementation of repository.CalendarConnectionRepository.
type CalendarConnectionRepo struct {
	db *pgxpool.Pool
}

func NewCalendarConnectionRepo(db *pgxpool.Pool) *CalendarConnectionRepo {
	return &CalendarConnectionRepo{db: db}
}

const calendarCols = `
	id, org_id, user_id, provider,
	access_token, refresh_token, token_expiry,
	sync_cursor, calendar_id,
	created_at, updated_at
`

func scanCalendarConnection(row pgx.Row) (*domain.CalendarConnection, error) {
	var c domain.CalendarConnection
	err := row.Scan(
		&c.ID, &c.OrgID, &c.UserID, &c.Provider,
		&c.AccessToken, &c.RefreshToken, &c.TokenExpiry,
		&c.SyncCursor, &c.CalendarID,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CalendarConnectionRepo) Upsert(ctx context.Context, c *domain.CalendarConnection) (*domain.CalendarConnection, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO calendar_connections
			(id, org_id, user_id, provider,
			 access_token, refresh_token, token_expiry,
			 sync_cursor, calendar_id,
			 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (org_id, user_id, provider) DO UPDATE SET
			access_token  = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expiry  = EXCLUDED.token_expiry,
			calendar_id   = EXCLUDED.calendar_id,
			updated_at    = NOW()
		RETURNING `+calendarCols,
		c.ID, c.OrgID, c.UserID, c.Provider,
		c.AccessToken, c.RefreshToken, c.TokenExpiry,
		c.SyncCursor, c.CalendarID,
		c.CreatedAt, c.UpdatedAt,
	)
	return scanCalendarConnection(row)
}

func (r *CalendarConnectionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CalendarConnection, error) {
	q := `SELECT ` + calendarCols + ` FROM calendar_connections WHERE id=$1`
	args := []any{id}

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		q += ` AND org_id=$2`
		args = append(args, orgID)
	}

	return scanCalendarConnection(r.db.QueryRow(ctx, q, args...))
}

func (r *CalendarConnectionRepo) GetByUserAndProvider(
	ctx context.Context,
	orgID, userID uuid.UUID,
	provider domain.CalendarProvider,
) (*domain.CalendarConnection, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+calendarCols+`
		 FROM calendar_connections
		 WHERE org_id=$1 AND user_id=$2 AND provider=$3`,
		orgID, userID, provider,
	)
	return scanCalendarConnection(row)
}

func (r *CalendarConnectionRepo) List(ctx context.Context, filter domain.CalendarConnectionFilter) ([]*domain.CalendarConnection, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	addWhere := func(expr string, val any) {
		where = append(where, fmt.Sprintf("%s = $%d", expr, i))
		args = append(args, val)
		i++
	}

	orgID, hasCtxOrg := domain.OrgIDFromContext(ctx)
	if !hasCtxOrg {
		orgID = filter.OrgID
	}
	if orgID != uuid.Nil {
		addWhere("org_id", orgID)
	}
	if filter.UserID != nil {
		addWhere("user_id", *filter.UserID)
	}
	if filter.Provider != nil {
		addWhere("provider", *filter.Provider)
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+calendarCols+` FROM calendar_connections WHERE `+strings.Join(where, " AND ")+` ORDER BY created_at ASC`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.CalendarConnection
	for rows.Next() {
		c, err := scanCalendarConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CalendarConnectionRepo) Update(ctx context.Context, id uuid.UUID, patch domain.CalendarConnectionPatch) (*domain.CalendarConnection, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	i := 1

	addArg := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	if patch.AccessToken != nil {
		addArg("access_token", *patch.AccessToken)
	}
	if patch.RefreshToken != nil {
		addArg("refresh_token", *patch.RefreshToken)
	}
	if patch.TokenExpiry != nil {
		addArg("token_expiry", *patch.TokenExpiry)
	}
	if patch.SyncCursor != nil {
		addArg("sync_cursor", *patch.SyncCursor)
	}
	if patch.CalendarID != nil {
		addArg("calendar_id", *patch.CalendarID)
	}

	whereClause := fmt.Sprintf(`id=$%d`, i)
	args = append(args, id)
	i++

	if orgID, ok := domain.OrgIDFromContext(ctx); ok {
		whereClause += fmt.Sprintf(` AND org_id=$%d`, i)
		args = append(args, orgID)
	}

	query := fmt.Sprintf(
		`UPDATE calendar_connections SET %s WHERE %s RETURNING %s`,
		strings.Join(sets, ", "), whereClause, calendarCols,
	)
	return scanCalendarConnection(r.db.QueryRow(ctx, query, args...))
}

func (r *CalendarConnectionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM calendar_connections WHERE id=$1`
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

// ListAllActive returns every connection — called by the sync worker without an org scope.
func (r *CalendarConnectionRepo) ListAllActive(ctx context.Context) ([]*domain.CalendarConnection, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+calendarCols+` FROM calendar_connections ORDER BY org_id, user_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.CalendarConnection
	for rows.Next() {
		c, err := scanCalendarConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
