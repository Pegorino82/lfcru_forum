---
title: "ADR-006: Статусная машина статьи — draft / in_review / published"
doc_kind: adr
doc_function: canonical
purpose: "Фиксирует предлагаемое решение о замене булевого is_published статусным enum-полем для поддержки workflow черновик → ревью → публикация."
derived_from:
  - ../features/FT-008/feature.md
  - ../use-cases/UC-001-article-publishing.md
status: active
decision_status: proposed
date: 2026-04-13
audience: humans_and_agents
must_not_define:
  - current_system_state
  - implementation_plan
---

# ADR-006: Статусная машина статьи — draft / in_review / published

## Контекст

Текущая таблица `news` использует булевый флаг `is_published` (`true`/`false`). Для admin-панели требуется workflow:
- **draft** — черновик, не виден публично
- **in_review** — отправлен на ревью другому Admin/Moderator, не виден публично
- **published** — опубликован, виден всем

Существующий флаг `is_published` не способен выразить промежуточный статус `in_review`. Нужно расширить модель.

## Драйверы решения

- UC-001 требует 3 различимых статуса.
- FT-006 (список статей) сейчас фильтрует `WHERE is_published = true` — нужна совместимость или миграция.
- Простота: enum в PostgreSQL (`news_status`) vs отдельная таблица статусов vs несколько булевых флагов.

## Рассмотренные варианты

| Вариант | Плюсы | Минусы | Почему рассмотрен |
| --- | --- | --- | --- |
| Заменить `is_published` на `status TEXT` с CHECK constraint | Простота, один столбец, легко мигрировать | TEXT без enum — нет type-safety на уровне БД | Кандидат |
| PostgreSQL enum `news_status` | Type-safety, нет недопустимых значений | Сложнее добавлять статусы (ALTER TYPE) | **Предлагается как основной** |
| Несколько булевых флагов (`is_draft`, `is_in_review`, `is_published`) | Читаемо | Возможны взаимоисключающие конфликты, много флагов | Отклонён |
| Отдельная таблица `article_status_history` | История смен статуса | Избыточно для текущего масштаба | Отклонён |

## Решение (proposed)

Предлагается:
1. Добавить PostgreSQL enum `CREATE TYPE news_status AS ENUM ('draft', 'in_review', 'published')`.
2. Добавить колонку `status news_status NOT NULL DEFAULT 'draft'`.
3. Выполнить data migration: `UPDATE news SET status = CASE WHEN is_published THEN 'published' ELSE 'draft' END`.
4. Убрать `is_published` (или оставить как computed column для обратной совместимости — решается в FT-008).
5. Обновить все запросы `WHERE is_published = true` → `WHERE status = 'published'`.

Добавить `reviewer_id UUID REFERENCES users(id)` — кто взял на ревью (nullable).

**Это решение требует человеческого approval (AG-*) в FT-008, так как меняет схему БД и затрагивает FT-006.**

## Последствия

### Положительные

- Явная, расширяемая модель статусов.
- Type-safety на уровне БД.
- Чистая data migration через goose.

### Отрицательные

- Breaking change: все места, использующие `is_published`, нужно обновить (FT-006 handler/repo/tests).
- PostgreSQL enum сложно расширять (нужен `ALTER TYPE ... ADD VALUE`).

### Нейтральные / организационные

- FT-006 implementation-plan должен быть проверен на совместимость с новой схемой.
- `internal/news/` — обновить модели, запросы, тесты.

## Риски и mitigation

- **Регрессия FT-006** — список новостей перестанет работать, если не обновить запрос. Mitigation: migration + обновление запроса + регрессионный тест в одном PR.
- **Необратимость** — удаление `is_published` нельзя откатить без потери данных. Mitigation: AG-* gate перед выполнением.

## Follow-up

- После принятия: создать goose-миграцию в FT-008.
- Обновить `internal/news/` модели и репозиторий.
- Обновить FT-006 (или проверить, что существующий код покрыт regression test).

## Связанные ссылки

- `memory-bank/use-cases/UC-001-article-publishing.md`
- `memory-bank/features/FT-008/feature.md`
- `memory-bank/features/FT-006/feature.md` — затрагивается migration
