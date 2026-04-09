//go:build integration

package forum_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/forum"
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

func insertUser(t *testing.T, pool *pgxpool.Pool, email, username string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, username, pass_hash) VALUES ($1, $2, '\x68617368') RETURNING id`,
		email, username,
	).Scan(&id)
	if err != nil {
		err2 := pool.QueryRow(context.Background(),
			`SELECT id FROM users WHERE email = $1`, email,
		).Scan(&id)
		if err2 != nil {
			t.Fatalf("insertUser: %v / %v", err, err2)
		}
	}
	return id
}

func insertSection(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO forum_sections (title) VALUES ('Test Section') RETURNING id`,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertSection: %v", err)
	}
	return id
}

func insertTopic(t *testing.T, pool *pgxpool.Pool, sectionID, authorID int64, title string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO forum_topics (section_id, author_id, title) VALUES ($1, $2, $3) RETURNING id`,
		sectionID, authorID, title,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertTopic: %v", err)
	}
	return id
}

func insertPost(t *testing.T, pool *pgxpool.Pool, topicID, authorID int64, content string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO forum_posts (topic_id, author_id, content) VALUES ($1, $2, $3)`,
		topicID, authorID, content,
	)
	if err != nil {
		t.Fatalf("insertPost: %v", err)
	}
}

func cleanForum(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool.Exec(ctx, `DELETE FROM forum_posts`)
	pool.Exec(ctx, `DELETE FROM forum_topics`)
	pool.Exec(ctx, `DELETE FROM forum_sections`)
}

func TestLatestActive_Empty(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestLatestActive_Limit(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest@example.com", "forumtest")
	sectionID := insertSection(t, pool)

	for i := 0; i < 7; i++ {
		topicID := insertTopic(t, pool, sectionID, authorID, "topic")
		insertPost(t, pool, topicID, authorID, "content")
	}

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5, got %d", len(result))
	}
}

func TestLatestActive_ExcludesTopicsWithoutPosts(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest2@example.com", "forumtest2")
	sectionID := insertSection(t, pool)

	// Тема с постом
	topicWithPost := insertTopic(t, pool, sectionID, authorID, "active topic")
	insertPost(t, pool, topicWithPost, authorID, "content")

	// Тема без поста
	insertTopic(t, pool, sectionID, authorID, "empty topic")

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
	if result[0].Title != "active topic" {
		t.Errorf("expected 'active topic', got %q", result[0].Title)
	}
}

func TestLatestActive_SortedDesc(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest3@example.com", "forumtest3")
	sectionID := insertSection(t, pool)

	topic1 := insertTopic(t, pool, sectionID, authorID, "first")
	topic2 := insertTopic(t, pool, sectionID, authorID, "second")

	insertPost(t, pool, topic1, authorID, "post1")
	time.Sleep(10 * time.Millisecond)
	insertPost(t, pool, topic2, authorID, "post2")

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 2 {
		t.Fatalf("expected at least 2 results")
	}
	if result[0].Title != "second" {
		t.Errorf("expected 'second' first (most recent), got %q", result[0].Title)
	}
}

func TestLatestActive_LastPostByName(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "forumtest4@example.com", "forumtest4")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic with name")
	insertPost(t, pool, topicID, authorID, "content")

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if result[0].LastPostByName != "forumtest4" {
		t.Errorf("expected 'forumtest4', got %q", result[0].LastPostByName)
	}
}

func TestLatestActive_DeletedUserShowsPlaceholder(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "tobedeleted@example.com", "tobedeleted")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic orphan")
	insertPost(t, pool, topicID, authorID, "content")

	// Обновим last_post_by на несуществующий id (эмулируем SET NULL)
	_, err := pool.Exec(ctx, `UPDATE forum_topics SET last_post_by = NULL WHERE id = $1`, topicID)
	if err != nil {
		t.Fatalf("update last_post_by: %v", err)
	}

	repo := forum.NewRepo(pool)
	result, err := repo.LatestActive(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range result {
		if r.ID == topicID {
			found = true
			if r.LastPostByName != "[удалён]" {
				t.Errorf("expected '[удалён]', got %q", r.LastPostByName)
			}
		}
	}
	if !found {
		t.Error("topic not found in results")
	}
}

func TestLatestActive_TriggerUpdatesFields(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "triggertest@example.com", "triggertest")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "trigger topic")

	// До поста — last_post_at IS NULL, post_count = 0
	var postCount int
	pool.QueryRow(ctx, `SELECT post_count FROM forum_topics WHERE id = $1`, topicID).Scan(&postCount)
	if postCount != 0 {
		t.Errorf("expected post_count=0 before insert, got %d", postCount)
	}

	insertPost(t, pool, topicID, authorID, "first post")

	// После поста — триггер должен обновить поля
	var lastPostBy *int64
	pool.QueryRow(ctx, `SELECT post_count, last_post_by FROM forum_topics WHERE id = $1`, topicID).Scan(&postCount, &lastPostBy)
	if postCount != 1 {
		t.Errorf("expected post_count=1 after insert, got %d", postCount)
	}
	if lastPostBy == nil || *lastPostBy != authorID {
		t.Errorf("expected last_post_by=%d, got %v", authorID, lastPostBy)
	}
}

// Tests for ListSections

func TestListSections_Empty(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	repo := forum.NewRepo(pool)
	sections, err := repo.ListSections(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sections) != 0 {
		t.Errorf("expected empty slice, got %d", len(sections))
	}
}

func TestListSections_Sorting(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	// Create sections with different sort_order
	var id1, id2, id3 int64
	pool.QueryRow(ctx, `INSERT INTO forum_sections (title, sort_order) VALUES ($1, $2) RETURNING id`,
		"section1", 2).Scan(&id1)
	pool.QueryRow(ctx, `INSERT INTO forum_sections (title, sort_order) VALUES ($1, $2) RETURNING id`,
		"section2", 1).Scan(&id2)
	pool.QueryRow(ctx, `INSERT INTO forum_sections (title, sort_order) VALUES ($1, $2) RETURNING id`,
		"section3", 1).Scan(&id3)

	repo := forum.NewRepo(pool)
	sections, err := repo.ListSections(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	// Should be sorted by sort_order ASC, id ASC
	if sections[0].ID != id2 || sections[1].ID != id3 || sections[2].ID != id1 {
		t.Errorf("incorrect sorting: %v, %v, %v", sections[0].ID, sections[1].ID, sections[2].ID)
	}
}

// Tests for GetSection

func TestGetSection_Exists(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	var id int64
	pool.QueryRow(ctx, `INSERT INTO forum_sections (title, description, sort_order) VALUES ($1, $2, $3) RETURNING id`,
		"test section", "description", 5).Scan(&id)

	repo := forum.NewRepo(pool)
	section, err := repo.GetSection(ctx, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section == nil {
		t.Fatal("expected section, got nil")
	}
	if section.Title != "test section" || section.Description != "description" || section.SortOrder != 5 {
		t.Errorf("incorrect section data: %+v", section)
	}
}

func TestGetSection_NotExists(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	repo := forum.NewRepo(pool)
	section, err := repo.GetSection(context.Background(), 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section != nil {
		t.Errorf("expected nil, got %+v", section)
	}
}

// Tests for ListTopicsBySection

func TestListTopicsBySection_Sorting(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "section-test@example.com", "sectiontest")
	sectionID := insertSection(t, pool)

	topic1 := insertTopic(t, pool, sectionID, authorID, "older")
	topic2 := insertTopic(t, pool, sectionID, authorID, "newer")

	insertPost(t, pool, topic1, authorID, "content1")
	time.Sleep(10 * time.Millisecond)
	insertPost(t, pool, topic2, authorID, "content2")

	repo := forum.NewRepo(pool)
	topics, err := repo.ListTopicsBySection(ctx, sectionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(topics) < 2 {
		t.Fatalf("expected at least 2 topics, got %d", len(topics))
	}
	// topic2 should be first (newer last_post_at)
	if topics[0].ID != topic2 {
		t.Errorf("expected topic2 first, got %d", topics[0].ID)
	}
}

// Tests for GetTopic

func TestGetTopic_Exists(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "gettopic@example.com", "gettopic")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "test topic")

	repo := forum.NewRepo(pool)
	topic, err := repo.GetTopic(ctx, topicID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic == nil {
		t.Fatal("expected topic, got nil")
	}
	if topic.Title != "test topic" || topic.SectionID != sectionID {
		t.Errorf("incorrect topic data: %+v", topic)
	}
}

func TestGetTopic_NotExists(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	repo := forum.NewRepo(pool)
	topic, err := repo.GetTopic(context.Background(), 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic != nil {
		t.Errorf("expected nil, got %+v", topic)
	}
}

// Tests for ListPostsByTopic

func TestListPostsByTopic_Sorting(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "listposts@example.com", "listposts")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic")

	insertPost(t, pool, topicID, authorID, "post1")
	time.Sleep(10 * time.Millisecond)
	insertPost(t, pool, topicID, authorID, "post2")

	repo := forum.NewRepo(pool)
	posts, err := repo.ListPostsByTopic(ctx, topicID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	// Should be sorted by created_at ASC
	if posts[0].Content != "post1" || posts[1].Content != "post2" {
		t.Errorf("incorrect sorting")
	}
}

// Tests for CreateSection

func TestCreateSection_ReturnsID(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	repo := forum.NewRepo(pool)

	s := &forum.Section{Title: "new section", Description: "desc", SortOrder: 10}
	id, err := repo.CreateSection(ctx, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Errorf("expected non-zero id, got %d", id)
	}

	// Verify it exists
	var title string
	pool.QueryRow(ctx, `SELECT title FROM forum_sections WHERE id = $1`, id).Scan(&title)
	if title != "new section" {
		t.Errorf("section not found in db")
	}
}

// Tests for CreateTopic

func TestCreateTopic_ReturnsID(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "createtopic@example.com", "createtopic")
	sectionID := insertSection(t, pool)

	repo := forum.NewRepo(pool)
	topic := &forum.Topic{SectionID: sectionID, AuthorID: authorID, Title: "new topic"}
	id, err := repo.CreateTopic(ctx, topic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Errorf("expected non-zero id")
	}
}

func TestCreateTopic_UpdatesTopicCount(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "topiccount@example.com", "topiccount")
	sectionID := insertSection(t, pool)

	// Check initial topic_count
	var count1 int
	pool.QueryRow(ctx, `SELECT topic_count FROM forum_sections WHERE id = $1`, sectionID).Scan(&count1)
	if count1 != 0 {
		t.Fatalf("expected topic_count=0 initially, got %d", count1)
	}

	repo := forum.NewRepo(pool)
	topic := &forum.Topic{SectionID: sectionID, AuthorID: authorID, Title: "topic"}
	_, err := repo.CreateTopic(ctx, topic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check topic_count after creation
	var count2 int
	pool.QueryRow(ctx, `SELECT topic_count FROM forum_sections WHERE id = $1`, sectionID).Scan(&count2)
	if count2 != 1 {
		t.Errorf("expected topic_count=1, got %d", count2)
	}
}

func TestCreateTopic_SectionNotFound(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "nosection@example.com", "nosection")

	repo := forum.NewRepo(pool)
	topic := &forum.Topic{SectionID: 99999, AuthorID: authorID, Title: "topic"}
	_, err := repo.CreateTopic(ctx, topic)
	if err != forum.ErrSectionNotFound {
		t.Errorf("expected ErrSectionNotFound, got %v", err)
	}
}

// Tests for CreatePost

func TestCreatePost_RootPost(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "rootpost@example.com", "rootpost")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic")

	repo := forum.NewRepo(pool)
	post := &forum.Post{TopicID: topicID, AuthorID: authorID, ParentID: nil, Content: "root content"}
	id, err := repo.CreatePost(ctx, post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Errorf("expected non-zero id")
	}

	// Check post_count was updated
	var count int
	pool.QueryRow(ctx, `SELECT post_count FROM forum_topics WHERE id = $1`, topicID).Scan(&count)
	if count != 1 {
		t.Errorf("expected post_count=1, got %d", count)
	}
}

func TestCreatePost_WithParent(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "withparent@example.com", "withparent")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic")

	insertPost(t, pool, topicID, authorID, "parent content")
	var parentID int64
	pool.QueryRow(ctx, `SELECT id FROM forum_posts WHERE topic_id = $1 LIMIT 1`, topicID).Scan(&parentID)

	repo := forum.NewRepo(pool)
	post := &forum.Post{
		TopicID:  topicID,
		AuthorID: authorID,
		ParentID: &parentID,
		Content:  "reply content",
	}
	id, err := repo.CreatePost(ctx, post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Errorf("expected non-zero id")
	}

	// Check snapshot was filled
	var snapshot *string
	pool.QueryRow(ctx, `SELECT parent_content_snapshot FROM forum_posts WHERE id = $1`, id).Scan(&snapshot)
	if snapshot == nil || *snapshot != "parent content" {
		t.Errorf("expected snapshot 'parent content', got %v", snapshot)
	}
}

func TestCreatePost_SnapshotTruncation(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "truncate@example.com", "truncate")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic")

	// Create long content with valid characters
	longContent := ""
	for i := 0; i < 150; i++ {
		longContent += "x"
	}
	insertPost(t, pool, topicID, authorID, longContent)
	var parentID int64
	pool.QueryRow(ctx, `SELECT id FROM forum_posts WHERE topic_id = $1 LIMIT 1`, topicID).Scan(&parentID)

	repo := forum.NewRepo(pool)
	post := &forum.Post{
		TopicID:  topicID,
		AuthorID: authorID,
		ParentID: &parentID,
		Content:  "reply",
	}
	id, err := repo.CreatePost(ctx, post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var snapshot string
	pool.QueryRow(ctx, `SELECT parent_content_snapshot FROM forum_posts WHERE id = $1`, id).Scan(&snapshot)
	if len([]rune(snapshot)) != 100 {
		t.Errorf("expected snapshot of 100 runes, got %d", len([]rune(snapshot)))
	}
}

func TestCreatePost_ParentNotFound(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "noparent@example.com", "noparent")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic")

	repo := forum.NewRepo(pool)
	fakeParentID := int64(99999)
	post := &forum.Post{
		TopicID:  topicID,
		AuthorID: authorID,
		ParentID: &fakeParentID,
		Content:  "reply",
	}
	_, err := repo.CreatePost(ctx, post)
	if err != forum.ErrParentNotFound {
		t.Errorf("expected ErrParentNotFound, got %v", err)
	}
}

func TestCreatePost_ParentFromDifferentTopic(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "diffopic@example.com", "difftopic")
	sectionID := insertSection(t, pool)
	topic1 := insertTopic(t, pool, sectionID, authorID, "topic1")
	topic2 := insertTopic(t, pool, sectionID, authorID, "topic2")

	insertPost(t, pool, topic1, authorID, "parent")
	var parentID int64
	pool.QueryRow(ctx, `SELECT id FROM forum_posts WHERE topic_id = $1 LIMIT 1`, topic1).Scan(&parentID)

	repo := forum.NewRepo(pool)
	post := &forum.Post{
		TopicID:  topic2, // Different topic!
		AuthorID: authorID,
		ParentID: &parentID,
		Content:  "reply",
	}
	_, err := repo.CreatePost(ctx, post)
	if err != forum.ErrParentNotFound {
		t.Errorf("expected ErrParentNotFound, got %v", err)
	}
}

func TestCreatePost_ReplyToReply(t *testing.T) {
	pool := setupPool(t)
	cleanForum(t, pool)
	defer cleanForum(t, pool)

	ctx := context.Background()
	authorID := insertUser(t, pool, "replytoreply@example.com", "replytoreply")
	sectionID := insertSection(t, pool)
	topicID := insertTopic(t, pool, sectionID, authorID, "topic")

	// Insert root post
	insertPost(t, pool, topicID, authorID, "root")
	var rootID int64
	pool.QueryRow(ctx, `SELECT id FROM forum_posts WHERE topic_id = $1 LIMIT 1`, topicID).Scan(&rootID)

	// Insert reply to root (this should work)
	var replyID int64
	pool.QueryRow(ctx, `INSERT INTO forum_posts (topic_id, author_id, parent_id, content) VALUES ($1, $2, $3, $4) RETURNING id`,
		topicID, authorID, rootID, "reply to root").Scan(&replyID)

	// Try to reply to reply (should fail)
	repo := forum.NewRepo(pool)
	post := &forum.Post{
		TopicID:  topicID,
		AuthorID: authorID,
		ParentID: &replyID,
		Content:  "reply to reply",
	}
	_, err := repo.CreatePost(ctx, post)
	if err != forum.ErrReplyToReply {
		t.Errorf("expected ErrReplyToReply, got %v", err)
	}
}
