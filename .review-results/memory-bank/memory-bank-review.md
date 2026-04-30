# Memory-Bank Review — Агрегированный отчёт

**Дата**: 2026-04-11
**Источники**: review-0 (2026-04-10), review-1..4 (2026-04-11)
**Охват**: все файлы `memory-bank/` + `CLAUDE.md`

---

## Исправлено до этого отчёта

| Файл | Проблема | Коммит |
|---|---|---|
| `ops/runbooks/README.md` | `doc_kind: engineering` → `doc_kind: ops` | 75f6133 |
| `domain/README.md` | `frontend.md` описан как «шаблон описания» | 75f6133 |

---

## Противоречия между review-файлами

| Проблема | review-0 | review-1 | review-2/3/4 | Решение |
|---|---|---|---|---|
| Дублирование секций в `testing-policy` ↔ `feature-flow` | CRITICAL #2 | INFO: намеренная выжимка | чисто | **RESOLVED** — intentional design, документировано явно |
| `go test ./...` без Docker | — | — | review-2: HIGH; review-4: LOW «новое» | **Принять HIGH** — review-4 ошибочно не заметил, что review-2 уже зафиксировал; оперативный риск остаётся |
| Дублирование docker-команд в testing-policy и development.md | — | чисто | review-2: MEDIUM | **LOW** — review-1, 3, 4 считают это нормальным; review-2 — единственный outlier |

---

## HIGH

### H-1. Модули `forum` и `content` отмечены `*(планируется)*`

**Файл:** `memory-bank/domain/architecture.md:23–24`
**Источники:** review-0 HIGH #3, review-2 HIGH #2

Форум реализован (коммиты `fix: replies...`, `feat(005-iter2): HTTP Layer`). Статус устарел.

**Нужно:** убрать `*(планируется)*` у `forum`. Уточнить статус `content`.

---

### H-2. `go test ./...` без Docker-обёртки

**Файл:** `memory-bank/engineering/testing-policy.md:107`
**Источники:** review-2 HIGH #1, review-4 LOW #8

Строка: `«агент прогоняет unit-тесты (go test ./...)»`. Выше в том же документе (раздел Stack) canonical команда — через `docker run ... golang:1.23-alpine go test ./...`. Согласно `PCON-04`, Go на хосте не нужен. Агент может запустить команду напрямую — ошибка.

**Нужно:** заменить на `«(Docker-командой из раздела Stack выше)»` или вставить полную команду.

---

### H-3. `dna/lifecycle.md` — правила 4 и 5 читаются как взаимоисключающие

**Файл:** `memory-bank/dna/lifecycle.md`
**Источники:** review-3 HIGH #1, review-4 HIGH #1

Правило 4: «Расхождение устраняется сразу.» Правило 5: «Агент сообщает человеку, самостоятельное исправление — только если задача явно требует.» По смыслу комплементарны, но на одном уровне без разграничения ролей читаются как противоречие.

**Нужно:** явно разбить: «Системный принцип» (4) vs «Поведение агента» (5).

---

### H-4. `governance.md` vs `frontmatter.md` — разные формулировки обязательности `derived_from`

**Файлы:** `memory-bank/dna/governance.md`, `memory-bank/dna/frontmatter.md`
**Источники:** review-0 HIGH #5, review-3 HIGH #2, review-4 HIGH #2

`governance.md`: «обязательно для active non-root». `frontmatter.md`: «условно обязательное, если есть upstream». Теоретически позволяют опустить поле, если «посчитать», что upstream нет.

**Нужно:** в `frontmatter.md` уточнить: «для active non-root документов upstream всегда есть».

---

## MEDIUM

### M-1. `CLAUDE.md` не содержит строки для `git-workflow.md`

**Файлы:** `CLAUDE.md`, `memory-bank/engineering/git-workflow.md`
**Источники:** review-2 LOW #6, review-3 MEDIUM #3, review-4 MEDIUM #3

Документ существует и описан в `engineering/README.md`, но в навигационной таблице `CLAUDE.md` отсутствует — агент не найдёт его при первичной ориентации.

**Нужно:** добавить строку `| Git workflow, коммиты, PR | memory-bank/engineering/git-workflow.md |`.

---

### M-2. Все файлы `dna/` не имеют поля `title` во frontmatter

**Файлы:** `dna/principles.md`, `dna/governance.md`, `dna/lifecycle.md`, `dna/frontmatter.md`, `dna/cross-references.md`, `dna/README.md`
**Источники:** review-2 LOW #7, review-3 MEDIUM #4, review-4 MEDIUM #4

Все остальные секции (`domain/`, `engineering/`, `ops/`, `flows/`) имеют `title`. `dna/` — нет. `title` не обязательное поле по schema, но отсутствие исключительно в `dna/` нарушает однородность навигации.

**Нужно:** добавить `title:` в frontmatter шести файлов `dna/`.

---

### M-3. CI не настроен, но `feature-flow.md` требует CI suites

**Файлы:** `memory-bank/engineering/testing-policy.md:44`, `memory-bank/flows/feature-flow.md:98`
**Источник:** review-0 HIGH #4

`testing-policy.md`: «CI: не настроен (задача backlog)». `feature-flow.md`: требует «required CI suites» в implementation-plan.

**Нужно:** либо убрать требование CI из `feature-flow.md`, либо обновить статус CI в `testing-policy.md`.

---

### M-4. Sentinel errors — два canonical документа

**Файлы:** `domain/architecture.md:59`, `engineering/coding-style.md:17`
**Источники:** review-0 LOW #13, review-2 MEDIUM #5

Оба документа описывают sentinel errors как паттерн. `coding-style.md` должен быть owner синтаксиса объявления. `architecture.md` — описывать только ownership (`domain/errors.go`) и паттерн использования (`errors.Is` в handler).

**Нужно:** разграничить ownership: синтаксис — в `coding-style.md`, структура файла — в `architecture.md`.

---

## LOW

### L-1. `features/README.md` — leftover фраза «шаблонный репозиторий»

**Файл:** `memory-bank/features/README.md`
**Источники:** review-3 LOW #6, review-4 LOW #5

«В шаблонном репозитории этот каталог может быть пустым.» Memory-bank уже развёрнут в конкретном проекте.

**Нужно:** «Если feature packages пока не созданы, каталог может быть пустым. Это нормально.»

---

### L-2. `ops/README.md` — `purpose` сформулирован как будущее действие

**Файл:** `memory-bank/ops/README.md`, поле `purpose`
**Источники:** review-3 LOW #8, review-4 LOW #6

«Читать при адаптации dev/prod workflow под проект» — ops-документы уже наполнены реальными данными.

**Нужно:** «Читать при работе с dev/prod workflow, релизами, конфигурацией и runbooks проекта.»

---

### L-3. `flows/README.md` — «reusable process-layer для шаблона»

**Файл:** `memory-bank/flows/README.md`
**Источники:** review-3 LOW #7, review-4 LOW #7

Формулировка «для шаблона» — след template-origin, не отражает текущий статус проектного документа.

**Нужно:** убрать «для шаблона» → «содержит process-layer проекта».

---

### L-4. Нет ссылки на `migrations/` в архитектурных документах

**Файл:** `memory-bank/domain/architecture.md`
**Источник:** review-0 LOW #11

Migrations-слой есть в стеке проекта, но в `architecture.md` не упомянут.

---

### L-5. Дублирование docker-команд в `testing-policy.md` и `ops/development.md`

**Файлы:** `memory-bank/engineering/testing-policy.md:28–44`, `memory-bank/ops/development.md:51–74`
**Источник:** review-2 MEDIUM #3 (остальные review считают чистым)

`ops/development.md` — логичный canonical owner команд. `testing-policy.md` дублирует их дословно вместо ссылки.

**Рекомендация:** `testing-policy.md` ссылается на `ops/development.md`, не дублирует. Спорный пункт — другие review не считают это проблемой.

---

### L-6. Отсутствие ADR для ключевых архитектурных решений

**Файлы:** `memory-bank/adr/README.md`, `memory-bank/dna/principles.md:9`
**Источник:** review-1 замечание #3

Принцип 9 требует ADR для каждого архитектурного решения. `adr/` пуст. Принятые решения (Echo, pgx, goose, HTMX, SSE) нигде не задокументированы в ADR-формате.

**Рекомендация:** рассмотреть ретроспективные ADR для ключевых решений. Не блокер.

---

### L-7. Inconsistent markdown table formatting

**Файлы:** несколько файлов
**Источник:** review-0 LOW #12

`| --- |` vs `|---|` — косметика.

---

## Итог

| Severity | Кол-во | Ключевые файлы |
|---|---|---|
| HIGH | 4 | `architecture.md`, `testing-policy.md`, `dna/lifecycle.md`, `dna/governance.md` + `dna/frontmatter.md` |
| MEDIUM | 4 | `CLAUDE.md`, `dna/` (6 файлов), `testing-policy.md`, `feature-flow.md` |
| LOW | 7 | `features/README.md`, `ops/README.md`, `flows/README.md`, `architecture.md`, несколько |
| FIXED | 2 | `ops/runbooks/README.md`, `domain/README.md` |

**Приоритет исправления:** H-1 (форум живой, статус устарел) и H-2 (риск сломать агент-workflow) — можно исправить за один коммит. H-3 и H-4 требуют редактирования `dna/` — отдельный коммит.
