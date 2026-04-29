# Ревью implementation-plan.md (FT-023) vs. memory-bank

**Дата:** 2026-04-29
**Файл:** `memory-bank/features/FT-023/implementation-plan.md`
**Сверялось с:** `architecture.md`, `testing-policy.md`, `coding-style.md`, `autonomy-boundaries.md`, `development.md`, `feature-flow.md`, `feature.md`

---

## БЛОКЕРЫ

### B-01 — `status: draft` вместо `active`

**Источник:** `feature-flow.md` § "Design Ready → Plan Ready"

Gate Plan Ready требует: `implementation-plan.md → status: active`. Фронтматтер плана содержит `status: draft` — нарушение перехода в Plan Ready. Все ревью пройдены (судя по git log), статус необходимо обновить.

---

### B-02 — Отсутствует шаг Simplify Review

**Источник:** `testing-policy.md` § "Simplify Review", `feature-flow.md` § "Execution → Done"

Policy требует: *"Выполняется после прохождения tests, но до closure gate."* В плане нет ни STEP, ни CP для simplify review. Execution → Done gate (`feature-flow.md`) явно включает: `simplify review выполнен`. Нужно добавить STEP (или явный CP) между CP-03/CP-04 и closure.

---

### B-03 — Отсутствует шаг обновления UC-001

**Источник:** `feature-flow.md` § Package Rules п. 11, § "Execution → Done"

`feature.md` Change Surface явно включает:
> `memory-bank/use-cases/UC-001-article-publishing.md` | docs | Добавить FT-023 в `Implemented by` — до closure gate

Closure gate feature-flow.md требует:
> если feature добавляет новый stable flow или materially changes существующий, соответствующий UC-* обновлён

Ни STEP, ни CP для этого в плане нет.

---

## ВЫСОКИЙ

### H-01 — AG-01 конфликтует с `autonomy-boundaries.md`

**Источник:** `autonomy-boundaries.md` § "Автопилот", § "UI-верификация"; `feature-flow.md` § "Boundary Rules" п. 9

`autonomy-boundaries.md`:
> UI-верификация через Playwright — **автопилот** (делай без подтверждения). Evidence: скриншот + вывод Playwright-теста прикрепляются как EVID-* в feature.md.

AG-01 ставит blocking human gate ("дать ок перед STEP-08") именно для проверки TipTap в браузере. Это не рискованное/необратимое действие (`feature-flow.md` § "Boundary Rules" п. 9 — AG-* только для таких). Агент обязан выполнить UI-верификацию автономно через Playwright. AG-01 либо переформулировать как автономный CP (агент открывает Playwright, проверяет консоль и инициализацию), либо исключить.

---

### H-02 — Integration tests в Environment Contract без предупреждения агенту

**Источник:** `testing-policy.md` § "Go-тесты"

Policy:
> Перед handoff агент прогоняет **только unit-тесты**. Integration-тесты запускаются только в CI.
> ⛔ Не изобретать `docker run` вручную.

Environment Contract показывает integration test команду без пометки, что агент не запускает её локально. Без явного disclaimер-а агент может нарушить policy.

---

### H-03 — E2E Environment Contract неполный

**Источник:** `development.md` § "E2E Tests"

`development.md` требует два шага:
```bash
# Шаг 1: поднять dev-стек (postgres)
docker compose -f docker-compose.dev.yml up -d

# Шаг 2: поднять e2e-контейнер
docker compose -f docker-compose.e2e.yml up -d
```

Plan Environment Contract указывает только `docker compose -f docker-compose.e2e.yml up -d`. dev-stack (postgres) — обязательный prerequisite для e2e-контейнера. Без него e2e упадёт.

---

## НИЗКИЙ / ИНФОРМАЦИОННЫЙ

### L-01 — `templates/news/article.html` исключён из плана без явной трассировки

**Источник:** `feature.md` Change Surface, `feature-flow.md` Boundary Rules

`feature.md` Change Surface включает этот файл как `code`-изменение ("рендеринг через safeHTML"). STEP-03 говорит "шаблон изменений не требует". Отклонение объяснено в OQ-07, но само OQ-07 не содержит явной ссылки на Change Surface feature.md — читатель плана не поймёт, почему changed surface из feature.md пропущен.

**Правка:** в OQ-07 добавить ссылку: *"feature.md Change Surface включает templates/news/article.html — данный план от него отклоняется: safeHTML реализуется через template.HTML() в handler, шаблон не меняется."*

---

### L-02 — OQ-07 не ссылается на `coding-style.md`

**Источник:** `coding-style.md` § "Templates"

`coding-style.md`:
> Не использовать `template.HTML()` без явной проверки.

OQ-07 обосновывает `template.HTML(article.Content)` паттерном кодовой базы, но не упоминает это ограничение и не объясняет явно, почему оно не нарушено (sanitize при записи в STEP-02 делает content уже safe). Обоснование верно, но источник ограничения не указан.

---

### L-03 — `$(pwd)` без кавычек в go build команде

**Источник:** `development.md` § "Running Tests"

```bash
# В плане:
-v $(pwd):/app

# В development.md (canonical):
-v "$(pwd)":/app
```

Пути с пробелами приведут к ошибке. Несоответствие canonical команде из development.md.

---

### L-04 — `feature.md → delivery_status: in_progress` не оформлен как STEP/PRE

**Источник:** `feature-flow.md` § "Plan Ready → Execution" (HARD STOP)

В разделе "Готово для приёмки" написано: *"перед началом execution"*. Но в STEP-ах нет шага "обновить delivery_status". По `feature-flow.md` это HARD STOP до первого коммита с кодом — нужно добавить PRE или STEP-00.

---

## Итого

| Severity | ID | Коротко |
|---|---|---|
| БЛОКЕР | B-01 | `status: draft` → должен быть `active` |
| БЛОКЕР | B-02 | Нет Simplify Review шага/CP |
| БЛОКЕР | B-03 | Нет шага обновления UC-001 |
| ВЫСОКИЙ | H-01 | AG-01 — human gate вместо Playwright автопилота |
| ВЫСОКИЙ | H-02 | Integration test команда без disclaimер агенту |
| ВЫСОКИЙ | H-03 | E2E contract пропускает dev-stack prerequisite |
| НИЗКИЙ | L-01 | Отклонение от Change Surface не трассировано в OQ-07 |
| НИЗКИЙ | L-02 | OQ-07 не ссылается на coding-style.md ограничение |
| НИЗКИЙ | L-03 | `$(pwd)` без кавычек в go build |
| НИЗКИЙ | L-04 | `delivery_status: in_progress` не оформлен как STEP/PRE |
