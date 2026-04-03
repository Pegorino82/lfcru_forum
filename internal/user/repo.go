package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
	ErrNotFound          = errors.New("user not found")
)

type Repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, u *User) (int64, error) {
	const q = `
		INSERT INTO users (username, email, pass_hash, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, q,
		u.Username, u.Email, u.PassHash, u.Role, u.IsActive,
	).Scan(&id)
	if err != nil {
		return 0, mapUniqueViolation(err)
	}
	return id, nil
}

func (r *Repo) GetByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
		SELECT id, username, email, pass_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)`

	u := &User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PassHash,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*User, error) {
	const q = `
		SELECT id, username, email, pass_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1`

	u := &User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PassHash,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "idx_users_email":
			return ErrDuplicateEmail
		case "idx_users_username":
			return ErrDuplicateUsername
		}
	}
	return fmt.Errorf("create user: %w", err)
}
