---
title: Architecture Patterns
doc_kind: domain
doc_function: canonical
purpose: Каноничное место для архитектурных границ проекта. Читать при изменениях, затрагивающих модули, фоновые процессы, интеграции или конфигурацию.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Architecture Patterns

## Module Boundaries

| Context | Owns | Must not depend on directly |
| --- | --- | --- |
| `auth` | регистрация, вход, выход, управление сессиями, rate-limit | детали форума и контента |
| `user` | модель `User`, хранение и поиск по email/ID | сессионный state |
| `session` | жизненный цикл сессий: Create, GetByID, Delete, Touch, CleanupExpired | бизнес-логика auth |
| `ratelimit` | запись и подсчёт login-попыток по IP | auth-логика |
| `cleanup` | фоновая очистка истёкших сессий (1ч) и старых login_attempts (10м) | любые бизнес-домены |
| `forum` *(планируется)* | разделы, темы, сообщения, SSE-события форума | детали auth, кроме UserID |
| `content` *(планируется)* | новости, статьи, полнотекстовый поиск | форумный state |

Минимальные правила:

- Межмодульные зависимости проходят только через именованные интерфейсы (`UserRepo`, `SessionRepo`, `AttemptRepo`), не через прямые импорты репо-слоя чужого модуля.
- UI, jobs и интеграции не читают внутренние детали чужого модуля в обход его owner-интерфейса.

## Layer Stack (канонический порядок)

```
Handler (HTTP/Echo) → Service (бизнес-логика) → Repo (SQL/pgx) → PostgreSQL
```

- **Handler** — парсит запрос, вызывает Service, рендерит шаблон или отдаёт редирект. Зависит только от Service.
- **Service** — принимает интерфейсы репозиториев; содержит всю domain-логику: валидацию, хеширование, rate-limit политику, управление сессиями.
- **Repo** — параметризованные SQL-запросы через `pgxpool.Pool`; маппит pgx-ошибки в sentinel errors домена.
- **Middleware** — `LoadSession` (cookie → user в контексте), `RequireAuth` (редирект на `/login`), `CSRF`.

DI и инициализация в `cmd/forum/main.go`: config → pool → goose migrations → repos → services → handlers → Echo routes → cleanup goroutine → graceful shutdown.

## Concurrency And Critical Sections

- `pgxpool.Pool` — connection pool; конкурентный доступ из горутин безопасен по умолчанию.
- **SSE fan-out**: одна горутина слушает `LISTEN/NOTIFY` от PostgreSQL, рассылает события по каналам всем активным подписчикам.
- **Cursor-based догонялка при реконнекте**: клиент передаёт `Last-Event-ID`, сервер досылает пропущенные события из PG по cursor.
- `LISTEN/NOTIFY` payload ≤ 8000 байт — передавать только ID события; полные данные клиент получает через отдельный запрос по cursor.
- **Cleanup-горутина** запускается как `go cleanup.Run(ctx, pool)` — не блокирует основной поток; завершается по context cancellation при graceful shutdown.

Запрещено:

- Не строить SQL через `fmt.Sprintf` — только параметризованные запросы.
- Не запускать интеграционные тесты параллельно (`-p 1` обязателен): каждый пакет вызывает `goose.Up()` и параллельный запуск вызывает race condition.

## Failure Handling And Error Tracking

- **Sentinel errors**: `var ErrXxx = errors.New(...)` объявляются в пакете домена (например, `auth/errors.go`).
- Repo маппит pgx `unique_violation` → `ErrDuplicateEmail`; Handler сравнивает через `errors.Is` и транслирует в HTTP-статус.
- Handler не дублирует логирование ошибок, если Echo `Recover` middleware уже обрабатывает панику.
- Ошибки конфигурации и `goose.Up()` — `log.Fatal` на старте; recovery не предусмотрен, это ожидаемое поведение.

## Configuration Ownership

1. **Canonical schema**: `internal/config/config.go` — единственный owner всех env-переменных проекта.
2. **Defaults** задаются там же, в `config.Load()`.
3. **Environment overlays**: `docker-compose.dev.yml` и `docker-compose.prod.yml` — передают env в контейнер.
4. При добавлении новой переменной: сначала обновить `internal/config/config.go`, затем `../ops/config.md`.
