---
title: Task Workflows
doc_kind: governance
doc_function: canonical
purpose: Маршрутизация задач по типам и базовый цикл разработки. Читать при получении новой задачи для выбора подхода.
derived_from:
  - ../dna/governance.md
  - feature-flow.md
canonical_for:
  - task_routing_rules
  - base_development_cycle
  - workflow_type_selection
  - autonomy_gradient
status: active
audience: humans_and_agents
---

# Task Workflows

## Базовый цикл

Любой workflow — цепочка повторений одного цикла:

```text
Артефакт → Ревью → Полировка
                  → Декомпозиция
                  → Принят
```

Артефакт — то, что создаётся на каждом этапе: спецификация, дизайн-док, план, код, PR, runbook.

## Градиент участия человека

Чем ближе к бизнес-требованиям, тем больше участия человека. Чем ближе к коду и локальному verify, тем больше агент работает автономно.

```text
Бизнес-требования  ← человек  |  агент →  Код
  PRD, Use Cases      Спека, План           PR, Тесты
```

## Типы Workflow

### 1. Малая фича

Когда:

- задача понятна;
- scope локален;
- решение помещается в одну сессию или один компактный change set.

Flow:

`issue/task -> routing -> implementation -> review -> merge`

### 2. Средняя или большая фича

Когда:

- затрагивает несколько слоёв;
- требует design choices;
- нужны checkpoints и явный execution plan.

Flow:

`issue/task -> spec -> feature package -> implementation plan -> execution -> review -> handoff`

### 3. Баг-фикс

Источники могут быть любыми: error tracker, support, QA, прямой report от пользователя, инцидентный анализ.

Flow:

`report -> reproduction -> analysis -> fix -> regression coverage -> review`

**FT-пакет:** облегчённый. Создаётся `FT-XXX/README.md` без `feature.md` и `implementation-plan.md`. README содержит:

- описание бага и условия воспроизведения
- корневую причину (после анализа)
- ссылку на коммит с фиксом
- добавленный regression-тест

Если в ходе анализа выясняется, что баг требует design choices или меняет контракт — поднимается до workflow «Средняя или большая фича» с полным feature package.

### 4. Рефакторинг

Разделяй минимум на три класса:

- по ходу delivery-задачи;
- исследовательский;
- системный, с большим change surface.

Исследовательский и системный refactoring обычно требуют явного плана и checkpoints.

### 5. Инцидент / PIR

Flow:

`incident -> timeline -> root cause analysis -> fixes -> prevention work`

Здесь человек обычно подтверждает RCA и приоритеты follow-up задач.

## Routing Rules

Используй минимальный workflow, который не теряет контроль над риском.

- Если задача маленькая и понятная, не раздувай её до большого feature package.
- Если задача меняет контракт, rollout или требует approvals, поднимай её до feature flow.
- Если замечания не уменьшаются от итерации к итерации, проблема может быть upstream, а не в коде.

## Session Protocol

### В начале сеанса

1. Прочитай `HANDOFF.md` в корне проекта (если существует) — там контекст от предыдущего агента.
2. Прочитай задачу до конца.
3. Найди затронутые файлы, прочитай их.

### В конце сеанса

Применяется ко всем типам workflow:

1. Запусти тесты (unit + integration для затронутых пакетов), убедись, что зелёные.
2. Simplify review — нет ли premature abstractions, dead code, дублирования логики.
3. Сделай коммит согласно [git-workflow.md](../engineering/git-workflow.md).
4. Обнови `HANDOFF.md` в корне проекта по шаблону:

```markdown
## Что сделано
- <краткий список>

## Что сделать следующим
- <конкретные шаги>

## Проблемы и решения
- <проблема> → <как решили>
```

Файл предназначен для следующего агента — пиши коротко и конкретно.

**Только для «Средняя или большая фича»:** если сессия закрывает feature package, выполни gate «Execution → Done» из [feature-flow.md](feature-flow.md).
