---
title: Eval Framework
doc_kind: governance
doc_function: canonical
purpose: "Eval-слой поверх feature-flow: уровни оценки, формы, evaluator agent protocol и gate-чеклисты. Читать при переходе между стадиями feature flow."
derived_from:
  - ../dna/governance.md
  - feature-flow.md
canonical_for:
  - eval_levels
  - eval_forms
  - evaluator_agent_protocol
  - gate_eval_checklists
must_not_define:
  - feature_flow_gate_predicates
status: active
audience: humans_and_agents
---

# Eval Framework

Eval отвечает на вопрос: **можно ли доверить этот результат следующему шагу?**

Не "умная ли модель", а "годится ли конкретный результат для продолжения работы".

## Уровни оценки

| Уровень | Вопрос | Gate в feature-flow |
|---|---|---|
| **Spec-level** | Спецификация достаточно зрелая для планирования? | Draft → Design Ready |
| **Artifact-level** | Артефакт (план, код) хорош сам по себе? | Design Ready → Plan Ready |
| **Execution-level** | Что фактически произошло? Подтверждено ли evidence? | Execution → Done |
| **Workflow-level** | Можно ли двигаться дальше? | Каждый gate-переход |

## Формы оценки

| Форма | Когда применяется |
|---|---|
| **Self-check агента** | `short.md`; любой gate при малой фиче |
| **Evaluator agent** | `large.md`: DR → PR и Done gates — обязательно |
| **Human checklist** | `AG-*` gates; Plan Ready → Execution (HARD STOP) |
| **Executable checks** | Тесты, lint, CI — на Execution-level всегда |
| **Hybrid** | Done gate для `large.md`: CI + evaluator agent + human AG-* |

## Outcomes

Каждый eval заканчивается одним из трёх решений:

- **`accept`** — переход разрешён; записать EVID-* с семантикой eval (см. ниже).
- **`revise`** — артефакт возвращается с конкретными пронумерованными замечаниями.
- **`escalate`** — upstream-проблема (конфликт требований, неясный scope, ломаный фундамент) → остановиться, поднять к человеку.

**Правило:** если revise повторяется > 2 раз для одного артефакта — это сигнал escalate.

## Evaluator Agent Protocol

Evaluator agent — агент, читающий артефакт в новом контексте без истории его создания.

**Изоляция сессии.** Evaluator agent вызывается через **Agent tool** — дочерний агент не имеет
контекста builder-сессии. Builder не может выступать evaluator для своего артефакта: это
антипаттерн «builder сам себе подтверждает готовность».

**Когда вызывать:**

- `short.md` → self-check агента достаточен для всех gates.
- `large.md`, gate DR → PR → evaluator agent обязателен.
- `large.md`, gate Execution → Done → evaluator agent обязателен.
- `large.md`, gate DR → PR для малых планов (≤ 3 STEP-*) → self-check допустим.

**Что делает:**

1. Читает артефакт и соответствующий чеклист из этого документа.
2. Проходит по каждому пункту чеклиста.
3. Возвращает: `accept` / `revise` (с пронумерованными замечаниями) / `escalate`.
4. Если `accept` — фиксирует EVID-* в документе (см. «Evidence для eval»).

**Что НЕ делает:**

- Не создаёт код или план.
- Не переписывает проверяемый артефакт.
- Не принимает решения об upstream-конфликтах (это `escalate`).

**Промпт для Agent tool** (подставить путь к артефакту и название gate):

```
Ты evaluator agent. Работай в режиме строгой независимой оценки — без доступа к истории создания артефакта.

Прочитай:
- [путь к артефакту: feature.md или implementation-plan.md]
- memory-bank/flows/eval.md — чеклист для gate [gate-name: DR→PR | Execution→Done]

Пройди по каждому пункту чеклиста.
Верни: accept / revise / escalate.
- revise → пронумерованные замечания
- accept → запиши EVID-* в артефакт

Запрещено: создавать код, переписывать артефакт, принимать upstream-решения (это escalate).
```

Если `revise` повторяется > 2 раз для одного артефакта — это сигнал `escalate` к человеку.

## Evidence для eval

При `accept`-решении evaluator добавляет строку в секцию Evidence артефакта:

```
EVID-XX: Eval [gate-name] — accept. YYYY-MM-DD. [форма: self-check / evaluator / CI]
```

Это обычный `EVID-*` с семантикой eval-evidence. Новый тип идентификатора не нужен.

## Gate Чеклисты

### Draft → Design Ready

Форма: **self-check агента** перед переводом `delivery_status`.

```
REQ coverage:
- [ ] каждый REQ-* описывает конкретное поведение, а не намерение
- [ ] каждый REQ-* однозначен: два независимых агента прочитают его одинаково
- [ ] нет REQ-*, дублирующего другой

SC coverage:
- [ ] каждый REQ-* прослеживается к ≥ 1 SC-*
- [ ] SC-* описывает наблюдаемый результат, а не внутреннюю реализацию
- [ ] SC-* читается как: Given / When / Then (или эквивалент)

Verify readiness:
- [ ] каждый CHK-* имеет команду или ручную процедуру (не "проверить вручную" без инструкции)
- [ ] каждый EVID-* имеет конкретный path contract (не "где-нибудь")
- [ ] NS-* достаточно, чтобы агент не додумывал scope
```

### Design Ready → Plan Ready

Форма: **evaluator agent** для `large.md`; self-check для `short.md`.

```
Grounding:
- [ ] discovery context содержит реальные пути из репозитория (не шаблонные заглушки)
- [ ] OQ-* зафиксированы явно, а не скрыты в prose шагов
- [ ] test strategy покрывает все change surfaces из feature.md

Plan completeness:
- [ ] каждый STEP-* имеет actor, goal, artifact, check command
- [ ] каждый AG-* имеет approver и ожидаемое evidence
- [ ] CHK-* в плане ссылаются на canonical CHK-* из feature.md
  (не создают параллельную систему acceptance checks)

Sequencing:
- [ ] нет шага, создающего downstream-артефакт раньше upstream-зависимости
- [ ] PAR-* не создают write-surface конфликт
```

### Plan Ready → Execution

Форма: **human approval** (AG-*); все предикаты — в HARD STOP feature-flow.md.

```
- [ ] все HARD STOP предикаты из feature-flow.md выполнены
- [ ] каждый manual-only gap имеет AG-* с approver и причиной
```

Eval для этого gate встроен в HARD STOP: если все предикаты истинны — outcome `accept` по умолчанию.

### Execution → Done

Форма: **hybrid** — executable checks + evaluator agent (`large.md`).

```
Evidence completeness:
- [ ] каждый EVID-* из feature.md имеет конкретный carrier (path, CI run URL, screenshot)
- [ ] каждый CHK-* из feature.md имеет статус pass или fail с кратким обоснованием
- [ ] ни один CHK-* не помечен pass без соответствующего EVID-* carrier

Test coverage:
- [ ] все required automated tests добавлены или обновлены
- [ ] local test suites зелёные (команды из Test Strategy)
- [ ] CI зелёный (если настроен)

Manual gaps:
- [ ] каждый manual-only gap имеет AG-* с approval ref
- [ ] нет manual-only gap для критичных путей (auth, sessions, CSRF) без явного обоснования

Simplify review:
- [ ] simplify review выполнен (не пропущен)

Workflow gate:
- [ ] все EC-* из feature.md истинны
- [ ] PR переведён из draft в ready for review
- [ ] `[human]` PR merged в `main` (closure-шаги после этого)
```

## Антипаттерны

| Антипаттерн | Что ломается |
|---|---|
| "Проверь, всё ли ок" без критериев | агенту не на что опираться |
| Eval без проверочных фактов (EVID-*) | решение принимается по впечатлению |
| Только self-check для Done gate на `large.md` | builder сам себе подтверждает готовность |
| Evaluator agent переписывает артефакт | нарушение separation of roles |
| HITL на каждом шаге | автономность падает без выигрыша в качестве |
