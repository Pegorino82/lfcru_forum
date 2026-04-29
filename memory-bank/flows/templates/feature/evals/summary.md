---
title: FT-XXX Eval Summary Template
doc_kind: feature
doc_function: template
purpose: "Wrapper-шаблон eval summary фичи. Инстанцируется при Bootstrap, обновляется при каждом gate-переходе, финализируется при Done."
derived_from:
  - ../../feature-flow.md
  - ../../../flows/eval.md
status: active
audience: humans_and_agents
template_for: feature_eval_summary
template_target_path: ../../../features/FT-XXX/evals/summary.md
---

# FT-XXX Eval Summary Template

Этот файл — wrapper-шаблон. Инстанцируется при Bootstrap в `evals/summary.md` feature package.
Обновляется после каждого gate-перехода. При Done: `status: final`.

## Instantiated Frontmatter

```yaml
title: "FT-XXX: Eval Summary"
doc_kind: feature
doc_function: eval-summary
ft_id: FT-XXX
status: active
audience: humans_and_agents
```

## Instantiated Body

```markdown
# FT-XXX Eval Summary

## Gates

| Gate | Форма | Outcome | Итерации | Date | EVID | Детали |
|---|---|---|---|---|---|---|
| Draft → Design Ready | — | — | — | — | — | [DR-eval.md](DR-eval.md) |
| Design Ready → Plan Ready | — | — | — | — | — | [PR-eval.md](PR-eval.md) |
| Plan Ready → Execution | human approval | — | — | — | — | — |
| Execution → Done | — | — | — | — | — | [Done-eval.md](Done-eval.md) |

> Заполнять по мере прохождения gates. При Done: обновить `status: final` в frontmatter.

## Notes

<!-- Наблюдения по eval-процессу этой фичи: неожиданные revise, escalate, паттерны проблем. -->
```
