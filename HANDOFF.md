# HANDOFF.md

## FT-014 — Дублирование ошибок при логине (HTMX outerHTML swap) ✅

**Commit:** fix(FT-014) @ b32cb8c

### Что сделано

1. **`templates/auth/login.html`** — добавлен `id="login-wrapper"` на внешний `<div>`, `hx-target` изменён с `#login-form` на `#login-wrapper`. Root cause: partial-ответ содержал весь content-блок (div + h1 + form), а target указывал на вложенную `<form>` — outerHTML-swap заменял форму на div+form, создавая вложение при каждой ошибке.
2. **`templates/auth/register.html`** — аналогичный фикс (`id="register-wrapper"`, `hx-target="#register-wrapper"`).
3. **`internal/auth/handler_integration_test.go`** — regression-тест `TestLogin_HTMX_InvalidCredentials_NoNestedForm`: два последовательных HTMX POST /login с неверными данными → каждый ответ содержит ровно один `id="login-wrapper"` и один `id="login-form"`.

### Проблемы и решения

- Нет.

### Что сделать следующим

- Нет незакрытых зависимостей.

---

## FT-013 — Форум не отображает залогиненного пользователя в навигации ✅

**Commit:** fix(FT-013) @ 481eb97

### Что сделано

1. **`internal/forum/handler.go`** — во все `data map` добавлен ключ `"User": auth.UserFromContext(c)` (методы `Index`, `ShowSection`, `ShowTopic`, `NewSection`, `NewTopic`, `CreateSection` (ошибка), `CreateTopic` (ошибка), `CreatePost` (ошибка + HTMX-успех)). Причина бага: `base.html` рендерит `{{if .User}}` для отображения имени пользователя в навигации, но форум-хендлер передавал `map[string]interface{}` без этого ключа.
2. **`internal/forum/handler_test.go`** — добавлен regression-тест `TestIndex_AuthUser_ShowsUsername`: `GET /forum` залогиненным пользователем → 200, тело содержит username.

### Проблемы и решения

- Нет.

### Что сделать следующим

- Нет незакрытых зависимостей.

---

## FT-012 — Навигация «Новости» и 500 на /news ✅

**Commit:** fix(FT-012) @ 7fbea0f

### Что сделано

1. **`templates/layouts/base.html`** — ссылка «Новости» в навигации исправлена: `href="#"` → `href="/news"`.
2. **`internal/news/handler.go`** — в структуру `ListData` добавлено поле `CSRFToken string`; заполняется в `ShowList` через `appMiddleware.CSRFToken(c)`. Причина бага: `html/template` выбрасывал ошибку на `{{.CSRFToken}}` в шаблоне `nav` (внутри `{{if .User}}`), когда пользователь залогинен, — поле отсутствовало в типе `news.ListData`.
3. **`internal/news/handler_test.go`** — добавлен regression-тест `TestShowList_AuthUser_OK`: `GET /news` залогиненным пользователем → 200.

### Проблемы и решения

- Нет.

### Что сделать следующим

- Нет незакрытых зависимостей.

---

## FT-008 — Управление статьями (CRUD + превью + review workflow) ✅

**Commit:** feat(FT-008) @ 72882e4

### Что сделано

1. **`migrations/011_article_status_machine.sql`** — `CREATE TYPE news_status ENUM ('draft','in_review','published')`, ADD COLUMN `status`, `reviewer_id`, data migration, DROP `is_published`.
2. **`internal/news/model.go`** — `ArticleStatus` тип + константы, убран `IsPublished`.
3. **`internal/news/repo.go`** — все публичные запросы переведены на `status = 'published'`; новые методы: `CreateDraft`, `UpdateArticle`, `ChangeStatus`, `ListByStatus`, `GetByIDAdmin`.
4. **`internal/news/markdown.go`** — `RenderMarkdown(content) template.HTML` через goldmark.
5. **`internal/news/handler.go`** — `ContentHTML template.HTML` в `ArticleData`, рендер Markdown в `ShowArticle`.
6. **`templates/news/article.html`** — использует `{{.ContentHTML}}` вместо `{{.Article.Content}}`.
7. **`internal/admin/articles_handler.go`** — `ArticlesHandler`: List, New, Create, Edit, Update, Preview, ChangeStatus. Валидация переходов статусов.
8. **`templates/admin/articles/list.html`**, **`edit.html`** — списки с фильтром, форма редактирования с кнопками смены статуса и загрузкой изображений.
9. **`cmd/forum/main.go`** — 7 новых admin-маршрутов для статей.
10. Все тесты `is_published` → `status` обновлены в `news`, `home`, `comment`, `admin` пакетах.
11. **`internal/admin/articles_handler_test.go`** — 7 интеграционных тестов (SC-01..SC-06, EC-01, EC-03, EC-05), все зелёные.
12. goldmark v1.8.2 добавлен в зависимости; создана директория `vendor/`.

### Что сделать следующим

- Нет незакрытых зависимостей. Все фичи admin-панели реализованы (FT-007..FT-011).

### Проблемы и решения

- Миграция data migration требует явного каста: `(CASE WHEN is_published THEN 'published' ELSE 'draft' END)::news_status` — без каста PostgreSQL отклоняет присвоение text к enum.
- Сетевые проблемы при скачивании Go-модулей в test-контейнере → создан `vendor/` через `go mod vendor`, тесты запускаются с флагом `-mod=vendor`.

---

## FT-011 — Управление пользователями (бан/разбан) ✅

**Commit:** feat(FT-011) @ 118aea1

### Что сделано

1. **`migrations/010_add_banned_at_to_users.sql`** — `ALTER TABLE users ADD COLUMN banned_at TIMESTAMPTZ`.
2. **`internal/user/model.go`** — добавлено поле `BannedAt *time.Time`.
3. **`internal/user/repo.go`** — обновлены `GetByEmail`/`GetByID` (сканируют `banned_at`), добавлены `ListAll`, `BanUser`, `UnbanUser`.
4. **`internal/user/service.go`** — новый `Service` с методами `ListAll`, `BanUser` (проверка self-ban), `UnbanUser`.
5. **`internal/auth/errors.go`** — добавлен `ErrUserBanned`.
6. **`internal/auth/service.go`** — `Login` и `GetSession` возвращают `ErrUserBanned` если `banned_at IS NOT NULL`.
7. **`internal/auth/handler.go`** — `Login` обрабатывает `ErrUserBanned` → 403 "Ваш аккаунт заблокирован".
8. **`internal/admin/users_handler.go`** — `UsersHandler`: `List`, `Ban`, `Unban`.
9. **`templates/admin/users/list.html`** — список пользователей с кнопками бан/разбан.
10. **`cmd/forum/main.go`** — зарегистрированы маршруты `/admin/users`, `/admin/users/:id/ban`, `/admin/users/:id/unban`.
11. **`internal/admin/users_handler_test.go`** — 5 интеграционных тестов (SC-01, EC-01, EC-03, EC-04, FM-02), все зелёные.
12. **`internal/auth/integration_test.go`** — добавлены EC-02 (бан при логине) и EC-02b (бан при GetSession), все зелёные.

### Что сделать следующим

1. **FT-008** — Управление статьями (CRUD + превью + review workflow). ADR-006 принят ✅. Все зависимости (FT-009 ✅) закрыты. Готово к реализации.
   - Миграция `011`: enum `news_status`, колонка `status`, `reviewer_id BIGINT`, data migration, drop `is_published`
   - Обновить `internal/news/` (model, repo, тесты) — убрать `is_published`, добавить `status`
   - Обновить тесты в `home`, `admin/images`, `comment` (INSERT-ы с `is_published`)
   - Markdown рендерер: `github.com/yuin/goldmark`
   - Создание статьи → всегда `draft`, публикация — явный статусный переход

### Проблемы и решения

- Нет: реализация прошла без blockers.

---

## FT-010 — Управление структурой форума ✅

**Commit:** feat(FT-010) @ 40863c8

### Что сделано

1. **`internal/forum/repo.go`** — `UpdateSection`, `UpdateTopic`.
2. **`internal/forum/service.go`** — `UpdateSection`, `UpdateTopic`, `GetTopic`, `ListTopicsBySection` + обновлён `RepoInterface`.
3. **`internal/forum/service_test.go`** — добавлены заглушки `UpdateSection`/`UpdateTopic` в `mockRepo`.
4. **`internal/forum/handler.go`** — исправлена обработка `ErrSectionNotFound`/`ErrTopicNotFound` → 404 (pre-existing bug).
5. **`internal/admin/forum_handler.go`** — `ForumHandler` с методами: `ListSections`, `NewSection`, `CreateSection`, `EditSection`, `UpdateSection`, `ListTopics`, `NewTopic`, `CreateTopic`, `EditTopic`, `UpdateTopic`.
6. **Шаблоны** `templates/admin/forum/` — `sections_list.html`, `section_edit.html`, `topics_list.html`, `topic_edit.html`.
7. **`cmd/forum/main.go`** — зарегистрированы 10 admin-маршрутов для форума.
8. **`internal/admin/handler_test.go`** — `cleanAdminData` расширена: удаляет `article_images` и `news` перед удалением пользователей (fix FK-проблемы при повторных запусках).
9. **`internal/admin/forum_handler_test.go`** — 8 интеграционных тестов (SC-01..SC-06, EC-04, FM-02), все зелёные.

### Что сделать следующим

1. **FT-011** — Управление пользователями (бан/разбан, независима).
2. **FT-008** — Управление статьями (требует принятия ADR-006 + зависит от FT-009).

### Проблемы и решения

- `GetSectionWithTopics`/`GetTopicWithPosts` возвращают sentinel-ошибку при отсутствии записи → public forum handler не распознавал 404 → исправлено проверкой `errors.Is` перед 500.
- `cleanAdminData` при повторных запусках не мог удалить пользователей из-за FK `news.author_id ON DELETE RESTRICT` → добавлена очистка `article_images` и `news` перед удалением пользователей.

---

## FT-009 — Загрузка и нормализация изображений ✅

**Commit:** feat(FT-009) @ d477571

### Что сделано

1. **`migrations/009_create_article_images.sql`** — таблица `article_images` (article_id FK → news, filename, original_filename).
2. **`internal/admin/images_repo.go`** — `ImagesRepo`: Create, ListByArticleID, GetByID, Delete.
3. **`internal/admin/image_service.go`** — `ImageService`: resize до 1200px (CatmullRom из `golang.org/x/image/draw`), encode JPEG (stdlib), хранение в `UPLOADS_DIR/{article_id}/{uuid}.jpg`. Поддержка входных форматов: JPEG, PNG, WebP.
4. **`internal/admin/images_handler.go`** — `POST /admin/articles/:id/images` (multipart), `DELETE /admin/articles/:id/images/:image_id`. Возвращает partial `image-item` для HTMX.
5. **`internal/news/model.go`** — добавлен `ImageView`.
6. **`internal/news/repo.go`** — `ListImagesByArticleID`.
7. **`internal/news/handler.go`** — `ShowArticle` загружает изображения и передаёт в `ArticleData.Images`.
8. **`templates/news/article.html`** — отображение изображений: `aspect-ratio: 16/9; object-fit: cover`.
9. **`templates/admin/articles/image_item.html`** — partial для HTMX-ответа после upload.
10. **`internal/config/config.go`** — поле `UploadsDir` (env `UPLOADS_DIR`, default `./uploads`).
11. **`docker-compose.dev.yml`** — volume `uploads_data:/app/uploads`, env `UPLOADS_DIR`.
12. **`internal/middleware/csrf.go`** — `TokenLookup` расширен: `"form:_csrf,header:X-CSRF-Token"` для поддержки HTMX DELETE.
13. **`go.mod`** — добавлена зависимость `golang.org/x/image v0.22.0`.
14. **Тесты** — 11 unit + 4 integration, все зелёные.

### Что сделать следующим

1. **FT-008** — Управление статьями (CRUD + превью + review workflow). Зависит от ADR-006 (принять enum статус) + использует `ImagesRepo`/`ImageService` из FT-009.
2. **FT-010** — Управление структурой форума (независимо).
3. **FT-011** — Управление пользователями (независимо).

### Проблемы и решения

- OQ-02 закрыт: выбран JPEG-вывод (stdlib), без CGO. WebP decode поддерживается через `golang.org/x/image/webp`.
- OQ-01 закрыт: max_width = 1200px.
- CSRF для HTMX DELETE: `TokenLookup` теперь включает `header:X-CSRF-Token`. Шаблон image_item.html передаёт токен через `hx-headers`.

---

## FT-007 — Admin-панель инфраструктура ✅

**Commit:** feat(FT-007) @ 00b2d1c

### Что сделано

1. **`internal/admin/middleware.go`** — `RequireAdminOrMod`: гость/неактивный → redirect `/login`; role != admin/moderator → 403.
2. **`internal/admin/handler.go`** — `Handler.Dashboard`: рендерит `templates/admin/dashboard.html`.
3. **`templates/admin/layouts/layout.html`** — admin layout: основная навигация (Статьи, Форум, Пользователи) + `{{block "admin-content"}}`.
4. **`templates/admin/dashboard.html`** — дашборд-заглушка.
5. **`cmd/forum/main.go`** — `adminGroup` с `RequireAdminOrMod`, маршрут `GET /admin`.
6. **`internal/admin/handler_test.go`** — 4 интеграционных теста SC-01..SC-04, все зелёные.

### Что сделать следующим

1. **Принять ADR-006** — подтвердить замену `is_published → status enum` для FT-008.
2. **FT-009** — загрузка и нормализация изображений (независима от ADR-006).
3. **FT-010** — управление структурой форума (независима от ADR-006).
4. **FT-011** — управление пользователями (независима от ADR-006).
5. **FT-008** — управление статьями (зависит от ADR-006 + FT-009).

### Проблемы и решения

- OQ-02 закрыт: поле `role` в таблице `users` существует (миграция 001).
- Layout-файл размещён в `templates/admin/layouts/` (а не `templates/admin/`) — так renderer подхватывает его как shared layout, доступный в template set dashboard.html.

---

## Инициатива: Admin-панель — документация создана ✅

### Что сделано

Создана вся upstream-документация для инициативы «Админ-панель»:

**Use Cases:**
- `UC-001` — Публикация статьи (draft → in_review → published)
- `UC-002` — Управление структурой форума
- `UC-003` — Управление пользователями (бан/разбан)

**ADR:**
- `ADR-005` — Хранение изображений: файловая система + Docker volume (`accepted`)
- `ADR-006` — Статусная машина статьи: enum `news_status` (`proposed` — требует approval перед реализацией FT-008)

**Feature Packages (Design Ready):**
- `FT-007` — Admin-панель инфраструктура (RBAC middleware, routing, layout)
- `FT-008` — Управление статьями (CRUD + превью + review workflow) — **блокирован ADR-006**
- `FT-009` — Загрузка и нормализация изображений
- `FT-010` — Управление структурой форума (разделы, темы)
- `FT-011` — Управление пользователями (бан/разбан)

### Что сделать следующим

1. **Принять ADR-006** — человек должен подтвердить замену `is_published → status enum` перед реализацией FT-008.
2. **Начать с FT-007** — инфраструктура, не блокирована ничем. Промпт: «Реализуй FT-007 (Admin-панель инфраструктура), код не написан».
3. Затем FT-009, FT-010, FT-011 (независимы от ADR-006).
4. FT-008 — последним (зависит от ADR-006 и FT-009).

### Проблемы и решения

- OQ в FT-007: поле `role` в таблице `users` — проверить наличие перед реализацией.
- OQ в FT-009: WebP encode в Go требует внешней библиотеки (CGO). Если нежелательно — хранить JPEG.
- OQ в FT-008: формат текста статьи (plain/Markdown) — уточнить у владельца продукта.

---

## FT-006 — News: список статей ✅

**Commit:** feat(FT-006) @ 29ce5d2

### Что сделано

1. **`internal/news/repo.go`** — добавлен `ListPublished(ctx, limit, offset int) ([]News, int64, error)`:
   - `COUNT(*)` + `SELECT … LIMIT/OFFSET` (два запроса)
   - Сортировка `published_at DESC`, только опубликованные

2. **`internal/news/handler.go`** — добавлен `ShowList`:
   - `GET /news?page=N` (default 1, невалидный → 1)
   - `const pageSize = 20` хардкод
   - HTMX-совместимость: при `HX-Request: true` → partial `content`-блок
   - Зарегистрирован в `RegisterRoutes`

3. **`templates/news/list.html`** — новый шаблон:
   - Список заголовков-ссылок с датой
   - Компактная пагинация (первая/последняя страница, ±2 от текущей, `…` для gaps)

4. **`internal/tmpl/renderer.go`** — добавлены FuncMap: `add`, `sub`, `paginate`

5. **Тесты**: 4 `TestListPublished_*` (repo) + 4 `TestShowList_*` (handler) — все 32 теста зелёные

### Что сделать следующим

- Создание/редактирование новостей (NS-01 в FT-006) — отдельная задача
- Ссылка «Новости» в навигации (`base.html`) ведёт на `/news` — можно добавить

### Проблемы и решения

- Docker Compose был остановлен → `docker compose -f docker-compose.dev.yml up -d postgres` перед тестами

---

> Файл передачи контекста между агентами. Обновляется в конце каждого сеанса.

---

## Итерация 1 (005) — Data Layer завершена ✅

### Что сделано (005-iter1)

**Commit:** feat(005-iter1)

1. **Миграции 007 и 008** — добавлены колонки и триггеры:
   - `forum_sections`: description, topic_count + триггер для счётчика тем
   - `forum_posts`: parent_id, parent_author_snapshot, parent_content_snapshot

2. **Модели и ошибки** (internal/forum/):
   - Section, Post, SectionView, TopicView, PostView
   - 10 sentinel-ошибок (ErrSectionNotFound, ErrParentNotFound, ErrReplyToReply, валидационные)

3. **Репозиторий** — 8 новых методов:
   - ListSections, GetSection, ListTopicsBySection, GetTopic, ListPostsByTopic
   - CreateSection, CreateTopic, CreatePost (с транзакцией, snapshot, проверка depth ≤ 1)
   - Маппинг PG-ошибок (23503 FK → ErrSectionNotFound/ErrParentNotFound)

4. **Сервис** — валидация + делегирование репо:
   - Использует интерфейс RepoInterface для mockability
   - CreateSection/CreateTopic/CreatePost с валидацией строк и рун

5. **Тесты**:
   - Юнит-тесты (service_test.go): 13 тест-кейсов с mock-репо
   - Интеграционные (repo_test.go): 22 тест-кейса (ListSections, GetSection, ListTopicsBySection, CreateTopic, CreatePost с ошибками)
   - **Все зелёные** ✅

---

## Что сделано в предыдущих итерациях (004)

- **Шаг 0**: Скаффолдинг — `go.mod`, `cmd/forum/main.go`, `internal/config/config.go`, `Dockerfile`, `docker-compose.dev.yml`, `.env.example`
- **Шаг 1**: Миграции — `users`, `sessions`, `login_attempts` через goose
- **Шаг 2**: Репозитории — `user.Repo`, `session.Repo`, `ratelimit.LoginAttemptRepo` (pgx)
- **Шаг 3**: Сервис аутентификации — `auth.Service` с Register/Login/Logout/GetSession + юнит-тесты
- **Шаг 4**: CSRF middleware (echo built-in, `_csrf` form field)
- **Шаг 5**: Session middleware — `LoadSession`, `RequireAuth`, `UserFromContext`
- **Шаг 6**: HTTP handlers — `/register`, `/login`, `/logout` с HTMX-совместимыми ответами
- **Шаг 7**: Шаблоны — `base.html`, `register.html`, `login.html`
- **Шаг 8**: Фоновые задачи — `cleanup.Run` (goroutine с ticker)
- Всё подключено в `main.go`, сервер стартует и отдаёт страницы
- **Шаг 9**: Интеграционные тесты — `internal/auth/integration_test.go` (17 сценариев), `internal/ratelimit/repo_test.go` (6 сценариев); build tag `integration`; запуск: `DATABASE_URL=... go test -tags integration ./internal/...`
- **Фича 002 (базовый layout)**: реализована полностью
  - `internal/tmpl/renderer.go`: `truncate` FuncMap + `RenderPartial` метод
  - `templates/layouts/base.html`: `<header>`, `<footer>`, skip-link, `<main id="content">`, `.main-nav`, `aria-label` на nav, `.nav-username` с truncate
  - `internal/home/handler.go` + `templates/home/index.html`: `GET /` home handler с HTMX partial поддержкой
  - `internal/auth/handler.go`: `ShowLogin` и `ShowRegister` теперь рендерят partial при `HX-Request: true`
  - `internal/layout/layout_test.go`: 12 интеграционных тестов (build tag `integration`), все зелёные
- **Фича 003 (главная страница)**:
  - Миграции `004_create_news.sql` и `005_create_forum_and_matches.sql` (matches, forum_sections, forum_topics, forum_posts + триггер)
  - `internal/news/model.go` + `repo.go` — `LatestPublished(ctx, limit)`
  - `internal/match/model.go` + `repo.go` — `NextUpcoming(ctx, asOf)`
  - `internal/forum/model.go` + `repo.go` — `LatestActive(ctx, limit)`
  - `internal/home/handler.go` рефакторен в структуру `Handler` с тремя репозиториями
  - `templates/home/index.html` — три секции (новости, матч, форум) с empty-state + inline CSS
  - `cmd/forum/main.go` обновлён: инициализация новых репозиториев и `home.NewHandler`
  - Интеграционные тесты: `internal/news/repo_test.go` (4), `internal/match/repo_test.go` (4), `internal/forum/repo_test.go` (7), `internal/home/handler_test.go` (4) — все зелёные

- **Фича 004 (страница статьи + комментарии) — Итерация 1 (data layer):** завершена, коммит `b9afb32`
  - `migrations/006_create_news_comments.sql` — таблица `news_comments` с snapshot-колонками и self-ref FK
  - `internal/news/repo.go` — добавлен `GetPublishedByID` (nil,nil для 404/draft)
  - `internal/user/repo.go` — добавлен `GetByUsernames` (case-insensitive, batch)
  - `internal/comment/` — пакет: `model.go`, `errors.go`, `repo.go` (ListByNewsID + Create с транзакцией depth≤1)
  - Интеграционные тесты: comment/repo (11), news/GetPublishedByID (3) — зелёные

- **Фича 004 (страница статьи + комментарии) — Итерация 2 (HTTP layer):** завершена, коммит `5c593aa`
  - `internal/comment/model.go` — `ContentHTML` изменён на `template.HTML`
  - `internal/comment/service.go` — `Service.Create` (trim/validate) + `Service.RenderMentions` (@mention → span, XSS-safe)
  - `internal/comment/service_test.go` — 5 юнит-тестов Create + 5 юнит-тестов RenderMentions
  - `internal/tmpl/renderer.go` — добавлен FuncMap `deref func(*string) string`
  - `internal/news/handler.go` — `ShowArticle` + `CreateComment` (HTMX + non-HTMX, гость → /login)
  - `internal/news/handler_test.go` — 15 интеграционных тестов (все зелёные)
  - `templates/news/article.html` — шаблон статьи + комментарии + reply-форма (Alpine.js + HTMX)
  - `cmd/forum/main.go` — comment repo/service, news handler, роуты `GET /news/:id` и `POST /news/:id/comments`

## Итерация 2 (005) — HTTP Layer + Templates завершена ✅

**Commit:** feat(005-iter2) @ 020019e

### Что сделано:

1. **Middleware** (`internal/auth/middleware.go`):
   - ✅ RequireRole(renderer, roles...string) — проверка роли, 403 при недостаточных правах

2. **Сервис** (`internal/forum/service.go`):
   - ✅ GetSection(ctx, id) — публичный метод для получения раздела (требуется в handler)

3. **Обработчики** (`internal/forum/handler.go`):
   - ✅ 8 методов реализованы: Index, ShowSection, ShowTopic, NewSection, CreateSection, NewTopic, CreateTopic, CreatePost
   - ✅ Маппинг ошибок на HTTP-коды (404, 422, 403)
   - ✅ HTMX поддержка: CreatePost возвращает partial #posts-list (201) или 422

4. **Шаблоны** (`templates/forum/`):
   - ✅ index.html — список разделов с inline CSS
   - ✅ section.html — темы в разделе
   - ✅ topic.html — тема с постами, reply-форма на Alpine.js + HTMX, якоря #post-{id}
   - ✅ new_section.html — форма создания раздела
   - ✅ new_topic.html — форма создания темы
   - ✅ templates/errors/403.html — страница ошибки прав (Forbidden)

5. **Wire up** (`cmd/forum/main.go`):
   - ✅ Создан forumSvc и forumHandler
   - ✅ Роуты зарегистрированы (modGroup для mod/admin, authGroup для auth, public routes)

6. **Правки** существующих файлов:
   - ✅ templates/home/index.html — добавлена ссылка "Все разделы" → /forum
   - ✅ templates/layouts/base.html — обновлена ссылка в навигации /forum

7. **Интеграционные тесты** (`internal/forum/handler_test.go`):
   - 🔄 Создан файл с 11+ тест-кейсами (в процессе отладки)

### Оставшееся:

- Завершить отладку handler_test.go (конфликты имен функций между repo_test и handler_test)
- Запустить все тесты и убедиться, что зелёные
- Создать коммит feat(005-iter2)
- Обновить HANDOFF.md с финальными результатами

## fix: forum reply form (commit 112fed4)

**Проблема:** форма ответа на пост отправлялась GET-запросом с данными в query string.

**Причина:** reply-форма была внутри `<template x-if="replyTo !== null">`. Alpine вставляет элемент в DOM после инициализации HTMX → `hx-post` не обрабатывается → нативный GET-сабмит.

**Решение:** перешли на паттерн из `article.html` — inline reply-форма для каждого поста с `x-show="replyTo === {{.ID}}"` + `x-cloak`. Форма всегда в DOM, HTMX инициализирует при загрузке. `@submit="replyTo = null"` сбрасывает состояние синхронно до HTMX-свопа (так Alpine инициализирует новые элементы скрытыми).

Также: `handler.go CreatePost` теперь передаёт `Topic`/`CanReply`/`CSRFToken` в partial, `base.html` получил `[x-cloak]` стиль.

---

## Проблемы и решения

- **fix/auth-htmx-422**: HTMX 1.9 не делает swap на 4xx по умолчанию → ошибки логина/регистрации не показывались. Исправлено двумя изменениями: (1) `htmx:beforeSwap` listener в `base.html` разрешает swap на 422/409; (2) добавлен `renderForm` хелпер в `auth/handler.go` — для HTMX-запросов возвращает только partial (`content`-блок), а не полный HTML-документ.

- `goose.Up` падал на пустой директории → добавили проверку "no migration files found" с graceful skip
- `go mod tidy` убрал прямую зависимость `google/uuid` (нет прямого импорта в момент tidy) → добавлена обратно через `go get`
- `auth.Config` не содержал `CookieSecure` → добавили поле
- pgx v5 не умеет сканировать INET в `string` → исправили в `session/repo.go` через `ip_addr::text`
- **fix/001**: `GET /login` → 500 из-за конфликта `{{define "content"}}` между login.html и register.html → `internal/tmpl/renderer.go` переписан: каждый page-файл получает собственный `*template.Template` (layouts + page); сигнатура `New` расширена параметром `prefix string` для соответствия именам в хендлерах
- **fix/003-1**: `$$`-функция в `005_create_forum_and_matches.sql` не парсилась goose → добавлены `-- +goose StatementBegin` / `-- +goose StatementEnd` вокруг `CREATE FUNCTION`
- **fix/003-2**: Тесты использовали `password_hash` — реальная колонка называется `pass_hash BYTEA` → исправлено во всех тестах
- **fix/003-3**: Параллельный `goose.Up()` из разных пакетов вызывает race condition → интеграционные тесты запускать с `-p 1`; добавлена заметка в CLAUDE.md
- **004-iter1/seed-contamination**: `TestLatestPublished_Empty` и `TestLatestPublished_ExcludesDrafts` падают из-за seed-данных в тестовой БД (добавлены в `f34f9ba`). `cleanNews` чистит только записи test-пользователя. Pre-existing проблема, не связана с фичей 004.
- **004-iter2/csrf-in-tests**: POST-тесты получали 403 — CSRF middleware требует `_csrf` куку + form-поле с одинаковым токеном. Исправлено в `doPost` хелпере: GET → извлечь `_csrf` из Set-Cookie, включить в куку + форму POST.
- **fix/reply-form-not-closing**: форма ответа на комментарий не закрывалась после сабмита. Причина: Alpine v3 MutationObserver инициализирует новые элементы после HTMX outerHTML-свопа с текущим значением `replyTo` (ещё не сброшен) и не устанавливает реактивную подписку на последующие изменения. Попытка сбросить через `hx-on:htmx:before-swap` не сработала — в HTMX 1.9 этот ивент стреляет на target (`#comments-list`), а не на elt (форму). Решение: `@submit="replyTo = null"` на reply form — Alpine-обработчик срабатывает до HTMX-запроса, DOM меняется уже с `replyTo = null`, новые элементы инициализируются скрытыми. Файл: `templates/news/article.html`.
