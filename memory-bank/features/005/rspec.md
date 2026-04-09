# 05_forum_structure — Спецификация (SDD) [APPROVED 2026-04-08]

## 1. Контекст

Главная страница уже отображает блок «Активность форума», но ссылки ведут в никуда: страниц `/forum`, `/forum/sections/:id`, `/forum/topics/:id` не существует, гость и пользователь не могут читать или писать на форуме. Цель — реализовать базовую функциональность форума «разделы → темы → посты» с ролевым доступом.

**Текущее состояние кода:**

- Миграция `005_create_forum_and_matches.sql` создаёт таблицы `forum_sections` (`id`, `title`, `sort_order`, `created_at`), `forum_topics` (`id`, `section_id`, `title`, `author_id`, `post_count`, `last_post_at`, `last_post_by`, `created_at`, `updated_at`), `forum_posts` (`id`, `topic_id`, `author_id`, `content`, `created_at`, `updated_at`). Триггер `forum_topics_update_last_post` поддерживает счётчик и «последнее сообщение» при INSERT/DELETE постов.
- `internal/forum/` содержит только `Repo.LatestActive` и `TopicWithLastAuthor` для главной.
- Роли хранятся в `users.role` (`user` | `moderator` | `admin`), но middleware проверки роли нет.
- Поддержка цитирования постов (`parent_id`, snapshot полей) отсутствует — будет добавлена аналогично feature 004.
- В `forum_sections` отсутствуют колонки `description` и `topic_count` — требуются брифом.

**Ссылка на бриф:** [brief.md](brief.md)
**GitHub Issue:** [#5](https://github.com/Pegorino82/lfcru_forum/issues/5)

> **Первичная настройка (первый деплой):** Создание разделов и тем требует роли `moderator` или `admin`. После первой регистрации на сайте назначьте себе роль вручную:
> ```sql
> UPDATE users SET role = 'admin' WHERE email = 'your@email.com';
> ```

---

## 2. Акторы

| Актор | Описание | Права |
|---|---|---|
| Гость | Незарегистрированный посетитель | Читает разделы, темы, посты |
| User (`role = 'user'`) | Зарегистрированный пользователь | Всё, что Гость + создание постов, ответ на пост с цитатой |
| Moderator (`role = 'moderator'`) | Модератор | Всё, что User + создание разделов и тем |
| Admin (`role = 'admin'`) | Администратор | То же, что Moderator (в рамках этой фичи) |

> В данной фиче Moderator и Admin обладают одинаковыми правами. Разделение оставлено на будущие фичи (модерация контента, бан пользователей).

---

## 3. Сценарии использования (Use Cases)

### UC-1: Просмотр списка разделов

**Актор:** Гость / User / Moderator / Admin

**Основной поток:**

1. Пользователь переходит на `/forum`
2. Система загружает все разделы, отсортированные по `sort_order ASC, id ASC`
3. Для каждого раздела отображаются: название, описание, счётчик тем (`topic_count`)
4. Клик по разделу ведёт на `/forum/sections/{id}`

**Альтернативные потоки:**

- **2a.** Разделов нет → отображается пустое состояние «Разделы пока не созданы»

**Acceptance Criteria:**

- **Given** в БД существуют 3 раздела, **When** пользователь открывает `/forum`, **Then** отображаются 3 раздела в порядке `sort_order`
- **Given** в БД нет разделов, **When** пользователь открывает `/forum`, **Then** отображается пустое состояние
- Каждый раздел в списке содержит название, описание и `topic_count`
- Страница доступна гостям (HTTP 200)
- Moderator/Admin дополнительно видит кнопку «Создать раздел»

### UC-2: Просмотр раздела со списком тем

**Актор:** Гость / User / Moderator / Admin

**Предусловие:** Раздел существует.

**Основной поток:**

1. Пользователь переходит на `/forum/sections/{id}`
2. Система загружает раздел по `id`
3. Система загружает темы раздела, отсортированные по `last_post_at DESC NULLS LAST, id DESC`
4. Для каждой темы отображаются: название, автор (username), дата последнего поста (или «—», если постов нет), счётчик постов (`post_count`)
5. Moderator/Admin дополнительно видит кнопку «Создать тему»

**Альтернативные потоки:**

- **2a.** Раздел не существует → HTTP 404 «Раздел не найден»
- **2b.** `id` не число → HTTP 404
- **3a.** Тем нет → пустое состояние «В разделе пока нет тем»

**Acceptance Criteria:**

- **Given** раздел с `id=1` существует и содержит 5 тем, **When** пользователь открывает `/forum/sections/1`, **Then** отображаются 5 тем, отсортированных по последней активности
- **Given** раздел не существует, **When** пользователь открывает `/forum/sections/999`, **Then** HTTP 404
- **Given** `/forum/sections/abc`, **Then** HTTP 404
- Каждая тема показывает название, автора, дату последнего поста (или «—», если постов нет), счётчик постов
- Страница доступна гостям
- Кнопка «Создать тему» видна только Moderator/Admin

### UC-3: Просмотр темы со списком постов

**Актор:** Гость / User / Moderator / Admin

**Предусловие:** Тема существует.

**Основной поток:**

1. Пользователь переходит на `/forum/topics/{id}`
2. Система загружает тему по `id`
3. Система загружает все посты темы, отсортированные по `created_at ASC, id ASC`, лимит 500
4. Для каждого поста отображается: имя автора (`username`), дата/время создания, текст поста
5. Если пост — ответ, отображается блок-цитата: имя автора родителя + первые 100 рун текста родителя, с якорной ссылкой `#post-{parent_id}`
6. Под списком постов отображается форма создания поста (для авторизованных)

**Альтернативные потоки:**

- **2a.** Тема не существует / `id` не число → HTTP 404
- **3a.** Постов нет → пустое состояние «В теме пока нет сообщений»
- **6a.** Гость → вместо формы текст «Войдите, чтобы оставить сообщение» со ссылкой на `/login?next=/forum/topics/{id}`

**Acceptance Criteria:**

- **Given** тема с `id=1` и 3 постами, **When** пользователь открывает `/forum/topics/1`, **Then** отображаются 3 поста в хронологическом порядке
- **Given** тема не существует, **Then** HTTP 404
- Каждый пост показывает username автора, дату/время и текст
- Пост-ответ содержит цитату с автором и фрагментом текста родителя
- Если родительский пост удалён через `SET NULL` — цитата без `<a>`, текст «[сообщение удалено]»
- Страница доступна гостям
- Форма создания поста видна только авторизованным
- Якоря постов: `<div id="post-{id}">`

### UC-4: Создание поста в теме

**Актор:** User / Moderator / Admin

**Предусловие:** Пользователь авторизован и находится на странице темы.

**Основной поток:**

1. Пользователь вводит текст в `<textarea name="content">` формы
2. Пользователь нажимает «Отправить»
3. Система валидирует текст (см. §4)
4. Система сохраняет пост (`topic_id`, `author_id`, `content`, `parent_id = NULL`)
5. Триггер `forum_topics_update_last_post` обновляет `post_count`, `last_post_at`, `last_post_by`
6. Сервер возвращает обновлённый блок `#posts-list` (HTMX swap `innerHTML`) с новым постом в конце + заголовок `HX-Trigger: postAdded`

**Альтернативные потоки:**

- **3a.** Пустой content → HTTP 422, «Сообщение не может быть пустым»
- **3b.** `utf8.RuneCountInString(content) > 20000` → HTTP 422, «Сообщение слишком длинное (максимум 20 000 символов)»
- **Гость:** POST-запрос → `RequireAuth` редиректит на `/login?next=/forum/topics/{id}` (HTTP 302)
- Невалидный CSRF-токен → HTTP 403 (стандартный CSRF-middleware)

**Acceptance Criteria:**

- **Given** авторизованный пользователь на странице темы, **When** вводит текст и отправляет форму, **Then** пост появляется в конце списка без перезагрузки
- **Given** гость отправляет POST на `/forum/topics/1/posts`, **Then** HTTP 302 на `/login?next=/forum/topics/1`
- **Given** пользователь отправляет пустой content, **Then** HTTP 422 «Сообщение не может быть пустым»
- **Given** content > 20 000 рун, **Then** HTTP 422 «Сообщение слишком длинное»
- После успешного создания `forum_topics.post_count` увеличился на 1, `last_post_at` и `last_post_by` обновлены (через триггер)
- Форма требует CSRF-токен

### UC-5: Ответ-цитата на пост

**User Story:** Как авторизованный пользователь, я хочу ответить конкретному сообщению с цитатой, чтобы мой ответ был понятен в контексте диалога, даже если между сообщениями появились другие посты.

**Актор:** User / Moderator / Admin

**Предусловие:** Авторизованный пользователь на странице темы, в теме есть хотя бы один пост.

**Основной поток:**

1. Пользователь нажимает кнопку «Ответить» у поста
2. Alpine.js показывает форму ответа с цитатой (username + первые 100 рун родителя), скрывая другие открытые формы ответа
3. Пользователь вводит текст и нажимает «Отправить»
4. Форма отправляет `POST /forum/topics/{id}/posts` с `parent_id = {post_id}`
5. Сервис валидирует текст (§4), открывает транзакцию, проверяет существование родителя в той же теме и что `parent_id родителя IS NULL` (depth ≤ 1), заполняет snapshot-поля, вставляет пост
6. Сервер возвращает обновлённый блок `#posts-list`, ответ отображается в хронологической позиции с цитатой родителя

**Альтернативные потоки:**

- **5a.** Родитель не найден или из другой темы → HTTP 422, «Сообщение, на которое вы отвечаете, не найдено»
- **5b.** Родитель сам является ответом (`parent_id IS NOT NULL`) → HTTP 422, «Нельзя отвечать на ответ» (depth ≤ 1)
- **5c.** Родитель удалён между проверкой и INSERT → PG вернёт FK `23503`, маппится в `ErrParentNotFound` → HTTP 422
- **2a.** Нажатие «Отмена» скрывает форму ответа

**Acceptance Criteria:**

- **Given** авторизованный пользователь нажимает «Ответить» на пост, **Then** появляется форма с цитатой
- **Given** форма ответа открыта на одном посте, пользователь нажимает «Ответить» на другом, **Then** первая форма скрывается, открывается новая
- **Given** пользователь отправляет ответ на корневой пост, **Then** ответ сохраняется с `parent_id`, `parent_author_snapshot`, `parent_content_snapshot` заполнены
- **Given** пользователь пытается ответить на пост-ответ (`parent_id IS NOT NULL`), **Then** HTTP 422
- **Given** родительский пост из другой темы, **Then** HTTP 422
- Гость не видит кнопку «Ответить»

### UC-6: Создание раздела

**Актор:** Moderator / Admin

**Предусловие:** Пользователь авторизован и имеет роль `moderator` или `admin`.

**Основной поток:**

1. Пользователь на `/forum` нажимает «Создать раздел» → переход на `/forum/sections/new`
2. Система отображает форму: `title`, `description`, `sort_order` (опционально, по умолчанию 0)
3. Пользователь заполняет форму и отправляет `POST /forum/sections`
4. Сервис валидирует (§4), вставляет запись в `forum_sections`
5. Редирект 303 на `/forum`

**Альтернативные потоки:**

- **Гость:** HTTP 302 на `/login?next=/forum/sections/new`
- **User:** HTTP 403 «Недостаточно прав»
- **4a.** Ошибка валидации → HTTP 422, форма с сообщением и сохранёнными значениями

**Acceptance Criteria:**

- **Given** модератор на форме создания раздела, заполняет валидные данные, **Then** раздел создан, редирект на `/forum`, новый раздел виден в списке
- **Given** обычный пользователь открывает `/forum/sections/new`, **Then** HTTP 403
- **Given** гость открывает `/forum/sections/new`, **Then** HTTP 302 на `/login?next=/forum/sections/new`
- **Given** модератор отправляет пустой `title`, **Then** HTTP 422 с сообщением об ошибке
- CSRF-токен обязателен

### UC-7: Создание темы в разделе

**Актор:** Moderator / Admin

**Предусловие:** Раздел существует, пользователь авторизован с ролью `moderator` или `admin`.

**Основной поток:**

1. Пользователь на `/forum/sections/{id}` нажимает «Создать тему» → переход на `/forum/sections/{id}/topics/new`
2. Система отображает форму: `title`
3. Пользователь отправляет `POST /forum/sections/{id}/topics`
4. Сервис валидирует (§4), вставляет запись в `forum_topics` (`section_id`, `title`, `author_id = current_user.id`)
5. Триггер увеличивает `forum_sections.topic_count` на 1
6. Редирект 303 на `/forum/topics/{new_id}`

**Альтернативные потоки:**

- **Гость:** HTTP 302 на `/login?next=/forum/sections/{id}/topics/new`
- **User:** HTTP 403
- **1a.** Раздел не существует → HTTP 404
- **4a.** Ошибка валидации → HTTP 422 с сохранением введённых значений

**Acceptance Criteria:**

- **Given** модератор создаёт тему в разделе с `id=1`, **Then** тема создана, `forum_sections.topic_count` увеличился на 1, редирект на страницу темы
- **Given** user на `/forum/sections/1/topics/new`, **Then** HTTP 403
- **Given** модератор отправляет пустой `title`, **Then** HTTP 422

### UC-8: Переходы с главной страницы

**Актор:** Гость / User / Moderator / Admin

**Предусловие:** Главная страница отображает блок «Активность форума» (feature 003).

**Основной поток:**

1. Пользователь видит список последних активных тем
2. Клик по теме ведёт на `/forum/topics/{id}`
3. Опционально — есть ссылка «Все разделы» на `/forum`

**Acceptance Criteria:**

- **Given** на главной отображается тема с `id=5`, **When** пользователь кликает по ней, **Then** открывается `/forum/topics/5`
- В шаблоне главной нет `href="#"` или ссылок, ведущих в никуда — все ведут на реальные роуты форума

### UC-9: Отказ в доступе

**Acceptance Criteria (сводно):**

- Гость получает 302 на `/login?next={original_path}` при попытке открыть любой `POST` или защищённый `GET` (формы создания)
- User получает 403 при попытке `GET /forum/sections/new`, `POST /forum/sections`, `GET /forum/sections/:id/topics/new`, `POST /forum/sections/:id/topics`
- 403-страница — брендированная, в рамках `base.html`, с сообщением «Недостаточно прав»

---

## 4. Правила валидации

### Section

| Поле | Правило | Сообщение об ошибке |
|---|---|---|
| title | После `TrimSpace` не пустой, длина ≤ 255 байт | «Название раздела не может быть пустым» / «Название раздела слишком длинное» |
| description | После `TrimSpace` длина ≤ 2000 байт (допустим пустой) | «Описание раздела слишком длинное» |
| sort_order | Целое число ≥ 0 (по умолчанию 0) | «Порядок сортировки должен быть неотрицательным» |

### Topic

| Поле | Правило | Сообщение об ошибке |
|---|---|---|
| title | После `TrimSpace` не пустой, длина ≤ 255 байт | «Название темы не может быть пустым» / «Название темы слишком длинное» |
| section_id | Существует в `forum_sections` | HTTP 404 |

### Post

| Поле | Правило | Сообщение об ошибке |
|---|---|---|
| content | После `TrimSpace` не пустой | «Сообщение не может быть пустым» |
| content | `utf8.RuneCountInString ≤ 20000` | «Сообщение слишком длинное (максимум 20 000 символов)» |
| topic_id | Существует в `forum_topics` | HTTP 404 |
| parent_id | Если указан — существует, принадлежит тому же `topic_id`, `parent.parent_id IS NULL` | «Сообщение, на которое вы отвечаете, не найдено» / «Нельзя отвечать на ответ» |

### Параметры URL

| Поле | Правило | Сообщение об ошибке |
|---|---|---|
| `:id` | Целое положительное число | HTTP 404 |

---

## 5. Модель данных

### 5.1. Миграция `007_forum_sections_description.sql`

Добавляет описание и денормализованный счётчик тем.

```sql
-- +goose Up
ALTER TABLE forum_sections
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    ADD COLUMN topic_count INT  NOT NULL DEFAULT 0;

-- Инициализация topic_count для существующих разделов
UPDATE forum_sections s
SET topic_count = (SELECT COUNT(*) FROM forum_topics t WHERE t.section_id = s.id);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION forum_sections_update_topic_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE forum_sections
        SET topic_count = topic_count + 1
        WHERE id = NEW.section_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE forum_sections
        SET topic_count = GREATEST(topic_count - 1, 0)
        WHERE id = OLD.section_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_forum_topics_update_section_count
AFTER INSERT OR DELETE ON forum_topics
FOR EACH ROW EXECUTE FUNCTION forum_sections_update_topic_count();

-- Индексы для критичных запросов форума
CREATE INDEX idx_forum_posts_topic ON forum_posts(topic_id, created_at ASC);
CREATE INDEX idx_forum_topics_section ON forum_topics(section_id, last_post_at DESC NULLS LAST);
CREATE INDEX idx_forum_posts_parent ON forum_posts(parent_id) WHERE parent_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_forum_posts_parent;
DROP INDEX IF EXISTS idx_forum_topics_section;
DROP INDEX IF EXISTS idx_forum_posts_topic;
DROP TRIGGER IF EXISTS trg_forum_topics_update_section_count ON forum_topics;
DROP FUNCTION IF EXISTS forum_sections_update_topic_count;
ALTER TABLE forum_sections
    DROP COLUMN IF EXISTS topic_count,
    DROP COLUMN IF EXISTS description;
```

### 5.2. Миграция `008_forum_posts_quotes.sql`

Добавляет поддержку цитирования постов.

```sql
-- +goose Up
ALTER TABLE forum_posts
    ADD COLUMN parent_id                BIGINT REFERENCES forum_posts(id) ON DELETE SET NULL,
    ADD COLUMN parent_author_snapshot   TEXT,
    ADD COLUMN parent_content_snapshot  TEXT;

-- +goose Down
ALTER TABLE forum_posts
    DROP COLUMN IF EXISTS parent_content_snapshot,
    DROP COLUMN IF EXISTS parent_author_snapshot,
    DROP COLUMN IF EXISTS parent_id;
```

> **Примечания:**
> - `parent_id` — самоссылка, `ON DELETE SET NULL`: при удалении родителя поле обнуляется, snapshot-данные сохраняются для отображения «[сообщение удалено]».
> - `parent_author_snapshot` — username автора родителя на момент ответа.
> - `parent_content_snapshot` — первые 100 рун контента родителя на момент ответа.
> - Счётчик и последняя активность темы уже обновляются триггером из миграции 005.

---

## 6. Структура Go-пакетов

### Изменения в `internal/forum/`

```
internal/forum/
├── model.go           # Section, Topic, Post, SectionView, TopicView, PostView (+ существующий TopicWithLastAuthor)
├── repo.go            # + ListSections, GetSection, ListTopicsBySection, GetTopic, ListPostsByTopic,
│                      #   CreateSection, CreateTopic, CreatePost
├── service.go         # NEW: Service — валидация Section/Topic/Post, mention-обработка НЕ используется
├── handler.go         # NEW: HTTP-хэндлеры
├── errors.go          # NEW: sentinel-ошибки
├── service_test.go    # NEW: юнит-тесты
├── repo_test.go       # + интеграционные тесты для новых методов (существует, расширить)
└── handler_test.go    # NEW: интеграционные тесты HTTP
```

### Модели

```go
// internal/forum/model.go

type Section struct {
    ID          int64
    Title       string
    Description string
    SortOrder   int
    TopicCount  int
    CreatedAt   time.Time
}

// Topic — существующая структура, без изменений.

type Post struct {
    ID                    int64
    TopicID               int64
    AuthorID              int64
    ParentID              *int64
    ParentAuthorSnapshot  *string
    ParentContentSnapshot *string
    Content               string
    CreatedAt             time.Time
    UpdatedAt             time.Time
}

// SectionView — проекция для списка разделов.
type SectionView struct {
    ID          int64
    Title       string
    Description string
    TopicCount  int
}

// TopicView — проекция для списка тем в разделе.
type TopicView struct {
    ID             int64
    Title          string
    AuthorID       int64
    AuthorUsername string
    PostCount      int
    LastPostAt     *time.Time
    CreatedAt      time.Time
}

// PostView — проекция для отображения поста.
type PostView struct {
    ID             int64
    TopicID        int64
    AuthorID       int64
    AuthorUsername string
    ParentID       *int64
    ParentAuthor   *string
    ParentSnippet  *string
    Content        string
    CreatedAt      time.Time
}
```

### Ошибки

```go
// internal/forum/errors.go
var (
    ErrSectionNotFound   = errors.New("forum: section not found")
    ErrTopicNotFound     = errors.New("forum: topic not found")
    ErrPostNotFound      = errors.New("forum: post not found")
    ErrParentNotFound    = errors.New("forum: parent post not found")
    ErrReplyToReply      = errors.New("forum: cannot reply to reply")

    ErrEmptyTitle        = errors.New("forum: empty title")
    ErrTitleTooLong      = errors.New("forum: title too long")
    ErrDescriptionTooLong= errors.New("forum: description too long")
    ErrEmptyContent      = errors.New("forum: empty content")
    ErrContentTooLong    = errors.New("forum: content too long")
)
```

---

## 7. Репозитории — SQL-запросы

Все запросы — параметризованные (`$1`, `$2`, ...). Методы принимают `context.Context` первым аргументом.

### `Repo.ListSections(ctx) ([]SectionView, error)`

```sql
SELECT id, title, description, topic_count
FROM forum_sections
ORDER BY sort_order ASC, id ASC
```

При пустом результате — `[]SectionView{}`, не `nil`.

### `Repo.GetSection(ctx, id int64) (*Section, error)`

```sql
SELECT id, title, description, sort_order, topic_count, created_at
FROM forum_sections
WHERE id = $1
```

Возвращает `nil, nil` если не найдено.

### `Repo.ListTopicsBySection(ctx, sectionID int64) ([]TopicView, error)`

```sql
SELECT
    t.id,
    t.title,
    t.author_id,
    COALESCE(u.username, '[удалён]') AS author_username,
    t.post_count,
    t.last_post_at,
    t.created_at
FROM forum_topics t
LEFT JOIN users u ON u.id = t.author_id
WHERE t.section_id = $1
ORDER BY t.last_post_at DESC NULLS LAST, t.id DESC
LIMIT 500
```

При пустом результате — `[]TopicView{}`.

### `Repo.GetTopic(ctx, id int64) (*Topic, error)`

```sql
SELECT id, section_id, title, author_id, post_count, last_post_at, last_post_by, created_at, updated_at
FROM forum_topics
WHERE id = $1
```

Возвращает `nil, nil` если не найдено.

### `Repo.ListPostsByTopic(ctx, topicID int64) ([]PostView, error)`

```sql
SELECT
    p.id,
    p.topic_id,
    p.author_id,
    COALESCE(u.username, '[удалён]') AS author_username,
    p.parent_id,
    p.parent_author_snapshot  AS parent_author,
    p.parent_content_snapshot AS parent_snippet,
    p.content,
    p.created_at
FROM forum_posts p
LEFT JOIN users u ON u.id = p.author_id
WHERE p.topic_id = $1
ORDER BY p.created_at ASC, p.id ASC
LIMIT 500
```

При пустом результате — `[]PostView{}`.

> **Примечания:**
> - `LEFT JOIN users` + `COALESCE` защищает на случай, если в будущем `author_id` может стать `NULL`. В текущей схеме `ON DELETE RESTRICT` — LEFT JOIN эквивалентен INNER JOIN, но менее хрупкий.
> - `LIMIT 500` — защитный потолок. Пагинация — вне scope.

### `Repo.CreateSection(ctx, s *Section) (int64, error)`

```sql
INSERT INTO forum_sections (title, description, sort_order)
VALUES ($1, $2, $3)
RETURNING id
```

### `Repo.CreateTopic(ctx, t *Topic) (int64, error)`

```sql
INSERT INTO forum_topics (section_id, title, author_id)
VALUES ($1, $2, $3)
RETURNING id
```

При отсутствии `section_id` (FK violation `23503`) → `ErrSectionNotFound`.

### `Repo.CreatePost(ctx, p *Post) (int64, error)`

Если `p.ParentID == nil` — одиночный INSERT.

Если `p.ParentID != nil` — **транзакция**:

**Шаг 1 — проверка родителя и загрузка snapshot-данных:**

```sql
SELECT u.username, LEFT(fp.content, 100), fp.parent_id
FROM forum_posts fp
JOIN users u ON u.id = fp.author_id
WHERE fp.id = $1 AND fp.topic_id = $2
```

- Строка не найдена → `ErrParentNotFound` (родитель не существует или из другой темы)
- `fp.parent_id IS NOT NULL` → `ErrReplyToReply` (depth ≤ 1)
- Заполнить `p.ParentAuthorSnapshot` и `p.ParentContentSnapshot`

**Шаг 2 — INSERT:**

```sql
INSERT INTO forum_posts
    (topic_id, author_id, parent_id, parent_author_snapshot, parent_content_snapshot, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id
```

Маппинг ошибок PG:
- `23503` (FK violation на `parent_id` из-за race condition с удалением родителя) → `ErrParentNotFound`
- `23503` на `topic_id` → `ErrTopicNotFound`

> Триггер `forum_topics_update_last_post` обновляет `post_count`, `last_post_at`, `last_post_by` автоматически при INSERT.

---

## 8. Сервисный слой

### `forum.Service`

```go
type Service struct {
    repo *Repo
}

func NewService(repo *Repo) *Service { return &Service{repo: repo} }
```

### `Service.ListSections(ctx) ([]SectionView, error)`

Делегирует `repo.ListSections`.

### `Service.GetSectionWithTopics(ctx, id int64) (*Section, []TopicView, error)`

1. `section, err := repo.GetSection(ctx, id)`; если `nil` → `ErrSectionNotFound`
2. `topics, err := repo.ListTopicsBySection(ctx, id)`
3. Вернуть `(section, topics, nil)`

### `Service.GetTopicWithPosts(ctx, id int64) (*Topic, []PostView, error)`

1. `topic, err := repo.GetTopic(ctx, id)`; если `nil` → `ErrTopicNotFound`
2. `posts, err := repo.ListPostsByTopic(ctx, id)`
3. Вернуть `(topic, posts, nil)`

### `Service.CreateSection(ctx, title, description string, sortOrder int) (int64, error)`

1. `title = strings.TrimSpace(title)`; если пустой → `ErrEmptyTitle`
2. `len(title) > 255` → `ErrTitleTooLong`
3. `description = strings.TrimSpace(description)`
4. `len(description) > 2000` → `ErrDescriptionTooLong`
5. `sortOrder < 0` → ошибка (опционально — clamp в 0)
6. `repo.CreateSection(ctx, &Section{...})`

### `Service.CreateTopic(ctx, sectionID, authorID int64, title string) (int64, error)`

1. `title = TrimSpace`; пустой → `ErrEmptyTitle`; длина > 255 → `ErrTitleTooLong`
2. `repo.CreateTopic(ctx, &Topic{SectionID: sectionID, AuthorID: authorID, Title: title})`
3. FK violation → `ErrSectionNotFound`

### `Service.CreatePost(ctx, topicID, authorID int64, parentID *int64, content string) (int64, error)`

1. `content = TrimSpace`; пустой → `ErrEmptyContent`
2. `utf8.RuneCountInString(content) > 20000` → `ErrContentTooLong`
3. `repo.CreatePost(ctx, &Post{TopicID: topicID, AuthorID: authorID, ParentID: parentID, Content: content})`
4. Пробросить `ErrParentNotFound`, `ErrReplyToReply`, `ErrTopicNotFound` наружу без изменений

> **Проверка ролей — НЕ в сервисе.** Ролевая защита реализована через middleware `RequireRole` перед вызовом хэндлера. Сервис доверяет вызывающему.

---

## 9. API-контракты

Все эндпоинты рендерят HTML (server-side rendering). При заголовке `HX-Request: true` возвращается HTML-фрагмент вместо полной страницы.

### Сводная таблица

| Метод | Путь | Auth | Роль | Успех | Ошибки |
|---|---|---|---|---|---|
| GET | `/forum` | Нет | — | 200 HTML | 500 |
| GET | `/forum/sections/:id` | Нет | — | 200 HTML | 404, 500 |
| GET | `/forum/topics/:id` | Нет | — | 200 HTML | 404, 500 |
| GET | `/forum/sections/new` | Да | moderator, admin | 200 HTML | 302 (guest), 403, 500 |
| POST | `/forum/sections` | Да | moderator, admin | 303 → `/forum` | 302, 403, 422, 500 |
| GET | `/forum/sections/:id/topics/new` | Да | moderator, admin | 200 HTML | 302, 403, 404, 500 |
| POST | `/forum/sections/:id/topics` | Да | moderator, admin | 303 → `/forum/topics/:new_id` | 302, 403, 404, 422, 500 |
| POST | `/forum/topics/:id/posts` | Да | user, moderator, admin | 201 (HTMX partial) / 303 → `/forum/topics/:id#post-:new_id` (без HTMX) | 302, 404, 422, 500 |

### GET /forum

**Данные шаблона:**
```go
type ForumIndexData struct {
    User      *user.User
    CSRFToken string
    Sections  []SectionView
    CanManage bool // User.Role in {moderator, admin}
}
```

### GET /forum/sections/:id

**Данные шаблона:**
```go
type SectionPageData struct {
    User      *user.User
    CSRFToken string
    Section   *Section
    Topics    []TopicView
    CanManage bool
}
```

### GET /forum/topics/:id

**Данные шаблона:**
```go
type TopicPageData struct {
    User        *user.User
    CSRFToken   string
    Topic       *Topic
    Posts       []PostView
    FormError   string
    FormContent string
}
```

### POST /forum/topics/:id/posts

**Запрос:** `application/x-www-form-urlencoded`

| Параметр | Тип | Обязательно | Описание |
|---|---|---|---|
| content | string | да | Текст поста |
| parent_id | int64 | нет | ID родительского поста |
| _csrf | string | да | CSRF-токен |

**Ответ (HTMX):**

- 201 + обновлённый HTML блока `#posts-list` + `HX-Trigger: postAdded`
- 422 + HTML формы с ошибкой + `HX-Retarget: #post-form`, `HX-Reswap: innerHTML`

**Ответ (без HTMX):**

- 303 → `/forum/topics/{id}#post-{new_id}`
- 422 → полная страница темы с `FormError` и `FormContent`

### POST /forum/sections, POST /forum/sections/:id/topics

Стандартные формы без HTMX. При успехе — 303 redirect. При 422 — полная страница формы с сохранёнными значениями и сообщением.

### Middleware

- `LoadSession` — глобально, загружает `*user.User` в контекст
- `CSRF` — глобально для POST/PUT/DELETE
- `RequireAuth` — на `POST /forum/topics/:id/posts`, `GET/POST /forum/sections/new`, `GET/POST /forum/sections/:id/topics/new`
- `RequireRole("moderator", "admin")` — NEW, поверх `RequireAuth`, на управление структурой
- Гости: `RequireAuth` → 302 на `/login?next={path}`
- Нехватка роли: `RequireRole` → 403 с брендированной страницей

---

## 10. Шаблоны

### Файлы

```
templates/forum/
├── index.html           # /forum — список разделов
├── section.html         # /forum/sections/:id — раздел + темы
├── topic.html           # /forum/topics/:id — тема + посты + форма
├── new_section.html     # /forum/sections/new
└── new_topic.html       # /forum/sections/:id/topics/new
```

### Целевая HTML-структура `topic.html`

Аналогично `templates/news/article.html` из feature 004. Ключевые блоки:

```html
{{define "content"}}
<article class="forum-topic">
  <header class="topic-header">
    <nav class="breadcrumbs" aria-label="Навигация">
      <a href="/forum">Форум</a> ›
      <a href="/forum/sections/{{.Topic.SectionID}}">Раздел</a> ›
      <span>{{.Topic.Title}}</span>
    </nav>
    <h1>{{.Topic.Title}}</h1>
  </header>

  <section class="posts-section" x-data="{ replyTo: null, replyAuthor: '' }"
           @reset-reply.window="replyTo = null; replyAuthor = ''">
    <div id="posts-list">
      {{if .Posts}}
        {{range .Posts}}
        <div class="post" id="post-{{.ID}}">
          {{if .ParentAuthor}}
            {{if .ParentID}}
            <a href="#post-{{.ParentID}}" class="post-quote" data-post-id="{{.ParentID}}">
              <span class="quote-author">{{deref .ParentAuthor}}</span>
              <span class="quote-text">{{if .ParentSnippet}}{{deref .ParentSnippet}}{{end}}</span>
            </a>
            {{else}}
            <div class="post-quote post-quote--deleted">
              <span class="quote-author">[удалён]</span>
              <span class="quote-text">[сообщение удалено]</span>
            </div>
            {{end}}
          {{end}}
          <div class="post-header">
            <span class="post-author">{{.AuthorUsername}}</span>
            <time class="post-date" datetime="{{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}">
              {{.CreatedAt.Format "02.01.2006 15:04"}}
            </time>
          </div>
          <div class="post-body">{{.Content}}</div>
          {{if $.User}}
          <button type="button"
                  class="post-reply-btn"
                  data-post-id="{{.ID}}"
                  data-post-author="{{.AuthorUsername}}"
                  aria-label="Ответить {{.AuthorUsername}}"
                  @click="replyTo = $el.dataset.postId; replyAuthor = $el.dataset.postAuthor">
            Ответить
          </button>
          {{end}}
        </div>
        {{end}}
      {{else}}
        <p class="empty-state">В теме пока нет сообщений.</p>
      {{end}}
    </div>

    {{if .User}}
    <!-- Форма ответа (reply) -->
    <template x-if="replyTo !== null">
      <form class="reply-form"
            id="post-form"
            hx-post="/forum/topics/{{.Topic.ID}}/posts"
            hx-target="#posts-list"
            hx-swap="innerHTML"
            hx-disabled-elt="find button[type=submit]">
        <input type="hidden" name="_csrf" value="{{.CSRFToken}}">
        <input type="hidden" name="parent_id" :value="replyTo">
        <div class="reply-quote">Ответ для <span x-text="replyAuthor"></span></div>
        <textarea name="content" rows="3" required maxlength="20000"
                  placeholder="Ваш ответ..." aria-label="Текст ответа"></textarea>
        <div class="form-actions">
          <button type="submit">Отправить</button>
          <button type="button" @click="replyTo = null">Отмена</button>
        </div>
      </form>
    </template>

    <!-- Основная форма создания поста -->
    <form class="post-form"
          id="post-form"
          hx-post="/forum/topics/{{.Topic.ID}}/posts"
          hx-target="#posts-list"
          hx-swap="innerHTML"
          hx-disabled-elt="find button[type=submit]">
      <input type="hidden" name="_csrf" value="{{.CSRFToken}}">
      {{if .FormError}}<p class="form-error" role="alert">{{.FormError}}</p>{{end}}
      <textarea name="content" rows="4" required maxlength="20000"
                placeholder="Ваше сообщение..." aria-label="Текст сообщения">{{.FormContent}}</textarea>
      <button type="submit">Отправить</button>
    </form>
    {{else}}
    <p class="login-prompt">
      <a href="/login?next=/forum/topics/{{.Topic.ID}}">Войдите</a>, чтобы оставить сообщение
    </p>
    {{end}}
  </section>
</article>
{{end}}
```

### Правила отображения

| Элемент | Формат |
|---|---|
| Дата/время поста | `02.01.2006 15:04` |
| `<time datetime>` | ISO 8601 с таймзоной |
| Цитата `parent_snippet` | Первые 100 рун + `…` если был обрезан |
| Якорь поста | `id="post-{id}"` |
| Подсветка `:target` | CSS `animation` 2s (аналогично 004) |

### Шаблон `new_section.html` / `new_topic.html`

Простые формы без HTMX: `method="POST"`, `_csrf`, поля ввода, сообщение об ошибке `FormError`, сохранённые значения.

---

## 11. Стилизация

Минимальные стили в `<style>` блоках шаблонов, консистентные с feature 004 (красный акцент `#c8102e`).

| Элемент | Стили |
|---|---|
| `.forum-topic`, `.forum-section` | `max-width: 900px; margin: 0 auto;` |
| `.breadcrumbs` | `font-size: 0.875rem; color: #666; margin-bottom: 0.5rem;` |
| `.section-card`, `.topic-row` | `padding: 0.75rem; border-bottom: 1px solid #eee;` |
| `.topic-count`, `.post-count` | `color: #666; font-size: 0.875rem;` |
| `.post` | `padding: 1rem 0; border-bottom: 1px solid #eee;` |
| `.post-author` | `font-weight: bold;` |
| `.post-quote` | `display: block; background: #f5f5f5; border-left: 3px solid #c8102e; padding: 0.5rem 0.75rem; margin-bottom: 0.5rem; font-size: 0.875rem; color: #555; text-decoration: none;` |
| `.post-quote--deleted` | `opacity: 0.6; cursor: default;` |
| `.post:target` | `animation: highlight 2s ease-out;` |
| Кнопки `.post-form button`, `.reply-form button[type=submit]` | `background: #c8102e; color: #fff; border: none; padding: 0.5rem 1rem; border-radius: 4px;` |
| `.login-prompt` | `color: #666; font-style: italic;` |

### Mobile (<768px)

- `.forum-topic { padding: 0 1rem; }`
- Кнопки: `min-height: 44px;` (touch target)
- `textarea { font-size: 1rem; }` (против iOS zoom)

---

## 12. Accessibility

| Требование | Реализация |
|---|---|
| Семантика | `<article>`, `<section>`, `<nav aria-label>` для breadcrumbs |
| Заголовки | `<h1>` для темы/раздела, `<h2>` для подсекций |
| Машиночитаемые даты | `<time datetime="ISO 8601">` |
| Якоря | `id="post-{id}"`, `id="topic-{id}"` |
| Формы | `<textarea aria-label>`, `required`, `maxlength` |
| Ошибки | `role="alert"` на `.form-error` |
| Ссылки навигации | Стандартные `<a href="#...">` |

---

## 13. HTMX и Alpine.js

### HTMX

- `POST /forum/topics/:id/posts`:
  - Успех: возвращает обновлённый `#posts-list` + `HX-Trigger: postAdded`
  - Ошибка (422): возвращает форму с ошибкой + `HX-Retarget: #post-form`, `HX-Reswap: innerHTML`
- Формы создания раздела/темы — без HTMX (стандартные POST + redirect)
- Страницы `GET /forum/...` при `HX-Request: true` возвращают только блок `content`

### Alpine.js

Секция постов — обёртка с `x-data="{ replyTo: null, replyAuthor: '' }"`.

- `replyTo` — ID поста-родителя (`null` = форма ответа скрыта)
- `replyAuthor` — username автора родителя (отображается в цитате формы)
- «Ответить»: `@click="replyTo = {id}; replyAuthor = '{username}'"`
- «Отмена»: `@click="replyTo = null"`
- Форма ответа: `x-show="replyTo !== null"` (или `<template x-if>`)

**Сброс формы после успешного POST:**

Сервер отдаёт `HX-Trigger: postAdded`. Хук:

```javascript
htmx.on('#posts-list', 'postAdded', function() {
    window.dispatchEvent(new Event('reset-reply'));
});
```

Обёртка слушает: `@reset-reply.window="replyTo = null; replyAuthor = ''"`.

> **Известная особенность (из памяти проекта):** Alpine не обновляет `x-show` реактивно для элементов, вставленных HTMX через `outerHTML` swap. Сбрасывать состояние в `htmx:before-swap`, не `htmx:after-request`. В данной фиче используется `innerHTML` swap `#posts-list`, обёртка Alpine остаётся, событие `postAdded` сбрасывает состояние — проблема не возникает, но соблюдать паттерн.

> **`hx-disabled-elt="find button[type=submit]"`** — защита от двойной отправки.

> **Разделение HTMX/Alpine:** на одном элементе `hx-*` и `x-*` не смешиваются. HTMX — на формах, Alpine — на обёртке секции.

---

## 14. Изменения в существующих файлах

### `internal/auth/middleware.go`

Добавить:

```go
// RequireRole — middleware, разрешающее доступ только пользователям с одной из указанных ролей.
// Должен располагаться ПОСЛЕ RequireAuth в цепочке.
func RequireRole(roles ...string) echo.MiddlewareFunc {
    allowed := make(map[string]struct{}, len(roles))
    for _, r := range roles {
        allowed[r] = struct{}{}
    }
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            u := UserFromContext(c)
            if u == nil {
                // Страховка: RequireAuth должен был сработать раньше
                return c.Redirect(http.StatusFound, "/login?next="+url.QueryEscape(c.Request().URL.RequestURI()))
            }
            if _, ok := allowed[u.Role]; !ok {
                return c.Render(http.StatusForbidden, "templates/errors/403.html", ...)
            }
            return next(c)
        }
    }
}
```

### `internal/forum/repo.go`

Расширить новыми методами из §7. `LatestActive` остаётся без изменений.

### `cmd/forum/main.go`

1. Создание `forum.Service`
2. Создание `forum.Handler` с зависимостями `(service, renderer)`
3. Регистрация роутов:
   ```go
   // ВАЖНО: статические маршруты (/forum/sections/new) регистрировать ДО параметрических
   // (/forum/sections/:id) — Echo приоритизирует статику над параметром автоматически,
   // но порядок вызовов e.GET / Group должен это обеспечивать.
   // modGroup регистрируется первым, поэтому /forum/sections/new встаёт раньше :id.

   modGroup := e.Group("", auth.RequireAuth, auth.RequireRole("moderator", "admin"))
   modGroup.GET("/forum/sections/new", forumHandler.NewSection)       // статический — первым
   modGroup.POST("/forum/sections", forumHandler.CreateSection)
   modGroup.GET("/forum/sections/:id/topics/new", forumHandler.NewTopic)
   modGroup.POST("/forum/sections/:id/topics", forumHandler.CreateTopic)

   e.GET("/forum", forumHandler.Index)
   e.GET("/forum/sections/:id", forumHandler.ShowSection)             // параметрический — после
   e.GET("/forum/topics/:id", forumHandler.ShowTopic)

   authGroup := e.Group("", auth.RequireAuth)
   authGroup.POST("/forum/topics/:id/posts", forumHandler.CreatePost)
   ```
4. Миграции 007 и 008 применяются автоматически через `goose.Up`

### `internal/home/handler.go` и шаблон главной страницы

Линки блока «Активность форума» должны вести на `/forum/topics/{id}` (сейчас — на несуществующий путь или `#`). Добавить ссылку «Все разделы» → `/forum`.

### `templates/layouts/base.html`

Добавить в навигацию ссылку «Форум» → `/forum` (если ещё нет).

### `internal/auth/handler.go`

`RequireAuth` уже поддерживает `?next=` (подтверждено спецификацией 004). Если нет — доработать: добавлять текущий путь (санитизированный: только `/`-относительные) и читать его при успешном логине.

---

## 15. Структура файлов (целевая)

```
migrations/
├── 007_forum_sections_description.sql   # NEW
└── 008_forum_posts_quotes.sql           # NEW

internal/
├── auth/
│   └── middleware.go              # + RequireRole
├── forum/
│   ├── model.go                   # + Section, Post, SectionView, TopicView, PostView
│   ├── repo.go                    # + новые методы
│   ├── service.go                 # NEW
│   ├── handler.go                 # NEW
│   ├── errors.go                  # NEW
│   ├── service_test.go            # NEW
│   ├── repo_test.go               # расширить
│   └── handler_test.go            # NEW
└── home/
    └── handler.go                 # обновить ссылки на форум (если хранится в коде)

templates/
├── forum/
│   ├── index.html                 # NEW
│   ├── section.html               # NEW
│   ├── topic.html                 # NEW
│   ├── new_section.html           # NEW
│   └── new_topic.html             # NEW
├── errors/
│   └── 403.html                   # NEW (если ещё нет)
├── home/
│   └── index.html                 # обновить линки блока «Активность форума»
└── layouts/
    └── base.html                  # + линк «Форум» в nav (опционально)
```

---

## 16. Тест-план

### Юнит-тесты (`service_test.go`, без БД)

| Компонент | Что проверяем |
|---|---|
| `Service.CreateSection` | Пустой title → `ErrEmptyTitle` |
| `Service.CreateSection` | Title только из пробелов → `ErrEmptyTitle` |
| `Service.CreateSection` | Title > 255 → `ErrTitleTooLong` |
| `Service.CreateSection` | Description > 2000 → `ErrDescriptionTooLong` |
| `Service.CreateSection` | Валидные данные → вызов `repo.CreateSection` |
| `Service.CreateTopic` | Пустой title → `ErrEmptyTitle` |
| `Service.CreateTopic` | Title > 255 → `ErrTitleTooLong` |
| `Service.CreateTopic` | Валидные данные → вызов `repo.CreateTopic` |
| `Service.CreatePost` | Пустой content → `ErrEmptyContent` |
| `Service.CreatePost` | Content только из пробелов → `ErrEmptyContent` |
| `Service.CreatePost` | Content > 20000 рун → `ErrContentTooLong` |
| `Service.CreatePost` | Валидные данные → вызов `repo.CreatePost` |
| `Service.CreatePost` | Пробрасывает `ErrParentNotFound` и `ErrReplyToReply` |

### Интеграционные тесты репозитория (`repo_test.go`, build tag `integration`)

| Метод | Что проверяем |
|---|---|
| `ListSections` | Пустой результат → `[]`, сортировка по `sort_order` |
| `ListSections` | После INSERT через `CreateTopic` — `topic_count` увеличен (триггер) |
| `GetSection` | Существующий → `*Section`; несуществующий → `nil, nil` |
| `ListTopicsBySection` | Сортировка по `last_post_at DESC NULLS LAST`, темы без постов — в конце |
| `GetTopic` | Существующий → `*Topic`; несуществующий → `nil, nil` |
| `ListPostsByTopic` | Сортировка по `created_at ASC, id ASC`, пустой → `[]` |
| `ListPostsByTopic` | Ответ содержит `ParentAuthor`, `ParentSnippet` из snapshot |
| `ListPostsByTopic` | После удаления родителя — `ParentID = nil`, snapshot-поля сохранены |
| `CreateSection` | Возвращает ID, запись в БД |
| `CreateTopic` | Возвращает ID, триггер увеличил `forum_sections.topic_count` |
| `CreateTopic` | Несуществующий `section_id` → `ErrSectionNotFound` (FK 23503) |
| `CreatePost` | Корневой → возвращает ID, триггер обновил `post_count`, `last_post_at`, `last_post_by` |
| `CreatePost` | С `parent_id` → snapshot-поля заполнены, запись в БД |
| `CreatePost` | `parent_id` из другой темы → `ErrParentNotFound` |
| `CreatePost` | Ответ на ответ (`parent.parent_id IS NOT NULL`) → `ErrReplyToReply` |
| `CreatePost` | Несуществующий `topic_id` → `ErrTopicNotFound` |
| `CreatePost` (UC-5) | Родитель с content ровно 100 рун → `parent_content_snapshot` = весь текст (100 рун) |
| `CreatePost` (UC-5) | Родитель с content 101 руна → `parent_content_snapshot` = первые 100 рун |
| `CreatePost` (UC-5) | После создания ответа: `parent_author_snapshot` совпадает с username на момент ответа |
| Миграция 007 | `topic_count` корректно инициализируется для существующих разделов (если есть) |

### Интеграционные тесты HTTP (`handler_test.go`, build tag `integration`)

| Сценарий | Метод/URL | Ожидаемый результат |
|---|---|---|
| Список разделов (гость) | GET `/forum` | 200, HTML с разделами |
| Пустой список | GET `/forum` (БД без разделов) | 200, «Разделы пока не созданы» |
| Раздел существует | GET `/forum/sections/1` | 200, HTML с темами |
| Раздел не существует | GET `/forum/sections/999` | 404 |
| Невалидный ID | GET `/forum/sections/abc` | 404 |
| Тема существует | GET `/forum/topics/1` | 200, HTML с постами |
| Тема не существует | GET `/forum/topics/999` | 404 |
| Создание поста (user) | POST `/forum/topics/1/posts` (content=«hi») | 201 (HTMX) / 303 (без HTMX), пост в БД |
| Создание поста (гость) | POST `/forum/topics/1/posts` | 302 → `/login?next=/forum/topics/1` |
| Пустой content | POST `/forum/topics/1/posts` (content="") | 422 |
| Ответ на пост (UC-5) | POST с `parent_id=5` (`parent.parent_id = null`) | 201, `parent_author_snapshot` и `parent_content_snapshot` заполнены |
| Ответ на ответ (UC-5) | POST с `parent_id=10` (`parent.parent_id != null`) | 422 «Нельзя отвечать на ответ» |
| Ответ с parent из другой темы (UC-5) | POST с `parent_id` поста из другой темы | 422 «Сообщение, на которое вы отвечаете, не найдено» |
| Ответ на несуществующий пост (UC-5) | POST с `parent_id=99999` | 422 |
| HTMX partial поста | POST + `HX-Request: true` | Ответ содержит только `#posts-list` |
| HTMX 422 ретаргетинг | POST + `HX-Request: true`, пустой content | 422 + `HX-Retarget: #post-form` |
| Модератор создаёт раздел | POST `/forum/sections` (role=moderator, title=«Test») | 303 → `/forum`, раздел в БД |
| User создаёт раздел | POST `/forum/sections` (role=user) | 403 |
| Гость на форме раздела | GET `/forum/sections/new` (без cookie) | 302 → `/login?next=/forum/sections/new` |
| Модератор создаёт тему | POST `/forum/sections/1/topics` (role=moderator, title=«T») | 303 → `/forum/topics/{new_id}`, `topic_count` увеличился |
| User создаёт тему | POST `/forum/sections/1/topics` (role=user) | 403 |
| Создание темы в несуществующем разделе | POST `/forum/sections/999/topics` (role=moderator) | 404 |
| Пустой title при создании раздела | POST `/forum/sections` (title="") | 422 |
| CSRF отсутствует | POST без `_csrf` | 403 |

### Ручная визуальная проверка

| Сценарий | Ожидаемый результат |
|---|---|
| Открыть `/forum` | Список разделов с названием, описанием, счётчиком тем |
| Открыть раздел | Список тем, отсортированных по последней активности |
| Открыть тему | Посты в хронологическом порядке, форма под списком |
| Создать пост | Пост появляется без перезагрузки |
| Ответить на пост | Форма с цитатой, ответ сохраняется, клик на цитату → скролл к родителю с подсветкой |
| С главной кликнуть на тему | Переход на `/forum/topics/{id}` работает |
| Модератор создаёт раздел | Раздел появляется в `/forum` |
| User пытается открыть `/forum/sections/new` | Страница 403 |
| Мобильный вид (<768px) | Контент читаем, формы удобны |

---

## 17. Не входит в scope

- Редактирование и удаление постов, тем, разделов
- Модерация контента (скрытие постов, бан пользователей, пин/замок тем)
- Пагинация постов и тем (потолок `LIMIT 500`)
- Полнотекстовый поиск по форуму
- SSE / real-time обновления списка постов
- Уведомления (email, in-app)
- Mentions `@username` в постах (перенос паттерна из feature 004 — отдельная задача)
- Markdown / rich text / BBCode в постах
- Like / dislike / реакции
- Аватары пользователей
- Вложения (файлы, изображения)
- Вложенные ответы глубже 1 уровня
- Счётчики «новые посты с последнего визита»
- RSS-фид форума
- Rate-limiting на создание постов
- SEO-метатеги

---

_Spec v1.0 | 2026-04-07_
