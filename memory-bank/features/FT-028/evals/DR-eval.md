---
title: "FT-028: Design Ready Eval"
doc_kind: feature
doc_function: gate-eval
ft_id: FT-028
gate: "Draft → Design Ready"
status: open
date: 2026-05-18
audience: humans_and_agents
---

# FT-028: Design Ready Eval

## Checklist

### Brief Loop (## What)
- [x] REQ-* описывают конкретное поведение
- [x] REQ-* однозначны
- [x] Нет дублирующих REQ-*
- [x] NS-* достаточно
- [x] Problem специфичен для delivery-единицы
- [x] MET-* имеют baseline, target, method
- [x] ASM-*/CON-*/DEC-* непротиворечивы

### Spec Loop (## How + ## Verify)
- [x] Solution описывает конкретный подход с trade-off
- [x] Change Surface содержит реальные пути
- [x] CTR-* с producer/consumer
- [x] FM-* покрывают XSS, API unavailability, idempotency
- [x] Каждый REQ-* прослеживается к SC-*
- [x] NEG-* присутствуют (NEG-01, NEG-02)
- [x] CHK-* содержат executable команды
- [x] EVID-* имеют конкретные path contracts
- [x] CSRF зафиксирован (CON-02, Flow step 1)

## Iterations

| # | Date | Outcome | Findings |
|---|---|---|---|
| brief-01 | 2026-05-18 | revise | 3 замечания: REQ-01 impl detail, REQ-03/NS-02 conflict, ASM-02 is OQ |
| brief-02 | 2026-05-18 | revise | 2 замечания (MEDIUM/LOW): REQ-01 ambiguity, DEC ordering |
| spec-01 | 2026-05-18 | revise | 8 замечаний: XSS blocker, missing Change Surface, no NEG-*, CHK commands |
| spec-02 | 2026-05-18 | accept | All criteria passed |

## Decision

**Outcome:** accept
**Date:** 2026-05-18
**EVID:** EVID-04 (spec loop accept in feature.md)
