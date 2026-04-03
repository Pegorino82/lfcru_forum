//go:build integration

package ratelimit_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ─── DB setup ────────────────────────────────────────────────────────────────

var (
	dbOnce     sync.Once
	sharedPool *pgxpool.Pool
	dbSetupErr error
)

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

func truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "TRUNCATE login_attempts CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// Record + Count: запись попытки увеличивает счётчик
func TestRepo_Record_IncreasesCount(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	repo := ratelimit.NewLoginAttemptRepo(pool)

	ip := "10.0.0.1"
	if err := repo.Record(context.Background(), ip); err != nil {
		t.Fatalf("Record: %v", err)
	}

	n, err := repo.Count(context.Background(), ip, 10*time.Minute)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 1 {
		t.Errorf("expected count=1, got %d", n)
	}
}

// Count: старые записи (вне окна) не учитываются
func TestRepo_Count_OldRecordsNotCounted(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	repo := ratelimit.NewLoginAttemptRepo(pool)

	ip := "10.0.0.2"

	// Вставляем старую запись напрямую (за пределами окна 10 минут)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO login_attempts (ip_addr, attempted_at) VALUES ($1, now() - interval '20 minutes')", ip)
	if err != nil {
		t.Fatalf("insert old attempt: %v", err)
	}

	n, err := repo.Count(context.Background(), ip, 10*time.Minute)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("expected old record to be outside window, count=%d", n)
	}
}

// Count: возвращает 0 для незнакомого IP
func TestRepo_Count_UnknownIP_Zero(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	repo := ratelimit.NewLoginAttemptRepo(pool)

	n, err := repo.Count(context.Background(), "1.2.3.4", 10*time.Minute)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

// Count: записи разных IP не влияют друг на друга
func TestRepo_Count_IsolatedByIP(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	repo := ratelimit.NewLoginAttemptRepo(pool)

	ip1, ip2 := "10.0.1.1", "10.0.1.2"
	for i := 0; i < 3; i++ {
		_ = repo.Record(context.Background(), ip1)
	}

	n, err := repo.Count(context.Background(), ip2, 10*time.Minute)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("ip2 should have 0 attempts, got %d", n)
	}
}

// Cleanup: удаляет записи старше 1 часа
func TestRepo_Cleanup_RemovesOldRecords(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	repo := ratelimit.NewLoginAttemptRepo(pool)

	ip := "10.0.0.3"

	// Свежая запись (не должна удаляться)
	_ = repo.Record(context.Background(), ip)

	// Старая запись (должна удаляться)
	_, err := pool.Exec(context.Background(),
		"INSERT INTO login_attempts (ip_addr, attempted_at) VALUES ($1, now() - interval '2 hours')", ip)
	if err != nil {
		t.Fatalf("insert old record: %v", err)
	}

	deleted, err := repo.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted record, got %d", deleted)
	}

	// Свежая запись остаётся
	n, _ := repo.Count(context.Background(), ip, 10*time.Minute)
	if n != 1 {
		t.Errorf("expected 1 fresh record remaining, got %d", n)
	}
}

// Cleanup: ничего не удаляет, если все записи свежие
func TestRepo_Cleanup_NoOldRecords(t *testing.T) {
	pool := testDB(t)
	truncate(t, pool)
	repo := ratelimit.NewLoginAttemptRepo(pool)

	_ = repo.Record(context.Background(), "10.0.0.4")

	deleted, err := repo.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}
