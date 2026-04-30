# Memory-Bank Review 4 — Консистентность и непротиворечивость

Дата: 2026-04-11
Охват: все файлы `memory-bank/` (35+ документов)
База: review-3 как отправная точка; фиксирует что исправлено, что осталось, что появилось нового.

---

## Что исправлено с review-3

| # | Файл | Проблема | Статус |
|---|---|---|---|
| 1 | `domain/README.md` | `frontend.md` описана как «шаблон описания» | **FIXED** — теперь «Описание UI-поверхностей, interaction patterns и правил HTMX/Alpine.js» |

---

## HIGH — Не исправлено с review-3

### 1. `dna/lifecycle.md` — правила 4 и 5 читаются как противоречие

**Файл:** `memory-bank/dna/lifecycle.md`

Правило 4: «Расхождение внутри authoritative set **устраняется сразу**.»
Правило 5: «Агент, обнаруживший расхождение, **сообщает человеку**. Самостоятельное исправление — только если текущая задача явно требует.»

По смыслу они комплементарны (4 — системный принцип, 5 — поведение агента), но расположены на одном уровне нумерации без явного разграничения ролей. Агент, читая подряд, может воспринять их как взаимоисключающие инструкции.

**Рекомендация:** разбить на два явных блока: «Системный принцип» (правило 4) и «Поведение агента» (правило 5), либо добавить пояснение что 4 описывает цель системы, 5 — операционное поведение агента при её нарушении.

---

### 2. `governance.md` vs `frontmatter.md` — разные формулировки обязательности `derived_from`

**Файлы:** `memory-bank/dna/governance.md`, `memory-bank/dna/frontmatter.md`

`governance.md`: «Для каждого `active` non-root документа `derived_from` **обязательно**.»
`frontmatter.md` schema: `derived_from` — «условно обязательное» с условием «Есть upstream-документ».

Теоретически допускают различное прочтение: `frontmatter.md` позволяет опустить `derived_from`, если «посчитать», что upstream нет. `governance.md` такой лазейки не даёт.

**Рекомендация:** в `frontmatter.md` уточнить: «для `active` non-root документов upstream всегда есть — условие всегда выполняется» или добавить в `governance.md` явную отсылку к `frontmatter.md` как canonical schema этого поля.

---

## MEDIUM — Не исправлено с review-3

### 3. `CLAUDE.md` — нет строки для `engineering/git-workflow.md`

**Файл:** `CLAUDE.md`, таблица маршрутизации

В таблице присутствуют `coding-style.md`, `testing-policy.md`, `autonomy-boundaries.md`, но нет строки для `git-workflow.md`. Документ существует, указан в `engineering/README.md`, однако через `CLAUDE.md` к нему нельзя попасть.

**Рекомендация:** добавить строку:

```
| Git workflow, коммиты, PR | `memory-bank/engineering/git-workflow.md` |
```

---

### 4. Все файлы `dna/` — отсутствует поле `title` во frontmatter

**Файлы:** `dna/principles.md`, `dna/governance.md`, `dna/lifecycle.md`, `dna/frontmatter.md`, `dna/cross-references.md`, `dna/README.md`

Все остальные разделы (`domain/`, `engineering/`, `ops/`, `flows/`, `features/`) имеют `title`. Отсутствие в `dna/` нарушает единообразие навигации агентом (нет человекочитаемого заголовка для идентификации документа без чтения body).

**Рекомендация:** добавить `title:` в frontmatter всех шести файлов `dna/`.

---

## LOW — Не исправлено с review-3

### 5. `features/README.md` — leftover фраза «шаблонный репозиторий»

**Файл:** `memory-bank/features/README.md`

«В **шаблонном репозитории** этот каталог может быть пустым.»

Memory-bank уже развёрнут в конкретном проекте; фраза — остаток времён template-репозитория.

**Рекомендация:** заменить на «Если feature packages пока не созданы, каталог может быть пустым. Это нормально.»

---

### 6. `ops/README.md` — `purpose` сформулирован как будущее действие

**Файл:** `memory-bank/ops/README.md`, поле `purpose`

«Читать при **адаптации** dev/prod workflow... **под проект**» — подразумевает незавершённую адаптацию, хотя ops-документы уже наполнены реальными данными проекта.

**Рекомендация:** переформулировать: «Читать при работе с dev/prod workflow, релизами, конфигурацией и runbooks проекта.»

---

### 7. `flows/README.md` — «reusable process-layer для шаблона»

**Файл:** `memory-bank/flows/README.md`, body

«Каталог `memory-bank/flows/` содержит reusable process-layer **для шаблона**: lifecycle rules...»

«Для шаблона» — след template-origin документа; не отражает текущий статус как проектного документа.

**Рекомендация:** убрать «для шаблона», например: «содержит process-layer проекта: lifecycle rules, taxonomy стабильных идентификаторов и governed templates.»

---

## LOW — Новое (не было в предыдущих review)

### 8. `engineering/testing-policy.md` — `go test ./...` без Docker-обёртки

**Файл:** `memory-bank/engineering/testing-policy.md`, раздел «Project-Specific Conventions», строка 107

«Перед handoff агент прогоняет unit-тесты (`go test ./...`)»

Выше в том же документе (раздел «Stack») canonical команда запуска unit-тестов — через `docker run ... golang:1.23-alpine go test ./...`. Согласно `PCON-04` из `domain/problem.md`, Go-тулчейн на хосте не требуется и не предполагается.

Голый `go test ./...` без Docker-обёртки создаёт двусмысленность: агент может воспринять его как команду для запуска на хосте, что нарушит `PCON-04`.

**Рекомендация:** уточнить формулировку: «прогоняет unit-тесты (Docker-командой из раздела Stack выше)» или добавить явную ссылку на полную команду.

---

## Что чисто

- Dependency tree (`derived_from`) корректен, циклов нет
- `PCON-*` в `domain/problem.md` и соответствующие ограничения в `domain/architecture.md` согласованы
- `testing-policy.md` и `development.md` описывают одни и те же Docker-команды без расхождений (кроме п.8 выше)
- `config.md` и `architecture.md` согласованы в ownership конфигурации
- `feature-flow.md` и `testing-policy.md` корректно разделяют ownership (SSoT не нарушен)
- `workflows.md` и `feature-flow.md` не дублируют определения lifecycle gates
- `ops/runbooks/README.md` — корректный frontmatter и derived_from после изменений
- `feature-flow.md` — корректно разграничивает acceptance-level и execution-level `CHK-*`/`EVID-*`
- Все README-индексы покрывают реальные документы в своих каталогах
- `frontend.md` описание — исправлено (см. «Что исправлено»)

---

## Итог

| Приоритет | Кол-во | Кратко |
|---|---|---|
| HIGH | 2 | `dna/lifecycle.md` правила 4+5; `governance.md`/`frontmatter.md` `derived_from` |
| MEDIUM | 2 | `CLAUDE.md` без git-workflow; `dna/` без `title` |
| LOW | 4 | `features/README.md`; `ops/README.md` purpose; `flows/README.md`; `testing-policy.md` |
| FIXED | 1 | `domain/README.md` frontend описание |
