# Memory-Bank Review 3 — Консистентность и непротиворечивость

Дата: 2026-04-11
Охват: все файлы `memory-bank/` (35 документов)

---

## HIGH — Противоречие или неоднозначность

### 1. `dna/lifecycle.md` — правила 4 и 5 читаются как противоречие

**Файл:** `memory-bank/dna/lifecycle.md`

Правило 4: "Расхождение внутри authoritative set **устраняется сразу**."
Правило 5: "Агент, обнаруживший расхождение, **сообщает человеку**. Самостоятельное исправление — только если текущая задача явно требует."

На одном уровне нумерации "устраняется сразу" противоречит "сообщает человеку". По смыслу они комплементарны (4 — принцип системы, 5 — правило поведения агента), но формулировка допускает двойное прочтение.

**Рекомендация:** явно разграничить в тексте: правило 4 — системный принцип, правило 5 — поведение агента.

---

### 2. `governance.md` vs `frontmatter.md` — разные формулировки обязательности `derived_from`

**Файлы:** `memory-bank/dna/governance.md`, `memory-bank/dna/frontmatter.md`

`governance.md`: "Для каждого `active` non-root документа `derived_from` **обязательно**."
`frontmatter.md` schema: `derived_from` — "условно обязательное" с условием "Есть upstream-документ".

Формально совместимы (у всех non-root документов есть upstream), но второе определение теоретически позволяет считать поле необязательным, если "посчитать", что upstream нет.

**Рекомендация:** привести формулировки к одной: либо добавить в `frontmatter.md` явное уточнение, что у всех `active` non-root документов upstream всегда есть, либо дать в `governance.md` отсылку на `frontmatter.md` как на canonical schema.

---

## MEDIUM — Пробелы в навигации

### 3. `CLAUDE.md` не упоминает `engineering/git-workflow.md`

**Файл:** `CLAUDE.md` (таблица маршрутизации)

В таблице маршрутизации нет строки для `memory-bank/engineering/git-workflow.md`. Документ существует и указан в `engineering/README.md`, но попасть на него через CLAUDE.md нельзя.

**Рекомендация:** добавить строку в таблицу CLAUDE.md, например:

```
| Git workflow, коммиты, PR | `memory-bank/engineering/git-workflow.md` |
```

---

### 4. Все файлы в `dna/` не имеют поля `title` в frontmatter

**Файлы:** `dna/principles.md`, `dna/governance.md`, `dna/lifecycle.md`, `dna/frontmatter.md`, `dna/cross-references.md`, `dna/README.md`

Все остальные секции (domain, engineering, ops, flows, features, prd, use-cases, adr) имеют `title` в frontmatter. `title` не входит в required schema по `frontmatter.md`, но отсутствие исключительно в `dna/` нарушает паттерн остальных документов.

**Рекомендация:** добавить `title` в frontmatter всех шести файлов `dna/` для единообразия.

---

## LOW — Стейтменты из шаблона, не актуальные для проекта

### 5. `domain/README.md` — `frontend.md` описан как "шаблон описания"

**Файл:** `memory-bank/domain/README.md`, строка описания `frontend.md`

"[Frontend](frontend.md) — **шаблон** описания UI-поверхностей, design system и i18n-слоя."

`frontend.md` — это уже реальный canonical-документ проекта, не шаблон.

**Рекомендация:** убрать слово "шаблон", например: "Описание UI-поверхностей, interaction patterns и правил HTMX/Alpine.js."

---

### 6. `features/README.md` — фраза "шаблонный репозиторий"

**Файл:** `memory-bank/features/README.md`

"В **шаблонном репозитории** этот каталог может быть пустым. Это нормально."

Memory-bank уже развёрнут в конкретном проекте; фраза — leftover из времён template-репозитория.

**Рекомендация:** заменить на нейтральную формулировку, не привязанную к "шаблонному репозиторию".

---

### 7. `flows/README.md` — "reusable process-layer для шаблона"

**Файл:** `memory-bank/flows/README.md`, body документа

Формулировка "reusable process-layer **для шаблона**" — след template-origin, не отражает текущий статус как проектного документа.

**Рекомендация:** заменить на нейтральную формулировку без слова "шаблона".

---

### 8. `ops/README.md` — `purpose` сформулирован как будущее действие

**Файл:** `memory-bank/ops/README.md`, поле `purpose` во frontmatter

"Читать при **адаптации** dev/prod workflow... **под проект**" — подразумевает, что адаптация ещё не выполнена, хотя ops-документы уже наполнены реальными данными.

**Рекомендация:** переформулировать purpose в настоящем времени — "Читать при работе с dev/prod workflow, релизами, конфигурацией и runbooks."

---

## Что чисто

- Dependency tree (`derived_from`) корректен, циклов нет
- `testing-policy.md` и `development.md` описывают одни и те же команды без расхождений
- `config.md` и `architecture.md` описывают конфигурацию согласованно
- `feature-flow.md` и `testing-policy.md` корректно разделяют ownership (нет SSoT-дублей)
- Все индексы (`README.md`) покрывают реальные документы в своих каталогах
- `lifecycle.md` sync checklist соответствует требованиям `governance.md` и `frontmatter.md`
- `workflows.md` и `feature-flow.md` не дублируют определения lifecycle gates

---

## Итог

| Приоритет | Кол-во | Файлы |
|---|---|---|
| HIGH | 2 | `dna/lifecycle.md`, `dna/governance.md` + `dna/frontmatter.md` |
| MEDIUM | 2 | `CLAUDE.md`, `dna/` (6 файлов) |
| LOW | 4 | `domain/README.md`, `features/README.md`, `flows/README.md`, `ops/README.md` |
