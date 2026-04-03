package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("session not found")

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, s *Session) (uuid.UUID, error) {
	const q = `
		INSERT INTO sessions (user_id, ip_addr, user_agent, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, s.UserID, s.IPAddr, s.UserAgent, s.ExpiresAt).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create session: %w", err)
	}
	return id, nil
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*Session, error) {
	const q = `
		SELECT id, user_id, ip_addr, user_agent, created_at, expires_at
		FROM sessions
		WHERE id = $1 AND expires_at > now()`

	s := &Session{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.UserID, &s.IPAddr, &s.UserAgent, &s.CreatedAt, &s.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return s, nil
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *Repo) Touch(ctx context.Context, id uuid.UUID, newExpiry time.Time) error {
	const q = `UPDATE sessions SET expires_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, q, newExpiry, id)
	return err
}

func (r *Repo) CountByUser(ctx context.Context, userID int64) (int, error) {
	const q = `SELECT count(*) FROM sessions WHERE user_id = $1 AND expires_at > now()`
	var n int
	err := r.db.QueryRow(ctx, q, userID).Scan(&n)
	return n, err
}

func (r *Repo) DeleteOldestByUser(ctx context.Context, userID int64) error {
	const q = `
		DELETE FROM sessions
		WHERE id = (
			SELECT id FROM sessions
			WHERE user_id = $1
			ORDER BY created_at ASC
			LIMIT 1
		)`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

func (r *Repo) DeleteExpired(ctx context.Context) (int64, error) {
	const q = `DELETE FROM sessions WHERE expires_at < now()`
	tag, err := r.db.Exec(ctx, q)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
