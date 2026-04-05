# Implementation Plan — Feature 003: Homepage

**Spec:** `memory-bank/features/003/rspec.md`
**Issue:** #3

---

## Обзор

Наполнить главную страницу тремя блоками данных: последние новости, ближайший матч, последняя активность форума. Текущий хэндлер `internal/home/handler.go` возвращает статичный шаблон — нужно добавить репозитории и данные.

---

## Шаги реализации

### Шаг 1. Миграции

**Файл:** `migrations/004_create_news.sql`

```sql
-- +goose Up
CREATE TABLE news (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title        VARCHAR(255)  NOT NULL,
    content      TEXT          NOT NULL DEFAULT '',
    is_published BOOLEAN       NOT NULL DEFAULT false,
    author_id    BIGINT        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    published_at TIMESTAMPTZ,
    CONSTRAINT chk_news_published CHECK (is_published = false OR published_at IS NOT NULL),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_published ON news (published_at DESC)
    WHERE is_published = true;

-- +goose Down
DROP TABLE IF EXISTS news;
```

**Файл:** `migrations/005_create_forum_and_matches.sql`

```sql
-- +goose Up
CREATE TABLE matches (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    opponent     VARCHAR(255)  NOT NULL,
    match_date   TIMESTAMPTZ   NOT NULL,
    tournament   VARCHAR(255)  NOT NULL,
    is_home      BOOLEAN       NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_matches_date ON matches (match_date ASC);

CREATE TABLE forum_sections (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE forum_topics (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section_id     BIGINT        NOT NULL REFERENCES forum_sections(id) ON DELETE RESTRICT,
    title          VARCHAR(255)  NOT NULL,
    author_id      BIGINT        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    post_count     INT           NOT NULL DEFAULT 0,
    last_post_at   TIMESTAMPTZ,
    last_post_by   BIGINT        REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_topics_last_post ON forum_topics (last_post_at DESC NULLS LAST);

CREATE TABLE forum_posts (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic_id    BIGINT       NOT NULL REFERENCES forum_topics(id) ON DELETE CASCADE,
    author_id   BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content     TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_posts_topic ON forum_posts (topic_id, created_at ASC);

-- Триггер: обновление денормализованных полей forum_topics при INSERT/DELETE в forum_posts
CREATE OR REPLACE FUNCTION forum_topics_update_last_post()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE forum_topics
        SET post_count  = post_count + 1,
            last_post_at = NEW.created_at,
            last_post_by = NEW.author_id,
            updated_at   = now()
        WHERE id = NEW.topic_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE forum_topics
        SET post_count   = GREATEST(post_count - 1, 0),
            last_post_at = (SELECT MAX(created_at) FROM forum_posts WHERE topic_id = OLD.topic_id AND id != OLD.id),
            last_post_by = (SELECT author_id FROM forum_posts WHERE topic_id = OLD.topic_id AND id != OLD.id ORDER BY created_at DESC LIMIT 1),
            updated_at   = now()
        WHERE id = OLD.topic_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_forum_posts_update_topic
AFTER INSERT OR DELETE ON forum_posts
FOR EACH ROW EXECUTE FUNCTION forum_topics_update_last_post();

-- +goose Down
DROP TRIGGER IF EXISTS trg_forum_posts_update_topic ON forum_posts;
DROP FUNCTION IF EXISTS forum_topics_update_last_post;
DROP TABLE IF EXISTS forum_posts;
DROP TABLE IF EXISTS forum_topics;
DROP TABLE IF EXISTS forum_sections;
DROP TABLE IF EXISTS matches;
```

---

### Шаг 2. Пакет `internal/news`

**`internal/news/model.go`**

```go
package news

import "time"

type News struct {
    ID          int64
    Title       string
    Content     string
    IsPublished bool
    AuthorID    int64
    PublishedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**`internal/news/repo.go`**

```go
package news

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// LatestPublished возвращает до limit опубликованных новостей,
// отсортированных по published_at DESC. При отсутствии данных — пустой слайс.
func (r *Repo) LatestPublished(ctx context.Context, limit int) ([]News, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT id, title, published_at
        FROM news
        WHERE is_published = true
        ORDER BY published_at DESC
        LIMIT $1
    `, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []News
    for rows.Next() {
        var n News
        if err := rows.Scan(&n.ID, &n.Title, &n.PublishedAt); err != nil {
            return nil, err
        }
        result = append(result, n)
    }
    if result == nil {
        result = []News{}
    }
    return result, rows.Err()
}
```

---

### Шаг 3. Пакет `internal/match`

**`internal/match/model.go`**

```go
package match

import "time"

type Match struct {
    ID         int64
    Opponent   string
    MatchDate  time.Time
    Tournament string
    IsHome     bool
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**`internal/match/repo.go`**

```go
package match

import (
    "context"
    "time"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// NextUpcoming возвращает ближайший будущий матч (match_date >= asOf)
// или nil, если будущих матчей нет.
func (r *Repo) NextUpcoming(ctx context.Context, asOf time.Time) (*Match, error) {
    var m Match
    err := r.pool.QueryRow(ctx, `
        SELECT id, opponent, match_date, tournament, is_home
        FROM matches
        WHERE match_date >= $1
        ORDER BY match_date ASC
        LIMIT 1
    `, asOf).Scan(&m.ID, &m.Opponent, &m.MatchDate, &m.Tournament, &m.IsHome)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &m, nil
}
```

---

### Шаг 4. Пакет `internal/forum`

**`internal/forum/model.go`**

```go
package forum

import "time"

type Topic struct {
    ID         int64
    SectionID  int64
    Title      string
    AuthorID   int64
    PostCount  int
    LastPostAt *time.Time
    LastPostBy *int64
    CreatedAt  time.Time
}

// TopicWithLastAuthor — проекция для главной страницы.
type TopicWithLastAuthor struct {
    ID             int64
    Title          string
    LastPostAt     time.Time
    LastPostByName string
}
```

**`internal/forum/repo.go`**

```go
package forum

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// LatestActive возвращает до limit тем с последней активностью.
// Темы без сообщений (last_post_at IS NULL) не включаются.
func (r *Repo) LatestActive(ctx context.Context, limit int) ([]TopicWithLastAuthor, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT t.id, t.title, t.last_post_at, COALESCE(u.username, '[удалён]') AS last_post_by_name
        FROM forum_topics t
        LEFT JOIN users u ON u.id = t.last_post_by
        WHERE t.last_post_at IS NOT NULL
        ORDER BY t.last_post_at DESC
        LIMIT $1
    `, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []TopicWithLastAuthor
    for rows.Next() {
        var t TopicWithLastAuthor
        if err := rows.Scan(&t.ID, &t.Title, &t.LastPostAt, &t.LastPostByName); err != nil {
            return nil, err
        }
        result = append(result, t)
    }
    if result == nil {
        result = []TopicWithLastAuthor{}
    }
    return result, rows.Err()
}
```

---

### Шаг 5. Обновление `internal/home/handler.go`

Текущий хэндлер — функция `ShowHome(c echo.Context) error` без зависимостей. Нужно преобразовать в структуру с полями репозиториев.

```go
package home

import (
    "log/slog"
    "net/http"
    "time"

    "github.com/Pegorino82/lfcru_forum/internal/auth"
    "github.com/Pegorino82/lfcru_forum/internal/forum"
    appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
    "github.com/Pegorino82/lfcru_forum/internal/match"
    "github.com/Pegorino82/lfcru_forum/internal/news"
    "github.com/Pegorino82/lfcru_forum/internal/tmpl"
    "github.com/Pegorino82/lfcru_forum/internal/user"
    "github.com/labstack/echo/v4"
)

type HomeData struct {
    User      *user.User
    CSRFToken string
    News      []news.News
    NextMatch *match.Match
    Topics    []forum.TopicWithLastAuthor
}

type Handler struct {
    newsRepo  *news.Repo
    matchRepo *match.Repo
    topicRepo *forum.Repo
}

func NewHandler(newsRepo *news.Repo, matchRepo *match.Repo, topicRepo *forum.Repo) *Handler {
    return &Handler{newsRepo: newsRepo, matchRepo: matchRepo, topicRepo: topicRepo}
}

func (h *Handler) ShowHome(c echo.Context) error {
    ctx := c.Request().Context()

    newsList, err := h.newsRepo.LatestPublished(ctx, 5)
    if err != nil {
        slog.Error("home: failed to load news", "err", err)
        return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
    }

    nextMatch, err := h.matchRepo.NextUpcoming(ctx, time.Now())
    if err != nil {
        slog.Error("home: failed to load match", "err", err)
        return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
    }

    topics, err := h.topicRepo.LatestActive(ctx, 5)
    if err != nil {
        slog.Error("home: failed to load topics", "err", err)
        return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
    }

    data := HomeData{
        User:      auth.UserFromContext(c),
        CSRFToken: appMiddleware.CSRFToken(c),
        News:      newsList,
        NextMatch: nextMatch,
        Topics:    topics,
    }

    if c.Request().Header.Get("HX-Request") == "true" {
        r := c.Echo().Renderer.(*tmpl.Renderer)
        return r.RenderPartial(c.Response(), "templates/home/index.html", "content", data)
    }
    return c.Render(http.StatusOK, "templates/home/index.html", data)
}
```

> **Примечание:** Шаблона `templates/errors/500.html` нет, используется `c.String(500, ...)` напрямую.

---

### Шаг 6. Обновление `templates/home/index.html`

Заменить файл целиком (сохранить внешний `{{define}}`, без него `tmpl.Renderer` не найдёт шаблон):

```html
{{define "templates/home/index.html"}}
{{template "templates/layouts/base.html" .}}
{{end}}

{{define "content"}}
<div class="home-grid">

  <section class="home-news" aria-labelledby="news-heading">
    <h2 id="news-heading">Последние новости</h2>
    {{if .News}}
      <ul class="news-list">
        {{range .News}}
        <li>
          <a href="/news/{{.ID}}">{{.Title}}</a>
          {{if .PublishedAt}}
          <time datetime="{{.PublishedAt.Format "2006-01-02"}}">
            {{.PublishedAt.Format "02.01.2006"}}
          </time>
          {{end}}
        </li>
        {{end}}
      </ul>
    {{else}}
      <p class="empty-state">На сайте еще не добавлены новости</p>
    {{end}}
  </section>

  <section class="home-match" aria-labelledby="match-heading">
    <h2 id="match-heading">Ближайший матч</h2>
    {{if .NextMatch}}
      <div class="match-card">
        <p class="match-opponent">{{.NextMatch.Opponent}}</p>
        <p class="match-date">
          <time datetime="{{.NextMatch.MatchDate.Format "2006-01-02T15:04:05Z07:00"}}">
            {{.NextMatch.MatchDate.Format "02.01.2006 15:04"}} МСК
          </time>
        </p>
        <p class="match-tournament">{{.NextMatch.Tournament}}</p>
      </div>
    {{else}}
      <p class="empty-state">Ближайших матчей нет</p>
    {{end}}
  </section>

  <section class="home-forum" aria-labelledby="forum-heading">
    <h2 id="forum-heading">Последнее на форуме</h2>
    {{if .Topics}}
      <ul class="forum-list">
        {{range .Topics}}
        <li>
          <a href="/forum/topics/{{.ID}}">{{.Title}}</a>
          <span class="forum-meta">
            {{.LastPostByName}} —
            <time datetime="{{.LastPostAt.Format "2006-01-02T15:04:05Z07:00"}}">
              {{.LastPostAt.Format "02.01.2006 15:04"}}
            </time>
          </span>
        </li>
        {{end}}
      </ul>
    {{else}}
      <p class="empty-state">В форуме пока нет активных обсуждений</p>
    {{end}}
  </section>

</div>
{{end}}
```

**CSS** добавить в `base.html` (или отдельным `<style>` в `index.html`):

```css
.home-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; padding: 1rem 0; }
.home-news  { grid-column: 1 / 2; grid-row: 1 / 2; }
.home-match { grid-column: 2 / 3; grid-row: 1 / 2; }
.home-forum { grid-column: 1 / -1; grid-row: 2 / 3; }
.news-list, .forum-list { list-style: none; padding: 0; margin: 0; }
.news-list li, .forum-list li { padding: 0.5rem 0; border-bottom: 1px solid #eee; }
.news-list time, .forum-meta { color: #666; font-size: 0.875rem; }
.match-card { background: #f9f9f9; border: 1px solid #eee; border-radius: 8px; padding: 1rem; text-align: center; }
.match-opponent { font-size: 1.25rem; font-weight: bold; }
.match-tournament { color: #666; font-size: 0.875rem; }
.empty-state { color: #999; font-style: italic; }
section h2 { font-size: 1.25rem; margin-bottom: 0.75rem; border-bottom: 2px solid #c8102e; padding-bottom: 0.5rem; }
@media (max-width: 768px) {
  .home-grid { grid-template-columns: 1fr; }
  .home-news, .home-match, .home-forum { grid-column: 1 / -1; }
}
```

---

### Шаг 7. Обновление `cmd/forum/main.go`

Добавить создание репозиториев и передачу в хэндлер:

```go
// Репозитории для главной страницы
newsRepo  := news.NewRepo(pool)
matchRepo := match.NewRepo(pool)
topicRepo := forum.NewRepo(pool)

// Хэндлер главной страницы
homeHandler := home.NewHandler(newsRepo, matchRepo, topicRepo)

// ...

e.GET("/", homeHandler.ShowHome)  // вместо home.ShowHome
```

Добавить импорты:
```go
"github.com/Pegorino82/lfcru_forum/internal/forum"
"github.com/Pegorino82/lfcru_forum/internal/match"
"github.com/Pegorino82/lfcru_forum/internal/news"
```

---

### Шаг 8. Тесты

#### Интеграционные тесты репозиториев (build tag `integration`)

**`internal/news/repo_test.go`** — проверить:
- пустой слайс при отсутствии данных
- возвращает не более `limit` записей (при 7 published → 5)
- не включает черновики (`is_published = false`)
- сортировка по `published_at DESC`

**`internal/match/repo_test.go`** — проверить:
- `nil` при отсутствии будущих матчей
- возвращает ближайший из трёх будущих
- не возвращает прошедшие
- детерминированность через параметр `asOf`

**`internal/forum/repo_test.go`** — проверить:
- пустой слайс при отсутствии данных
- не более `limit` записей
- не включает темы без сообщений (`last_post_at IS NULL`)
- сортировка по `last_post_at DESC`
- корректный `LastPostByName` (username)
- `[удалён]` при удалённом пользователе (LEFT JOIN)
- триггер обновляет `last_post_at`/`last_post_by`/`post_count` при INSERT в `forum_posts`

#### HTTP-интеграционные тесты

**`internal/home/handler_test.go`** (или отдельный файл) — проверить:
- `GET /` без данных → 200, содержит все три empty-state сообщения
- `GET /` с данными → содержит заголовок новости, имя соперника, название темы
- `GET /` + `HX-Request: true` → ответ не содержит `<html>`, `<head>`
- `GET /` без cookie → 200 (гость имеет доступ)

---

## Порядок выполнения

```
1. Миграции (004, 005)
2. internal/news/model.go + repo.go
3. internal/match/model.go + repo.go
4. internal/forum/model.go + repo.go
5. internal/home/handler.go (рефакторинг в структуру)
6. templates/home/index.html (замена шаблона + CSS)
7. cmd/forum/main.go (инициализация новых репозиториев)
8. Тесты репозиториев (news, match, forum)
9. HTTP-тесты главной страницы
10. Запустить все тесты, убедиться в зелёном статусе
11. Коммит + обновить HANDOFF.md
```

---

## Открытые вопросы

1. **Timezone:** Время матча в шаблоне выводится как `02.01.2006 15:04 МСК` — статичная метка. Для корректной конвертации UTC→Moscow нужно добавить `MatchDate.In(loc).Format(...)` в хэндлере, где `loc = time.LoadLocation("Europe/Moscow")`. Конфигурировать через `DISPLAY_TIMEZONE` (по спеке).
