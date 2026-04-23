---
title: Testing Policy
doc_kind: engineering
doc_function: canonical
purpose: Описывает testing policy репозитория LFC.ru forum — обязательность test case design, требования к automated regression coverage и допустимые manual-only gaps.
derived_from:
  - ../dna/governance.md
  - ../flows/feature-flow.md
status: active
canonical_for:
  - repository_testing_policy
  - feature_test_case_inventory_rules
  - automated_test_requirements
  - sufficient_test_coverage_definition
  - manual_only_verification_exceptions
  - simplify_review_discipline
  - verification_context_separation
must_not_define:
  - feature_acceptance_criteria
  - feature_scope
audience: humans_and_agents
---

# Testing Policy

## Stack

### Go-тесты

- **Framework:** `go test` (stdlib)
- **Data:** тестовая БД `lfcru_test` (postgres); изолирована от `lfcru` — dev-данные не затрагиваются
- **Запуск unit-тестов** (без БД):
  ```bash
  docker run --rm -v "$(pwd)":/app -w /app golang:1.23-alpine go test ./...
  ```
- **Запуск integration-тестов** (нужна запущенная БД из `docker-compose.dev.yml`):
  ```bash
  docker run --rm \
    -v "$(pwd)":/app -w /app \
    --network lfcru_forum_default \
    -e DATABASE_URL="postgres://postgres:postgres@postgres:5432/lfcru_test?sslmode=disable" \
    golang:1.23-alpine \
    go test -tags integration -p 1 ./internal/...
  ```
- **Флаг `-p 1` обязателен** для integration-тестов: каждый пакет вызывает `goose.Up()` независимо, параллельный запуск вызывает race condition
- Тестовая БД создаётся автоматически скриптом `scripts/init-test-db.sql` при первом старте postgres-контейнера

### E2E-тесты (Playwright)

- **Framework:** `@playwright/test` (TypeScript)
- **Data:** та же тестовая БД `lfcru_test`; seed/teardown через `e2e/global-setup.ts` и `e2e/global-teardown.ts`
- **App:** отдельный контейнер `app-e2e` на порту `8081`, указывает на `lfcru_test` (`docker-compose.e2e.yml`)
- **Запуск:**
  ```bash
  # Предварительно: docker compose -f docker-compose.dev.yml up -d
  #                 docker compose -f docker-compose.e2e.yml up -d
  npm install
  npx playwright install chromium
  npx playwright test
  ```
- **Артефакты:** `e2e/test-results/` (скриншоты при падении), `e2e/test-report/` (HTML-отчёт) — в `.gitignore`

### CI

GitHub Actions (`.github/workflows/ci.yml`). Запускается на каждый push и PR.

**Jobs:**

- **Lint** — actionlint, shellcheck, shfmt, markdownlint
- **Go Tests** — unit + integration тесты:
  - postgres запускается как service container (postgres:17-alpine)
  - `go test ./...` — unit
  - `go test -tags integration -p 1 ./internal/...` — integration
  - `DATABASE_URL` указывает на `lfcru_test` в service container
- **E2E (Playwright)** — playwright на раннере:
  - postgres + app-e2e поднимаются через docker compose
  - `.env` создаётся из `.env.example` перед стартом compose
  - `npm ci` + `npx playwright install chromium --with-deps`
  - `npx playwright test`
  - Артефакты (скриншоты, HTML-отчёт) загружаются при падении

Все три job-а обязательны для merge (branch protection на `main`).

## Core Rules

- Любое изменение поведения, которое можно проверить детерминированно, обязано получить automated regression coverage.
- Любой новый или измененный contract обязан получить contract-level automated verification.
- Любой bugfix обязан добавить regression test на воспроизводимый сценарий.
- Required automated tests считаются закрывающими риск только если они проходят локально (и в CI, когда CI настроен).
- Manual-only verify допустим только как явное исключение и не заменяет automated coverage там, где automation реалистична.

## Ownership Split

- Canonical test cases delivery-единицы задаются в `feature.md` через `SC-*`, feature-specific `NEG-*`, `CHK-*` и `EVID-*`.
- `implementation-plan.md` владеет только стратегией исполнения: какие test surfaces будут добавлены или обновлены, какие gaps временно остаются manual-only и почему.

## Feature Flow Expectations

Canonical lifecycle gates живут в [../flows/feature-flow.md](../flows/feature-flow.md):

- к `Design Ready` `feature.md` уже фиксирует test case inventory;
- к `Plan Ready` `implementation-plan.md` содержит `Test Strategy` с planned automated coverage и manual-only gaps;
- к `Done` required tests добавлены, локальные команды зелёные (CI не противоречит, если настроен).

## Что Считается Sufficient Coverage

- Покрыт основной changed behavior и ближайший regression path.
- Покрыты новые или измененные contracts, события, schema или integration boundaries.
- Покрыты критичные failure modes из `FM-*`, bug history или acceptance risks.
- Покрыты feature-specific negative/edge scenarios, если они меняют verdict.
- Процент line coverage сам по себе недостаточен: нужен scenario- и contract-level coverage.

## Когда Manual-Only Допустим

- Сценарий зависит от live infra, внешних систем, hardware, недетерминированной среды или human оценки UI.
- Сценарий зависит от SSE/real-time поведения или визуального рендеринга, которое Playwright не покрывает (анимации, шрифты, пиксельный layout).
- Browser-специфика и HTMX/Alpine.js-взаимодействия — **покрываются Playwright**, не являются основанием для manual-only.
- Для каждого manual-only gap: причина, ручная процедура, owner follow-up.
- Если manual-only gap оставляет без regression protection критичный путь (auth, сессии, CSRF), feature не считается завершённой.

## Simplify Review

Отдельный проход верификации после функционального тестирования. Цель: убедиться, что реализация минимально сложна.

- Выполняется после прохождения tests, но до closure gate.
- Паттерны: premature abstractions, глубокая вложенность, дублирование логики, dead code, overengineering.
- Три похожие строки лучше premature abstraction. Абстракция оправдана только когда она реально уменьшает риск или повтор.

## Verification Context Separation

Разные этапы верификации — отдельные проходы:

1. **Функциональная верификация** — tests проходят, acceptance scenarios покрыты
2. **Simplify review** — код минимально сложен
3. **Acceptance test** — end-to-end по `SC-*`

Для small features допустимо в одной сессии, но simplify review не пропускается.

## Project-Specific Conventions

### Go-тесты
- Unit-тесты (`*_test.go` без build tag) живут рядом с тестируемым пакетом в `internal/`
- Integration-тесты помечаются build tag `//go:build integration` и также живут в пакете рядом с кодом
- Каждый пакет самостоятельно вызывает `goose.Up()` в `TestMain` — setup изолирован
- Моки репозиториев используются в unit-тестах сервисов; integration-тесты обязаны попадать в реальную БД
- Перед handoff агент прогоняет unit-тесты (Docker-командой из раздела Stack выше) и integration-тесты затронутых пакетов

### E2E-тесты (Playwright)
- Спеки (`*.spec.ts`) живут в `e2e/<домен>/` — зеркалируют структуру `internal/<домен>/`
- Глобальный seed/teardown — `e2e/global-setup.ts` / `e2e/global-teardown.ts`
- Тестовые данные вставляются с фиксированными ID через `OVERRIDING SYSTEM VALUE`; teardown чистит их по тому же ID
- Конфиг: `playwright.config.ts` в корне проекта

