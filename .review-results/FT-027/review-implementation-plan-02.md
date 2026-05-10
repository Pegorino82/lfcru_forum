# Review: implementation-plan.md (FT-027) — iteration 2

**Gate:** Design Ready -> Plan Ready
**Date:** 2026-05-07
**Reviewer:** evaluator agent
**Verdict:** `revise`

---

## Checklist

### A. Связь с feature.md

| ID | Verdict | Details |
|----|---------|---------|
| A-1 | OK | STEP-*.Implements ссылаются на `REQ-*` и `CTR-*` (STEP-01..05). STEP-06.Implements ссылается на `SC-*` и `NEG-*` — допустимо, это E2E-тест, верифицирующий acceptance scenarios. Plan использует собственные execution-level checks `CHK-E01`..`CHK-E07` (строка 100), не путая с feature-level `CHK-01`..`CHK-05`. Исправлено по сравнению с итерацией 1. |
| A-2 | OK | Все файлы из Change Surface (`templates/news/list.html`, `internal/news/handler.go`, `templates/forum/index.html`, `templates/forum/section.html`, `internal/news/repo.go`) отражены в Discovery Context и STEP-*. |
| A-3 | OK | Расхождение с `ASM-01` зафиксировано через `OQ-01`. |
| A-4 | OK | Нет конфликта с `ASM-02`, `CON-*`, `NS-*`. |

### B. Discovery Context

| ID | Verdict | Details |
|----|---------|---------|
| B-1 | OK | Все пути реальны. Проверено через Glob/Bash: `internal/news/handler.go` (строки 66-101 соответствуют ShowList), `internal/news/repo.go` (строки 44-82 — LatestPublishedForHome, 84-116 — ListPublished), `internal/news/model.go` (строки 38-56 — HomeNewsItem), `templates/home/index.html`, `templates/news/list.html`, `templates/forum/index.html`, `templates/forum/section.html`, `templates/layouts/base.html` (строка 46 — `.container { max-width: 960px }`). |
| B-2 | OK | `OQ-01` зафиксирован явно. |
| B-3 | OK | Нет дублирования. `OQ-01` — вопрос сигнатуры, `ER-01` — риск других потребителей. Связаны, но не дублируют. |

### C. Test Strategy

| ID | Verdict | Details |
|----|---------|---------|
| C-1 | OK | Unit test команда в Environment Contract (строка 63) дословно совпадает с ops/development.md: `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...`. Исправлено по сравнению с итерацией 1. |
| C-2 | OK | Integration-тесты не включены в локальный запуск — только CI. |
| C-3 | OK | E2E — Required CI suites = E2E job. |
| C-4 | OK | Указан тип нового файла: `e2e/sections-design/unified-design.spec.ts` (TypeScript) в STEP-06. |
| C-5 | OK | Assertions конкретны: element assertions, screenshots at 3 viewports, grid column count, console error listener, edge cases (page=0, -1, 999). |
| C-6 | LOW | E2E seeding: план заявляет, что тесты используют существующие страницы `/news`, `/forum`, `/forum/sections/X`. Однако для надежных assertions (карточки с изображениями, анонсами, comment count) может потребоваться seed-данные с заполненными полями. Существующий global-setup.ts создает `E2E_NEWS_ID=9997` и `E2E_SECTION_ID=9999`, но не вставляет `article_images` для cover_image. Рекомендуется: добавить в STEP-06 явную проверку seed-данных и при необходимости расширить global-setup.ts. Не блокер — при отсутствии image рендерится placeholder, assertions остаются валидными. |
| C-7 | OK | Все `CHK-01`..`CHK-05` из feature.md покрыты в Test Strategy. Фантомный `CHK-06` устранен. |

### D. Environment Contract

| ID | Verdict | Details |
|----|---------|---------|
| D-1 | OK | Go unit test команда дословна из ops/development.md. Исправлено. |
| D-2 | OK | E2E prerequisite (строка 64) полон: `docker compose -f docker-compose.dev.yml up -d && docker compose -f docker-compose.e2e.yml up -d && npm install && npx playwright install chromium`. Исправлено. |

### E. Lifecycle

| ID | Verdict | Details |
|----|---------|---------|
| E-1 | OK | `status: draft` корректен. |
| E-2 | OK | Переход `delivery_status: in_progress` оформлен как `PRE-04` (строка 74) и описан в шапке плана (строка 19): "выполняется как PRE-04 до первого коммита с кодом". Исправлено. |

### F. Качество STEP-*

| ID | Verdict | Details |
|----|---------|---------|
| F-1 | OK | Каждый STEP атомарен. |
| F-2 | HIGH | **Sequencing не учитывает существующие тесты.** STEP-01 меняет сигнатуру `ListPublished()`: возвращаемый тип с `[]News` на `[]HomeNewsItem`. Однако в `internal/news/repo_test.go` (integration tests, строки 238-354) есть 4 теста `TestListPublished_*`, которые обращаются к полям `News` (например, `item.Title` в строке 325). При изменении return type на `[]HomeNewsItem` эти тесты **не сломаются** (оба типа имеют поле `Title`), но тесты в `handler_test.go` (также integration) зависят от типа `ListData.Items`, который меняется в STEP-02. **План не упоминает необходимость обновления существующих тестов** — ни repo_test.go, ни handler_test.go не указаны в touchpoints STEP-01 или STEP-02. Поскольку оба файла тестов — integration (build tag `//go:build integration`), они не запускаются локально и проверяются в CI. Однако plan должен зафиксировать, что STEP-01/02 затрагивают эти файлы, чтобы CI не падал. |
| F-3 | OK | Нет отклонений от Layer Stack. |
| F-4 | OK | AG-* не нужны — только UI/template изменения. |
| F-5 | OK | N/A. |

### G. ER-* и STOP-*

| ID | Verdict | Details |
|----|---------|---------|
| G-1 | OK | `ER-01` -> `STOP-01`. `ER-02` -> `STOP-02` (эскалация при 2 итерациях). |
| G-2 | OK | `PAR-01` — разные файлы, нет write-surface конфликта. |

### H. Обязательные STEP

| ID | Verdict | Details |
|----|---------|---------|
| H-1 | OK | STEP-07 — Simplify Review. Исправлено. |
| H-2 | OK | Change Surface не включает UC-* или docs. |

---

## Summary of Findings

### BLOCKER (0)

Все 4 BLOCKERа из итерации 1 исправлены.

### HIGH (1)

1. **[F-2]** STEP-01 и STEP-02 меняют сигнатуру `ListPublished()` и тип `ListData.Items`, но план не упоминает обновление существующих integration-тестов (`repo_test.go`: 4 теста `TestListPublished_*`, `handler_test.go`: тесты `TestShowList_*`). Файлы тестов должны быть указаны в Touchpoints соответствующих STEP и/или добавлен отдельный STEP для обновления тестов. Без этого CI сломается.

   **Цитата (строка 92):**
   > `STEP-01` | agent | `REQ-01`, `CTR-01` | Расширить repo: ListPublished возвращает HomeNewsItem | `internal/news/repo.go`

   **Норма:** feature-flow.md § Boundary Rules, п. 8: "discovery context содержит: relevant paths, local reference patterns..." — тесты являются relevant path. Также testing-policy.md § Core Rules: "Любое изменение поведения... обязано получить automated regression coverage" — существующие тесты должны быть обновлены при изменении контракта.

### LOW (1)

2. **[C-6]** E2E seed-данные могут не содержать `article_images` для cover_image assertion. Рекомендуется зафиксировать в STEP-06 явную проверку seed-данных.

---

## Verdict: `revise`

Один HIGH не позволяет выставить `accept`. Исправление минимальное: добавить `internal/news/repo_test.go` в touchpoints STEP-01 и `internal/news/handler_test.go` в touchpoints STEP-02 (или создать отдельный STEP для обновления тестов). После исправления план готов к accept.
