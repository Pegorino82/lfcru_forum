# Memory Bank Review — Consistency & Contradictions

Date: 2026-04-10

---

## CRITICAL

### 1. `doc_kind: engineering` в ops/runbooks/README.md
**Файл:** `memory-bank/ops/runbooks/README.md`
Frontmatter содержит `doc_kind: engineering`, хотя файл живёт в `ops/` и описывает операционные runbooks.
**Должно быть:** `doc_kind: ops`

### 2. Тестовые правила размножены без синхронизации
**Файлы:** `engineering/testing-policy.md` (lines 47–53) и `flows/feature-flow.md` (lines 131–139)
Оба файла описывают "sufficient coverage" с разными формулировками. `feature-flow.md` сам признаёт проблему ("необходимо обновлять синхронно"), но признание — не решение. SSoT нарушен.

---

## HIGH

### 3. Модули `forum` и `content` — противоречие статуса
**Файлы:** `domain/architecture.md` (lines 23–24) vs `domain/problem.md` (lines 30–33)
В `architecture.md` модули отмечены как `*(планируется)*`. В `problem.md` они описаны как часть существующего продукта. Судя по git-истории, форум уже реализован.
**Нужно:** убрать `*(планируется)*` или выровнять статусы.

### 4. CI не настроен vs. CI требуется в implementation-plan
**Файлы:** `engineering/testing-policy.md:44` и `flows/feature-flow.md:98`
`testing-policy.md` говорит "CI: не настроен (задача backlog)". `feature-flow.md` требует "required CI suites" в implementation-plan.
**Нужно:** либо убрать требование CI из feature-flow, либо обновить статус в testing-policy.

### 5. `derived_from` — амбивалентность правила
**Файлы:** `dna/lifecycle.md` (line 25) vs шаблоны в `flows/templates/`
`lifecycle.md` требует `derived_from` для `active` non-root документов. Шаблоны показывают `status: draft` без `derived_from` — допустимо ли это нигде явно не написано.
**Нужно:** добавить явное правило: `derived_from` не обязателен для `status: draft`.

---

## MEDIUM — Дублирование / SSoT нарушен

| # | Проблема | Файлы-участники |
|---|---|---|
| 6 | `DATABASE_URL` и `APP_PORT` описаны дважды | `ops/config.md` + `ops/development.md` |
| 7 | Stack definition размножен (нет canonical view) | `domain/architecture.md` + `domain/frontend.md` |
| 8 | CSRF упомянут в трёх местах без canonical источника | `domain/problem.md`, `domain/frontend.md` — нет в `architecture.md` |
| 9 | `doc_kind` / `doc_function` рекомендованы, но отсутствуют | `domain/architecture.md`, `domain/frontend.md` |
| 10 | `features/`, `prd/`, `use-cases/`, `adr/` — полностью пустые | Governance описывает структуру для несуществующих документов |

---

## LOW — Косметика

| # | Проблема | Файл |
|---|---|---|
| 11 | Нет ссылки на `migrations/` в архитектурных документах | `domain/architecture.md` |
| 12 | Inconsistent markdown table formatting (`| --- |` vs `|---|`) | Несколько файлов |
| 13 | Sentinel errors перечислены как паттерн, но нет canonical списка | `domain/architecture.md:59`, `engineering/coding-style.md:17` |

---

## Итого

| Severity | Кол-во |
|---|---|
| CRITICAL | 2 |
| HIGH | 3 |
| MEDIUM | 5 |
| LOW | 3 |

**Главный вывод:** SSoT нарушен в нескольких местах (тест-правила, стек, env vars). Статус модулей `forum`/`content` в `architecture.md` не соответствует реальности. Governance описывает структуру документов, которых ещё нет.
