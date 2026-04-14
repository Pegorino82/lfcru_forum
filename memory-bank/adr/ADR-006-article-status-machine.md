---
title: "ADR-006: Статусная машина статьи — draft / in_review / published"
doc_kind: adr
doc_function: canonical
purpose: "Фиксирует предлагаемое решение о замене булевого is_published статусным enum-полем для поддержки workflow черновик → ревью → публикация."
derived_from:
  - ../features/FT-008/feature.md
  - ../use-cases/UC-001-article-publishing.md
status: active
decision_status: accepted
date: 2026-04-13
accepted_date: 2026-04-14
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

## Решение (accepted 2026-04-14)

1. Добавить PostgreSQL enum `CREATE TYPE news_status AS ENUM ('draft', 'in_review', 'published')`.
2. Добавить колонку `status news_status NOT NULL DEFAULT 'draft'`.
3. Выполнить data migration: `UPDATE news SET status = CASE WHEN is_published THEN 'published' ELSE 'draft' END`.
4. Удалить `is_published` полностью — колонка из БД и поле из Go-модели убираются без замены. Нигде в коде не используется.
5. Обновить все запросы `WHERE is_published = true` → `WHERE status = 'published'`.
6. Добавить `reviewer_id BIGINT REFERENCES users(id)` — кто взял на ревью (nullable). Тип BIGINT соответствует `users.id`.

**Дополнительные принятые решения:**
- **OQ-01 закрыт:** формат текста статьи — Markdown. Рендерер: `github.com/yuin/goldmark`.
- **OQ-02 закрыт:** создание статьи всегда сохраняет `draft`. Прямая публикация из формы создания исключена — переход в `published` только через явный статусный переход. Это исключает случайную публикацию сырой статьи.

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
