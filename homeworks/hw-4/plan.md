# HW-4: Исполнение больших планов — План реализации

**Дата:** 2026-04-30
**Ветка:** `hw/4-execution-loops`

---

## Решение по ключевым вопросам

### Brief / Spec без разделения файлов

`feature.md` не разбивается на `brief.md` + `spec.md` (оценка затрат — `.protocols/brief-spec-split-assessment.md`; будет использован как отправная точка для плана разделения после отладки процессов).

Два цикла таргетируют разные секции одного файла:

| Цикл | Секция feature.md |
|---|---|
| brief improve loop | `## What` (Problem, Scope/REQ-*, NS-*, Constraints) |
| spec improve loop | `## How` + `## Verify` |

### Размещение артефактов

Все process specs и шаблоны — в `memory-bank/flows/` и `memory-bank/flows/templates/`.
HW-4 продолжает выстраивание процесса разработки, артефакты — проектные, не учебные.

### State-pack (3 артефакта)

| Артефакт | Роль | Читатель |
|---|---|---|
| `run-state/FT-XXX/active-context.md` | текущий stage, pending/blocked | runner при resume |
| `HANDOFF.md` (корень) | сессионный entry point | агент при старте сессии |
| `run-state/FT-XXX/stage-log.md` | журнал этапов с ссылками на evidence | runner + человек |

`.review-results/FT-XXX/` — linked из `stage-log.md` как evidence, не входят в state-pack формально.

---

## Артефакты к созданию

```
memory-bank/flows/
  brief-improve-loop.md          # process spec: диаграмма, entry/exit criteria, escalation, artifacts
  spec-improve-loop.md           # то же для spec
  feature-execution-loop.md      # большой цикл: этапы + state-обновления + HITL-моменты
  templates/
    prompts/
      brief-loop.md              # prompt: фокус на ## What
      spec-loop.md               # prompt: фокус на ## How + ## Verify

scripts/
  improve-loop.sh                # ./improve-loop.sh <prompt-file> <artifact-path>

run-state/
  FT-XXX/                        # шаблон — при реальном прогоне FT-XXX → конкретный ID
    active-context.md
    stage-log.md

homeworks/hw-4/
  plan.md                        # этот файл
  trace.md                       # трасса реального прогона большого цикла (FT-023)
  report.md                      # итоговый отчёт
```

---

## Этапы реализации

### 1. `memory-bank/flows/brief-improve-loop.md`

Process spec малого цикла:
- Mermaid-диаграмма
- Entry criteria (когда применять)
- Exit criteria (accept / revise / escalate)
- Escalation rules (> 2 revise → escalate)
- Артефакты, которые runner обновляет или возвращает

### 2. `memory-bank/flows/spec-improve-loop.md`

Аналогично, фокус на `## How` + `## Verify`, проверки из `eval.md` (traceability, CHK-*, EVID-*).

### 3. `memory-bank/flows/templates/prompts/brief-loop.md`

Новый промпт, фокус на `## What` — entry criteria, проверки REQ-*/NS-*/ASM-*, outcome.

### 4. `memory-bank/flows/templates/prompts/spec-loop.md`

Адаптация `review-feature-md.md`: убрать What-проверки, усилить How/Verify проверки.

### 5. `scripts/improve-loop.sh`

Единый bash-скрипт:
```bash
./scripts/improve-loop.sh <prompt-file> <artifact-path>
```
Подставляет путь к артефакту в промпт, запускает `claude --print`, сохраняет результат в `.review-results/`.

### 6. `memory-bank/flows/feature-execution-loop.md`

Большой цикл — этапы:
1. brief improve loop (`improve-loop.sh`)
2. spec improve loop (`improve-loop.sh`)
3. **[HITL]** подтверждение Design Ready → Plan Ready
4. implementation по плану
5. local verify: unit-тесты (`docker run golang:1.23-alpine go test ./...`)
6. e2e smoke: `docker-compose.e2e.yml` как безопасный контур
7. verification по SC-* из feature.md
8. fix-цикл по замечаниям
9. closure: PR ready for review

После каждого этапа: обновить `run-state/FT-XXX/active-context.md` + `stage-log.md`.

### 7. `run-state/` — шаблоны state-pack

Шаблоны `active-context.md` и `stage-log.md` с форматом и инструкцией.

### 8. `homeworks/hw-4/trace.md` + `report.md`

- `trace.md` — реальный прогон на FT-023: материал из `.review-results/FT-023/` (4 итерации ревью, stop/resume между сессиями)
- `report.md` — финальный отчёт: этапы, runner-ы, state-файлы, outcome

---

## Безопасный контур для deploy-этапа

Настоящего stage нет. Эквивалент — `docker-compose.e2e.yml` (app на :8081 против `lfcru_test`). Verification по SC-* выполняется Playwright-тестами на этом контуре.

---

## Что остаётся неизменным

- Структура `memory-bank/` — только добавляем, не меняем существующее
- `feature.md` — не разбиваем
- Существующие `eval.md`, `feature-flow.md` — не меняем, новые процессы их используют
