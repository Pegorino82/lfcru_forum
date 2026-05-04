---
title: "FT-024: Plan Ready Eval"
doc_kind: feature
doc_function: gate-eval
ft_id: FT-024
gate: "DR→Plan Ready"
status: open
date: 2026-05-03
audience: humans_and_agents
---

# FT-024: Plan Ready Eval

## Checklist

### A. Связь с feature.md
- [x] Каждый `STEP-*.Implements` ссылается на существующие ID из `feature.md`
- [x] Все файлы из Change Surface отражены в Discovery Context или STEP-*
- [x] Расхождения зафиксированы как `OQ-*`
- [x] Нет конфликта с `ASM-*` / `CON-*` / `NS-*`

### B. Discovery Context
- [x] Пути реальны (проверены через filesystem)
- [x] `OQ-*` явно в Open Questions
- [x] Нет дублирования между `OQ-*` и `ER-*` без связи

### C. Test Strategy
- [x] Unit: Required local suites = дословная команда из ops/development.md
- [x] Integration: Required local suites = "—"
- [x] E2E: Required CI suites = E2E job
- [x] Тип теста явен для каждого нового test-файла
- [x] Assertions конкретны (struct field, HTTP body, DOM)
- [x] E2E seeding описан (fixture, teardown)
- [x] Каждый CHK-* из feature.md покрыт

### D. Environment Contract
- [x] Команды скопированы дословно из ops/development.md
- [x] E2E prerequisite включает оба шага (dev-stack + e2e-stack)

### E. Lifecycle
- [x] `status: draft` корректен во время ревью
- [x] `delivery_status: in_progress` оформлен как STEP-00 (HARD STOP)

### F. STEP-* качество
- [x] Каждый STEP атомарен
- [x] Sequencing корректен
- [x] Отклонения от Layer Stack задокументированы
- [x] Рискованные действия имеют AG-*
- [x] AG-* не стоит на автопилоте

### G. ER-* и STOP-*
- [x] Все 4 ER-* покрыты STOP-* (STOP-01..STOP-05)
- [x] PAR-* без write-conflict

### H. Обязательные STEP
- [x] Simplify Review STEP-10 присутствует
- [x] UC-*/docs STEP не требуется (не входит в Change Surface)

## Iterations

| # | Date | Outcome | Findings |
|---|---|---|---|
| 1 | 2026-05-03 | revise | 3 BLOCKER (D-1 wrong docker cmd, C-2 integration в local suites, D-2 e2e prerequisite); 1 HIGH; 5 MEDIUM |
| 2 | 2026-05-03 | revise | 2 MEDIUM (D-1 пакетная вариация команды, G-1 STOP-05 отсутствует для ER-04) |
| 3 | 2026-05-03 | accept | self-check — все замечания устранены |

## Decision

**Outcome:** accept
**Date:** 2026-05-03
**EVID:** EVID-06 (записан в feature.md)
