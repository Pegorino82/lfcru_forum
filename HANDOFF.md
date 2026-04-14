# HANDOFF.md

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
