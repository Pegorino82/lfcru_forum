//go:build integration

package match_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/match"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsPath = "../../migrations"

func setupPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(pool.Close)

	db, err := goose.OpenDBWithDriver("pgx", url)
	if err != nil {
		t.Fatalf("goose open: %v", err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(db, migrationsPath); err != nil && !strings.Contains(err.Error(), "no migration files found") {
		t.Fatalf("goose up: %v", err)
	}

	return pool
}

func cleanMatches(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `DELETE FROM matches`)
	if err != nil {
		t.Fatalf("cleanMatches: %v", err)
	}
}

func insertMatch(t *testing.T, pool *pgxpool.Pool, opponent string, matchDate time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO matches (opponent, match_date, tournament) VALUES ($1, $2, $3)`,
		opponent, matchDate, "Premier League",
	)
	if err != nil {
		t.Fatalf("insertMatch: %v", err)
	}
}

func TestNextUpcoming_NoMatches(t *testing.T) {
	pool := setupPool(t)
	cleanMatches(t, pool)
	defer cleanMatches(t, pool)

	repo := match.NewRepo(pool)
	result, err := repo.NextUpcoming(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil when no matches")
	}
}

func TestNextUpcoming_ReturnsClosest(t *testing.T) {
	pool := setupPool(t)
	cleanMatches(t, pool)
	defer cleanMatches(t, pool)

	now := time.Now()
	insertMatch(t, pool, "Chelsea", now.Add(72*time.Hour))
	insertMatch(t, pool, "Arsenal", now.Add(24*time.Hour))
	insertMatch(t, pool, "Man City", now.Add(48*time.Hour))

	repo := match.NewRepo(pool)
	result, err := repo.NextUpcoming(context.Background(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected match, got nil")
	}
	if result.Opponent != "Arsenal" {
		t.Errorf("expected Arsenal (closest), got %q", result.Opponent)
	}
}

func TestNextUpcoming_ExcludesPast(t *testing.T) {
	pool := setupPool(t)
	cleanMatches(t, pool)
	defer cleanMatches(t, pool)

	now := time.Now()
	insertMatch(t, pool, "Past Team", now.Add(-24*time.Hour))

	repo := match.NewRepo(pool)
	result, err := repo.NextUpcoming(context.Background(), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for past matches, got %q", result.Opponent)
	}
}

func TestNextUpcoming_Deterministic(t *testing.T) {
	pool := setupPool(t)
	cleanMatches(t, pool)
	defer cleanMatches(t, pool)

	asOf := time.Date(2030, 1, 15, 12, 0, 0, 0, time.UTC)
	insertMatch(t, pool, "Future A", asOf.Add(1*time.Hour))
	insertMatch(t, pool, "Future B", asOf.Add(2*time.Hour))

	repo := match.NewRepo(pool)
	r1, err := repo.NextUpcoming(context.Background(), asOf)
	if err != nil || r1 == nil {
		t.Fatalf("unexpected: err=%v, result=%v", err, r1)
	}
	r2, err := repo.NextUpcoming(context.Background(), asOf)
	if err != nil || r2 == nil {
		t.Fatalf("unexpected: err=%v, result=%v", err, r2)
	}
	if r1.Opponent != r2.Opponent {
		t.Errorf("non-deterministic: %q vs %q", r1.Opponent, r2.Opponent)
	}
	if r1.Opponent != "Future A" {
		t.Errorf("expected Future A, got %q", r1.Opponent)
	}
}
