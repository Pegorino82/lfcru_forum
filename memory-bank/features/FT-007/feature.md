---
title: "FT-007: Admin-панель — инфраструктура"
doc_kind: feature
doc_function: canonical
purpose: "RBAC middleware, роутинг /admin/*, базовый layout admin-панели. Фундамент для всех последующих admin-фич."
derived_from:
  - ../../domain/problem.md
  - ../../domain/architecture.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-007: Admin-панель — инфраструктура

## What

### Problem

Нет выделенного маршрутного пространства и middleware для admin-функций. Добавление FT-008..011 без общего фундамента приведёт к дублированию проверок ролей в каждом handler и отсутствию единого layout.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Доступность admin-раздела | Нет маршрута `/admin` | `GET /admin` → 200 для Admin/Mod, 403 для User, 302 для гостя | Интеграционный тест |

### Scope

- `REQ-01` Middleware `RequireAdminOrMod`: проверяет, что аутентифицированный пользователь имеет роль `admin` или `moderator`. Гостей редиректит на `/login`, зарегистрированных User — 403.
- `REQ-02` Группа маршрутов `GET /admin` и `/admin/*` с применённым `RequireAdminOrMod`.
- `REQ-03` Базовый HTML-шаблон admin-layout (`templates/admin/layout.html`) с навигацией: Статьи, Форум, Пользователи.
- `REQ-04` Дашборд `GET /admin` — заглушка-страница (список разделов навигации), без бизнес-логики.

### Non-Scope

- `NS-01` Любая бизнес-логика (статьи, форум, пользователи) — в FT-008..011.
- `NS-02` Разграничение прав внутри admin-панели между Admin и Moderator — OQ-01, в этом FT оба получают одинаковый доступ.
- `NS-03` Кастомный дизайн/стили — используем inline CSS как в существующих шаблонах.
- `NS-04` Pagination, поиск, фильтрация на дашборде.

### Constraints / Assumptions

- `ASM-01` Роли пользователей уже хранятся в таблице `users` (поле `role`). Если нет — нужно добавить миграцию (OQ-02).
- `ASM-02` `LoadSession` middleware уже добавляет `User` в контекст.
- `CON-01` Шаблоны через `html/template` stdlib; HTMX используется по необходимости.
- `CON-02` Go на хосте не установлен — все команды через Docker.

## How

### Solution

Добавить `internal/admin/` пакет с `Handler`, зарегистрировать группу `/admin` в `main.go` с `RequireAdminOrMod` middleware. Создать базовый layout-шаблон. Дашборд — пустая страница с навигацией.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/admin/handler.go` | code | Новый пакет: Handler, Dashboard |
| `internal/admin/middleware.go` | code | RequireAdminOrMod middleware |
| `templates/admin/layout.html` | code | Base admin layout с навигацией |
| `templates/admin/dashboard.html` | code | Дашборд-заглушка |
| `cmd/forum/main.go` | code | Регистрация admin-группы маршрутов |

### Flow

1. Пользователь открывает `GET /admin`.
2. `LoadSession` → user в ctx.
3. `RequireAdminOrMod`: если нет user — redirect `/login`; если role не admin/moderator — 403; иначе пропускает.
4. Handler рендерит `dashboard.html` через `layout.html`.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `GET /admin` → 200 HTML | Handler / Browser | Только для admin/mod |
| `CTR-02` | `GET /admin` (guest) → 302 /login | Middleware / Browser | — |
| `CTR-03` | `GET /admin` (User role) → 403 | Middleware / Browser | — |

### Failure Modes

- `FM-01` Пользователь с активной сессией, но заблокированный — `RequireAdminOrMod` проверяет `is_banned` и редиректит.
- `FM-02` Роль в сессии не совпадает с БД (stale session) — middleware может опираться только на данные из session; при смене роли требуется перелогин.

## Verify

### Exit Criteria

- `EC-01` `GET /admin` возвращает 200 для пользователя с ролью `admin`.
- `EC-02` `GET /admin` возвращает 200 для пользователя с ролью `moderator`.
- `EC-03` `GET /admin` гостя → 302 `/login`.
- `EC-04` `GET /admin` User (не admin/mod) → 403.
- `EC-05` Автоматические тесты зелёные.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `ASM-02`, `CTR-02`, `CTR-03`, `FM-01`, `FM-02` | `EC-01`..`EC-04`, `SC-01`..`SC-04` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-01` | `EC-01`, `EC-02` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CON-01` | `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CTR-01` | `SC-01` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Admin открывает `/admin` — видит страницу дашборда с навигацией.
- `SC-02` Moderator открывает `/admin` — видит страницу дашборда.
- `SC-03` Гость переходит на `/admin` — перенаправляется на `/login`.
- `SC-04` User (без прав admin/mod) открывает `/admin` — получает 403.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`..`EC-05`, `SC-01`..`SC-04` | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/admin/...` | Все тесты зелёные | stdout теста |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | stdout `go test` |

### Evidence

- `EVID-01` Вывод `go test` с `ok internal/admin` и без FAIL.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | stdout go test | docker test run | stdout | `CHK-01` |

## Open Questions

- `OQ-01` Нужно ли различать права Admin и Moderator внутри admin-панели (например, только Admin может банить пользователей)? Пока оба получают одинаковый доступ.
- `OQ-02` Поле `role` в таблице `users` уже существует? Проверить перед реализацией.
