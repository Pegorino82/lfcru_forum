---
title: "FT-025: Implementation Plan Review — Gate DR→PR"
doc_kind: eval
doc_function: gate-eval
gate: Design Ready → Plan Ready
feature: FT-025
reviewed: implementation-plan.md
date: 2026-05-05
outcome: revise
evaluator: evaluator agent
---

# Review: FT-025 Implementation Plan (Gate DR→PR)

## Outcome

**revise**

Два блокера (статус плана и команда unit-тестов) требуют исправления до перехода в Plan Ready.

---

## Замечания

### BLOCKER-1 — E-1: `implementation-plan.md` статус `draft` вместо `active`

**Цитата:**
```yaml
status: draft
```

**Норма:** feature-flow.md § «Design Ready → Plan Ready»:
> `implementation-plan.md` → `status: active`

Переход в Plan Ready не допустим с `status: draft`. Это явное условие gate.

**Исправление:** изменить `status: draft` → `status: active` в frontmatter `implementation-plan.md`.

---

### BLOCKER-2 — D-1: команда Go unit-тестов не совпадает с канонической (ops/development.md)

**Цитата (Environment Contract):**
```
docker run golang:1.23-alpine go test ./...
```

**Норма:** ops/development.md § «Go-тесты» (единственный источник, использовать дословно):
```bash
docker run --rm \
  -v "$(pwd)":/app -w /app \
  -v lfcru_gomod:/root/go/pkg/mod \
  golang:1.23-alpine \
  go test ./...
```

Также testing-policy.md:
> ⛔ **Не изобретать `docker run` вручную.** Без `-v lfcru_gomod` модули скачиваются заново при каждом запуске.

Команда в плане не монтирует ни исходники (`-v "$(pwd)":/app -w /app`), ни кэш модулей (`-v lfcru_gomod:/root/go/pkg/mod`). При выполнении STEP-08 тест либо провалится, либо скачает зависимости заново.

**Исправление:** заменить команду в Environment Contract (и в строке Check STEP-08) на каноническую дословно из ops/development.md.

---

### HIGH-1 — C-6: E2E seeding для CHK-02 и CHK-03 не описан

**Цитата (Test Strategy):**
> `e2e/forum/avatar-display.spec.ts`: post-avatar visible + click → modal
> `e2e/news/avatar-display.spec.ts`: comment-avatar visible + click → modal

**Норма:** testing-policy.md § «E2E-тесты (Playwright)»:
> Тестовые данные вставляются с фиксированными ID через `OVERRIDING SYSTEM VALUE`; teardown чистит их по тому же ID

CHK-02 проверяет аватарки у постов форума, CHK-03 — у комментариев к новостям. Для этих сценариев требуется наличие записей в `forum_posts` и `news_comments` с конкретными `author_id`. Без описания seeding-стратегии (кто вставляет тестовые посты/комментарии, с каким `author_id`, в каком `topic_id`/`news_id`) план оставляет реализацию E2E-спеков неопределённой.

**Исправление:** добавить в STEP-09 (или в отдельный раздел Test Data Contract) описание seeding: какие фиксированные IDs используются, как `global-setup.ts` создаёт тестовые посты и комментарии, как `global-teardown.ts` их удаляет.

---

### HIGH-2 — C-2 / C-4: Integration-тесты отсутствуют в Test Strategy, хотя repo/model изменяются

**Норма:** testing-policy.md § «CI» — Go Tests job включает:
> `go test -tags integration -p 1 ./internal/...` — integration

testing-policy.md § «Что Считается Sufficient Coverage»:
> Покрыты новые или измененные contracts, события, schema или integration boundaries.

План добавляет `AuthorAvatarURL *string` в `PostView`/`CommentView` и расширяет Scan в repo — это integration boundary (repo ↔ БД). Test Strategy не содержит строки для integration-тестов. В строке Go unit написано «новые unit на scan не требуются — поведение покрыто integration», но integration-row в таблице отсутствует, а Required CI suites не упоминают integration.

**Исправление:** добавить строку в Test Strategy:

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI suites | Manual-only gap | Approval ref |
|---|---|---|---|---|---|---|---|
| Go integration (forum/comment repo) | CON-02 | Существующие integration-тесты | Существующие тесты проходят (Scan расширен, структура JOIN не нарушена) | — | Go Tests (CI), integration tag | — | none |

---

### MEDIUM-1 — B-2: расхождение `comment/repo.go` — INNER JOIN vs заявленный LEFT JOIN не зафиксировано как OQ-*

**Цитата (Discovery Context):**
> `internal/comment/repo.go:19` | `ListByNewsID` — JOIN users уже есть

**Факт (код):**
```sql
JOIN users u ON u.id = c.author_id
```

В `comment/repo.go` используется `JOIN` (INNER JOIN), а не `LEFT JOIN` как в `forum/repo.go`. Это означает, что комментарии без валидного `author_id` в таблице `users` будут исключены из результата (что может быть желаемым поведением, но отличается от паттерна в forum). Разница не задокументирована: ни как `OQ-*`, ни как обоснованное расхождение.

**Норма:** feature-flow.md § «Traceability Contract»:
> Если sequencing блокируется неопределённостью, план фиксирует её как `OQ-*`, а не прячет в prose.

**Исправление:** явно указать в Discovery Context, что `comment/repo.go` использует INNER JOIN (в отличие от LEFT JOIN в forum). Если при добавлении `u.avatar_url` тип JOIN меняться не будет — зафиксировать это решение явно. Если потребуется изменить на LEFT JOIN — зафиксировать как `OQ-01`.

---

### MEDIUM-2 — H-1: Simplify Review не зафиксирован как STEP-* или CP-*

**Норма:** testing-policy.md § «Simplify Review»:
> Отдельный проход верификации после функционального тестирования. Цель: убедиться, что реализация минимально сложна. Выполняется после прохождения tests, но до closure gate.

feature-flow.md § «Execution → Done»:
> simplify review выполнен: код минимально сложен или complexity обоснована ссылкой на `CON-*`, `FM-*` или `DEC-*`

Ни один из STEP-* и ни один CP-* не упоминает Simplify Review как явный шаг. CP-03 содержит «Unit-тесты pass, E2E спеки созданы и проходят в CI» — но не Simplify Review.

**Исправление:** добавить `STEP-11` (actor: agent, Goal: Simplify Review — проверить что изменения минимально сложны, нет dead code, нет дублирования) или расширить CP-03 явной ссылкой на Simplify Review.

---

### LOW-1 — E-2: переход `delivery_status: planned → in_progress` не зафиксирован в плане

**Цитата (feature.md):**
```yaml
delivery_status: planned
```

**Норма:** feature-flow.md § «Plan Ready → Execution» (HARD STOP):
> `feature.md` → `delivery_status: in_progress`

Это шаг Execution gate, выполняемый до первого коммита. В плане нет ни `STEP-*`, ни `PRE-*`, ни секции, явно фиксирующей этот переход как обязательное действие до начала STEP-01.

**Исправление:** добавить `PRE-04: feature.md delivery_status: in_progress` (или `STEP-00` / примечание в Preconditions) с указанием, что это обязательное действие до первого коммита с кодом согласно feature-flow.md.

---

## Checklist Summary

| Проверка | Статус | Краткое описание |
|---|---|---|
| A-1 STEP-*.Implements → feature.md IDs | OK | Все REQ-*, CON-*, CHK-* существуют |
| A-2 Change Surface → Discovery Context | OK | Все 8 поверхностей из Change Surface отражены |
| A-3 Расхождения → OQ-* | MEDIUM | INNER JOIN в comment/repo не зафиксирован |
| A-4 Нет конфликта с ASM-*/CON-*/NS-* | OK | Конфликтов нет |
| B-1 Пути реальны | OK | Все пути проверены Glob/Read, существуют |
| B-2 OQ-* зафиксированы явно | MEDIUM | «Нет» — но расхождение JOIN не задокументировано |
| B-3 Нет дублирования OQ-*/ER-* | OK | ER-01, ER-02 различны и связаны с STOP-01/STOP-02 |
| C-1 Unit-тесты: local = canonical cmd | BLOCKER | Команда неполная (D-1) |
| C-2 Integration: local = "—" | HIGH | Строка integration отсутствует в Test Strategy |
| C-3 E2E: Required CI = E2E job | OK | E2E (CI) указан |
| C-4 Тип теста указан | HIGH | Integration-тип не указан в Test Strategy |
| C-5 Test assertions конкретны | OK | data-testid assertions + click → modal |
| C-6 E2E seeding описан | HIGH | Seeding для forum posts / news comments не описан |
| C-7 Каждый CHK-* покрыт в Test Strategy | OK | CHK-01..04 все покрыты строками |
| D-1 Команды из ops/development.md дословно | BLOCKER | Go unit cmd неполная |
| D-2 E2E prerequisite: dev + e2e стеки | OK | Оба стека указаны в Environment Contract |
| E-1 status: draft + момент перехода | BLOCKER | status должен быть active для Plan Ready |
| E-2 delivery_status in_progress зафиксирован | LOW | Переход не зафиксирован в плане |
| F-1 Каждый STEP-* атомарен | OK | Атомарность соблюдена |
| F-2 Sequencing корректен | OK | Blocked by цепочки корректны |
| F-3 Отклонения от Layer Stack | OK | Нет отклонений |
| F-4 Рискованные действия → AG-* | OK | Нет рискованных необратимых действий |
| F-5 AG-* не используется для автопилота | OK | AG-* отсутствуют (нет рисков) |
| G-1 ER-* → STOP-* | OK | ER-01→STOP-01, ER-02→STOP-02 |
| G-2 PAR-* write-surface конфликт | OK | PAR-01/02/03 не создают конфликтов |
| H-1 Simplify Review STEP-* или CP-* | MEDIUM | Отсутствует явный шаг |
| H-2 Change Surface UC-*/docs → STEP-* | OK | UC-*/docs не в Change Surface |
