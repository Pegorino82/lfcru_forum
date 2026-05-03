---
gate: "Design Ready → Plan Ready"
artifact: "memory-bank/features/FT-024/implementation-plan.md"
date: "2026-05-03"
outcome: revise
---

# Результат ревью: implementation-plan.md (FT-024)

**Gate:** Design Ready → Plan Ready
**Artifact:** `memory-bank/features/FT-024/implementation-plan.md`
**Date:** 2026-05-03
**Outcome:** revise

---

## Замечания

### 1. BLOCKER — Environment Contract: команда `go test` скопирована неверно (D-1)

**Цитата из плана** (секция `Environment Contract`, строка `go test`):
```
docker run --rm --network lfcru_forum_default -v lfcru_gomod:/go/pkg/mod golang:1.23-alpine go test -tags integration -p 1 ./...
```

**Норма:** `memory-bank/ops/development.md` § «Running Tests» — unit-команда:
```bash
docker run --rm \
  -v "$(pwd)":/app -w /app \
  -v lfcru_gomod:/root/go/pkg/mod \
  golang:1.23-alpine \
  go test ./...
```

**Нарушения:**
1. Volume path неверен: в плане `/go/pkg/mod`, в каноничном документе `/root/go/pkg/mod`.
2. В плане присутствуют флаги `--network lfcru_forum_default` и `-tags integration -p 1` — это команда integration-тестов, а не unit. В Environment Contract она названа просто `go test`.
3. Отсутствует `-v "$(pwd)":/app -w /app`.

**Требуемое исправление:** скопировать дословно из `ops/development.md`. Если контракт описывает обе команды (unit и integration) — указать их раздельно, каждую verbatim.

---

### 2. BLOCKER — Test Strategy: integration тесты указаны в колонке "Required local suites" (C-2)

**Цитата из плана** (секция `Test Strategy`, строка `GET /profile/{username}`):
```
Required local suites: go test -tags integration -p 1 ./internal/profile/...
```

**Цитата из плана** (строка `POST /profile/{username}/avatar`):
```
Required local suites: обе suites
```
(где "обе suites" = integration + Playwright E2E)

**Норма:** `memory-bank/engineering/testing-policy.md` § «Project-Specific Conventions → Go-тесты»:
> «Перед handoff агент прогоняет **только unit-тесты** (Docker-командой из раздела Stack выше). Integration-тесты запускаются только в CI.»

**Норма:** `testing-policy.md` § «Stack → Go-тесты»:
> «Причина split: integration-тесты требуют docker-compose-сети; CI гарантирует правильное окружение.»

**Требуемое исправление:** в колонке `Required local suites` для integration-строк поставить `—` (integration не запускается локально). Перенести integration команду в `Required CI suites` = `Go Tests job`.

---

### 3. HIGH — E2E prerequisite: dev-stack не указан как обязательный шаг перед e2e-stack (D-2)

**Цитата из плана** (секция `Environment Contract`, строка `e2e`):
```
docker compose -f docker-compose.e2e.yml up -d + npx playwright test
```

**Норма:** `memory-bank/ops/development.md` § «E2E Tests (Playwright)»:
```bash
# Шаг 1: поднять dev-стек (postgres + app на 8080), если ещё не запущен
docker compose -f docker-compose.dev.yml up -d

# Шаг 2: поднять e2e-контейнер (app на 8081 → lfcru_test)
docker compose -f docker-compose.e2e.yml up -d
```
> dev-stack — обязательный prerequisite для e2e-контейнера.

**Требуемое исправление:** в строке `e2e` Environment Contract явно добавить оба шага в правильном порядке: сначала `docker compose -f docker-compose.dev.yml up -d`, затем `docker compose -f docker-compose.e2e.yml up -d`.

---

### 4. MEDIUM — Discovery Context: `RequireAuth`/`LoadSession` указаны в неверном пакете (B-1)

**Цитата из плана** (секция `Discovery Context`, строка про middleware):
```
internal/middleware/ | LoadSession (cookie → user в ctx), RequireAuth | ...
```

**Факт:** `LoadSession` и `RequireAuth` определены в `internal/auth/middleware.go` (пакет `auth`), а не в `internal/middleware/`. Директория `internal/middleware/` содержит только `csrf.go`.

**PRE-04** в плане:
```
ASM-04 | RequireAuth middleware существует в internal/middleware/
```
— также неверно: `RequireAuth` находится в `internal/auth/`.

**Требуемое исправление:** в Discovery Context и PRE-04 заменить `internal/middleware/` на `internal/auth/middleware.go`.

---

### 5. MEDIUM — Lifecycle: нет явного STEP-00 или PRE-* для перевода статусов до первого коммита (E-1, E-2)

**Цитата из плана:** таблица `Preconditions` и таблица `Порядок работ` — отсутствует явный шаг/precondition для:
- `implementation-plan.md → status: active`
- `feature.md → delivery_status: in_progress`

**Норма:** `memory-bank/flows/feature-flow.md` § «Plan Ready → Execution»:
> **⛔ HARD STOP.** Все пункты ниже выполняются **до первого коммита с кодом**:
> - `feature.md → delivery_status: in_progress`
> - `implementation-plan.md → status: active`

**Норма:** `feature-flow.md` — evaluator check E-1:
> Момент перевода в `active` должен быть явно зафиксирован в плане — в `STEP-00`, `PRE-*` или преамбуле.

**Требуемое исправление:** добавить `STEP-00` (или отдельные PRE-*) с явным указанием: перед первым коммитом обновить `feature.md → delivery_status: in_progress` и `implementation-plan.md → status: active`.

---

### 6. MEDIUM — Test assertions: "данные" без указания конкретного контекста (C-5)

**Цитата из плана** (секция `Test Strategy`, строка `GET /profile/{username}`, колонка `Planned automated coverage`):
```
Go integration test: 200 + данные, 404 для несуществующего
```

**Норма:** `testing-policy.md` — «Sufficient Coverage»:
> Покрыты новые или измененные contracts, события, schema или integration boundaries. Scenario- и contract-level coverage — не line coverage.

"Данные" — слишком абстрактно. Не указано, что именно проверяется (например: HTTP-body содержит `username`, `avatar_url`, поля профиля; HTTP-статус 404 для несуществующего пользователя).

**Требуемое исправление:** уточнить assertion: например, «200 + body содержит username, AvatarURL, дату регистрации; 404 если user не найден».

---

### 7. MEDIUM — E2E seeding: не описан для тестовых данных профиля (C-6)

**Цитата из плана** (секция `Test Strategy`, строки Playwright):
Нет описания: откуда берётся `{username}` в тестах `e2e/profile.spec.ts`, как создаётся fixture пользователя, как выполняется teardown.

**Норма:** `memory-bank/engineering/testing-policy.md` § «E2E-тесты (Playwright)»:
> «Тестовые данные вставляются с фиксированными ID через `OVERRIDING SYSTEM VALUE`; teardown чистит их по тому же ID»
> «Глобальный seed/teardown — `e2e/global-setup.ts` / `e2e/global-teardown.ts`»

**Требуемое исправление:** добавить в Test Strategy или в STEP-09 описание seeding: fixture пользователя создаётся в `e2e/global-setup.ts` с фиксированным ID/username, teardown — через тот же механизм.

---

### 8. MEDIUM — ER-03 без STOP-* и без escalation threshold (G-1)

**Цитата из плана** (секция `Execution Risks`, строка `ER-03`):
```
ER-03 | avatarColor возвращает разные цвета при изменении палитры → fallback нестабилен | ... | Зафиксировать палитру как const; unit-тест на детерминизм | тест в STEP-08
```

Нет соответствующего `STOP-03`... (имеющийся `STOP-03` ссылается на `FM-03`, не на `ER-03`).

**Норма:** `memory-bank/engineering/autonomy-boundaries.md` § «Правило эскалации»:
> «Если замечания не уменьшаются после 2–3 итераций — проблема upstream.»

Каждый `ER-*` должен иметь `STOP-*` или escalation threshold (feature-flow.md G-1 check).

**Требуемое исправление:** добавить STOP-* для ER-03: trigger — unit-тест на детерминизм падает после N итераций исправлений → остановиться и эскалировать.

---

### 9. MEDIUM — Отсутствует STEP-* для Simplify Review (H-1)

**Цитата из плана:** в таблице `Порядок работ` нет шага Simplify Review. В секции `Готово для приёмки` упоминается финальный verify, но нет явного STEP-* после прохождения тестов.

**Норма:** `memory-bank/engineering/testing-policy.md` § «Simplify Review»:
> «Отдельный проход верификации после функционального тестирования. Цель: убедиться, что реализация минимально сложна. Выполняется после прохождения tests, но до closure gate.»

**Норма:** `memory-bank/flows/feature-flow.md` § «Execution → Done»:
> «simplify review выполнен: код минимально сложен или complexity обоснована ссылкой на CON-*, FM-* или DEC-*»

**Требуемое исправление:** добавить `STEP-11` (или `CP-*`) явно для Simplify Review — после STEP-09/CP-03, до STEP-10 (PR push/CI).

---

## Итог

**Outcome: revise**

Блокеры (3 пункта):
1. Environment Contract: неверная docker-команда (volume path и флаги).
2. Test Strategy: integration тесты в `Required local suites` — нарушение testing-policy.md.
3. (производное от 1+2) Environment Contract нарушает "Не изобретать docker run вручную".

Высокие (1 пункт):
4. E2E prerequisite: dev-stack не указан как обязательный шаг 1.

Средние (5 пунктов):
5. Discovery Context / PRE-04: неверный путь к RequireAuth middleware.
6. Lifecycle: нет STEP-00 для HARD STOP статусных изменений.
7. Test assertions: недостаточно конкретны для integration row.
8. E2E seeding не описан.
9. ER-03 без STOP-*.
10. Отсутствует STEP-* для Simplify Review.
