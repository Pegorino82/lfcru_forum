---
title: "FT-028: Раздел «Команда» на форуме"
doc_kind: feature
doc_function: canonical
purpose: "Создание форумного раздела «Команда» с автоматически сгенерированными темами по игрокам из football-data.org API."
derived_from:
  - ../../domain/problem.md
  - ../../prd/PRD-001-forum-content-sections.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
---

# FT-028: Раздел «Команда» на форуме

## What

### Problem

LFC.ru не содержит структурированной информации о текущем составе Ливерпуля. Болельщики не могут найти список игроков, их позиции и базовые данные, а также не имеют тематических тем для обсуждения конкретных игроков. Upstream PRD: [PRD-001](../../prd/PRD-001-forum-content-sections.md).

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Форумный раздел «Команда» с темами по игрокам | 0 | 1 раздел + ≥ 20 тем (по числу игроков в ростере) | COUNT topics WHERE section = «Команда» |

### Scope

- `REQ-01` Система содержит данные о составе команды, загруженные из football-data.org API на момент запуска генерации: имя, позиция, дата рождения, национальность каждого игрока.
- `REQ-02` Создать форумный раздел «Команда» (открытый для гостей).
- `REQ-03` Автоматически сгенерировать тему на каждого игрока. Первый пост содержит карточку игрока: имя, позиция, дата рождения, национальность. Поля, недоступные через API (номер, контракт), оставить пустыми с placeholder.
- `REQ-04` Темы создаются от имени системного пользователя (admin). Обычные пользователи могут отвечать в этих темах.
- `REQ-05` Graceful degradation: если API недоступен, раздел создаётся пустым; генерация тем откладывается.

### Non-Scope

- `NS-01` Автоматическая периодическая синхронизация ростера с API (PRD-001 `NG-03`).
- `NS-02` Фото игроков — API не предоставляет на бесплатном tier.
- `NS-03` Отдельная страница сайта «Команда» вне форума (PRD-001 `NG-01`).
- `NS-04` Разделы «Тренерский штаб» и «Руководство» — отдельные фичи FT-029, FT-030.
- `NS-05` Номер на футболке (`shirtNumber`) — API не возвращает это поле; placeholder в карточке.

### Constraints / Assumptions

- `ASM-01` football-data.org API доступен и возвращает squad data для team ID 64 (Liverpool FC) — подтверждено тестовым вызовом.
- `ASM-02` API возвращает ~43 записи в squad (включая youth/reserve).
- `CON-01` Бесплатный tier football-data.org: 10 requests/minute. Генерация тем выполняется одним API-вызовом (`/v4/teams/64`), а не отдельным запросом на игрока.
- `CON-02` CSRF-токен обязателен для POST-запросов создания тем (PCON-02).
- `DEC-01` Формат карточки игрока в первом посте: HTML-шаблон, встраиваемый в тело поста. Решение принято — не требует отдельного UI-компонента.
- `DEC-02` Механизм запуска генерации тем: admin endpoint (кнопка в админке или CLI-команда). Конкретный вариант — OQ в плане.
- `DEC-03` Стратегия обработки youth/reserve записей (фильтровать или отображать все) — OQ для implementation-plan.

## How

### Solution

Расширить существующий `football.Client` методом `Squad()`, который возвращает список игроков из `/v4/teams/64`. Добавить admin HTTP endpoint для генерации тем: один вызов API → парсинг squad → создание раздела (если не существует) → bulk-создание тем с карточками игроков в первом посте. Кеширование аналогично существующим методам (TTL 24h). Карточка игрока — HTML-фрагмент, где статическая разметка передаётся как safeHTML, а данные игрока из API проходят стандартный Go template escaping для защиты от XSS.

**Trade-off:** один API-вызов на весь squad (вместо отдельных запросов на игрока) экономит rate-limit, но не даёт получить дополнительные данные (shirtNumber, contract) — они недоступны на endpoint `/v4/teams/{id}` бесплатного tier.

### Change Surface

| Surface | Type | Why it changes |
| --- | --- | --- |
| `internal/football/client.go` | code | Новый метод `Squad()` для получения данных о составе |
| `internal/football/models.go` (новый) | code | Структуры `Player`, `SquadResponse` |
| `internal/admin/handler.go` | code | Admin endpoint POST для генерации тем |
| `internal/forum/service.go` | code | Метод генерации тем по списку игроков |
| `internal/forum/repo.go` | code | Bulk-создание тем и первых постов |
| `cmd/forum/main.go` | code | Регистрация нового admin route |
| `templates/forum/player-card.html` (новый) | template | HTML-шаблон карточки игрока для первого поста |
| `migrations/` | data | Миграция для создания раздела «Команда» (если не через seed) |

### Flow

1. Администратор запускает генерацию (admin HTTP POST endpoint, защищён CSRF и RequireAuth + role check).
2. Система вызывает `footballClient.Squad(ctx)` → получает список игроков из API.
3. Система проверяет/создаёт раздел «Команда» в таблице `sections`.
4. Для каждого игрока: создаёт тему с заголовком `{name}` и первый пост с HTML-карточкой (имя, позиция, дата рождения, национальность).
5. Темы создаются от имени admin-пользователя (ID = 1 или конфигурируемый).
6. При ошибке API — раздел создаётся, темы не генерируются; логируется ошибка.

### Contracts

| Contract ID | Input / Output | Producer / Consumer | Notes |
| --- | --- | --- | --- |
| `CTR-01` | `GET /v4/teams/64` → JSON с полями `squad[].{id,name,position,dateOfBirth,nationality}`, `coach.{name,dateOfBirth,nationality,contract}` | football-data.org / `football.Client` | Бесплатный tier; 10 req/min |
| `CTR-02` | `Squad()` → `[]Player` | `football.Client` / `forum.Service` | Кешируется 24h; nil при ошибке |
| `CTR-03` | Тема форума с первым постом (HTML-карточка) | `forum.Service` / UI (шаблон темы) | Статическая разметка карточки — safeHTML; данные игрока из API — стандартный Go template escaping |

### Failure Modes

- `FM-01` API football-data.org недоступен → `Squad()` возвращает кеш или nil → генерация тем пропускается, логируется warning. Раздел «Команда» создаётся пустым.
- `FM-02` API возвращает пустой squad → генерация тем пропускается, логируется warning.
- `FM-03` Раздел «Команда» уже существует → не дублируется. Темы с совпадающими заголовками → пропускаются (idempotent).
- `FM-04` XSS через данные из внешнего API (имя игрока, национальность) → данные проходят стандартный Go template escaping; safeHTML применяется только к статической HTML-разметке карточки, не к данным из API.

### ADR Dependencies

Нет. Фича использует существующие паттерны (football-data.org client, forum sections/topics).

## Verify

### Exit Criteria

- `EC-01` Раздел «Команда» создан и виден в списке разделов форума (гостям и авторизованным).
- `EC-02` Темы по игрокам сгенерированы с корректными карточками (имя, позиция, дата рождения, национальность).
- `EC-03` Повторный запуск генерации не создаёт дубликатов.
- `EC-04` При недоступном API раздел создаётся пустым без ошибки 500.

### Traceability matrix

| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CTR-01`, `CTR-02`, `CON-01` | `EC-02`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | — | `EC-01`, `SC-01` | `CHK-02` | `EVID-02` |
| `REQ-03` | `CTR-03`, `DEC-01` | `EC-02`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `CON-02` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-05` | `FM-01`, `FM-02` | `EC-04`, `SC-03`, `NEG-01` | `CHK-03` | `EVID-03` |

### Acceptance Scenarios

- `SC-01` **Happy path:** Admin запускает генерацию → API возвращает squad → раздел «Команда» появляется в списке → каждый игрок имеет тему с карточкой (имя, позиция, дата рождения, национальность в первом посте).
- `SC-02` **Idempotent re-run:** повторный запуск генерации → новые темы не создаются, существующие не дублируются.
- `SC-03` **API unavailable:** API не отвечает → раздел «Команда» создаётся пустым → нет ошибки 500 → в логах warning.

### Negative / Edge Cases

- `NEG-01` API возвращает игрока с пустым именем или null полями → тема не создаётся, игрок пропускается, в логах warning.
- `NEG-02` Данные из API содержат HTML/script теги в имени игрока → Go template escaping экранирует их, XSS невозможен.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-02`, `SC-01` | `docker run --rm -v $(pwd):/app -w /app -v lfcru_gomod:/root/go/pkg/mod --network lfcru_forum_default golang:1.23-alpine go test -tags integration -p 1 ./internal/football/... ./internal/forum/...` | Темы созданы с корректными данными; Squad() возвращает список игроков | `EVID-01` |
| `CHK-02` | `EC-01`, `EC-03`, `SC-02`, `NEG-02` | `npx playwright test e2e/forum/team-section.spec.ts` | Раздел «Команда» виден, тема игрока содержит карточку с данными, повторный запуск не дублирует, XSS-строки экранированы | `EVID-02` |
| `CHK-03` | `EC-04`, `SC-03`, `NEG-01` | `docker run --rm -v $(pwd):/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./internal/football/... -run TestSquad` (mock HTTP 500 + пустые поля) | Squad() → nil при 500; пустые имена пропускаются | `EVID-03` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `.github/workflows/ci.yml` → Go Tests job |
| `CHK-02` | `EVID-02` | `e2e/test-report/` |
| `CHK-03` | `EVID-03` | `.github/workflows/ci.yml` → Go Tests job |

### Evidence

- `EVID-01` Go test pass для `Squad()` и генерации тем (unit + integration).
- `EVID-02` Playwright test pass: раздел «Команда» виден, тема игрока содержит карточку, повторный запуск idempotent.
- `EVID-03` Go test pass: graceful degradation при недоступном API.
- `EVID-04`: Spec loop — accept. 2026-05-18. evaluator agent
- `EVID-05`: Plan eval — accept. 2026-05-18. evaluator agent

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Go test output (unit + integration) | CI / docker run go test | `.github/workflows/ci.yml` → Go Tests job | `CHK-01` |
| `EVID-02` | Playwright test report | CI / npx playwright test | `e2e/test-report/` | `CHK-02` |
| `EVID-03` | Go test output (graceful degradation) | CI / docker run go test | `.github/workflows/ci.yml` → Go Tests job | `CHK-03` |
