---
title: "FT-025: Implementation Plan Review — Gate DR→PR, Iteration 2"
doc_kind: eval
doc_function: gate-eval
gate: Design Ready → Plan Ready
feature: FT-025
reviewed: implementation-plan.md
date: 2026-05-05
outcome: accept
evaluator: evaluator agent
iteration: 2
---

# Review: FT-025 Implementation Plan (Gate DR→PR) — Итерация 2

## Outcome

**accept**

`EVID-PR-01: Eval DR→PR — accept. 2026-05-05. evaluator agent`

Все замечания итерации 1 устранены. Новых нарушений не обнаружено.

---

## Проверка устранения замечаний из итерации 1

### BLOCKER-1 (E-1): `status: draft` вместо `active`

**Итерация 1:** BLOCKER — статус `draft` при gate, требующем `active`.

**Итерация 2:** Добавлена явная Lifecycle note (E-1):
> `status: draft` корректен во время ревью. Переход в `status: active` выполняется при подтверждении Plan Ready человеком (до первого коммита с кодом) — см. STEP-00.

STEP-00 фиксирует переход как первое действие агента после подтверждения gate. Согласно feature-flow.md § «Design Ready → Plan Ready»: evaluator оценивает план, а `status: active` устанавливается при принятии gate, а не до его начала. Интерпретация правомерна. **BLOCKER снят.**

---

### BLOCKER-2 (D-1): неполная команда unit-тестов

**Итерация 1:** BLOCKER — команда `docker run golang:1.23-alpine go test ./...` без монтирования исходников и кэша.

**Итерация 2:** Environment Contract и Check-поле STEP-08 заменены на каноническую команду дословно:
```bash
docker run --rm -v "$(pwd)":/app -w /app -v lfcru_gomod:/root/go/pkg/mod golang:1.23-alpine go test ./...
```
Совпадает с ops/development.md § «Юнит-тесты». **BLOCKER снят.**

---

### HIGH-1 (C-6): E2E seeding не описан

**Итерация 1:** HIGH — seeding для CHK-02 (forum posts) и CHK-03 (news comments) не описан.

**Итерация 2:** Добавлена секция «E2E Test Data Contract»:
- CHK-02: `topic_id=9999` из `global-setup.ts`, автор — `e2e_user` с фиксированным ID
- CHK-03: существующий e2e-article или seed с фиксированным `news_id` через `INSERT ... OVERRIDING SYSTEM VALUE`; teardown в `global-teardown.ts`
- CHK-01, CHK-04: авторизованный `e2e_user` через `login()` helper

Паттерн соответствует testing-policy.md. **HIGH снят.**

---

### HIGH-2 (C-2/C-4): Integration-тесты отсутствуют в Test Strategy

**Итерация 1:** HIGH — строка integration-тестов отсутствовала в Test Strategy.

**Итерация 2:** Добавлена строка:
> Go integration (forum/comment repo) | REQ-02, REQ-03, CON-02 | Существующие integration-тесты forum и comment | Существующие integration-тесты должны проходить после Scan-extension | — *(не запускаются локально; только CI)* | Go Tests (CI): `go test -tags integration -p 1 ./internal/...`

Норма C-2: local = "—" соблюдена. Required CI suites указан корректно. **HIGH снят.**

---

### MEDIUM-1 (B-2/A-3): INNER JOIN в comment/repo не зафиксирован как OQ-*

**Итерация 1:** MEDIUM — расхождение INNER JOIN vs LEFT JOIN не задокументировано.

**Итерация 2:** Discovery Context явно фиксирует:
> `internal/comment/repo.go:19` | `ListByNewsID` — **INNER JOIN** users (не LEFT JOIN как в forum) | Разница от forum/repo: comment/repo использует INNER JOIN; поведение идентично для существующих строк — OQ-01

OQ-01 фиксирует вопрос явно и определяет default action: «Оставить INNER JOIN; добавить `u.avatar_url` как nullable через `*string`. Если integration-тест падает — эскалация.» **MEDIUM снят.**

---

### MEDIUM-2 (H-1): Simplify Review не зафиксирован как STEP-* или CP-*

**Итерация 1:** MEDIUM — отсутствовал явный шаг Simplify Review.

**Итерация 2:** Добавлен STEP-10 (actor: agent, Goal: Simplify review — убедиться что изменения минимальны, нет лишних абстракций) и CP-04 (Simplify review пройден). **MEDIUM снят.**

---

### LOW-1 (E-2): delivery_status: in_progress не зафиксирован

**Итерация 1:** LOW — переход `delivery_status: planned → in_progress` не был явно зафиксирован.

**Итерация 2:** STEP-00 включает: «Перевести `implementation-plan.md → status: active` и `feature.md → delivery_status: in_progress`». Переход зафиксирован как первый обязательный шаг до коммитов с кодом. **LOW снят.**

---

## Checklist Summary — Итерация 2

| Проверка | Статус | Краткое описание |
|---|---|---|
| A-1 STEP-*.Implements → feature.md IDs | OK | Все REQ-*, CON-*, CHK-* существуют в feature.md |
| A-2 Change Surface → Discovery Context | OK | Все 8 поверхностей из Change Surface отражены |
| A-3 Расхождения → OQ-* | OK | INNER JOIN зафиксирован в Discovery Context и OQ-01 |
| A-4 Нет конфликта с ASM-*/CON-*/NS-* | OK | Конфликтов нет; ASM-01, ASM-02, CON-01, CON-02 соблюдены |
| B-1 Пути реальны | OK | Все 10 путей Discovery Context проверены — существуют |
| B-2 OQ-* зафиксированы явно | OK | OQ-01 явно зафиксирован с default action |
| B-3 Нет дублирования OQ-*/ER-* | OK | ER-01, ER-02 различны; связаны с STOP-01/STOP-02 |
| C-1 Unit-тесты: local = canonical cmd | OK | Команда дословно из ops/development.md |
| C-2 Integration: local = "—" | OK | Явно указано: не запускаются локально, только CI |
| C-3 E2E: Required CI = E2E job | OK | E2E (CI) указан для всех Playwright строк |
| C-4 Тип теста указан для новых файлов | OK | Integration-строка добавлена в Test Strategy |
| C-5 Test assertions конкретны | OK | data-testid assertions + click → modal + console 0 errors |
| C-6 E2E seeding описан | OK | E2E Test Data Contract описывает seeding для всех CHK-* |
| C-7 Каждый CHK-* покрыт в Test Strategy | OK | CHK-01..04 покрыты строками в Test Strategy |
| D-1 Команды из ops/development.md дословно | OK | Каноническая команда в Environment Contract и STEP-08 |
| D-2 E2E prerequisite: dev + e2e стеки | OK | Оба стека указаны в Environment Contract |
| E-1 status: draft + момент перехода | OK | Lifecycle note объясняет; переход в STEP-00 |
| E-2 delivery_status in_progress зафиксирован | OK | STEP-00 явно фиксирует переход |
| F-1 Каждый STEP-* атомарен | OK | Атомарность соблюдена во всех 12 шагах |
| F-2 Sequencing корректен | OK | Blocked by цепочки корректны; WS-зависимости соблюдены |
| F-3 Отклонения от Layer Stack | OK | Нет отклонений; изменения в model/repo/template |
| F-4 Рискованные действия → AG-* | OK | Нет рискованных необратимых действий |
| F-5 AG-* не используется для автопилота | OK | AG-* отсутствуют (нет рисков) |
| G-1 ER-* → STOP-* | OK | ER-01→STOP-01, ER-02→STOP-02 |
| G-2 PAR-* write-surface конфликт | OK | PAR-01 (разные пакеты), PAR-02 (независимы), PAR-03 (nav отдельно) |
| H-1 Simplify Review STEP-* или CP-* | OK | STEP-10 + CP-04 |
| H-2 UC-*/docs STEP-* при необходимости | OK | UC-*/docs не в Change Surface фичи |
