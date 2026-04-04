# HANDOFF.md

> Файл передачи контекста между агентами. Обновляется в конце каждого сеанса.

---

## Что сделано

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

## Что сделать следующим

- Реализовать следующую фичу согласно `PROJECT.md`

## Проблемы и решения

- `goose.Up` падал на пустой директории → добавили проверку "no migration files found" с graceful skip
- `go mod tidy` убрал прямую зависимость `google/uuid` (нет прямого импорта в момент tidy) → добавлена обратно через `go get`
- `auth.Config` не содержал `CookieSecure` → добавили поле
- pgx v5 не умеет сканировать INET в `string` → исправили в `session/repo.go` через `ip_addr::text`
- **fix/001**: `GET /login` → 500 из-за конфликта `{{define "content"}}` между login.html и register.html → `internal/tmpl/renderer.go` переписан: каждый page-файл получает собственный `*template.Template` (layouts + page); сигнатура `New` расширена параметром `prefix string` для соответствия именам в хендлерах
- **fix/003-1**: `$$`-функция в `005_create_forum_and_matches.sql` не парсилась goose → добавлены `-- +goose StatementBegin` / `-- +goose StatementEnd` вокруг `CREATE FUNCTION`
- **fix/003-2**: Тесты использовали `password_hash` — реальная колонка называется `pass_hash BYTEA` → исправлено во всех тестах
- **fix/003-3**: Параллельный `goose.Up()` из разных пакетов вызывает race condition → интеграционные тесты запускать с `-p 1`; добавлена заметка в CLAUDE.md
