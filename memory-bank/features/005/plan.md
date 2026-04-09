# Feature 005 — Forum Structure: Implementation Plan

Spec: `memory-bank/features/005/rspec.md`
GitHub Issue: #5

---

## Оценка объёма

| Категория | Кол-во |
|---|---|
| SQL-миграции | 2 |
| Новых Go-файлов | 5 (`errors.go`, `service.go`, `handler.go`, `service_test.go`, `handler_test.go`) |
| Расширяемых файлов | `model.go`, `repo.go`, `repo_test.go`, `middleware.go`, `main.go`, `base.html`, `home/index.html` |
| Новых шаблонов | 5 forum + 1 errors/403 |
| Юнит-тестов | ~13 |
| Интеграционных тестов (repo) | ~18 |
| Интеграционных тестов (handler) | ~22 |

**Итого: большая фича, разбита на 2 итерации.**

---

## Итерация 1 — Data Layer

**Цель:** реализовать слой данных (миграции, модели, репозиторий, сервис) и покрыть его тестами. HTTP не трогается.

### Шаги

#### 1.1 Миграции

- `migrations/007_forum_sections_description.sql`
  - `ALTER TABLE forum_sections ADD COLUMN description TEXT NOT NULL DEFAULT '', ADD COLUMN topic_count INT NOT NULL DEFAULT 0`
  - UPDATE topic_count для существующих разделов
  - Функция + триггер `trg_forum_topics_update_section_count` (AFTER INSERT OR DELETE ON forum_topics)
  - Индексы: `idx_forum_posts_topic`, `idx_forum_topics_section`, `idx_forum_posts_parent`
  - Down: DROP всего

- `migrations/008_forum_posts_quotes.sql`
  - `ALTER TABLE forum_posts ADD COLUMN parent_id BIGINT REFERENCES forum_posts(id) ON DELETE SET NULL`
  - `ADD COLUMN parent_author_snapshot TEXT, parent_content_snapshot TEXT`
  - Down: DROP колонок

#### 1.2 Модели — `internal/forum/model.go`

Добавить (существующий `TopicWithLastAuthor` — не трогать):

```go
type Section struct { ID, Title, Description, SortOrder, TopicCount, CreatedAt }
type Post struct { ID, TopicID, AuthorID, ParentID*, ParentAuthorSnapshot*, ParentContentSnapshot*, Content, CreatedAt, UpdatedAt }
type SectionView struct { ID, Title, Description, TopicCount }
type TopicView struct { ID, Title, AuthorID, AuthorUsername, PostCount, LastPostAt*, CreatedAt }
type PostView struct { ID, TopicID, AuthorID, AuthorUsername, ParentID*, ParentAuthor*, ParentSnippet*, Content, CreatedAt }
```

Расширить Topic: добавить поле `SectionID int64` если отсутствует.

#### 1.3 Ошибки — `internal/forum/errors.go` (новый файл)

```go
var (
    ErrSectionNotFound, ErrTopicNotFound, ErrPostNotFound, ErrParentNotFound, ErrReplyToReply,
    ErrEmptyTitle, ErrTitleTooLong, ErrDescriptionTooLong, ErrEmptyContent, ErrContentTooLong
)
```

#### 1.4 Репозиторий — `internal/forum/repo.go` (расширить)

Новые методы (SQL из §7 спеки):
- `ListSections(ctx) ([]SectionView, error)` — ORDER BY sort_order, id
- `GetSection(ctx, id int64) (*Section, error)` — nil,nil если не найдено
- `ListTopicsBySection(ctx, sectionID int64) ([]TopicView, error)` — LEFT JOIN users, ORDER BY last_post_at DESC NULLS LAST
- `GetTopic(ctx, id int64) (*Topic, error)` — nil,nil если не найдено
- `ListPostsByTopic(ctx, topicID int64) ([]PostView, error)` — LEFT JOIN users, ORDER BY created_at ASC, LIMIT 500
- `CreateSection(ctx, s *Section) (int64, error)` — RETURNING id
- `CreateTopic(ctx, t *Topic) (int64, error)` — RETURNING id, маппинг 23503 → ErrSectionNotFound
- `CreatePost(ctx, p *Post) (int64, error)` — если ParentID != nil: транзакция (проверка родителя + snapshot + INSERT), маппинг ошибок PG

#### 1.5 Сервис — `internal/forum/service.go` (новый файл)

```go
type Service struct { repo *Repo }
func NewService(repo *Repo) *Service
```

Методы:
- `ListSections(ctx)` — делегирует repo
- `GetSectionWithTopics(ctx, id)` — GetSection (nil → ErrSectionNotFound) + ListTopicsBySection
- `GetTopicWithPosts(ctx, id)` — GetTopic (nil → ErrTopicNotFound) + ListPostsByTopic
- `CreateSection(ctx, title, description string, sortOrder int)` — TrimSpace + валидация + repo.CreateSection
- `CreateTopic(ctx, sectionID, authorID int64, title string)` — TrimSpace + валидация + repo.CreateTopic
- `CreatePost(ctx, topicID, authorID int64, parentID *int64, content string)` — TrimSpace + валидация + repo.CreatePost

#### 1.6 Юнит-тесты — `internal/forum/service_test.go` (новый файл)

Mock-репозиторий (интерфейс или stub). 13 тест-кейсов:

| Тест | Что проверяем |
|---|---|
| CreateSection: пустой title | ErrEmptyTitle |
| CreateSection: только пробелы | ErrEmptyTitle |
| CreateSection: title > 255 | ErrTitleTooLong |
| CreateSection: description > 2000 | ErrDescriptionTooLong |
| CreateSection: валидные данные | вызов repo.CreateSection |
| CreateTopic: пустой title | ErrEmptyTitle |
| CreateTopic: title > 255 | ErrTitleTooLong |
| CreateTopic: валидные данные | вызов repo.CreateTopic |
| CreatePost: пустой content | ErrEmptyContent |
| CreatePost: только пробелы | ErrEmptyContent |
| CreatePost: content > 20000 рун | ErrContentTooLong |
| CreatePost: валидные данные | вызов repo.CreatePost |
| CreatePost: пробрасывает ErrParentNotFound, ErrReplyToReply | без изменений |

#### 1.7 Интеграционные тесты — `internal/forum/repo_test.go` (расширить)

~18 тест-кейсов для новых методов (build tag `integration`):

- ListSections: пустой → []; сортировка по sort_order
- ListSections: topic_count увеличивается триггером после CreateTopic
- GetSection: существующий → *Section; несуществующий → nil, nil
- ListTopicsBySection: сортировка last_post_at DESC NULLS LAST
- GetTopic: существующий → *Topic; несуществующий → nil, nil
- ListPostsByTopic: порядок ASC; пустой → []; snapshot-поля в ответе; после удаления родителя ParentID=nil, snapshot сохранён
- CreateSection: возвращает ID
- CreateTopic: возвращает ID; триггер topic_count; несуществующий section_id → ErrSectionNotFound
- CreatePost корневой: ID, триггер post_count/last_post_at/last_post_by
- CreatePost с parent_id: snapshot заполнен; несуществующий topic_id → ErrTopicNotFound
- CreatePost parent из другой темы → ErrParentNotFound
- CreatePost ответ на ответ → ErrReplyToReply
- CreatePost: content ровно 100 рун → snapshot = весь текст; 101 руна → первые 100

### Критерий завершения итерации 1

- Миграции применяются без ошибок
- Юнит-тесты зелёные: `go test ./internal/forum/...`
- Интеграционные тесты зелёные: `go test -tags integration -p 1 ./internal/forum/...`
- Коммит с тегом `feat(005-iter1)`
- HANDOFF.md обновлён

---

## Итерация 2 — HTTP Layer + Templates

**Цель:** реализовать HTTP-слой, все шаблоны, подключить к роутеру, проверить интеграционными тестами.

### Шаги

#### 2.1 Middleware — `internal/auth/middleware.go`

Добавить `RequireRole(roles ...string) echo.MiddlewareFunc`:
- Читает `u = UserFromContext(c)`
- Если nil → редирект на /login (страховка)
- Если роль не в списке → рендер `templates/errors/403.html` со статусом 403
- Иначе → `next(c)`

#### 2.2 Обработчики — `internal/forum/handler.go` (новый файл)

```go
type Handler struct { svc *Service; renderer echo.Renderer }
func NewHandler(svc *Service, renderer echo.Renderer) *Handler
```

8 методов:

| Метод | Путь | Описание |
|---|---|---|
| `Index` | GET /forum | ListSections → ForumIndexData → render index.html |
| `ShowSection` | GET /forum/sections/:id | GetSectionWithTopics → 404 или SectionPageData → render section.html |
| `ShowTopic` | GET /forum/topics/:id | GetTopicWithPosts → 404 или TopicPageData → render topic.html |
| `NewSection` | GET /forum/sections/new | render new_section.html |
| `CreateSection` | POST /forum/sections | валидация → CreateSection → 303 /forum или 422 с формой |
| `NewTopic` | GET /forum/sections/:id/topics/new | GetSection → 404 или render new_topic.html |
| `CreateTopic` | POST /forum/sections/:id/topics | CreateTopic → 303 /forum/topics/:id или 422 |
| `CreatePost` | POST /forum/topics/:id/posts | CreatePost → HTMX: 201 partial / 422; без HTMX: 303 или 422 полная страница |

Детали CreatePost:
- HTMX success: GetTopicWithPosts → render `#posts-list` фрагмент + `HX-Trigger: postAdded`
- HTMX 422: форма с ошибкой + `HX-Retarget: #post-form`, `HX-Reswap: innerHTML`
- Без HTMX success: 303 → `/forum/topics/{id}#post-{new_id}`
- Без HTMX 422: полная страница темы с FormError + FormContent

Ошибки маппятся:
- `ErrSectionNotFound`, `ErrTopicNotFound` → 404
- `ErrEmptyContent`, `ErrContentTooLong`, `ErrParentNotFound`, `ErrReplyToReply`, `ErrEmptyTitle`, `ErrTitleTooLong`, `ErrDescriptionTooLong` → 422 с сообщением
- Невалидный `:id` (strconv) → 404

Все GET поддерживают `HX-Request: true` → partial (content-блок).

#### 2.3 Шаблоны форума — `templates/forum/`

5 новых файлов:

**`index.html`** — список разделов:
- `{{define "content"}}`
- Хлебные крошки или заголовок «Форум»
- `{{range .Sections}}` → `.section-card` (название, описание, topic_count)
- Empty state «Разделы пока не созданы»
- `{{if .CanManage}}` → кнопка «Создать раздел» → `/forum/sections/new`
- Inline CSS: `.forum-index`, `.section-card`, `.topic-count`

**`section.html`** — список тем в разделе:
- Хлебные крошки: Форум › Название раздела
- `{{range .Topics}}` → `.topic-row` (название, автор, дата последнего поста / «—», post_count)
- Empty state «В разделе пока нет тем»
- `{{if .CanManage}}` → кнопка «Создать тему»
- Inline CSS

**`topic.html`** — тема + посты + форма (по шаблону из §10 спеки):
- Хлебные крошки: Форум › Раздел › Название темы
- `x-data="{ replyTo: null, replyAuthor: '' }"` обёртка
- `#posts-list` с постами + цитатами
- `id="post-{id}"` якоря
- Reply-форма через `<template x-if="replyTo !== null">`
- Основная форма с HTMX
- `htmx.on('#posts-list', 'postAdded', ...)` → `window.dispatchEvent(new Event('reset-reply'))`
- `@reset-reply.window="replyTo = null; replyAuthor = ''"` на обёртке
- `@submit="replyTo = null"` на reply-форме (паттерн из feature 004)
- Для гостей — login-prompt вместо форм, кнопок «Ответить» нет
- Inline CSS по §11 спеки

**`new_section.html`** — форма создания раздела:
- `method="POST" action="/forum/sections"`
- Поля: title, description (textarea), sort_order (number, default 0)
- `_csrf`, `FormError`, сохранённые значения
- Валидация через required + maxlength

**`new_topic.html`** — форма создания темы:
- `method="POST" action="/forum/sections/{{.Section.ID}}/topics"`
- Поле: title
- `_csrf`, `FormError`

#### 2.4 Шаблон ошибки — `templates/errors/403.html`

Брендированная страница в рамках `base.html`:
- `{{define "content"}}`
- Заголовок «403 — Недостаточно прав»
- Текст «У вас нет доступа к этой странице»
- Ссылка «Вернуться на главную»

#### 2.5 Интеграционные тесты — `internal/forum/handler_test.go` (новый файл)

~22 тест-кейса (build tag `integration`). Использовать паттерн из `news/handler_test.go`:
- doGet/doPost хелперы с CSRF (как в handler_test.go фичи 004)
- seed тестовых данных (section, topic, posts, users с разными ролями)

| Тест | Метод/URL | Ожидание |
|---|---|---|
| Список разделов (гость) | GET /forum | 200 |
| Пустой список | GET /forum без данных | 200 + «Разделы пока не созданы» |
| Раздел существует | GET /forum/sections/{id} | 200 |
| Раздел не существует | GET /forum/sections/999 | 404 |
| Невалидный ID | GET /forum/sections/abc | 404 |
| Тема существует | GET /forum/topics/{id} | 200 |
| Тема не существует | GET /forum/topics/999 | 404 |
| Создание поста (user) | POST /forum/topics/{id}/posts | 201 HTMX / 303 без HTMX |
| Создание поста (гость) | POST /forum/topics/{id}/posts | 302 → /login?next=... |
| Пустой content | POST posts (content="") | 422 |
| Ответ на пост | POST с parent_id (корневой) | 201, snapshot заполнен |
| Ответ на ответ | POST с parent_id (ответ) | 422 |
| parent из другой темы | POST с parent_id другой темы | 422 |
| parent несуществующий | POST с parent_id=99999 | 422 |
| HTMX partial | GET /forum + HX-Request | content-блок |
| HTMX 422 ретаргетинг | POST HTMX пустой content | 422 + HX-Retarget |
| Модератор создаёт раздел | POST /forum/sections | 303 → /forum |
| User создаёт раздел | POST /forum/sections (role=user) | 403 |
| Гость на форме раздела | GET /forum/sections/new | 302 → /login |
| Модератор создаёт тему | POST /forum/sections/{id}/topics | 303 → /forum/topics/{new_id} |
| User создаёт тему | POST /forum/sections/{id}/topics (role=user) | 403 |
| Создание темы в несуществующем разделе | POST /forum/sections/999/topics (mod) | 404 |
| Пустой title раздела | POST /forum/sections title="" | 422 |
| CSRF отсутствует | POST без _csrf | 403 |

#### 2.6 Wire up — `cmd/forum/main.go`

1. Создать `forumRepo := forum.NewRepo(pool)` (уже есть для LatestActive)
2. Создать `forumSvc := forum.NewService(forumRepo)`
3. Создать `forumHandler := forum.NewHandler(forumSvc, renderer)`
4. Зарегистрировать роуты (порядок важен — статические до параметрических):

```go
modGroup := e.Group("", auth.RequireAuth, auth.RequireRole("moderator", "admin"))
modGroup.GET("/forum/sections/new", forumHandler.NewSection)
modGroup.POST("/forum/sections", forumHandler.CreateSection)
modGroup.GET("/forum/sections/:id/topics/new", forumHandler.NewTopic)
modGroup.POST("/forum/sections/:id/topics", forumHandler.CreateTopic)

e.GET("/forum", forumHandler.Index)
e.GET("/forum/sections/:id", forumHandler.ShowSection)
e.GET("/forum/topics/:id", forumHandler.ShowTopic)

authGroup := e.Group("", auth.RequireAuth)
authGroup.POST("/forum/topics/:id/posts", forumHandler.CreatePost)
```

#### 2.7 Правки существующих файлов

**`templates/home/index.html`** — блок «Активность форума»:
- Ссылки на темы: `href="/forum/topics/{{.ID}}"` (убрать `href="#"` или несуществующие пути)
- Добавить ссылку «Все разделы» → `/forum`

**`templates/layouts/base.html`** — навигация:
- Добавить пункт «Форум» → `/forum` если ещё нет

### Критерий завершения итерации 2

- Юнит-тесты зелёные: `go test ./internal/...`
- Интеграционные тесты зелёные: `go test -tags integration -p 1 ./internal/...`
- `docker compose -f docker-compose.dev.yml up` — сервер стартует
- `/forum` открывается в браузере
- Коммит с тегом `feat(005-iter2)`
- HANDOFF.md обновлён

---

## Что НЕ входит в scope

(из §17 спеки: редактирование/удаление, пагинация, поиск, SSE, упоминания, markdown, реакции, аватары, вложения, rate-limiting на посты, RSS, SEO)

---

_Plan v1.0 | 2026-04-09_
