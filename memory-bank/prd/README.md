---
title: Product Requirements Documents Index
doc_kind: prd
doc_function: index
purpose: Навигация по instantiated PRD проекта. Читать, чтобы найти существующий Product Requirements Document или завести новый по шаблону.
derived_from:
  - ../dna/governance.md
  - ../flows/templates/prd/PRD-XXX.md
status: active
audience: humans_and_agents
---

# Product Requirements Documents Index

Каталог `memory-bank/prd/` хранит instantiated PRD проекта.

PRD нужен, когда задача живет на уровне продуктовой инициативы или capability, а не одного vertical slice. Обычно PRD стоит между общим контекстом из [`../domain/problem.md`](../domain/problem.md) и downstream feature packages из [`../features/README.md`](../features/README.md).

## Граница С `domain/problem.md`

- [`../domain/problem.md`](../domain/problem.md) остается project-wide документом и не превращается в PRD.
- PRD наследует этот контекст через `derived_from`, но фиксирует только initiative-specific проблему, users, goals и scope.
- Если документ нужен только для того, чтобы повторить общий background проекта, оставайся на уровне `domain/problem.md`.

## Когда Заводить PRD

- инициатива распадается на несколько feature packages;
- нужно зафиксировать users, goals, product scope и success metrics до проектирования реализации;
- есть риск смешать продуктовые требования с architecture/design detail.

## Когда PRD Не Нужен

- задача локальна и полностью помещается в один `feature.md`;
- общий продуктовый контекст уже покрыт [`../domain/problem.md`](../domain/problem.md), а feature не требует отдельного product-layer документа.

## Gate: draft → upstream-ready

PRD считается готовым как upstream-документ для downstream feature packages только когда выполнены все предикаты:

- [ ] секция «Problem» описывает пользовательскую или бизнес-проблему, а не решение
- [ ] секция «Users And Jobs» содержит ≥ 1 строку с явным Job To Be Done и текущей болью
- [ ] секция «Goals» содержит ≥ 1 обязательный outcome (G-01) с измеримым критерием
- [ ] секция «Non-Goals» содержит ≥ 1 явный NG-*, исключающий scope, который можно было бы молча додумать
- [ ] секция «Success Metrics» содержит ≥ 1 метрику с baseline и target (не "улучшить")
- [ ] секция «Downstream Features» перечисляет ожидаемые FT-XXX или явно фиксирует, что декомпозиция ещё не выполнена

После достижения upstream-ready:
- переведи `status: draft` → `status: active` во frontmatter PRD;
- зарегистрируй PRD в таблице Records (создай её при первом PRD);
- downstream `feature.md` добавляет этот PRD в свой `derived_from`.

## Naming

- Формат файла: `PRD-XXX-short-name.md`
- Вместо `XXX` используй идентификатор, принятый в проекте: initiative id, epic id или другой стабильный ключ
- Один PRD может быть upstream для нескольких feature packages

## Template

- Используй шаблон [`../flows/templates/prd/PRD-XXX.md`](../flows/templates/prd/PRD-XXX.md)
