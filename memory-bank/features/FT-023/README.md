---
title: "FT-023: Feature Package"
doc_kind: feature
doc_function: index
purpose: "Bootstrap-safe навигация по документации фичи. Читать, чтобы сначала перейти к canonical feature.md, а optional derived docs добавлять только после их появления."
derived_from:
  - ../../dna/governance.md
  - feature.md
status: active
audience: humans_and_agents
---

# FT-023: Feature Package

## О разделе

Каталог feature package хранит canonical `feature.md`, а optional derived/external routes добавляются только после появления соответствующих документов. Сначала читай `feature.md`, затем расширяй routing по мере появления execution и decision artifacts.

## Аннотированный индекс

- [`feature.md`](feature.md)
  Читать, когда нужно: открыть instantiated canonical feature-документ сразу после bootstrap нового feature package.
  Отвечает на вопрос: где находятся scope, design, verify, blockers и canonical IDs для этой фичи.

- [`../../../adr/ADR-007-wysiwyg-editor-html-storage.md`](../../../adr/ADR-007-wysiwyg-editor-html-storage.md)
  Читать, когда нужно: проверить текущий `decision_status` решения по формату хранения и выбору редактора.
  Отвечает на вопрос: почему TipTap + HTML-хранение, и на каком этапе ADR.
