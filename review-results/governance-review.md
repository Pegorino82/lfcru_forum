# Memory-bank: Document Governance Review

_Дата: 2026-04-10_

---

## Структура

```
memory-bank/
├── dna/          — ядро governance (principles, frontmatter, lifecycle, cross-refs)
├── domain/       — product context (problem, architecture, frontend)
├── engineering/  — практики (testing, autonomy, coding-style, git)
├── ops/          — операционный слой (dev, stages, release, config, runbooks)
├── flows/        — процессы + шаблоны (feature-flow, workflows, templates/)
└── prd/, use-cases/, adr/, features/  — пустые индексы (только README)
```

---

## Сильные стороны

- **SSoT с явным authority flow** — `derived_from` в frontmatter, нет конкурирующих canonical-owner
- **Полный dependency-tree** — читаемый граф зависимостей документов, нет циклов
- **Слоистая иерархия** — DNA → Domain → Engineering → Ops → Features; каждый слой изолирован
- **Index-first** — все документы достижимы из README-индексов; осиротевших файлов нет
- **Обёрточные шаблоны** с embedded-контрактом — предотвращают schema drift при копировании
- **Lifecycle rules** — явный sync-checklist, conflict resolution, правило "report, not self-fix"
- **HANDOFF.md** — рабочая передача контекста между агентами

---

## Проблемы / Гапы

| Приоритет | Проблема | Где |
|-----------|----------|-----|
| **High** | **Нет экземпляров** в `prd/`, `adr/`, `use-cases/` — только пустые README. Шаблоны есть, применения нет. Governance висит в воздухе без реальных артефактов. | `prd/`, `adr/`, `use-cases/` |
| **High** | **CLAUDE.md дублирует memory-bank** — архитектура, стек, coding conventions, тест-команды, autonomy — всё это есть и в `CLAUDE.md` и в `engineering/`, `ops/`, `domain/`. Два источника одной правды, нет чёткого canonical owner. | `CLAUDE.md` vs `engineering/*`, `ops/*` |
| **Medium** | **Старые features 001–005** удалены из `memory-bank/features/` (статус D в git), но в `CLAUDE.md` таблица "Реализованные фичи" ссылается на них. Ссылки битые. | `CLAUDE.md` → раздел "Реализованные фичи" |
| **Medium** | **Нет автоматической проверки** — governance только декларативный. Нет pre-commit hook, нет CI-шага для валидации frontmatter / `derived_from` / README-coverage. | Весь memory-bank |
| **Medium** | **Worktrees** — в `engineering/git-workflow.md` объявлены "не используются", но `CLAUDE.md` их не упоминает. Если политика поменялась — два документа не синхронизированы. | `engineering/git-workflow.md` vs `CLAUDE.md` |
| **Low** | **`runbooks/` пустой** — только README-шаблон. Для production-готового проекта отсутствие runbooks — риск. | `ops/runbooks/` |
| **Low** | **Сложность feature-flow** — 9+ стабильных ID-префиксов (REQ-*, SC-*, CHK-*, NT-*, INV-*...) могут быть избыточны для маленьких фич. | `flows/feature-flow.md` |

---

## Главный вывод

Governance-система зрелая и хорошо спроектированная, но **не заземлена** — шаблоны есть, реальных документов (PRD, ADR, use cases) нет. Самый серьёзный structural gap — дублирование между `CLAUDE.md` и `memory-bank`: непонятно, что canonical при расхождении. Нужно либо сделать `CLAUDE.md` читаемым "entry point → ссылки в memory-bank", либо убрать дублирование из memory-bank.
