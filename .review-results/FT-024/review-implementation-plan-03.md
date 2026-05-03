---
gate: "Design Ready → Plan Ready"
artifact: "memory-bank/features/FT-024/implementation-plan.md"
date: "2026-05-03"
iteration: 3
outcome: accept
---

# Результат ревью: implementation-plan.md (FT-024) — итерация 3 (self-check)

**Gate:** Design Ready → Plan Ready
**Artifact:** `memory-bank/features/FT-024/implementation-plan.md`
**Date:** 2026-05-03
**Iteration:** 3 (self-check после 2 revise-итераций)
**Outcome:** accept

---

## Статус исправлений из review-02

Оба средних замечания из итерации 2 устранены:

- **D-1 (MEDIUM):** Test Strategy — FuncMap строка, `Required local suites` теперь содержит каноничную команду из `ops/development.md`: `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...` ✓
- **G-1 (MEDIUM):** ER-04 покрыт добавленным `STOP-05` с trigger (`npx playwright test` connection refused), immediate action (проверка статуса обоих стеков) и escalation threshold (после 1 перезапуска — эскалировать человеку) ✓

---

## Полный проход по чеклисту

### A. Связь с feature.md

- **A-1** OK — все STEP-*.Implements ссылаются на существующие ID feature.md (REQ-*, CON-*, CTR-*, FM-*, SC-*, EC-*, CHK-*, ASM-*)
- **A-2** OK — все файлы из Change Surface отражены в Discovery Context или STEP-*; включая `cmd/forum/main.go`, `internal/config/config.go`, `templates/layouts/base.html`, inline стили (охвачены STEP-06)
- **A-3** OK — OQ-01 явно зафиксирован в Open Questions; других расхождений нет
- **A-4** OK — нет конфликта с ASM-*/CON-*/NS-*; NS-04 (удаление аватара) не реализуется

### B. Discovery Context

- **B-1** OK — все пути проверены через Glob/Read и существуют в worktree:
  - `internal/user/model.go`, `internal/user/repo.go` ✓
  - `internal/tmpl/renderer.go` ✓
  - `internal/admin/images_handler.go`, `internal/admin/image_service.go` ✓
  - `internal/config/config.go` ✓
  - `internal/forum/repo.go`, `internal/comment/repo.go` ✓
  - `internal/auth/middleware.go` ✓
  - `cmd/forum/main.go` ✓
  - `templates/layouts/base.html`, `templates/forum/topic.html`, `templates/forum/index.html`, `templates/forum/section.html`, `templates/news/article.html` ✓
  - `migrations/012_idx_posts_topic_id.sql` ✓
- **B-2** OK — OQ-01 явно в Open Questions
- **B-3** OK — нет дублирования OQ-* и ER-* без связи

### C. Test Strategy

- **C-1** OK — команда unit-тестов в FuncMap строке (`Required local suites`) совпадает символ-в-символ с `ops/development.md`: `docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...`
- **C-2** OK — integration-поверхности (`GET /profile/{username}`, `POST /profile/{username}/avatar`) имеют `Required local suites = —`; только CI
- **C-3** OK — E2E-поверхности указывают на `E2E job` в Required CI suites
- **C-4** OK — тип теста (unit/integration/E2E) явен в описании planned coverage для каждой строки
- **C-5** OK — assertions конкретны: struct fields (`User.Username`, `PostCount`, `CommentCount`), HTTP body (`"unsupported format"`), DOM selectors (`[data-testid="profile-modal"]`, `.nav-avatar`, `.avatar-initials`, `not.toBeVisible()`)
- **C-6** OK — E2E seeding: `e2e/global-setup.ts` via `POST /register`; teardown via `e2e/global-teardown.ts` с `lfcru_test`
- **C-7** OK — CHK-01 (все SC-*, NEG-04,05,06, EC-01..05,07) и CHK-02 (EC-06, NEG-01, NEG-02) покрыты в Test Strategy и STEP-08/09

### D. Environment Contract

- **D-1** OK — unit-команда совпадает дословно с `ops/development.md`; integration-команда совпадает дословно
- **D-2** OK — E2E prerequisite: Шаг 1 (dev-stack) + Шаг 2 (e2e-stack) оба присутствуют в правильном порядке

### E. Lifecycle

- **E-1** OK — `status: draft` корректен во время ревью
- **E-2** OK — `delivery_status: in_progress` оформлен как STEP-00 (HARD STOP, human confirm)

### F. STEP-* качество

- **F-1** OK — каждый STEP атомарен, имеет единственную цель и конкретные touchpoints
- **F-2** OK — sequencing логичен; PAR-01 (STEP-06 || STEP-08 после STEP-05) корректен; PAR-02 (STEP-07 зависит от STEP-06) правильно не параллелен
- **F-3** OK — отклонений от Layer Stack нет; profile следует handler→service→repo паттерну
- **F-4** OK — AG-01 для merge PR; STEP-00 требует `human confirm`
- **F-5** OK — AG-01 одобряется человеком; Playwright не стоит на автопилоте

### G. ER-* и STOP-*

- **G-1** OK — ER-01→STOP-01, ER-02→STOP-02, ER-03→STOP-04, ER-04→STOP-05; FM-03→STOP-03 ✓
- **G-2** OK — PAR-01: STEP-06 (шаблоны) и STEP-08 (тесты) — разные файлы, write-conflict невозможен ✓

### H. Обязательные STEP

- **H-1** OK — STEP-10 = Simplify Review присутствует
- **H-2** N/A — создание UC-* является требованием Done-gate (feature-flow.md §11), а не Plan Ready; на данном этапе не блокирует

---

## Итог

**Outcome: accept**

Все блокеры и высокие замечания отсутствуют. Оба средних замечания из review-02 устранены корректно. Discovery Context полный и верифицирован. Test Strategy соответствует testing-policy.md и ops/development.md. Environment Contract выверен дословно. Stop Conditions покрывают все ER-*. Simplify Review STEP присутствует.

EVID-06 добавлен в `memory-bank/features/FT-024/feature.md`.
