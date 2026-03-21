package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omnir/crm-api/internal/domain"
)

type totpRepo struct{ db *pgxpool.Pool }

func NewTOTPRepo(db *pgxpool.Pool) *totpRepo {
	return &totpRepo{db: db}
}

func (r *totpRepo) SetSecret(ctx context.Context, userID uuid.UUID, encryptedSecret string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET totp_secret = $1, totp_enabled = false WHERE id = $2`,
		encryptedSecret, userID)
	return err
}

func (r *totpRepo) GetSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	var secret string
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(totp_secret, '') FROM users WHERE id = $1`, userID).Scan(&secret)
	return secret, err
}

func (r *totpRepo) Activate(ctx context.Context, userID uuid.UUID, hashedCodes []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`UPDATE users SET totp_enabled = true WHERE id = $1`, userID); err != nil {
		return err
	}

	// Replace backup codes.
	if _, err := tx.Exec(ctx,
		`DELETE FROM totp_backup_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, h := range hashedCodes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO totp_backup_codes (user_id, code_hash) VALUES ($1, $2)`, userID, h); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *totpRepo) Disable(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`UPDATE users SET totp_secret = NULL, totp_enabled = false WHERE id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM totp_backup_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *totpRepo) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	var enabled bool
	err := r.db.QueryRow(ctx,
		`SELECT totp_enabled FROM users WHERE id = $1`, userID).Scan(&enabled)
	return enabled, err
}

func (r *totpRepo) FindUnusedBackupCode(ctx context.Context, userID uuid.UUID) ([]*domain.TOTPBackupCode, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, code_hash, used_at, created_at
		 FROM totp_backup_codes WHERE user_id = $1 AND used_at IS NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []*domain.TOTPBackupCode
	for rows.Next() {
		var c domain.TOTPBackupCode
		var usedAt *time.Time
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &usedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.UsedAt = usedAt
		codes = append(codes, &c)
	}
	return codes, rows.Err()
}

func (r *totpRepo) MarkBackupCodeUsed(ctx context.Context, codeID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE totp_backup_codes SET used_at = NOW() WHERE id = $1`, codeID)
	return err
}
