# HANDOFF.md

> Файл передачи контекста между агентами. Обновляется в конце каждого сеанса.

---

## Что сделано

- **Шаг 0**: Скаффолдинг — `go.mod`, `cmd/forum/main.go`, `internal/config/config.go`, `Dockerfile`, `docker-compose.dev.yml`, `.env.example`
- **Шаг 1**: Миграции — `users`, `sessions`, `login_attempts` через goose
- **Шаг 2**: Репозитории — `user.Repo`, `session.Repo`, `ratelimit.LoginAttemptRepo` (pgx)
- **Шаг 3**: Сервис аутентификации �� `auth.Service` с Register/Login/Logout/GetSession + юнит-тесты
- **Шаг 4**: CSRF middleware (echo built-in, `_csrf` form field)
- **Шаг 5**: Session middleware — `LoadSession`, `RequireAuth`, `UserFromContext`
- **Шаг 6**: HTTP handlers — `/register`, `/login`, `/logout` с HTMX-совместимыми ответами
- **Шаг 7**: Шаблоны — `base.html`, `register.html`, `login.html`
- **Шаг 8**: Фоновые задачи — `cleanup.Run` (goroutine с ticker)
- Всё подключено в `main.go`, сервер стартует и отдаёт страницы

## Что сделать следующим

- **Шаг 9**: Интеграционные тесты — `internal/auth/integration_test.go` с testcontainers-go (16 сценариев из спеки §10)
- Добавить `internal/ratelimit/repo_test.go` (интеграционные тесты repo)
- Убедиться, что goose down откатывает миграции корректно

## Проблемы и решения

- `goose.Up` падал на пустой директории → добавили проверку "no migration files found" с graceful skip
- `go mod tidy` убрал прямую зависимость `google/uuid` (нет прямого импорта в момент tidy) → добавлена обратно через `go get`
- `auth.Config` не содержал `CookieSecure` → добавили поле
