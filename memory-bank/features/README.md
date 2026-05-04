---
title: Feature Packages Index
doc_kind: feature
doc_function: index
purpose: Навигация по instantiated feature packages. Читать, чтобы найти существующую delivery-единицу или понять, где создавать новую.
derived_from:
  - ../dna/governance.md
  - ../flows/feature-flow.md
status: active
audience: humans_and_agents
---

# Feature Packages Index

Каталог `memory-bank/features/` хранит instantiated feature packages вида `FT-XXX/`.

## Rules

- Каждый package создается по правилам из [`../flows/feature-flow.md`](../flows/feature-flow.md).
- Для bootstrap используй шаблоны из [`../flows/templates/feature/`](../flows/templates/feature/).
- Этот файл обновляется из feature-ветки (worktree). При параллельных фичах конфликты решаются `git rebase main` перед merge — файл append-only, конфликт тривиальный.
- Если feature реализует или существенно меняет устойчивый сценарий проекта, соответствующий `UC-*` из [`../use-cases/README.md`](../use-cases/README.md) должен быть создан или обновлён, а `feature.md` — ссылаться на него. UC может создаваться вместе с фичей, но должен быть готов до Done gate.
- Если feature packages пока не созданы, каталог может быть пустым. Это нормально.

## Packages

| FT | Title | Status |
| --- | --- | --- |
| [FT-001](FT-001/) | Auth | `done` |
| [FT-002](FT-002/) | — | `done` |
| [FT-003](FT-003/) | — | `done` |
| [FT-004](FT-004/) | — | `done` |
| [FT-005](FT-005/) | Forum | `done` |
| [FT-006](FT-006/) | News list | `done` |
| [FT-007](FT-007/) | Admin-панель — инфраструктура | `done` |
| [FT-008](FT-008/) | Admin — управление статьями | `done` |
| [FT-009](FT-009/) | Admin — загрузка изображений | `done` |
| [FT-010](FT-010/) | Admin — управление форумом | `done` |
| [FT-011](FT-011/) | Admin — управление пользователями | `done` |
| [FT-012](FT-012/) | fix: навигация «Новости» и 500 на /news | `done` |
| [FT-013](FT-013/) | fix: форум не отображает залогиненного пользователя | `done` |
| [FT-014](FT-014/) | fix: дублирование ошибок при логине (HTMX) | `done` |
| [FT-015](FT-015/) | fix: UX редактора статей (превью + фидбек сохранения) | `done` |
| [FT-016](FT-016/) | feat: real-time обновление постов форума (SSE) | `done` |
| [FT-017](FT-017/) | fix: позиция цитирующего комментария на форуме | `done` |
| [FT-018](FT-018/) | feat: блок «Ближайший матч» на главной странице | `done` |
| [FT-019](FT-019/) | feat: блок «Последний матч» на главной странице | `done` |
| [FT-020](FT-020/) | feat: таблица чемпионата АПЛ на главной странице | `done` |
| [FT-021](FT-021/) | feat: редизайн главной страницы | `done` |
| [FT-022](FT-022/) | fix: некорректное отображение даты матча на главной | `done` |
| [FT-023](FT-023/) | WYSIWYG-редактор статей | `done` |
| [FT-024](FT-024/) | Профиль пользователя (quick view + страница + аватар) | `planned` |

## Naming

- Базовый формат: `FT-XXX/`
- Вместо `XXX` используй идентификатор, принятый в проекте: issue id, ticket id или другой стабильный ключ
- Один package = одна delivery-единица
