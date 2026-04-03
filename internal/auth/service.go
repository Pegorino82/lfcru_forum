package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// dummyHash is used for constant-time comparison when user not found.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.MinCost)

type RegisterInput struct {
	Username        string
	Email           string
	Password        string
	PasswordConfirm string
	IPAddr          string
	UserAgent       string
}

type LoginInput struct {
	Email     string
	Password  string
	IPAddr    string
	UserAgent string
}

type UserRepo interface {
	Create(ctx context.Context, u *user.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	GetByID(ctx context.Context, id int64) (*user.User, error)
}

type SessionRepo interface {
	Create(ctx context.Context, s *session.Session) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*session.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Touch(ctx context.Context, id uuid.UUID, newExpiry time.Time) error
	CountByUser(ctx context.Context, userID int64) (int, error)
	DeleteOldestByUser(ctx context.Context, userID int64) error
}

type AttemptRepo interface {
	Record(ctx context.Context, ip string) error
	Count(ctx context.Context, ip string, window time.Duration) (int, error)
}

type Config struct {
	BcryptCost         int
	SessionLifetime    time.Duration
	RateLimitWindow    time.Duration
	RateLimitMax       int
	SessionGracePeriod time.Duration
	MaxSessionsPerUser int
	CookieSecure       bool
}

type Service struct {
	users    UserRepo
	sessions SessionRepo
	attempts AttemptRepo
	cfg      Config
}

func NewService(users UserRepo, sessions SessionRepo, attempts AttemptRepo, cfg Config) *Service {
	return &Service{
		users:    users,
		sessions: sessions,
		attempts: attempts,
		cfg:      cfg,
	}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*user.User, *session.Session, error) {
	// Rate-limit check
	count, err := s.attempts.Count(ctx, in.IPAddr, s.cfg.RateLimitWindow)
	if err != nil {
		return nil, nil, fmt.Errorf("rate limit check: %w", err)
	}
	if count >= s.cfg.RateLimitMax {
		return nil, nil, ErrRateLimited
	}

	// Validation
	if verr := validateRegisterInput(in); verr != nil {
		return nil, nil, verr
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), s.cfg.BcryptCost)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	u := &user.User{
		Username: in.Username,
		Email:    strings.ToLower(strings.TrimSpace(in.Email)),
		PassHash: hash,
		Role:     "user",
		IsActive: true,
	}

	id, err := s.users.Create(ctx, u)
	if err != nil {
		return nil, nil, mapUserRepoErr(err)
	}
	u.ID = id

	sess, err := s.createSession(ctx, u.ID, in.IPAddr, in.UserAgent)
	if err != nil {
		return nil, nil, err
	}

	return u, sess, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*user.User, *session.Session, error) {
	// Rate-limit check
	count, err := s.attempts.Count(ctx, in.IPAddr, s.cfg.RateLimitWindow)
	if err != nil {
		return nil, nil, fmt.Errorf("rate limit check: %w", err)
	}
	if count >= s.cfg.RateLimitMax {
		return nil, nil, ErrRateLimited
	}

	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil {
		// Dummy compare to prevent timing attacks
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(in.Password))
		_ = s.attempts.Record(ctx, in.IPAddr)
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(u.PassHash, []byte(in.Password)); err != nil {
		_ = s.attempts.Record(ctx, in.IPAddr)
		return nil, nil, ErrInvalidCredentials
	}

	sess, err := s.createSession(ctx, u.ID, in.IPAddr, in.UserAgent)
	if err != nil {
		return nil, nil, err
	}

	return u, sess, nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessions.Delete(ctx, sessionID)
}

func (s *Service) GetSession(ctx context.Context, sessionID uuid.UUID) (*user.User, *session.Session, error) {
	sess, err := s.sessions.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, session.ErrNotFound) {
			return nil, nil, ErrSessionNotFound
		}
		return nil, nil, fmt.Errorf("get session: %w", err)
	}

	u, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}

	// Touch with grace period: only update if less than (lifetime - grace) remains
	newExpiry := time.Now().Add(s.cfg.SessionLifetime)
	threshold := newExpiry.Add(-s.cfg.SessionGracePeriod)
	if sess.ExpiresAt.Before(threshold) {
		_ = s.sessions.Touch(ctx, sessionID, newExpiry)
		sess.ExpiresAt = newExpiry
	}

	return u, sess, nil
}

// createSession enforces max-sessions policy and creates a new session.
func (s *Service) createSession(ctx context.Context, userID int64, ip, ua string) (*session.Session, error) {
	n, err := s.sessions.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count sessions: %w", err)
	}
	if n >= s.cfg.MaxSessionsPerUser {
		if err := s.sessions.DeleteOldestByUser(ctx, userID); err != nil {
			return nil, fmt.Errorf("delete oldest session: %w", err)
		}
	}

	sess := &session.Session{
		UserID:    userID,
		IPAddr:    ip,
		UserAgent: ua,
		ExpiresAt: time.Now().Add(s.cfg.SessionLifetime),
	}
	id, err := s.sessions.Create(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	sess.ID = id
	return sess, nil
}

func validateRegisterInput(in RegisterInput) ValidationErrors {
	errs := make(ValidationErrors)

	// username
	if utf8.RuneCountInString(in.Username) < 3 || utf8.RuneCountInString(in.Username) > 30 {
		errs["username"] = "Имя пользователя должно быть от 3 до 30 символов"
	} else if !usernameRe.MatchString(in.Username) {
		errs["username"] = "Имя пользователя может содержать только буквы, цифры, _ и -"
	}

	// email
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if len(email) > 254 || !strings.Contains(email, "@") || strings.Count(email, "@") != 1 {
		errs["email"] = "Введите корректный email"
	}

	// password
	if len(in.Password) < 8 {
		errs["password"] = "Пароль должен содержать не менее 8 символов"
	} else if len(in.Password) > 72 {
		errs["password"] = "Пароль слишком длинный (максимум 72 байта)"
	}

	// password_confirm
	if in.Password != in.PasswordConfirm {
		errs["password_confirm"] = "Пароли не совпадают"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func mapUserRepoErr(err error) error {
	if errors.Is(err, user.ErrDuplicateEmail) {
		return ErrDuplicateEmail
	}
	if errors.Is(err, user.ErrDuplicateUsername) {
		return ErrDuplicateUsername
	}
	return err
}
