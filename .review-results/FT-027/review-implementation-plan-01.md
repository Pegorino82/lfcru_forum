# Review: implementation-plan.md (FT-027)

**Gate:** Design Ready -> Plan Ready
**Date:** 2026-05-07
**Reviewer:** evaluator agent
**Verdict:** `revise`

---

## Checklist

### A. Связь с feature.md

| ID | Verdict | Details |
|----|---------|---------|
| A-1 | HIGH | STEP-*.Implements ссылается на `SC-01`..`SC-05`, `NEG-01`, `NEG-02` в STEP-06, но эти идентификаторы принадлежат feature.md (acceptance scenarios), а не plan-level IDs. Колонка `Implements` должна ссылаться на `REQ-*` / `CTR-*`. Формально корректно, но STEP-06.Verifies использует `CHK-01`..`CHK-05` из feature.md, что допустимо. **Однако:** STEP-03 Verifies ссылается на `CHK-01`, `CHK-02`, `CHK-03` — это acceptance-level checks из feature.md. По boundary rule 5 feature-flow.md, plan `CHK-*` — execution-level; plan не должен использовать feature-level `CHK-*` в колонке Verifies без создания собственных execution-level checks. Это не блокер, но замечание. |
| A-2 | OK | Все файлы из Change Surface (`templates/news/list.html`, `internal/news/handler.go`, `templates/forum/index.html`, `templates/forum/section.html`, `internal/news/repo.go`) отражены в Discovery Context и STEP-*. |
| A-3 | OK | Расхождение с `ASM-01` зафиксировано через `OQ-01` (ListPublished возвращает `[]News`, а нужно `[]HomeNewsItem`). |
| A-4 | OK | Нет конфликта с `ASM-02` (960px container сохраняется), `CON-01`, `CON-02`, `NS-*`. |

### B. Discovery Context

| ID | Verdict | Details |
|----|---------|---------|
| B-1 | OK | Все пути реальны: `internal/news/handler.go`, `internal/news/repo.go`, `internal/news/model.go`, `templates/home/index.html`, `templates/news/list.html`, `templates/forum/index.html`, `templates/forum/section.html`, `templates/layouts/base.html` существуют в worktree. |
| B-2 | OK | `OQ-01` зафиксирован явно в отдельной таблице. |
| B-3 | OK | Нет дублирования между `OQ-*` и `ER-*`. `OQ-01` про сигнатуру метода, `ER-01` про неизвестных потребителей — связаны, но не дублируют. |

### C. Test Strategy

| ID | Verdict | Details |
|----|---------|---------|
| C-1 | BLOCKER | **Unit тесты: Required local suites указана неверно.** В таблице Test Strategy для "News repo SQL" указана команда `docker run go test ./internal/news/...`. Canonical команда из ops/development.md для unit-тестов: `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...`. Команда в плане сокращена и не дословна. Кроме того, в Environment Contract (строка 60) указана команда с `--network` и `DATABASE_URL` — это формат integration-тестов, а не unit-тестов. Если тест SQL repo не требует БД (мок), используй unit-команду дословно. Если требует БД — это integration-тест, и тогда он не должен запускаться локально (см. C-2). |
| C-2 | OK | Integration тесты явно не включены в локальный запуск. Однако см. C-1 — Environment Contract содержит команду с `--network` и `DATABASE_URL`, что фактически является integration-тестом. |
| C-3 | OK | E2E Playwright указан как Required CI suites = E2E job. |
| C-4 | MEDIUM | Тип нового test-файла для Go unit test ("News repo SQL") не указан явно — непонятно, будет ли это `repo_test.go` (unit) или `repo_integration_test.go` (integration с build tag). Указать тип и путь файла. |
| C-5 | OK | Test assertions достаточно конкретны: screenshots, element assertions, console error listener, viewport breakpoints. |
| C-6 | OK | E2E тесты не требуют {id}-seeding — используют существующие страницы `/news`, `/forum`, `/forum/sections/X`. |
| C-7 | HIGH | `CHK-06` упоминается в Environment Contract (строка 60), но **не определен ни в feature.md, ни в implementation-plan.md**. В feature.md checks идут от `CHK-01` до `CHK-05`. `CHK-06` — фантомная ссылка. |

### D. Environment Contract

| ID | Verdict | Details |
|----|---------|---------|
| D-1 | BLOCKER | **Команда Go-тестов не дословна.** ops/development.md unit-тесты: `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...`. В плане (строка 60): `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod --network lfcru_forum_default -e DATABASE_URL="..." golang:1.23-alpine go test ./internal/news/...`. Расхождения: (1) добавлены `--network` и `-e DATABASE_URL` — это формат integration-тестов, (2) `go test ./internal/news/...` вместо `go test ./...`, (3) `DATABASE_URL` указан как `"..."` — placeholder. Если это unit-тест — убрать `--network` и `DATABASE_URL`, скопировать дословно. Если integration — не запускать локально. |
| D-2 | BLOCKER | **E2E prerequisite неполон.** В плане (строка 61): `docker compose -f docker-compose.e2e.yml up -d && npx playwright test`. Отсутствует первый шаг: `docker compose -f docker-compose.dev.yml up -d`. Canonical из ops/development.md требует оба шага: сначала dev-stack, затем e2e-stack. Также отсутствуют `npm install` и `npx playwright install chromium`. |

### E. Lifecycle

| ID | Verdict | Details |
|----|---------|---------|
| E-1 | OK | `status: draft` корректен во время ревью. |
| E-2 | BLOCKER | **`feature.md -> delivery_status: in_progress` не оформлен как явный STEP или PRE.** Текущий `delivery_status: planned` в feature.md. По feature-flow.md, переход Plan Ready -> Execution требует `delivery_status: in_progress`. В плане нет STEP или PRE, фиксирующего этот переход. |

### F. Качество STEP-*

| ID | Verdict | Details |
|----|---------|---------|
| F-1 | OK | Каждый STEP атомарен: repo, handler, шаблон новостей, шаблон форума (разделы), шаблон форума (темы), E2E тесты. |
| F-2 | OK | Sequencing корректен: STEP-01 (repo) -> STEP-02 (handler) -> STEP-03 (template) -> STEP-06 (E2E). STEP-04 и STEP-05 независимы. |
| F-3 | OK | Нет отклонений от Layer Stack. Repo -> Handler -> Template — канонический порядок. |
| F-4 | OK | Approval Gates = "Нет" — обоснование корректно: только UI/template изменения, нет рискованных/необратимых действий. |
| F-5 | OK | AG-* не используется — и не нужен (UI-верификация через Playwright = автопилот по autonomy-boundaries.md). |

### G. ER-* и STOP-*

| ID | Verdict | Details |
|----|---------|---------|
| G-1 | OK | `ER-01` имеет `STOP-01`. `ER-02` — визуальный риск с описанной mitigation (placeholder стилизован), не требует STOP-*. |
| G-2 | OK | `PAR-01` — STEP-04/05 и STEP-01..03 работают с разными файлами, нет write-surface конфликта. |

### H. Обязательные STEP

| ID | Verdict | Details |
|----|---------|---------|
| H-1 | MEDIUM | **Нет STEP или CP для Simplify Review.** По testing-policy.md и feature-flow.md (Execution -> Done gate), simplify review обязателен. В плане нет упоминания simplify review ни в STEP-*, ни в Checkpoints. Добавить CP или STEP. |
| H-2 | OK | Change Surface не включает UC-* или docs — дополнительных STEP не требуется. |

---

## Summary of Findings

### BLOCKERs (4)

1. **[C-1 / D-1]** Go-тест в Environment Contract использует формат integration-теста (`--network`, `DATABASE_URL`), но заявлен как unit-тест, запускаемый локально. Команда не дословна из ops/development.md. Решить: если это unit-тест на мок — скопировать unit-команду дословно. Если это integration-тест с реальной БД — убрать из Required local suites (integration-тесты только в CI по testing-policy.md).

2. **[D-2]** E2E prerequisite неполон — отсутствует шаг `docker compose -f docker-compose.dev.yml up -d` и одноразовые шаги `npm install` / `npx playwright install chromium`. Canonical из ops/development.md: сначала dev-stack, затем e2e-stack.

3. **[E-2]** Отсутствует STEP или PRE для перевода `feature.md -> delivery_status: in_progress`. Это обязательное условие перехода Plan Ready -> Execution.

4. **[C-7]** `CHK-06` в Environment Contract (строка 60) не определен ни в feature.md, ни в implementation-plan.md. Фантомная ссылка.

### HIGH (1)

5. **[A-1]** STEP-03..05 Verifies ссылаются на feature-level `CHK-*` (`CHK-01`..`CHK-03`). По boundary rule 5 feature-flow.md, plan использует execution-level checks. Рекомендуется: либо создать plan-level `CHK-*`, либо переместить ссылки в `Evidence IDs`.

### MEDIUM (2)

6. **[C-4]** Тип нового Go test-файла (unit vs integration, путь) не указан.

7. **[H-1]** Нет STEP или CP для Simplify Review.

---

## Verdict: `revise`

Исправить 4 BLOCKERа, рассмотреть HIGH и MEDIUM замечания. После исправления — повторное ревью.
