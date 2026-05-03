---
title: "FT-024: Feature Package"
doc_kind: feature
doc_function: index
purpose: "Bootstrap-safe навигация по документации фичи. Читать, чтобы сначала перейти к canonical `feature.md`, а optional derived docs добавлять только после их появления."
derived_from:
  - ../../dna/governance.md
  - feature.md
status: active
audience: humans_and_agents
---

# FT-024: Feature Package

## О разделе

Каталог feature package хранит canonical `feature.md`, а optional derived/external routes добавляются только после появления соответствующих документов. Сначала читай `feature.md`, затем расширяй routing по мере появления execution и decision artifacts.

## Аннотированный индекс

- [`feature.md`](feature.md)
  Читать, когда нужно: открыть instantiated canonical feature-документ сразу после bootstrap нового feature package.
  Отвечает на вопрос: где находятся scope, design, verify, blockers и canonical IDs для этой фичи.

- [`evals/strategy.md`](evals/strategy.md)
  Читать, когда нужно: узнать, какая форма eval применяется на каждом gate и какие risk areas выделены для этой фичи.
  Отвечает на вопрос: как организован eval-процесс для данной фичи.

- [`evals/summary.md`](evals/summary.md)
  Читать, когда нужно: получить сводку всех gate-переходов — исходы, итерации, EVID-ссылки.
  Отвечает на вопрос: каков итоговый eval-статус фичи по всем gates.
