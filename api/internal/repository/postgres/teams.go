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

// TeamsConnectionRepo is the Postgres implementation of repository.TeamsConnectionRepository.
type TeamsConnectionRepo struct {
	db *pgxpool.Pool
}

func NewTeamsConnectionRepo(db *pgxpool.Pool) *TeamsConnectionRepo {
	return &TeamsConnectionRepo{db: db}
}

const teamsCols = `id, org_id, tenant_id, bot_token, default_channel_id, channel_name, created_at, updated_at`

func scanTeamsConnection(row pgx.Row) (*domain.TeamsConnection, error) {
	var c domain.TeamsConnection
	err := row.Scan(
		&c.ID, &c.OrgID, &c.TenantID, &c.BotToken,
		&c.DefaultChannelID, &c.ChannelName,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *TeamsConnectionRepo) Upsert(ctx context.Context, c *domain.TeamsConnection) (*domain.TeamsConnection, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	row := r.db.QueryRow(ctx, `
		INSERT INTO teams_connections
			(id, org_id, tenant_id, bot_token, default_channel_id, channel_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (org_id) DO UPDATE SET
			tenant_id         = EXCLUDED.tenant_id,
			bot_token         = EXCLUDED.bot_token,
			default_channel_id = EXCLUDED.default_channel_id,
			channel_name      = EXCLUDED.channel_name,
			updated_at        = EXCLUDED.updated_at
		RETURNING `+teamsCols,
		c.ID, c.OrgID, c.TenantID, c.BotToken,
		c.DefaultChannelID, c.ChannelName,
		c.CreatedAt, c.UpdatedAt,
	)
	return scanTeamsConnection(row)
}

func (r *TeamsConnectionRepo) GetByOrgID(ctx context.Context, orgID uuid.UUID) (*domain.TeamsConnection, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+teamsCols+` FROM teams_connections WHERE org_id = $1`, orgID,
	)
	return scanTeamsConnection(row)
}

func (r *TeamsConnectionRepo) Delete(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM teams_connections WHERE org_id = $1`, orgID,
	)
	return err
}
