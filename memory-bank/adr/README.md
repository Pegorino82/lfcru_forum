---
title: Architecture Decision Records Index
doc_kind: adr
doc_function: index
purpose: Навигация по ADR проекта. Читать, чтобы найти уже принятые решения или завести новый ADR по шаблону.
derived_from:
  - ../dna/governance.md
  - ../flows/templates/adr/ADR-XXX.md
status: active
audience: humans_and_agents
---

# Architecture Decision Records Index

Каталог `memory-bank/adr/` хранит instantiated ADR проекта.

- Заводи новый ADR из шаблона [`../flows/templates/adr/ADR-XXX.md`](../flows/templates/adr/ADR-XXX.md).
- Держи в этом каталоге только реальные decision records, а не заметки или черновые исследования.
## Records

| ADR | Решение | Статус | Источник |
|---|---|---|---|
| [ADR-001](ADR-001-session-storage.md) | Сессии хранятся в PostgreSQL (не Redis, не JWT) | `accepted` | FT-001 |
| [ADR-002](ADR-002-rate-limiting-strategy.md) | IP-based rate limiting на `/login` | `accepted` | FT-001 |
| [ADR-003](ADR-003-email-only-auth-no-confirmation.md) | Email-only регистрация без подтверждения email в MVP | `accepted` | FT-001 |
| [ADR-004](ADR-004-forum-hierarchy-model.md) | Иерархическая модель форума sections→topics→posts | `accepted` | FT-005 |
| [ADR-005](ADR-005-image-storage.md) | Хранение изображений: файловая система + Docker volume | `accepted` | FT-009 |
| [ADR-006](ADR-006-article-status-machine.md) | Статусная машина статьи: draft/in_review/published enum | `proposed` | FT-008 |

## Naming

- Формат файла: `ADR-XXX-short-decision-name.md`
- Нумерация монотонная и не переиспользуется
- Заголовок файла должен совпадать с `title` во frontmatter

## Statuses

- `proposed` — решение сформулировано, но еще не принято
- `accepted` — решение принято и считается canonical input для downstream-документов
- `superseded` — решение заменено другим ADR
- `rejected` — решение рассмотрено и отклонено
