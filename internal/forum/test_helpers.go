//go:build integration

package forum

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestInsertSection inserts a test section for handler tests.
func TestInsertSection(t *testing.T, pool *pgxpool.Pool, title, description string, sortOrder int) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO forum_sections (title, description, sort_order) VALUES ($1, $2, $3) RETURNING id`,
		title, description, sortOrder,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert section: %v", err)
	}
	return id
}

// TestInsertTopic inserts a test topic for handler tests.
func TestInsertTopic(t *testing.T, pool *pgxpool.Pool, sectionID, authorID int64, title string) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO forum_topics (section_id, author_id, title) VALUES ($1, $2, $3) RETURNING id`,
		sectionID, authorID, title,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert topic: %v", err)
	}
	return id
}
