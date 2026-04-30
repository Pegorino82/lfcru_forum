# Ревью переноса memory-bank-legacy → memory-bank

**Дата:** 2026-04-12
**Ревьюер:** claude-sonnet-4-6
**Область:** `memory-bank/features/FT-001` – `FT-005`
**Источник требований:** `legacy-migration-guide.md`

---

## Итог по фичам

| Фича | Статус | Серьёзность |
|---|---|---|
| FT-001: Basic Authentication | ✅ OK | — |
| FT-002: Base Layout | ⚠️ Проблемы | BLOCKER + WARNING |
| FT-003: Homepage | ℹ️ Замечание | MINOR |
| FT-004: Article Page | ℹ️ Замечание | MINOR |
| FT-005: Forum Structure | ⚠️ Проблема | ISSUE |

---

## FT-001: Basic Authentication

**Вердикт: OK** — перенос выполнен корректно и полностью соответствует требованиям guide.

### Что проверено

- [x] Выбор шаблона: large.md — обоснован (5 REQ, 5 SC, 2 NEG, CON-*) ✓
- [x] `derived_from`: `domain/problem.md` + GitHub Issue #1 ✓
- [x] `status: draft`, `delivery_status: planned` ✓
- [x] `implementation-plan.md` отсутствует ✓
- [x] Желаемый результат → `MET-01` (не в REQ) ✓
- [x] Rate-limit AC → `REQ-4` + `CON-1` (инфра-ограничение отделено от функционального требования) ✓
- [x] Сессия → `REQ-5` + `CON-2` ✓
- [x] Email-only → `NS-1` + `CON-3` + ссылка на ADR-003 ✓
- [x] Traceability: все 5 REQ покрыты SC (REQ-1..5 → SC-1..5) ✓
- [x] NEG-1, NEG-2 присутствуют ✓
- [x] `memory-bank/features/README.md` не обновлён (нет записи в каталоге) ✓

### Замечание (MINOR)

**Problem-секция копирует текст из legacy-brief.** Guide требует давать ссылку, а не переписывать (`"Не копировать текст «Проблема» из brief.md — сослаться на domain/problem.md"`). В feature.md есть ссылка `"См. общий контекст: ../../domain/problem.md"`, но перед ней — параграф с пересказом legacy-brief. Рекомендуется убрать пересказ, оставив только feature-specific контекст и ссылку.

---

## FT-002: Base Layout

**Вердикт: Проблемы** — два независимых нарушения.

### Находка 1 — BLOCKER: неверная метка шаблона в README.md

`README.md` заявляет `"canonical feature spec (short)"`, но это противоречит правилам guide:

> «`short.md` допустим **только если выполнены все условия:**
> — ≤ 1 SC-*»

`feature.md` содержит `SC-1` **и** `SC-2` → нарушено пороговое условие short.md.

Дополнительно guide прямо указывает:

> «Если legacy-brief содержит больше одного AC — скорее всего large.md.»

Legacy-brief FT-002 содержит **3 AC**. Таким образом, фича должна быть классифицирована как **large.md**.

Тело `feature.md` де-факто уже написано по large.md — в нём есть секции `### Solution`, `### Change Surface`, `### Flow`, которые в short.md шаблоне отсутствуют. Содержимое корректно, метка — нет.

**Требуемое исправление:** изменить метку в README.md с `(short)` на `(large)`.

---

### Находка 2 — WARNING: формат Traceability отличается от остальных фич

`feature.md` использует расширенный формат из `large.md`-шаблона:

```
| Requirement ID | Design refs | Acceptance refs | Checks | Evidence IDs |
```

Все остальные фичи (FT-001, FT-003, FT-004, FT-005) следуют упрощённому формату из migration-guide Шаг 5:

```
| REQ | SC | NEG |
```

Оба формата технически допустимы (разные источники — template vs guide-skeleton), но внутрипроектная несогласованность создаёт читаемость-проблему для агентов. Рекомендуется привести к единому стилю.

---

### Что проверено (остальное)

- [x] `derived_from`: `domain/problem.md` + GitHub Issue #2 ✓
- [x] `status: draft`, `delivery_status: planned` ✓
- [x] `implementation-plan.md` отсутствует ✓
- [x] Инфра-ограничение (html/template stdlib) вынесено в `CON-1`, не в REQ ✓
- [x] NS-1, NS-2 соответствуют явным оговоркам legacy-brief ✓
- [x] How-секция заполнена (Solution, Change Surface, Flow) ✓

---

## FT-003: Homepage

**Вердикт: Замечание (MINOR)**

### Находка — MINOR: Problem-секция не содержит явной ссылки на domain/problem.md

В FT-001 Problem-секция заканчивается на:

> `"См. общий контекст: ../../domain/problem.md."`

В FT-003 такой ссылки нет — `domain/problem.md` упомянут только в frontmatter `derived_from`.

Guide:

> «Не копировать текст «Проблема» из brief.md — сослаться на `domain/problem.md`»

Текст Problem-секции — feature-specific (не копирование), это хорошо. Но явная in-body ссылка помогает агентам находить контекст без чтения frontmatter. Рекомендуется добавить одну строку по образцу FT-001.

### Что проверено (остальное)

- [x] Выбор шаблона: large.md ✓ (4 REQ, 7 SC, 3 CON)
- [x] `derived_from`: `domain/problem.md` + GitHub Issue #3 ✓
- [x] `status: draft`, `delivery_status: planned` ✓
- [x] `implementation-plan.md` отсутствует ✓
- [x] Желаемый результат → `MET-01` ✓
- [x] Viewport-ограничение → `CON-1` (не в REQ), покрыто `EC-2` ✓
- [x] SC-7 (viewport) не в Traceability (правильно — он покрывает CON-1 через EC-2, а не REQ) ✓
- [x] NS-1, NS-2, NS-3 соответствуют оговоркам legacy-brief ✓
- [x] Traceability: все 4 REQ покрыты SC ✓

---

## FT-004: Article Page

**Вердикт: Замечание (MINOR)**

### Находка — MINOR: Problem-секция не содержит явной ссылки на domain/problem.md

Аналогично FT-003 — `derived_from` корректен, но in-body ссылки нет. Рекомендуется добавить по образцу FT-001.

### Что проверено (остальное)

- [x] Выбор шаблона: large.md ✓ (3 REQ, ASM-1, CON-1, SC, NEG)
- [x] `ASM-1` (модель News уже есть) — правомерно для large.md ✓
- [x] CON-1 (404 для несуществующего id) — правильно вынесено в CON, не REQ ✓
- [x] `derived_from`: `domain/problem.md` + GitHub Issue #4 ✓
- [x] `status: draft`, `delivery_status: planned` ✓
- [x] `implementation-plan.md` отсутствует ✓
- [x] Желаемый результат → `MET-01` ✓
- [x] SC-2 + NEG-1 трассируются к CON-1 через EC-2 (не в Traceability-таблице — корректно) ✓
- [x] NS-1, NS-2, NS-3 соответствуют явным оговоркам legacy-brief ✓

---

## FT-005: Forum Structure

**Вердикт: Проблема (ISSUE)**

### Находка — ISSUE: REQ-7 не покрыт отдельным SC

В Traceability:

```
| `REQ-7` | `SC-6` | `NEG-2` |
```

Текст SC-6:

> «Администратор создаёт **раздел** → раздел появляется в `/forum`»

SC-6 тестирует только создание раздела (`REQ-6`), но не создание темы (`REQ-7`). Для `REQ-7` ("Модератор/Администратор может создать тему в разделе") отсутствует соответствующий сценарий.

Это нарушает чеклист migration-guide:

> «Каждый REQ-* упомянут минимум в одном SC-* (таблица Traceability заполнена)»

**Требуемое исправление:** добавить `SC-8` (или аналог) — «Администратор создаёт тему в разделе → тема появляется в `/forum/sections/:id`» — и обновить Traceability.

---

### Что проверено (остальное)

- [x] Выбор шаблона: large.md ✓ (7 REQ, 7 SC, 2 NEG, 3 CON, ADR dependency)
- [x] `derived_from`: `domain/problem.md` + GitHub Issue #5 ✓
- [x] `status: draft`, `delivery_status: planned` ✓
- [x] `implementation-plan.md` отсутствует ✓
- [x] Желаемый результат → `MET-01` ✓
- [x] Роли (Модератор/Администратор) → `CON-2` (инфра-ограничение, не REQ) ✓
- [x] Зависимость от FT-003 → `CON-3` (явная межфичевая зависимость) ✓
- [x] NS-1..4 соответствуют оговоркам legacy-brief ✓
- [x] ADR-004 (forum hierarchy) указан в ADR Dependencies ✓
- [x] NEG-1, NEG-2 присутствуют ✓
- [x] `memory-bank/features/README.md` не обновлён (нет записи в каталоге) ✓

---

## Сводная таблица нарушений

| # | FT | Тип | Серьёзность | Описание |
|---|---|---|---|---|
| 1 | FT-002 | Структура | **BLOCKER** | README.md помечает фичу как `short`, хотя имеет 2 SC-* и 3 legacy-AC — должна быть `large` |
| 2 | FT-005 | Traceability | **ISSUE** | REQ-7 (создание темы) не покрыт отдельным SC; SC-6 тестирует только создание раздела |
| 3 | FT-002 | Формат | WARNING | Traceability-таблица использует расширенный large.md-формат, остальные фичи — упрощённый guide-формат; несогласованность |
| 4 | FT-001 | Problem-секция | MINOR | Перед ссылкой на domain/problem.md присутствует пересказ legacy-brief |
| 5 | FT-003 | Problem-секция | MINOR | Нет in-body ссылки на domain/problem.md (есть только в frontmatter derived_from) |
| 6 | FT-004 | Problem-секция | MINOR | Нет in-body ссылки на domain/problem.md (есть только в frontmatter derived_from) |

---

## Общие наблюдения

**Что сделано хорошо:**
- Во всех 5 фичах корректно разделены CON-* (инфра-ограничения) и REQ-* (функциональные требования) — один из ключевых паттернов migration guide.
- `implementation-plan.md` не создан ни в одной фиче — правильно для статуса draft.
- `memory-bank/features/README.md` не обновлён — правильно (фичи не достигли Design Ready).
- ADR-ссылки в FT-001 и FT-005 структурированы через ADR Dependencies-таблицу.
- NS-* во всех фичах содержат только явные оговорки из legacy-brief, без «всего, что не упомянуто».

**Системная проблема:**
- Три фичи (FT-003, FT-004, FT-005) не имеют in-body ссылки на `domain/problem.md` в Problem-секции. FT-001 — единственный, кто следует образцу guide. Рекомендуется стандартизировать: добавлять финальную строку вида `"Общий контекст: [domain/problem.md](../../domain/problem.md)."` в каждую Problem-секцию.
