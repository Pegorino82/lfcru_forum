---
title: "FT-010: Admin — управление структурой форума"
doc_kind: feature
doc_function: canonical
purpose: "Создание и редактирование разделов и тем форума через admin-панель. Depends on FT-007, UC-002."
derived_from:
  - ../../domain/problem.md
  - ../../use-cases/UC-002-forum-structure-management.md
  - ../../adr/ADR-004-forum-hierarchy-model.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-010: Admin — управление структурой форума

## What

### Problem

Разделы и темы форума нельзя создать или изменить через UI. Сейчас это возможно только через прямые SQL-запросы.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Создание раздела через UI | Невозможно | `POST /admin/forum/sections` → 200, раздел виден в форуме | Интеграционный тест |
| `MET-02` | Создание темы через UI | Невозможно | `POST /admin/forum/sections/:id/topics` → 200 | Интеграционный тест |

### Scope

- `REQ-01` Список разделов форума (`GET /admin/forum/sections`) с возможностью перейти к редактированию.
- `REQ-02` Форма создания раздела: название, описание (опционально). `POST /admin/forum/sections`.
- `REQ-03` Форма редактирования раздела: изменение названия и описания. `POST /admin/forum/sections/:id`.
- `REQ-04` Список тем раздела в admin (`GET /admin/forum/sections/:id/topics`).
- `REQ-05` Форма создания темы: название. `POST /admin/forum/sections/:id/topics`.
- `REQ-06` Форма редактирования темы: изменение названия. `POST /admin/forum/topics/:id`.

### Non-Scope

- `NS-01` Удаление разделов и тем — требует решения по orphaned content (темы, сообщения).
- `NS-02` Изменение порядка разделов (drag-and-drop).
- `NS-03` Вложенные подразделы (ADR-004: только два уровня sections→topics).
- `NS-04` Права доступа на уровне раздела (публичный/приватный).

### Constraints / Assumptions

- `ASM-01` FT-007 реализован: `RequireAdminOrMod` middleware.
- `ASM-02` Таблицы `sections` и `topics` существуют (миграции FT-005). Репозиторий `internal/forum/` уже содержит базовые методы.
- `ASM-03` ADR-004: секция sections → topics (два уровня), без вложенности.
- `CON-01` Go на хосте не установлен — все команды через Docker.

## How

### Solution

Добавить admin-handlers в `internal/admin/` для секций и тем. Переиспользовать существующие методы `internal/forum/` repo через service-интерфейс (не импортировать repo напрямую). Добавить недостающие методы Create/Update в forum.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/forum/repo.go` | code | Методы: CreateSection, UpdateSection, CreateTopic, UpdateTopic |
| `internal/forum/service.go` | code | Интерфейс + бизнес-правила валидации |
| `internal/admin/forum_handler.go` | code | HTTP handlers для секций и тем |
| `templates/admin/forum/` | code | sections_list.html, section_edit.html, topics_list.html, topic_edit.html |

### Flow

1. Admin открывает `/admin/forum/sections` — список разделов.
2. Нажимает «Создать раздел» → форма → `POST /admin/forum/sections`.
3. Раздел появляется в публичном форуме немедленно.
4. Admin выбирает раздел → список тем → создаёт тему аналогично.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `POST /admin/forum/sections` body: name, description → redirect `/admin/forum/sections` | Handler / Browser | Валидация: name не пустой |
| `CTR-02` | `POST /admin/forum/sections/:id/topics` body: name → redirect | Handler / Browser | section must exist |

### Failure Modes

- `FM-01` Пустое название раздела или темы → 400, форма с ошибкой валидации.
- `FM-02` Раздел не найден (`:id` несуществующий) → 404.
- `FM-03` DB error → 500, slog.Error.

## Verify

### Exit Criteria

- `EC-01` Создание раздела через форму → раздел виден на публичном `/forum`.
- `EC-02` Редактирование названия раздела → изменение отражается немедленно.
- `EC-03` Создание темы в разделе → тема видна в публичном разделе.
- `EC-04` Попытка создать раздел с пустым именем → 400.
- `EC-05` Автоматические тесты зелёные.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `ASM-02` | `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CTR-01`, `FM-01` | `EC-01`, `EC-04`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-03` | `CTR-01` | `EC-02`, `SC-03` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-02` | `SC-04` | `CHK-01` | `EVID-01` |
| `REQ-05` | `CTR-02`, `FM-01`, `FM-02` | `EC-03`, `SC-05` | `CHK-01` | `EVID-01` |
| `REQ-06` | `CTR-02` | `SC-06` | `CHK-01` | `EVID-01` |

### Acceptance Scenarios

- `SC-01` Admin открывает `/admin/forum/sections` — видит список существующих разделов.
- `SC-02` Admin создаёт раздел → он виден в `/forum`.
- `SC-03` Admin редактирует название раздела → публичный форум отражает изменение.
- `SC-04` Admin выбирает раздел → видит список тем.
- `SC-05` Admin создаёт тему в разделе → тема появляется в публичном форуме.
- `SC-06` Admin редактирует название темы → изменение отражается.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`..`EC-05`, `SC-01`..`SC-06` | `docker compose -f docker-compose.dev.yml run --rm app go test -tags integration -p 1 ./internal/admin/... ./internal/forum/...` | Все тесты зелёные | stdout теста |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | stdout `go test` |

### Evidence

- `EVID-01` Вывод `go test` с `ok internal/admin`, `ok internal/forum` и без FAIL.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | stdout go test | docker test run | stdout | `CHK-01` |
