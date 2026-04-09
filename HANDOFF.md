# HANDOFF.md

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

## Что сделать следующим — Итерация 2 (005)

Spec: `memory-bank/features/005/rspec.md`
Plan: `memory-bank/features/005/plan.md` (раздел "Итерация 2 — HTTP Layer + Templates")

**Шаги:**

1. **Middleware** (`internal/auth/middleware.go`):
   - RequireRole(roles ...string) — редирект на /login или 403 → templates/errors/403.html

2. **Обработчики** (`internal/forum/handler.go`):
   - Index, ShowSection, ShowTopic, NewSection, CreateSection, NewTopic, CreateTopic, CreatePost (8 методов)
   - Маппинг ошибок на HTTP-коды (ErrSectionNotFound → 404, валидационные → 422)
   - HTMX: CreatePost возвращает partial #posts-list (201) или 422 с формой

3. **Шаблоны** (`templates/forum/`):
   - index.html, section.html, topic.html, new_section.html, new_topic.html (5 шаблонов)
   - topic.html: Alpine.js для reply-форм, HTMX для CreatePost, якоря #post-{id}
   - templates/errors/403.html — страница ошибки прав

4. **Wire up** (`cmd/forum/main.go`):
   - Создать forumRepo, forumSvc, forumHandler
   - Зарегистрировать роуты (порядок важен: статические до параметрических)
   - modGroup (CreateSection, CreateTopic) требует role moderator/admin
   - authGroup (CreatePost) требует auth

5. **Правки** существующих файлов:
   - templates/home/index.html — ссылки на темы, кнопка «Все разделы» → /forum
   - templates/layouts/base.html — навигация, ссылка на /forum

6. **Интеграционные тесты** (`internal/forum/handler_test.go`):
   - 22 тест-кейса: GET/POST на все эндпоинты, ошибки 404/422/403, HTMX-поведение
   - Паттерн: doGet/doPost helpers с CSRF (как в news/handler_test.go feature 004)

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
