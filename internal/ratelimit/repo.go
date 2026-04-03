package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LoginAttemptRepo struct {
	db *pgxpool.Pool
}

func NewLoginAttemptRepo(db *pgxpool.Pool) *LoginAttemptRepo {
	return &LoginAttemptRepo{db: db}
}

func (r *LoginAttemptRepo) Record(ctx context.Context, ip string) error {
	const q = `INSERT INTO login_attempts (ip_addr) VALUES ($1)`
	_, err := r.db.Exec(ctx, q, ip)
	if err != nil {
		return fmt.Errorf("record login attempt: %w", err)
	}
	return nil
}

func (r *LoginAttemptRepo) Count(ctx context.Context, ip string, window time.Duration) (int, error) {
	const q = `
		SELECT count(*) FROM login_attempts
		WHERE ip_addr = $1 AND attempted_at > now() - $2::interval`

	var n int
	err := r.db.QueryRow(ctx, q, ip, window.String()).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count login attempts: %w", err)
	}
	return n, nil
}

func (r *LoginAttemptRepo) Cleanup(ctx context.Context) (int64, error) {
	const q = `DELETE FROM login_attempts WHERE attempted_at < now() - interval '1 hour'`
	tag, err := r.db.Exec(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("cleanup login attempts: %w", err)
	}
	return tag.RowsAffected(), nil
}
