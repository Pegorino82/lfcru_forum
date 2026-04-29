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
| `forum` | разделы, темы, сообщения, SSE-события форума | детали auth, кроме UserID |
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

Миграции (`migrations/`) — SQL-файлы goose; применяются при старте через `goose.Up()` до инициализации репозиториев.

## Concurrency And Critical Sections

- `pgxpool.Pool` — connection pool; конкурентный доступ из горутин безопасен по умолчанию.
- **SSE fan-out (MVP, FT-016)**: in-process broadcast hub (`internal/forum/hub.go`) — `map[topicID][]chan string`, защищён `sync.RWMutex`. `CreatePost` пушит HTML-фрагмент в hub после записи в БД, исключая канал автора. Каждое SSE-соединение — одна горутина + один буферизованный канал (≥ 16). Hub ограничен 200 подписчиками/топик и 2000 глобально. **ADR**: PostgreSQL `LISTEN/NOTIFY` отложен до масштабирования на multi-pod (NS-04 в FT-016).
- **Cursor-based догонялка при реконнекте**: клиент передаёт `Last-Event-ID` (database post ID), сервер досылает пропущенные посты `WHERE topic_id = :id AND id > last_event_id ORDER BY id ASC LIMIT 50`.
- **Cleanup-горутина** запускается как `go cleanup.Run(ctx, pool)` — не блокирует основной поток; завершается по context cancellation при graceful shutdown.

Запрещено:

- Не строить SQL через `fmt.Sprintf` — только параметризованные запросы.
- Не запускать интеграционные тесты параллельно (`-p 1` обязателен): каждый пакет вызывает `goose.Up()` и параллельный запуск вызывает race condition.

## Failure Handling And Error Tracking

- **Sentinel errors** (синтаксис объявления — [coding-style.md](../engineering/coding-style.md)): объявляются в пакете домена (например, `auth/errors.go`).
- Repo маппит pgx `unique_violation` → `ErrDuplicateEmail`; Handler сравнивает через `errors.Is` и транслирует в HTTP-статус.
- Handler не дублирует логирование ошибок, если Echo `Recover` middleware уже обрабатывает панику.
- Ошибки конфигурации и `goose.Up()` — `log.Fatal` на старте; recovery не предусмотрен, это ожидаемое поведение.

## Article Body Storage (ADR-007)

Тело статьи (`news.content`) хранится как **HTML-строка**.

- Редактор — **TipTap** (vanilla JS, ESM CDN), генерирует HTML напрямую.
- При сохранении (Create/Update): HTML санитизируется через **bluemonday** с allowlist-политикой (разрешены `p`, `h1`-`h3`, `strong`, `em`, `s`, `a[href]`, `img[src,alt]`, `figure`, `figcaption`, `br`, `p[style="text-align:*"]`, `div[style="text-align:*"]`).
- При рендеринге: `template.HTML(article.Content)` — без повторной санитизации (sanitize-at-write).
- Формат Markdown больше не используется для новых статей; существующие Markdown-статьи деградируют без отдельной миграции (ADR-007 / ASM-03).

## Configuration Ownership

1. **Canonical schema**: `internal/config/config.go` — единственный owner всех env-переменных проекта.
2. **Defaults** задаются в helper-функциях `getEnv / getInt / getBool / getDuration` в `config.go`.
3. **Environment overlays**: `docker-compose.dev.yml` и `docker-compose.prod.yml` — передают env в контейнер.
4. При добавлении новой переменной: сначала обновить `internal/config/config.go`, затем `../ops/config.md`.
