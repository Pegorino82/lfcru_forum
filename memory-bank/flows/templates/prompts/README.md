# Промпты для праймеринга сессий

Шаблоны промптов для запуска агента под каждый тип workflow.
Вставляй нужный промпт в начало сессии, заменяя `[...]` на конкретную задачу.

| Файл | Workflow |
|---|---|
| `small-feature.md` | Малая фича (issue → implementation → merge) |
| `large-feature.md` | Средняя/большая фича (spec → plan → execution) |
| `bug-fix.md` | Баг-фикс (reproduction → fix → regression coverage) |
| `refactoring.md` | Рефакторинг (по ходу / исследовательский / системный) |
| `incident-pir.md` | Инцидент / PIR (timeline → RCA → fixes → prevention) |
| `review-feature-md.md` | Ревью `feature.md` evaluator agent-ом (gate Design Ready → Plan Ready) |
| `review-implementation-plan.md` | Ревью `implementation-plan.md` evaluator agent-ом (gate Design Ready → Plan Ready) |

Выбор workflow — по правилам из [`../workflows.md`](../workflows.md).
