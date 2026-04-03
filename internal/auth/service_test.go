package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- mocks ---

type mockUserRepo struct {
	createFn     func(ctx context.Context, u *user.User) (int64, error)
	getByEmailFn func(ctx context.Context, email string) (*user.User, error)
	getByIDFn    func(ctx context.Context, id int64) (*user.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, u)
	}
	return 1, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, user.ErrNotFound
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &user.User{ID: id}, nil
}

type mockSessionRepo struct {
	createFn            func(ctx context.Context, s *session.Session) (uuid.UUID, error)
	getByIDFn           func(ctx context.Context, id uuid.UUID) (*session.Session, error)
	deleteFn            func(ctx context.Context, id uuid.UUID) error
	touchFn             func(ctx context.Context, id uuid.UUID, newExpiry time.Time) error
	countByUserFn       func(ctx context.Context, userID int64) (int, error)
	deleteOldestByUser  func(ctx context.Context, userID int64) error
}

func (m *mockSessionRepo) Create(ctx context.Context, s *session.Session) (uuid.UUID, error) {
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return uuid.New(), nil
}
func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, session.ErrNotFound
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockSessionRepo) Touch(ctx context.Context, id uuid.UUID, newExpiry time.Time) error {
	if m.touchFn != nil {
		return m.touchFn(ctx, id, newExpiry)
	}
	return nil
}
func (m *mockSessionRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	if m.countByUserFn != nil {
		return m.countByUserFn(ctx, userID)
	}
	return 0, nil
}
func (m *mockSessionRepo) DeleteOldestByUser(ctx context.Context, userID int64) error {
	if m.deleteOldestByUser != nil {
		return m.deleteOldestByUser(ctx, userID)
	}
	return nil
}

type mockAttemptRepo struct {
	recordFn func(ctx context.Context, ip string) error
	countFn  func(ctx context.Context, ip string, window time.Duration) (int, error)
}

func (m *mockAttemptRepo) Record(ctx context.Context, ip string) error {
	if m.recordFn != nil {
		return m.recordFn(ctx, ip)
	}
	return nil
}
func (m *mockAttemptRepo) Count(ctx context.Context, ip string, window time.Duration) (int, error) {
	if m.countFn != nil {
		return m.countFn(ctx, ip, window)
	}
	return 0, nil
}

// --- helpers ---

func defaultCfg() Config {
	return Config{
		BcryptCost:         bcrypt.MinCost,
		SessionLifetime:    720 * time.Hour,
		RateLimitWindow:    10 * time.Minute,
		RateLimitMax:       5,
		SessionGracePeriod: 5 * time.Minute,
		MaxSessionsPerUser: 10,
	}
}

func newSvc(u UserRepo, s SessionRepo, a AttemptRepo) *Service {
	return NewService(u, s, a, defaultCfg())
}

func validRegisterInput() RegisterInput {
	return RegisterInput{
		Username:        "testuser",
		Email:           "test@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
		IPAddr:          "127.0.0.1",
		UserAgent:       "test-agent",
	}
}

// --- tests ---

func TestRegister_Success(t *testing.T) {
	svc := newSvc(&mockUserRepo{}, &mockSessionRepo{}, &mockAttemptRepo{})
	u, sess, err := svc.Register(context.Background(), validRegisterInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u == nil || sess == nil {
		t.Fatal("expected user and session")
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	svc := newSvc(&mockUserRepo{}, &mockSessionRepo{}, &mockAttemptRepo{})

	cases := []struct {
		name  string
		input RegisterInput
		field string
	}{
		{"short username", RegisterInput{Username: "ab", Email: "a@b.com", Password: "password123", PasswordConfirm: "password123"}, "username"},
		{"invalid username chars", RegisterInput{Username: "ab cd", Email: "a@b.com", Password: "password123", PasswordConfirm: "password123"}, "username"},
		{"bad email", RegisterInput{Username: "abc", Email: "notanemail", Password: "password123", PasswordConfirm: "password123"}, "email"},
		{"short password", RegisterInput{Username: "abc", Email: "a@b.com", Password: "short", PasswordConfirm: "short"}, "password"},
		{"password mismatch", RegisterInput{Username: "abc", Email: "a@b.com", Password: "password123", PasswordConfirm: "different"}, "password_confirm"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.input.IPAddr = "127.0.0.1"
			_, _, err := svc.Register(context.Background(), tc.input)
			var verr ValidationErrors
			if !errors.As(err, &verr) {
				t.Fatalf("expected ValidationErrors, got %v", err)
			}
			if _, ok := verr[tc.field]; !ok {
				t.Errorf("expected error for field %q, got %v", tc.field, verr)
			}
		})
	}
}

func TestRegister_RateLimited(t *testing.T) {
	attempts := &mockAttemptRepo{
		countFn: func(_ context.Context, _ string, _ time.Duration) (int, error) {
			return 5, nil
		},
	}
	svc := newSvc(&mockUserRepo{}, &mockSessionRepo{}, attempts)
	_, _, err := svc.Register(context.Background(), validRegisterInput())
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	users := &mockUserRepo{
		createFn: func(_ context.Context, _ *user.User) (int64, error) {
			return 0, user.ErrDuplicateEmail
		},
	}
	svc := newSvc(users, &mockSessionRepo{}, &mockAttemptRepo{})
	_, _, err := svc.Register(context.Background(), validRegisterInput())
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	users := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*user.User, error) {
			return &user.User{ID: 1, PassHash: hash}, nil
		},
	}
	svc := newSvc(users, &mockSessionRepo{}, &mockAttemptRepo{})
	u, sess, err := svc.Login(context.Background(), LoginInput{
		Email: "test@example.com", Password: "password123", IPAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u == nil || sess == nil {
		t.Fatal("expected user and session")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	users := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*user.User, error) {
			return &user.User{ID: 1, PassHash: hash}, nil
		},
	}
	recorded := false
	attempts := &mockAttemptRepo{
		recordFn: func(_ context.Context, _ string) error {
			recorded = true
			return nil
		},
	}
	svc := newSvc(users, &mockSessionRepo{}, attempts)
	_, _, err := svc.Login(context.Background(), LoginInput{
		Email: "test@example.com", Password: "wrong", IPAddr: "127.0.0.1",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if !recorded {
		t.Error("expected login attempt to be recorded")
	}
}

func TestLogin_UserNotFound_TimingAttack(t *testing.T) {
	// User not found — should still record attempt and return ErrInvalidCredentials
	recorded := false
	attempts := &mockAttemptRepo{
		recordFn: func(_ context.Context, _ string) error {
			recorded = true
			return nil
		},
	}
	svc := newSvc(&mockUserRepo{}, &mockSessionRepo{}, attempts)
	_, _, err := svc.Login(context.Background(), LoginInput{
		Email: "nobody@example.com", Password: "whatever", IPAddr: "127.0.0.1",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if !recorded {
		t.Error("expected attempt to be recorded for timing protection")
	}
}

func TestLogin_RateLimited(t *testing.T) {
	attempts := &mockAttemptRepo{
		countFn: func(_ context.Context, _ string, _ time.Duration) (int, error) {
			return 5, nil
		},
	}
	svc := newSvc(&mockUserRepo{}, &mockSessionRepo{}, attempts)
	_, _, err := svc.Login(context.Background(), LoginInput{
		Email: "x@x.com", Password: "p", IPAddr: "127.0.0.1",
	})
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	deleted := uuid.Nil
	sessions := &mockSessionRepo{
		deleteFn: func(_ context.Context, id uuid.UUID) error {
			deleted = id
			return nil
		},
	}
	svc := newSvc(&mockUserRepo{}, sessions, &mockAttemptRepo{})
	id := uuid.New()
	if err := svc.Logout(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if deleted != id {
		t.Errorf("expected session %v to be deleted", id)
	}
}

func TestGetSession_GracePeriod(t *testing.T) {
	id := uuid.New()
	// Session expires in > gracePeriod: no touch expected
	far := time.Now().Add(730 * time.Hour)
	touched := false
	sessions := &mockSessionRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*session.Session, error) {
			return &session.Session{ID: id, UserID: 1, ExpiresAt: far}, nil
		},
		touchFn: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
			touched = true
			return nil
		},
	}
	svc := newSvc(&mockUserRepo{getByIDFn: func(_ context.Context, _ int64) (*user.User, error) {
		return &user.User{ID: 1}, nil
	}}, sessions, &mockAttemptRepo{})

	_, _, err := svc.GetSession(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if touched {
		t.Error("session should not be touched within grace period")
	}
}

func TestGetSession_TouchedOutsideGrace(t *testing.T) {
	id := uuid.New()
	// Session expires soon (within grace): touch expected
	soon := time.Now().Add(1 * time.Minute)
	touched := false
	sessions := &mockSessionRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*session.Session, error) {
			return &session.Session{ID: id, UserID: 1, ExpiresAt: soon}, nil
		},
		touchFn: func(_ context.Context, _ uuid.UUID, _ time.Time) error {
			touched = true
			return nil
		},
	}
	svc := newSvc(&mockUserRepo{getByIDFn: func(_ context.Context, _ int64) (*user.User, error) {
		return &user.User{ID: 1}, nil
	}}, sessions, &mockAttemptRepo{})

	_, _, err := svc.GetSession(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if !touched {
		t.Error("session should be touched when expiry is near")
	}
}

func TestRegister_MaxSessions(t *testing.T) {
	deletedOldest := false
	sessions := &mockSessionRepo{
		countByUserFn: func(_ context.Context, _ int64) (int, error) {
			return 10, nil // at max
		},
		deleteOldestByUser: func(_ context.Context, _ int64) error {
			deletedOldest = true
			return nil
		},
	}
	svc := newSvc(&mockUserRepo{}, sessions, &mockAttemptRepo{})
	_, _, err := svc.Register(context.Background(), validRegisterInput())
	if err != nil {
		t.Fatal(err)
	}
	if !deletedOldest {
		t.Error("expected oldest session to be deleted when at max")
	}
}

// --- validation unit tests ---

func TestValidateUsername(t *testing.T) {
	cases := []struct {
		username string
		valid    bool
	}{
		{"ab", false},                // 2 chars — too short
		{"abc", true},               // 3 chars — ok
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true}, // 30 chars — ok
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false}, // 31 chars — too long
		{"validUser_1", true},
		{"invalid user", false},     // space
		{"invalid@user", false},     // @
	}
	for _, tc := range cases {
		in := RegisterInput{
			Username: tc.username, Email: "a@b.com",
			Password: "password123", PasswordConfirm: "password123",
		}
		verr := validateRegisterInput(in)
		_, hasErr := verr["username"]
		if tc.valid && hasErr {
			t.Errorf("username %q: expected valid, got error: %s", tc.username, verr["username"])
		}
		if !tc.valid && !hasErr {
			t.Errorf("username %q: expected error, got none", tc.username)
		}
	}
}

func TestValidatePassword_Boundaries(t *testing.T) {
	cases := []struct {
		password string
		valid    bool
	}{
		{"1234567", false},                    // 7 bytes — too short
		{"12345678", true},                    // 8 bytes — ok
		{string(make([]byte, 72)), true},      // 72 bytes — ok
		{string(make([]byte, 73)), false},     // 73 bytes — too long
	}
	// Fill with 'a' to avoid null bytes
	for i := range cases {
		if cases[i].password == string(make([]byte, 72)) {
			cases[i].password = string(make([]rune, 72))
			// replace with 'a' * n
		}
	}

	for _, tc := range cases {
		pw := tc.password
		if len(pw) == 0 {
			// Already set above in some cases
			continue
		}
		in := RegisterInput{
			Username: "abc", Email: "a@b.com",
			Password: pw, PasswordConfirm: pw,
		}
		verr := validateRegisterInput(in)
		_, hasErr := verr["password"]
		if tc.valid && hasErr {
			t.Errorf("password len=%d: expected valid, got error: %s", len(pw), verr["password"])
		}
		if !tc.valid && !hasErr {
			t.Errorf("password len=%d: expected error, got none", len(pw))
		}
	}
}
