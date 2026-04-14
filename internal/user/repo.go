package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		SELECT id, username, email, pass_hash, role, is_active, banned_at, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)`

	u := &User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PassHash,
		&u.Role, &u.IsActive, &u.BannedAt, &u.CreatedAt, &u.UpdatedAt,
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
		SELECT id, username, email, pass_hash, role, is_active, banned_at, created_at, updated_at
		FROM users
		WHERE id = $1`

	u := &User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PassHash,
		&u.Role, &u.IsActive, &u.BannedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetByUsernames возвращает активных пользователей по списку username (case-insensitive).
// При пустом входе возвращает пустой слайс без запроса к БД.
func (r *Repo) GetByUsernames(ctx context.Context, usernames []string) ([]User, error) {
	if len(usernames) == 0 {
		return []User{}, nil
	}
	lower := make([]string, len(usernames))
	for i, u := range usernames {
		lower[i] = strings.ToLower(u)
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, username
		FROM users
		WHERE lower(username) = ANY($1)
		AND is_active = true
	`, lower)
	if err != nil {
		return nil, fmt.Errorf("get users by usernames: %w", err)
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	if result == nil {
		result = []User{}
	}
	return result, rows.Err()
}

// ListAll returns all users ordered by created_at DESC.
func (r *Repo) ListAll(ctx context.Context) ([]User, error) {
	const q = `
		SELECT id, username, email, pass_hash, role, is_active, banned_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list all users: %w", err)
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PassHash,
			&u.Role, &u.IsActive, &u.BannedAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	if result == nil {
		result = []User{}
	}
	return result, rows.Err()
}

// BanUser sets banned_at = now() for the given user.
func (r *Repo) BanUser(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `UPDATE users SET banned_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("ban user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UnbanUser clears banned_at for the given user.
func (r *Repo) UnbanUser(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `UPDATE users SET banned_at = NULL WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("unban user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
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
