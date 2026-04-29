---
title: FT-XXX Eval Strategy Template
doc_kind: feature
doc_function: template
purpose: "Wrapper-шаблон eval-стратегии фичи. Инстанцируется при Bootstrap в evals/strategy.md — до начала работы с feature.md."
derived_from:
  - ../../feature-flow.md
  - ../../../flows/eval.md
status: active
audience: humans_and_agents
template_for: feature_eval_strategy
template_target_path: ../../../features/FT-XXX/evals/strategy.md
---

# FT-XXX Eval Strategy Template

Этот файл — wrapper-шаблон. Инстанцируемый документ живёт ниже как embedded contract и копируется в `evals/strategy.md` feature package без wrapper frontmatter.

## Instantiated Frontmatter

```yaml
title: "FT-XXX: Eval Strategy"
doc_kind: feature
doc_function: eval-strategy
ft_id: FT-XXX
status: active
audience: humans_and_agents
```

## Instantiated Body

```markdown
# FT-XXX Eval Strategy

## Gates & Forms

| Gate | Форма | Evaluator |
|---|---|---|
| Draft → Design Ready | self-check / evaluator agent | Agent tool |
| Design Ready → Plan Ready | evaluator agent | Agent tool |
| Plan Ready → Execution | human approval | [имя approver] |
| Execution → Done | hybrid: CI + evaluator agent + human AG-* | Agent tool + AG-* |

> Для `short.md` self-check достаточен на всех gates. Для `large.md` evaluator agent
> обязателен на DR→PR и Execution→Done (исключение: план ≤ 3 STEP-* → self-check допустим).

## Risk Areas

<!-- Специфические для этой фичи риски, требующие особого внимания при eval.
     Примеры: критичный path (auth, sessions, CSRF), сложный sequencing, manual-only gaps. -->

- (заполнить при Bootstrap или Draft)
```
