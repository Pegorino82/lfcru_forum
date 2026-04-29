---
title: FT-XXX Gate Eval Template
doc_kind: feature
doc_function: template
purpose: "Wrapper-шаблон для gate-eval файла (DR-eval.md / PR-eval.md / Done-eval.md). Инстанцируется evaluator agent при каждом gate-переходе."
derived_from:
  - ../../feature-flow.md
  - ../../../flows/eval.md
status: active
audience: humans_and_agents
template_for: feature_gate_eval
template_target_path: ../../../features/FT-XXX/evals/[gate]-eval.md
---

# FT-XXX Gate Eval Template

Этот файл — wrapper-шаблон. Инстанцируется evaluator agent как `DR-eval.md`, `PR-eval.md`
или `Done-eval.md` в `evals/` feature package. Имя файла — по gate: DR / PR / Done.

## Instantiated Frontmatter

```yaml
title: "FT-XXX: [Gate Name] Eval"
doc_kind: feature
doc_function: gate-eval
ft_id: FT-XXX
gate: "DR→PR | DR→Plan | Execution→Done"
status: open
date: YYYY-MM-DD
audience: humans_and_agents
```

## Instantiated Body

```markdown
# FT-XXX: [Gate Name] Eval

## Checklist

<!-- Скопировать соответствующий чеклист из memory-bank/flows/eval.md и проставить статусы. -->

### [Раздел чеклиста 1]
- [ ] пункт
- [ ] пункт

### [Раздел чеклиста 2]
- [ ] пункт

## Iterations

| # | Date | Outcome | Findings |
|---|---|---|---|
| 1 | YYYY-MM-DD | revise / accept / escalate | 1. ... |

## Decision

**Outcome:** accept / revise / escalate
**Date:** YYYY-MM-DD
**EVID:** EVID-XX (canonical carrier в feature.md)

<!-- При accept: EVID-* записывается в feature.md; здесь — ссылка на него. -->
<!-- При revise: итерации фиксируются выше, финальное решение — после последней итерации. -->
<!-- При escalate: описать upstream-проблему и передать человеку. -->
```
