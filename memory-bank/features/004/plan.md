# 04_article_page — Implementation Plan

Спека: `memory-bank/features/004/rspec.md`

---

## Объём

**Новые файлы:**
- `migrations/006_create_news_comments.sql`
- `internal/comment/model.go`, `errors.go`, `repo.go`, `service.go`
- `internal/comment/service_test.go`, `repo_test.go`
- `internal/news/handler.go`, `handler_test.go`
- `templates/news/article.html`

**Правки существующих файлов:**
- `internal/news/repo.go` — добавить `GetPublishedByID`
- `internal/user/repo.go` — добавить `GetByUsernames`
- `internal/auth/handler.go` — redirect по `?next=` после логина
- `internal/tmpl/renderer.go` — FuncMap `deref` для `*string`
- `cmd/forum/main.go` — регистрация comment repo/service/handler, роуты

---

## Итерация 1: Data layer

**Цель:** миграция + репозитории + интеграционные тесты репозиториев.

| # | Файл | Что делаем |
|---|---|---|
| 1 | `migrations/006_create_news_comments.sql` | CREATE TABLE news_comments + индекс |
| 2 | `internal/news/repo.go` | + `GetPublishedByID(ctx, id) (*News, error)` — WHERE id=$1 AND is_published=true; pgx.ErrNoRows → nil,nil |
| 3 | `internal/user/repo.go` | + `GetByUsernames(ctx, []string) ([]User, error)` — WHERE lower(username)=ANY($1) AND is_active=true; пустой вход → [] без запроса |
| 4 | `internal/comment/model.go` | structs `Comment`, `CommentView` |
| 5 | `internal/comment/errors.go` | ErrEmptyContent, ErrContentTooLong, ErrParentNotFound, ErrNewsNotFound |
| 6 | `internal/comment/repo.go` | `ListByNewsID` (JOIN users, LIMIT 500) + `Create` (транзакция: проверка parent + глубины ≤ 1 + snapshot; маппинг 23503 → ErrParentNotFound) |
| 7 | `internal/news/repo_test.go` | GetPublishedByID: опубликованная, несуществующая, неопубликованная |
| 8 | `internal/comment/repo_test.go` | Create (корневой, reply, wrong news_id, depth>1) + ListByNewsID (пустой, сортировка, snapshot) + GetByUsernames |

**Проверка:**
```bash
docker run --rm -v "$(pwd)":/app -w /app \
  --network lfcru_forum_default \
  -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru_test?sslmode=disable" \
  golang:1.23-alpine \
  go test -tags integration -p 1 ./internal/comment/... ./internal/news/... ./internal/user/...
```
→ коммит → обновить HANDOFF.md → **ждать подтверждения**

---

## Итерация 2: HTTP layer

**Цель:** сервис + хэндлер + шаблон + роуты.

| # | Файл | Что делаем |
|---|---|---|
| 1 | `internal/comment/service.go` | `Service.Create` (trim/validate → repo.Create) + `Service.RenderMentions` (regex → GetByUsernames → html.EscapeString → span) + интерфейс `UserRepo` |
| 2 | `internal/comment/service_test.go` | Create (пустой, пробелы, >10000, валидный) + RenderMentions (существующий/несуществующий mention, email≠mention, XSS, дедупликация) |
| 3 | `internal/news/handler.go` | `ShowArticle` (parse id→404, GetPublishedByID→404, ListByNewsID, RenderMentions, render partial/full) + `CreateComment` (RequireAuth, validate, 201/422/303) |
| 4 | `templates/news/article.html` | article+комментарии+формы, Alpine.js x-data replyTo, HTMX hx-post, HX-Trigger listener, CSS из спеки |
| 5 | `internal/tmpl/renderer.go` | добавить FuncMap `deref func(*string) string` |
| 6 | `internal/auth/handler.go` | после успешного Login: читать `?next=`, санитизировать (только `/`-пути, не `//`), redirect |
| 7 | `cmd/forum/main.go` | comment.NewRepo + comment.NewService + news.NewHandler + роуты GET /news/:id и POST /news/:id/comments |
| 8 | `internal/news/handler_test.go` | ~15 HTTP-сценариев из тест-плана спеки |

**Проверка:**
```bash
docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./internal/comment/...

docker run --rm -v "$(pwd)":/app -w /app \
  --network lfcru_forum_default \
  -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru_test?sslmode=disable" \
  golang:1.23-alpine \
  go test -tags integration -p 1 ./internal/...
```
→ коммит → обновить HANDOFF.md

---

## Ключевые решения

- `ContentHTML` — тип `template.HTML`; безопасность через `html.EscapeString` в `RenderMentions` до вставки `<span>`
- HTMX 422: заголовки `HX-Retarget: #comment-form` + `HX-Reswap: innerHTML`
- Глубина ответов: проверка `nc.parent_id IS NOT NULL` в транзакции → ErrParentNotFound
- `?next=` санитизация: только пути начинающиеся с `/` и не с `//`
- Интеграционные тесты: флаг `-p 1` обязателен
