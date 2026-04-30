# Отчёт HW-4: Исполнение больших планов

**Дата:** 2026-04-30
**Проект:** LFC.ru forum

---

## Что сделано

### Два малых цикла

**Brief Improve Loop** (`memory-bank/flows/brief-improve-loop.md`)
Таргетирует `## What` в `feature.md`. Содержит mermaid-диаграмму, entry/exit criteria, escalation rules, runner contract. Запускается через `improve-loop.sh` с промптом `brief-loop.md`.

**Spec Improve Loop** (`memory-bank/flows/spec-improve-loop.md`)
Таргетирует `## How` + `## Verify`. Аналогичная структура. Запускается с промптом `spec-loop.md`. Стартует только после `accept` brief loop.

Оба цикла работают с `feature.md` без разделения на отдельные файлы — brief и spec выделены как логические секции одного документа (обоснование: `.protocols/brief-spec-split-assessment.md`).

### Prompt-файлы

- `memory-bank/flows/templates/prompts/brief-loop.md` — проверки REQ-*/NS-*/ASM-*/MET-* с выводом accept/revise/escalate
- `memory-bank/flows/templates/prompts/spec-loop.md` — проверки Solution/Change Surface/Traceability/CHK-*/EVID-* с учётом архитектурных правил

### Runner

`scripts/improve-loop.sh <prompt-file> <artifact-path>`

Подставляет `{{ARTIFACT_PATH}}`, `{{FT_ID}}`, `{{DATE}}` в промпт, запускает `claude --print`, сохраняет результат в `.review-results/`. Единый для обоих малых циклов.

### Большой цикл

`memory-bank/flows/feature-execution-loop.md` — 12 этапов:

1. Brief loop (`improve-loop.sh`)
2. Spec loop (`improve-loop.sh`)
3. HITL — Design Ready approval
4. Plan + eval DR→PR (Agent tool)
5. HITL — Plan Ready approval
6. Implementation
7. Local verify (unit tests, docker)
8. E2E smoke (`docker-compose.e2e.yml` — безопасный контур)
9. Verification по SC-*
10. Fix cycle
11. Closure (PR ready)
12. HITL — ждать merge

После каждого этапа: обновление state-pack.

### State-pack

Три артефакта обеспечивают resume без пересказа:

| Артефакт | Роль |
|---|---|
| `run-state/FT-XXX/active-context.md` | текущий stage, blocked/pending, key decisions |
| `run-state/FT-XXX/stage-log.md` | журнал этапов с outcome и ссылками на evidence |
| `HANDOFF.md` (корень) | сессионный entry point, ссылается на run-state |

Шаблоны — `run-state/FT-XXX/`.

### Трасса

`homeworks/hw-4/trace.md` — ретроспективная трасса прогона FT-023 (WYSIWYG-редактор):

- 4 итерации brief/spec loop → `accept`
- 2 stop/resume между сессиями
- 5 итераций plan loop → `accept`
- Текущий статус: `blocked` на HITL pr-approval gate

---

## Уровень реализации

По критериям домашки:

| Критерий | Статус |
|---|---|
| Два малых process spec с диаграммами | ✅ |
| Prompt-файлы для циклов | ✅ |
| Runner (один общий) | ✅ |
| Большой цикл | ✅ |
| Большой цикл переиспользует малые циклы | ✅ |
| Verification — реальные проверки (unit + e2e) | ✅ |
| State-pack из 3 артефактов | ✅ |
| State обновляется по ходу нескольких этапов | ✅ |
| Trace/report с этапами, stop/resume, финальный статус | ✅ |
| Безопасный deploy-контур | ✅ (`docker-compose.e2e.yml`) |
| Явный HITL/escalation момент | ✅ (3 HITL gate в большом цикле) |
| Runner переиспользуем, не привязан к одному кейсу | ✅ |
