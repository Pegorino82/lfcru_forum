//go:build integration

package auth_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ─── DB setup ────────────────────────────────────────────────────────────────

var (
	dbOnce     sync.Once
	sharedPool *pgxpool.Pool
	dbSetupErr error
)

// migrationsPath is relative to the package dir (internal/auth/) at test time.
const migrationsPath = "../../migrations"

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	dbOnce.Do(func() {
		pool, err := pgxpool.New(context.Background(), dbURL)
		if err != nil {
			dbSetupErr = err
			return
		}
		if err := pool.Ping(context.Background()); err != nil {
			pool.Close()
			dbSetupErr = err
			return
		}
		if err := runMigrations(dbURL); err != nil {
			pool.Close()
			dbSetupErr = err
			return
		}
		sharedPool = pool
	})
	if dbSetupErr != nil {
		t.Fatalf("db setup: %v", dbSetupErr)
	}
	return sharedPool
}

func runMigrations(dbURL string) error {
	db, err := goose.OpenDBWithDriver("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, migrationsPath)
}

// truncate clears test tables before each test.
func truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"TRUNCATE login_attempts, sessions, users CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// ─── Service factory ─────────────────────────────────────────────────────────

func newSvc(pool *pgxpool.Pool) *auth.Service {
	return auth.NewService(
		user.NewRepo(pool),
		session.NewRepo(pool),
		ratelimit.NewLoginAttemptRepo(pool),
		auth.Config{
			BcryptCost:         bcrypt.MinCost,
			SessionLifetime:    30 * 24 * time.Hour,
			RateLimitWindow:    10 * time.Minute,
			RateLimitMax:       5,
			SessionGracePeriod: 5 * time.Minute,
			MaxSessionsPerUser: 10,
		},
	)
}

func validInput() auth.RegisterInput {
	return auth.RegisterInput{
		Username:        "testuser",
		Email:           "test@example.com",
		Password:        "password123",
		PasswordConfirm: "password123",
		IPAddr:          "127.0.0.1",
		UserAgent:       "go-test",
	}
}

// insertAttempts directly inserts N login_attempts for an IP to simulate rate-limiting.
func insertAttempts(t *testing.T, pool *pgxpool.Pool, ip string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		_, err := pool.Exec(context.Background(),
			"INSERT INTO login_attempts (ip_addr) VALUES ($1)", ip)
		if err != nil {
			t.Fatalf("insertAttempts: %v", err)
		}
	}
}

func countAttempts(t *testing.T, pool *pgxpool.Pool, ip string) int {
	t.Helper()
	var n int
	err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM login_attempts WHERE ip_addr = $1", ip).Scan(&n)
	if err != nil {
		t.Fatalf("countAttempts: %v", err)
	}
	return n
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// 1. Регистрация с валидными данными → user в БД, сессия в БД
func TestIntegration_Register_Success(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	u, sess, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}
	if sess.ID == uuid.Nil {
		t.Fatal("expected non-nil session ID")
	}

	// Проверяем запись в users
	dbUser, err := user.NewRepo(pool).GetByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("user not found in DB: %v", err)
	}
	if dbUser.Username != "testuser" {
		t.Errorf("username: got %q, want %q", dbUser.Username, "testuser")
	}

	// Проверяем запись в sessions
	dbSess, err := session.NewRepo(pool).GetByID(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("session not found in DB: %v", err)
	}
	if dbSess.UserID != u.ID {
		t.Errorf("session.user_id: got %d, want %d", dbSess.UserID, u.ID)
	}
}

// 2. Регистрация с занятым email → ErrDuplicateEmail
func TestIntegration_Register_DuplicateEmail(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	if _, _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatalf("first register: %v", err)
	}

	in2 := validInput()
	in2.Username = "otheruser"
	_, _, err := svc.Register(context.Background(), in2)
	if !errors.Is(err, auth.ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got: %v", err)
	}
}

// 3. Регистрация с занятым username → ErrDuplicateUsername
func TestIntegration_Register_DuplicateUsername(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	if _, _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatalf("first register: %v", err)
	}

	in2 := validInput()
	in2.Email = "other@example.com"
	_, _, err := svc.Register(context.Background(), in2)
	if !errors.Is(err, auth.ErrDuplicateUsername) {
		t.Fatalf("expected ErrDuplicateUsername, got: %v", err)
	}
}

// 4. Регистрация заблокирована по rate-limit (≥5 записей в login_attempts)
func TestIntegration_Register_RateLimited(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	insertAttempts(t, pool, "10.0.0.1", 5)
	svc := newSvc(pool)

	in := validInput()
	in.IPAddr = "10.0.0.1"
	_, _, err := svc.Register(context.Background(), in)
	if !errors.Is(err, auth.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got: %v", err)
	}
}

// 5. Регистрация с невалидными данными → ValidationErrors (без обращения к БД)
func TestIntegration_Register_ValidationError(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	in := validInput()
	in.Password = "short"
	in.PasswordConfirm = "short"
	_, _, err := svc.Register(context.Background(), in)

	var verr auth.ValidationErrors
	if !errors.As(err, &verr) {
		t.Fatalf("expected ValidationErrors, got: %v", err)
	}
	if _, ok := verr["password"]; !ok {
		t.Errorf("expected error for field 'password', got: %v", verr)
	}
}

// 6. Вход с верными credentials → user + сессия в БД
func TestIntegration_Login_Success(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	if _, _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatalf("register: %v", err)
	}

	u, sess, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
		IPAddr:   "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}

	// Сессия должна быть в БД
	dbSess, err := session.NewRepo(pool).GetByID(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("session not found in DB: %v", err)
	}
	if dbSess.UserID != u.ID {
		t.Errorf("session.user_id: got %d, want %d", dbSess.UserID, u.ID)
	}
}

// 7. Вход с неверным паролем → ErrInvalidCredentials + attempt записывается
func TestIntegration_Login_WrongPassword_RecordsAttempt(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	if _, _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatalf("register: %v", err)
	}
	truncate(t, pool) // очищаем — Register мог создать сессию, но нам нужен только user
	// Пересоздаём пользователя после truncate
	if _, _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatalf("re-register: %v", err)
	}
	// Сбрасываем login_attempts (Register не пишет попытки)
	_, _ = pool.Exec(context.Background(), "DELETE FROM login_attempts")

	before := countAttempts(t, pool, "127.0.0.1")
	_, _, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
		IPAddr:   "127.0.0.1",
	})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
	after := countAttempts(t, pool, "127.0.0.1")
	if after != before+1 {
		t.Errorf("expected attempt count %d, got %d", before+1, after)
	}
}

// 8. Вход с несуществующим email → ErrInvalidCredentials + attempt записывается
func TestIntegration_Login_UserNotFound_RecordsAttempt(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	before := countAttempts(t, pool, "192.168.1.1")
	_, _, err := svc.Login(context.Background(), auth.LoginInput{
		Email:    "nobody@example.com",
		Password: "whatever",
		IPAddr:   "192.168.1.1",
	})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
	after := countAttempts(t, pool, "192.168.1.1")
	if after != before+1 {
		t.Errorf("attempt not recorded: before=%d after=%d", before, after)
	}
}

// 9. 6 неудачных попыток входа → 6-я возвращает ErrRateLimited (429)
func TestIntegration_Login_RateLimit_After5Failures(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)
	ip := "10.1.2.3"

	if _, _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatalf("register: %v", err)
	}

	// 5 неудачных попыток
	for i := 0; i < 5; i++ {
		_, _, err := svc.Login(context.Background(), auth.LoginInput{
			Email: "test@example.com", Password: "wrong", IPAddr: ip,
		})
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", i+1, err)
		}
	}

	// 6-я попытка — rate-limit
	_, _, err := svc.Login(context.Background(), auth.LoginInput{
		Email: "test@example.com", Password: "wrong", IPAddr: ip,
	})
	if !errors.Is(err, auth.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited on 6th attempt, got: %v", err)
	}
}

// 10. Успешный вход не записывает попытку в login_attempts
func TestIntegration_Login_Success_NoAttemptRecorded(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	if _, _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, _ = pool.Exec(context.Background(), "DELETE FROM login_attempts")

	before := countAttempts(t, pool, "127.0.0.1")
	_, _, err := svc.Login(context.Background(), auth.LoginInput{
		Email: "test@example.com", Password: "password123", IPAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	after := countAttempts(t, pool, "127.0.0.1")
	if after != before {
		t.Errorf("expected no new attempts on success: before=%d after=%d", before, after)
	}
}

// 11. Выход → сессия удаляется из БД
func TestIntegration_Logout_DeletesSession(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	_, sess, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := svc.Logout(context.Background(), sess.ID); err != nil {
		t.Fatalf("logout: %v", err)
	}

	_, err = session.NewRepo(pool).GetByID(context.Background(), sess.ID)
	if !errors.Is(err, session.ErrNotFound) {
		t.Errorf("expected session to be deleted, got: %v", err)
	}
}

// 12. Повторный Logout (уже удалённая сессия) не возвращает ошибку
func TestIntegration_Logout_AlreadyDeleted_NoError(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	_, sess, _ := svc.Register(context.Background(), validInput())

	_ = svc.Logout(context.Background(), sess.ID)
	if err := svc.Logout(context.Background(), sess.ID); err != nil {
		t.Errorf("expected no error on double logout, got: %v", err)
	}
}

// 13. GetSession с валидной сессией → возвращает пользователя
func TestIntegration_GetSession_Valid(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	u, sess, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	gotUser, gotSess, err := svc.GetSession(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if gotUser.ID != u.ID {
		t.Errorf("user.ID: got %d, want %d", gotUser.ID, u.ID)
	}
	if gotSess.ID != sess.ID {
		t.Errorf("session.ID mismatch")
	}
}

// 14. GetSession с истёкшей сессией → ErrSessionNotFound
func TestIntegration_GetSession_Expired(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	u, _, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Вставляем сессию с expires_at в прошлом напрямую
	expiredSess := &session.Session{
		UserID:    u.ID,
		IPAddr:    "127.0.0.1",
		UserAgent: "test",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	id, err := session.NewRepo(pool).Create(context.Background(), expiredSess)
	if err != nil {
		t.Fatalf("create expired session: %v", err)
	}

	_, _, err = svc.GetSession(context.Background(), id)
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
	}
}

// 15. GetSession с несуществующим ID → ErrSessionNotFound
func TestIntegration_GetSession_NotFound(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	_, _, err := svc.GetSession(context.Background(), uuid.New())
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got: %v", err)
	}
}

// 16. GetSession для свежей сессии не обновляет expires_at (grace period)
func TestIntegration_GetSession_FreshSession_NotTouched(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	_, sess, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Запоминаем expires_at до вызова GetSession
	before, err := session.NewRepo(pool).GetByID(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("get session before: %v", err)
	}

	_, _, err = svc.GetSession(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	after, err := session.NewRepo(pool).GetByID(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("get session after: %v", err)
	}

	// Свежая сессия (expires_at ≈ now+30d) не должна обновляться
	diff := after.ExpiresAt.Sub(before.ExpiresAt)
	if diff > time.Second {
		t.Errorf("expected expires_at unchanged for fresh session, diff=%v", diff)
	}
}

// Bonus: MaxSessions — при достижении лимита старейшая сессия удаляется
func TestIntegration_MaxSessions_OldestDeleted(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)

	svc := auth.NewService(
		user.NewRepo(pool),
		session.NewRepo(pool),
		ratelimit.NewLoginAttemptRepo(pool),
		auth.Config{
			BcryptCost:         bcrypt.MinCost,
			SessionLifetime:    30 * 24 * time.Hour,
			RateLimitWindow:    10 * time.Minute,
			RateLimitMax:       5,
			SessionGracePeriod: 5 * time.Minute,
			MaxSessionsPerUser: 3, // маленький лимит для теста
		},
	)

	_, firstSess, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	firstID := firstSess.ID

	// Два дополнительных входа
	for i := 0; i < 2; i++ {
		_, _, err := svc.Login(context.Background(), auth.LoginInput{
			Email: "test@example.com", Password: "password123", IPAddr: "127.0.0.1",
		})
		if err != nil {
			t.Fatalf("login %d: %v", i+1, err)
		}
	}

	// Теперь лимит достигнут (3 сессии). Следующий вход должен удалить firstSess
	_, _, err = svc.Login(context.Background(), auth.LoginInput{
		Email: "test@example.com", Password: "password123", IPAddr: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("4th login: %v", err)
	}

	_, err = session.NewRepo(pool).GetByID(context.Background(), firstID)
	if !errors.Is(err, session.ErrNotFound) {
		t.Errorf("expected oldest session to be deleted, got: %v", err)
	}
}

// EC-02: Заблокированный пользователь не может войти → ErrUserBanned.
func TestIntegration_Login_BannedUser(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	u, _, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Ban the user directly
	if _, err := pool.Exec(context.Background(),
		`UPDATE users SET banned_at = now() WHERE id = $1`, u.ID,
	); err != nil {
		t.Fatalf("ban user: %v", err)
	}

	_, _, err = svc.Login(context.Background(), auth.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
		IPAddr:   "127.0.0.1",
	})
	if !errors.Is(err, auth.ErrUserBanned) {
		t.Errorf("want ErrUserBanned, got: %v", err)
	}
}

// EC-02b: Заблокированный пользователь с активной сессией — GetSession возвращает ошибку.
func TestIntegration_GetSession_BannedUser(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	svc := newSvc(pool)

	u, sess, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Ban the user after session was created
	if _, err := pool.Exec(context.Background(),
		`UPDATE users SET banned_at = now() WHERE id = $1`, u.ID,
	); err != nil {
		t.Fatalf("ban user: %v", err)
	}

	_, _, err = svc.GetSession(context.Background(), sess.ID)
	if !errors.Is(err, auth.ErrUserBanned) {
		t.Errorf("want ErrUserBanned, got: %v", err)
	}
}
