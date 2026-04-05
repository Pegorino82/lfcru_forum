//go:build integration

package comment_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Pegorino82/lfcru_forum/internal/comment"
	"github.com/Pegorino82/lfcru_forum/internal/user"
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
		`INSERT INTO users (email, username, pass_hash, is_active) VALUES ($1, $2, $3, true) RETURNING id`,
		email, username, []byte("hash"),
	).Scan(&id)
	if err != nil {
		// Пользователь уже существует — получим его id
		err2 := pool.QueryRow(context.Background(),
			`SELECT id FROM users WHERE email = $1`, email,
		).Scan(&id)
		if err2 != nil {
			t.Fatalf("insertUser: %v", err2)
		}
	}
	return id
}

func insertNews(t *testing.T, pool *pgxpool.Pool, authorID int64) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO news (title, content, is_published, author_id, published_at)
		 VALUES ('test news', 'content', true, $1, now()) RETURNING id`,
		authorID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insertNews: %v", err)
	}
	return id
}

func cleanComments(t *testing.T, pool *pgxpool.Pool, newsID int64) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `DELETE FROM news_comments WHERE news_id = $1`, newsID)
	if err != nil {
		t.Fatalf("cleanComments: %v", err)
	}
}

// --- comment.Repo.Create ---

func TestCreate_Root(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool, "creroot@example.com", "creroot")
	newsID := insertNews(t, pool, authorID)
	defer cleanComments(t, pool, newsID)

	repo := comment.NewRepo(pool)
	id, err := repo.Create(context.Background(), &comment.Comment{
		NewsID:   newsID,
		AuthorID: authorID,
		Content:  "Привет, это тест",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero id")
	}
}

func TestCreate_Reply(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool, "crereply@example.com", "crereply")
	newsID := insertNews(t, pool, authorID)
	defer cleanComments(t, pool, newsID)

	ctx := context.Background()
	repo := comment.NewRepo(pool)

	// Создаём корневой
	parentID, err := repo.Create(ctx, &comment.Comment{
		NewsID:   newsID,
		AuthorID: authorID,
		Content:  "Корневой комментарий с достаточным текстом для snapshot",
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}

	// Ответ на него
	replyID, err := repo.Create(ctx, &comment.Comment{
		NewsID:   newsID,
		AuthorID: authorID,
		ParentID: &parentID,
		Content:  "Ответ на корневой",
	})
	if err != nil {
		t.Fatalf("create reply: %v", err)
	}
	if replyID == 0 {
		t.Error("expected non-zero id")
	}

	// Проверяем snapshot-поля
	comments, err := repo.ListByNewsID(ctx, newsID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var reply *comment.CommentView
	for i := range comments {
		if comments[i].ID == replyID {
			reply = &comments[i]
			break
		}
	}
	if reply == nil {
		t.Fatal("reply not found in list")
	}
	if reply.ParentAuthor == nil || *reply.ParentAuthor != "crereply" {
		t.Errorf("expected ParentAuthor=crereply, got %v", reply.ParentAuthor)
	}
	if reply.ParentSnippet == nil {
		t.Error("expected ParentSnippet to be set")
	}
}

func TestCreate_ParentWrongNews(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool, "crewrong@example.com", "crewrong")
	newsID1 := insertNews(t, pool, authorID)
	newsID2 := insertNews(t, pool, authorID)
	defer cleanComments(t, pool, newsID1)
	defer cleanComments(t, pool, newsID2)

	ctx := context.Background()
	repo := comment.NewRepo(pool)

	// Комментарий к статье 1
	parentID, err := repo.Create(ctx, &comment.Comment{
		NewsID:   newsID1,
		AuthorID: authorID,
		Content:  "Комментарий к статье 1",
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}

	// Пытаемся ответить в контексте статьи 2
	_, err = repo.Create(ctx, &comment.Comment{
		NewsID:   newsID2,
		AuthorID: authorID,
		ParentID: &parentID,
		Content:  "Ответ в неправильной статье",
	})
	if err == nil {
		t.Fatal("expected error for wrong news_id parent")
	}
}

func TestCreate_ReplyToReply_Forbidden(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool, "credepth@example.com", "credepth")
	newsID := insertNews(t, pool, authorID)
	defer cleanComments(t, pool, newsID)

	ctx := context.Background()
	repo := comment.NewRepo(pool)

	// Корневой
	rootID, err := repo.Create(ctx, &comment.Comment{
		NewsID: newsID, AuthorID: authorID, Content: "Корневой",
	})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}

	// Ответ на корневой
	replyID, err := repo.Create(ctx, &comment.Comment{
		NewsID: newsID, AuthorID: authorID, ParentID: &rootID, Content: "Ответ первого уровня",
	})
	if err != nil {
		t.Fatalf("create reply: %v", err)
	}

	// Ответ на ответ — запрещено
	_, err = repo.Create(ctx, &comment.Comment{
		NewsID: newsID, AuthorID: authorID, ParentID: &replyID, Content: "Ответ второго уровня",
	})
	if err == nil {
		t.Fatal("expected ErrParentNotFound for depth > 1")
	}
}

// --- comment.Repo.ListByNewsID ---

func TestListByNewsID_Empty(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool, "listempty@example.com", "listempty")
	newsID := insertNews(t, pool, authorID)
	defer cleanComments(t, pool, newsID)

	repo := comment.NewRepo(pool)
	result, err := repo.ListByNewsID(context.Background(), newsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestListByNewsID_SortedAsc(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool, "listsort@example.com", "listsort")
	newsID := insertNews(t, pool, authorID)
	defer cleanComments(t, pool, newsID)

	ctx := context.Background()
	repo := comment.NewRepo(pool)

	for _, text := range []string{"первый", "второй", "третий"} {
		if _, err := repo.Create(ctx, &comment.Comment{
			NewsID: newsID, AuthorID: authorID, Content: text,
		}); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	comments, err := repo.ListByNewsID(ctx, newsID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(comments) != 3 {
		t.Fatalf("expected 3, got %d", len(comments))
	}
	if comments[0].Content != "первый" {
		t.Errorf("expected first, got %q", comments[0].Content)
	}
	if comments[2].Content != "третий" {
		t.Errorf("expected third, got %q", comments[2].Content)
	}
}

func TestListByNewsID_AuthorUsername(t *testing.T) {
	pool := setupPool(t)
	authorID := insertUser(t, pool, "listauthor@example.com", "listauthor")
	newsID := insertNews(t, pool, authorID)
	defer cleanComments(t, pool, newsID)

	ctx := context.Background()
	repo := comment.NewRepo(pool)
	if _, err := repo.Create(ctx, &comment.Comment{
		NewsID: newsID, AuthorID: authorID, Content: "текст",
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	comments, err := repo.ListByNewsID(ctx, newsID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(comments) == 0 {
		t.Fatal("expected at least one comment")
	}
	if comments[0].AuthorUsername != "listauthor" {
		t.Errorf("expected author listauthor, got %q", comments[0].AuthorUsername)
	}
}

// --- user.Repo.GetByUsernames ---

func TestGetByUsernames_Found(t *testing.T) {
	pool := setupPool(t)
	insertUser(t, pool, "usrnames1@example.com", "usrnames1")
	insertUser(t, pool, "usrnames2@example.com", "usrnames2")

	repo := user.NewRepo(pool)
	users, err := repo.GetByUsernames(context.Background(), []string{"usrnames1", "usrnames2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) < 2 {
		t.Errorf("expected >= 2 users, got %d", len(users))
	}
}

func TestGetByUsernames_Empty(t *testing.T) {
	pool := setupPool(t)

	repo := user.NewRepo(pool)
	users, err := repo.GetByUsernames(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty slice, got %d", len(users))
	}
}

func TestGetByUsernames_CaseInsensitive(t *testing.T) {
	pool := setupPool(t)
	insertUser(t, pool, "usrcase@example.com", "UsrCase")

	repo := user.NewRepo(pool)
	users, err := repo.GetByUsernames(context.Background(), []string{"usrcase"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}
