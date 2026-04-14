---
title: "FT-011: Admin — управление пользователями"
doc_kind: feature
doc_function: canonical
purpose: "Список пользователей, бан и разбан через admin-панель. Depends on FT-007, UC-003."
derived_from:
  - ../../domain/problem.md
  - ../../use-cases/UC-003-user-management.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-011: Admin — управление пользователями

## What

### Problem

Admin/Moderator не может заблокировать нарушителя через UI. Сейчас это возможно только через прямой SQL.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Бан пользователя через UI | Невозможен | `POST /admin/users/:id/ban` → 200, user не может войти | Интеграционный тест |
| `MET-02` | Разбан через UI | Невозможен | `POST /admin/users/:id/unban` → 200 | Интеграционный тест |

### Scope

- `REQ-01` Список пользователей (`GET /admin/users`) с email, датой регистрации, статусом (активен/заблокирован).
- `REQ-02` Бан пользователя: `POST /admin/users/:id/ban` → устанавливает `banned_at = now()`.
- `REQ-03` Разбан: `POST /admin/users/:id/unban` → обнуляет `banned_at`.
- `REQ-04` Заблокированный пользователь не может войти в систему (проверка в auth-middleware при `LoadSession` или в login handler).
- `REQ-05` Нельзя забанить самого себя.

### Non-Scope

- `NS-01` История банов (кто, когда, причина).
- `NS-02` Временный бан (auto-unban по таймауту).
- `NS-03` Редактирование email, пароля или роли пользователя через admin-панель.
- `NS-04` Удаление пользователей.
- `NS-05` Поиск/фильтрация пользователей — для MVP достаточно списка.

### Constraints / Assumptions

- `ASM-01` FT-007 реализован.
- `ASM-02` Таблица `users` существует. Поле `banned_at TIMESTAMPTZ` — может потребоваться новая миграция (OQ-01).
- `ASM-03` Auth (login handler или `LoadSession`) уже проверяет `banned_at IS NOT NULL` или нужно добавить эту проверку (OQ-01).
- `CON-01` Go на хосте не установлен — все команды через Docker.

## How

### Solution

Добавить endpoints бана/разбана в `internal/admin/`. Переиспользовать `internal/user/` repo через сервисный интерфейс. Если `banned_at` ещё не существует в схеме — добавить goose-миграцию. Обновить auth login-check.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `migrations/` | data | Миграция: `ALTER TABLE users ADD COLUMN banned_at TIMESTAMPTZ` (если не существует) |
| `internal/user/repo.go` | code | Методы: ListAll, BanUser, UnbanUser |
| `internal/user/service.go` | code | BanUser, UnbanUser, валидация «нельзя банить себя» |
| `internal/auth/` | code | Проверка `banned_at IS NOT NULL` при login и LoadSession |
| `internal/admin/users_handler.go` | code | HTTP handlers: список, бан, разбан |
| `templates/admin/users/` | code | list.html с кнопками бан/разбан |

### Flow

1. Admin открывает `/admin/users` — видит список с кнопками «Заблокировать» / «Разблокировать».
2. Нажимает «Заблокировать» → `POST /admin/users/:id/ban`.
3. Handler: проверяет `id != currentUserID`; вызывает `UserService.BanUser(id)` → `UPDATE users SET banned_at = now()`.
4. Пользователь при следующем входе получает отказ (login check: `banned_at IS NOT NULL → 401/403`).

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `POST /admin/users/:id/ban` → redirect `/admin/users` | Handler / Browser | Идемпотентно |
| `CTR-02` | `POST /admin/users/:id/unban` → redirect `/admin/users` | Handler / Browser | Идемпотентно |
| `CTR-03` | `POST /login` для заблокированного → 403 | Auth handler / Browser | banned_at IS NOT NULL |

### Failure Modes

- `FM-01` Попытка забанить себя → 400.
- `FM-02` Пользователь не найден → 404.
- `FM-03` DB error → 500, slog.Error.

## Verify

### Exit Criteria

- `EC-01` Admin банит User → `banned_at` заполнена в БД.
- `EC-02` Заблокированный пользователь не может войти (login → 403).
- `EC-03` Admin разбанивает User → `banned_at = NULL`, пользователь может войти.
- `EC-04` Попытка забанить самого себя → 400.
- `EC-05` Автоматические тесты зелёные.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01` | `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `ASM-02`, `CTR-01`, `FM-01`, `FM-02` | `EC-01`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CTR-02` | `EC-03`, `SC-04` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-03`, `CTR-03` | `EC-02`, `SC-03` | `CHK-01` | `EVID-01` |
| `REQ-05` | `FM-01` | `EC-04`, `SC-05` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Admin открывает `/admin/users` — видит список пользователей с email и статусом.
- `SC-02` Admin банит User → в списке статус меняется на «Заблокирован».
- `SC-03` Заблокированный пользователь пытается войти → получает сообщение об ошибке.
- `SC-04` Admin разбанивает User → пользователь снова может войти.
- `SC-05` Admin пытается забанить себя → получает ошибку.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`..`EC-05`, `SC-01`..`SC-05` | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/admin/... ./internal/user/... ./internal/auth/...` | Все тесты зелёные | stdout теста |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | stdout `go test` |

### Evidence

- `EVID-01` Вывод `go test` с `ok internal/admin`, `ok internal/user`, `ok internal/auth` и без FAIL.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | stdout go test | docker test run | stdout | `CHK-01` |

## Open Questions

- `OQ-01` Поле `banned_at` уже есть в таблице `users`? Если нет — нужна goose-миграция. Проверить перед реализацией.
- `OQ-02` Активные сессии заблокированного пользователя: аннулировать немедленно (`DELETE FROM sessions WHERE user_id = ?`) или при следующем запросе? LoadSession уже проверяет или нужно добавить check?
