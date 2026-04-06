# 04_article_page — Спецификация (SDD)

## 1. Контекст

Главная страница уже отображает список последних новостей со ссылками вида `/news/{id}`, но переход по ним ведёт на 404 — роута и страницы не существует. Это обрывает основной пользовательский сценарий: пришёл на сайт → увидел заголовок → хочет прочитать. Задача — создать страницу просмотра статьи и систему комментариев к ней.

Модель `News` с полем `Content` уже есть в коде (`internal/news/model.go`). Таблица `news` создана миграцией `004_create_news.sql`. Требуется: новый хэндлер, шаблон, репозиторный метод для получения статьи по ID, таблица комментариев и логика работы с ними.

**Ссылка на бриф:** [brief.md](brief.md)
**GitHub Issue:** [#4](https://github.com/Pegorino82/lfcru_forum/issues/4)

---

## 2. Акторы

| Актор | Описание |
|---|---|
| Гость | Незарегистрированный посетитель. Может просматривать статью и комментарии. Не может комментировать |
| Пользователь | Зарегистрированный аккаунт. Может просматривать статью, оставлять комментарии, отвечать на комментарии, упоминать других пользователей через `@username` |

---

## 3. Сценарии использования (Use Cases)

### UC-1: Просмотр статьи

**Актор:** Гость / Пользователь

**Предусловие:** Статья существует в БД и опубликована (`is_published = true`).

**Основной поток:**

1. Пользователь переходит по URL `/news/{id}` (клик по ссылке на главной или прямой ввод)
2. Система загружает статью из БД по `id` (только опубликованные)
3. Система загружает комментарии к статье (дерево комментариев)
4. Система отображает страницу: заголовок, дата публикации, полный текст статьи, блок комментариев

**Альтернативные потоки:**

- **2a.** Статья с указанным `id` не существует → HTTP 404, страница «Статья не найдена»
- **2b.** Статья существует, но не опубликована (`is_published = false`) → HTTP 404, страница «Статья не найдена» (не раскрывать существование черновика)
- **2c.** `id` не является числом → HTTP 404

**Acceptance Criteria:**

- **Given** в БД есть опубликованная статья с `id=1`, **When** пользователь открывает `/news/1`, **Then** отображается страница с заголовком, датой и полным текстом статьи
- **Given** в БД нет статьи с `id=999`, **When** пользователь открывает `/news/999`, **Then** HTTP 404 со страницей «Статья не найдена»
- **Given** статья с `id=2` существует, но `is_published = false`, **When** пользователь открывает `/news/2`, **Then** HTTP 404 (аналогично несуществующей)
- **Given** пользователь открывает `/news/abc`, **Then** HTTP 404
- Страница доступна гостям без авторизации (HTTP 200)

### UC-2: Просмотр комментариев

**Актор:** Гость / Пользователь

**Предусловие:** Пользователь находится на странице статьи.

**Основной поток:**

1. Система загружает все комментарии к статье, отсортированные по `created_at ASC` (хронологически)
2. Комментарии отображаются плоским списком под текстом статьи
3. Каждый комментарий содержит: имя автора (`username`), дату/время создания, текст комментария
4. Ответы на комментарии (reply) визуально отличаются: содержат цитату-ссылку на родительский комментарий

**Альтернативный поток:**

- **1a.** Комментариев нет → блок комментариев отображает текст «Пока нет комментариев. Войдите, чтобы оставить первый!» (для гостей) или «Пока нет комментариев. Будьте первым!» (для авторизованных)

**Acceptance Criteria:**

- **Given** у статьи 3 комментария, **When** пользователь открывает статью, **Then** все 3 комментария отображаются в хронологическом порядке
- **Given** у статьи нет комментариев, гость открывает статью, **Then** отображается «Пока нет комментариев. Войдите, чтобы оставить первый!»
- **Given** у статьи нет комментариев, авторизованный пользователь открывает статью, **Then** отображается «Пока нет комментариев. Будьте первым!»
- Каждый комментарий отображает username автора, дату/время и текст
- Комментарий-ответ содержит цитату с текстом и автором родительского комментария

### UC-3: Создание комментария

**Актор:** Пользователь (авторизован)

**Предусловие:** Пользователь авторизован и находится на странице статьи.

**Основной поток:**

1. Под блоком комментариев отображается форма: текстовое поле и кнопка «Отправить»
2. Пользователь вводит текст комментария
3. Пользователь нажимает «Отправить»
4. Система валидирует текст (см. [§4 Валидация](#4-правила-валидации))
5. Система сохраняет комментарий в БД (привязан к статье и пользователю, `parent_id = NULL`)
6. Система возвращает обновлённый блок комментариев с новым комментарием в конце списка (HTMX swap)

**Альтернативные потоки:**

- **4a.** Текст пустой или только пробелы → ошибка «Комментарий не может быть пустым»
- **4b.** Текст превышает 10 000 символов → ошибка «Комментарий слишком длинный (максимум 10 000 символов)»
- **Гость:** форма не отображается. Вместо неё текст «Войдите, чтобы оставить комментарий» с ссылкой на `/login?next=/news/{id}`

**Acceptance Criteria:**

- **Given** авторизованный пользователь на странице статьи, **When** вводит текст и нажимает «Отправить», **Then** комментарий появляется в конце списка без перезагрузки страницы
- **Given** гость на странице статьи, **Then** форма комментирования не отображается, показана ссылка на вход
- **Given** пользователь отправляет пустой комментарий, **Then** ошибка «Комментарий не может быть пустым»
- **Given** пользователь отправляет комментарий > 10 000 символов, **Then** ошибка «Комментарий слишком длинный (максимум 10 000 символов)»
- Для создания комментария требуется CSRF-токен

### UC-4: Ответ на комментарий (reply)

**Актор:** Пользователь (авторизован)

**Предусловие:** Пользователь авторизован, на странице статьи есть хотя бы один комментарий.

**Основной поток:**

1. Пользователь нажимает кнопку «Ответить» у комментария
2. Под целевым комментарием появляется форма ответа с цитатой родительского комментария (текст автора + первые 100 символов текста)
3. Пользователь вводит текст ответа
4. Пользователь нажимает «Отправить»
5. Система валидирует текст (§4)
6. Система сохраняет комментарий в БД с `parent_id` = ID родительского комментария
7. Система возвращает обновлённый блок комментариев; новый ответ отображается в хронологической позиции с цитатой родительского комментария

**Альтернативные потоки:**

- **1a.** Пользователь нажимает «Ответить» повторно или на другой комментарий → предыдущая форма ответа скрывается, открывается новая (Alpine.js, только одна форма ответа одновременно)
- **2a.** Пользователь нажимает «Отмена» в форме ответа → форма скрывается
- **6a.** Родительский комментарий был удалён до момента отправки → HTTP 422, ошибка «Комментарий, на который вы отвечаете, не найден»

**Acceptance Criteria:**

- **Given** пользователь нажимает «Ответить» у комментария, **Then** появляется форма ответа с цитатой
- **Given** пользователь отправляет ответ, **Then** ответ сохраняется с `parent_id`, отображается с цитатой родительского комментария
- **Given** пользователь нажимает «Ответить» на два разных комментария последовательно, **Then** открыта только одна форма ответа (последняя)
- **Given** форма ответа открыта, пользователь нажимает «Отмена», **Then** форма скрывается
- Гость не видит кнопку «Ответить»

### UC-5: Навигация к родительскому комментарию

**Актор:** Гость / Пользователь

**Предусловие:** На странице есть комментарий-ответ с цитатой.

**Основной поток:**

1. Пользователь кликает на цитату (блок цитирования) в комментарии-ответе
2. Браузер прокручивает страницу к родительскому комментарию (якорь `#comment-{id}`)
3. Родительский комментарий визуально подсвечивается на 2 секунды (CSS transition)

**Acceptance Criteria:**

- **Given** комментарий-ответ с `parent_id=5`, **When** пользователь кликает на цитату, **Then** страница прокручивается к элементу `#comment-5`
- Родительский комментарий подсвечивается с плавным затуханием (CSS `animation`)
- Если родительский комментарий удалён — цитата отображается, но ссылка неактивна (нет `href`), текст цитаты: «[комментарий удалён]»

### UC-6: Упоминание пользователя (@username)

**Актор:** Пользователь (авторизован)

**Предусловие:** Пользователь пишет комментарий или ответ.

**Основной поток:**

1. Пользователь вводит `@username` в тексте комментария (где `username` — имя зарегистрированного пользователя)
2. Система при сохранении комментария обрабатывает текст: находит все паттерны `@[a-zA-Z0-9_-]{3,30}` и проверяет существование пользователей с такими username
3. При отображении комментария: `@username` существующего пользователя оборачивается в `<span class="mention">@username</span>` (визуально выделен)
4. `@username` несуществующего пользователя остаётся обычным текстом (без оборачивания)

**Acceptance Criteria:**

- **Given** пользователь пишет комментарий с текстом `@admin привет`, пользователь `admin` существует, **Then** при отображении `@admin` выделен как mention
- **Given** пользователь пишет `@nonexistent`, такого пользователя нет, **Then** текст отображается как есть, без стилизации
- **Given** пользователь пишет `admin@mail.ru`, **Then** это НЕ распознаётся как mention (@ не в начале слова)
- Mentions обрабатываются только при рендеринге (не изменяют сохранённый текст в БД)
- Паттерн mention: `@` в начале слова (после пробела, начала строки или начала текста), за которым следует `[a-zA-Z0-9_-]{3,30}`

---

## 4. Правила валидации

### Комментарий

| Поле | Правило | Сообщение об ошибке |
|---|---|---|
| content | Не пустой после `strings.TrimSpace()` | «Комментарий не может быть пустым» |
| content | Длина ≤ 10 000 символов (`utf8.RuneCountInString`) | «Комментарий слишком длинный (максимум 10 000 символов)» |
| parent_id | Если указан — должен существовать в `news_comments` и принадлежать той же статье (`news_id`) | «Комментарий, на который вы отвечаете, не найден» |

### Параметр URL

| Поле | Правило | Сообщение об ошибке |
|---|---|---|
| id (в `/news/{id}`) | Целое положительное число | HTTP 404 |

---

## 5. Модель данных

### Таблица `news_comments`

```sql
CREATE TABLE news_comments (
    id                      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    news_id                 BIGINT       NOT NULL REFERENCES news(id) ON DELETE CASCADE,
    author_id               BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    parent_id               BIGINT       REFERENCES news_comments(id) ON DELETE SET NULL,
    parent_author_snapshot  TEXT,
    parent_content_snapshot TEXT,
    content                 TEXT         NOT NULL,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_comments_news ON news_comments (news_id, created_at ASC);
```

> **Примечания:**
> - `news_id` — FK на статью. `ON DELETE CASCADE`: при удалении статьи удаляются все комментарии.
> - `author_id` — FK на автора. `ON DELETE RESTRICT`: нельзя удалить пользователя, у которого есть комментарии (soft delete через `is_active = false`).
> - `parent_id` — FK на родительский комментарий (самоссылка). `NULL` для корневых комментариев. `ON DELETE SET NULL`: при удалении родительского комментария `parent_id` обнуляется. Данные цитаты при этом **сохраняются** в snapshot-колонках — отображение «[комментарий удалён]» остаётся возможным.
> - `parent_author_snapshot` — username автора родительского комментария на момент создания ответа. `NULL` для корневых комментариев. Не меняется при удалении родителя.
> - `parent_content_snapshot` — первые 100 рун текста родительского комментария на момент создания ответа. `NULL` для корневых комментариев. Не меняется при удалении родителя.
> - `content` — текст комментария. Хранится как есть (plain text), без HTML. Mentions (`@username`) обрабатываются на этапе рендеринга.
> - Нет поля `updated_at` — комментарии не редактируются в данной фиче.

### Денормализация в `news` (счётчик комментариев)

Не добавляется. Количество комментариев не отображается на главной странице и не используется в текущей фиче. При необходимости — добавить в отдельной задаче.

### Миграция

Файл: `006_create_news_comments.sql`

> **Нумерация:** миграция `005_create_forum_and_matches.sql` уже существует в репозитории. Нумерация `006` корректна.

```sql
-- +goose Up
CREATE TABLE news_comments (
    id                      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    news_id                 BIGINT       NOT NULL REFERENCES news(id) ON DELETE CASCADE,
    author_id               BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    parent_id               BIGINT       REFERENCES news_comments(id) ON DELETE SET NULL,
    parent_author_snapshot  TEXT,
    parent_content_snapshot TEXT,
    content                 TEXT         NOT NULL,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_comments_news ON news_comments (news_id, created_at ASC);

-- +goose Down
DROP TABLE IF EXISTS news_comments;
```

---

## 6. Структура Go-пакетов

### Новый пакет

```
internal/
└── comment/
    ├── model.go           # struct Comment, CommentView
    ├── repo.go            # CommentRepo: Create, ListByNewsID
    ├── service.go         # CommentService: валидация, mention-обработка
    └── errors.go          # sentinel-ошибки
```

### Модели

```go
// internal/comment/model.go

// Comment — запись комментария в БД.
type Comment struct {
    ID                    int64
    NewsID                int64
    AuthorID              int64
    ParentID              *int64  // nil для корневых комментариев
    ParentAuthorSnapshot  *string // nil для корневых; username родителя на момент создания ответа
    ParentContentSnapshot *string // nil для корневых; первые 100 рун контента родителя на момент создания ответа
    Content               string
    CreatedAt             time.Time
}

// CommentView — проекция для отображения.
// Содержит username автора и данные родительского комментария (если reply).
type CommentView struct {
    ID              int64
    NewsID          int64
    AuthorID        int64
    AuthorUsername  string
    ParentID        *int64  // nil если корневой ИЛИ если родитель был удалён (SET NULL)
    ParentAuthor    *string // nil если корневой; заполняется из parent_author_snapshot — сохраняется даже при удалении родителя
    ParentSnippet   *string // nil если корневой; заполняется из parent_content_snapshot — сохраняется даже при удалении родителя
    Content         string  // raw text из БД
    ContentHTML     string  // обработанный текст с mentions (заполняется в service)
    CreatedAt       time.Time
}
```

### Ошибки

```go
// internal/comment/errors.go
var (
    ErrEmptyContent   = errors.New("comment: empty content")
    ErrContentTooLong = errors.New("comment: content too long")
    ErrParentNotFound = errors.New("comment: parent not found")
    ErrNewsNotFound   = errors.New("comment: news not found")
)
```

---

## 7. Репозитории — SQL-запросы

### `news.Repo.GetPublishedByID(ctx, id int64) (*News, error)`

Новый метод в существующем `internal/news/repo.go`.

```sql
SELECT id, title, content, is_published, author_id, published_at, created_at, updated_at
FROM news
WHERE id = $1 AND is_published = true
```

Возвращает `*News` или `nil` если статья не найдена или не опубликована. Ошибка — только при проблемах с БД.

> **Примечание:** один запрос с `AND is_published = true` вместо двух (SELECT + проверка) — не раскрывает существование черновика через timing.

### `comment.Repo.ListByNewsID(ctx, newsID int64) ([]CommentView, error)`

```sql
SELECT
    c.id,
    c.news_id,
    c.author_id,
    u.username AS author_username,
    c.parent_id,
    c.parent_author_snapshot  AS parent_author,
    c.parent_content_snapshot AS parent_snippet,
    c.content,
    c.created_at
FROM news_comments c
JOIN users u ON u.id = c.author_id
WHERE c.news_id = $1
ORDER BY c.created_at ASC, c.id ASC
LIMIT 500
```

Возвращает `[]CommentView`. При пустом результате — пустой слайс `[]CommentView{}`, не `nil`.

> **Примечания:**
> - `JOIN users u` — inner join: комментарий всегда имеет автора (ON DELETE RESTRICT).
> - `parent_author_snapshot` и `parent_content_snapshot` заполняются при CREATE и сохраняются навсегда. При удалении родителя (SET NULL для `parent_id`) snapshot-данные остаются — это позволяет отображать «[комментарий удалён]» в цитате.
> - `ORDER BY created_at ASC, id ASC` — вторичная сортировка по `id` гарантирует стабильный порядок при одинаковом `created_at`.
> - `LIMIT 500` — защитный потолок для MVP. При росте до 200+ комментариев на статью — добавить cursor-based пагинацию в отдельной задаче.
> - Троеточие в `parent_snippet`: добавляется в Go при маппинге, если `len([]rune(snapshot)) == 100` (значит контент был длиннее и обрезан при сохранении).

### `comment.Repo.Create(ctx, c *Comment) (int64, error)`

Если `c.ParentID != nil`, выполняется в **одной транзакции**:

**Шаг 1 — получить данные родительского комментария:**

```sql
SELECT u.username, LEFT(nc.content, 100), nc.parent_id
FROM news_comments nc
JOIN users u ON u.id = nc.author_id
WHERE nc.id = $1 AND nc.news_id = $2
```

- Если строка не найдена → `ErrParentNotFound` (родитель не существует или из другой статьи)
- Если `nc.parent_id IS NOT NULL` → `ErrParentNotFound` (превышена глубина вложенности: ответ на ответ запрещён, только 1 уровень)
- Заполнить `c.ParentAuthorSnapshot` и `c.ParentContentSnapshot` из результата

**Шаг 2 — вставка:**

```sql
INSERT INTO news_comments
    (news_id, author_id, parent_id, parent_author_snapshot, parent_content_snapshot, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id
```

> **Race condition:** если родитель удаляется между шагом 1 и шагом 2, FK constraint вернёт ошибку `23503`. Маппинг: `pgconn.PgError{Code: "23503"}` → `ErrParentNotFound` (аналогично маппингу `23505` → `ErrDuplicateEmail` в `user/repo.go`).

> **Глубина вложенности:** проверка `nc.parent_id IS NOT NULL` ограничивает ответы только первым уровнем. Ответ на ответ возвращает `ErrParentNotFound`.

> **Обоснование:** FK constraint гарантирует ссылочную целостность, но не проверяет `news_id` и глубину. Транзакция устраняет оба gap.

### `user.Repo.GetByUsernames(ctx, usernames []string) ([]User, error)`

Новый метод в существующем `internal/user/repo.go`. Используется для валидации mentions.

```sql
SELECT id, username
FROM users
WHERE lower(username) = ANY($1)
AND is_active = true
```

Параметр `$1` — массив нормализованных (lower) usernames. Возвращает `[]User` (только `ID`, `Username` заполнены). При пустом входе — пустой слайс без запроса к БД.

---

## 8. Сервисный слой

### `comment.Service`

```go
type Service struct {
    commentRepo *Repo
    userRepo    UserRepo // интерфейс: GetByUsernames(ctx, []string) ([]user.User, error)
}
```

**Интерфейс `UserRepo`** (в пакете `comment`):

```go
type UserRepo interface {
    GetByUsernames(ctx context.Context, usernames []string) ([]user.User, error)
}
```

### Метод `Service.Create(ctx, newsID, authorID int64, parentID *int64, content string) (int64, error)`

1. `content = strings.TrimSpace(content)`
2. Если `content == ""` → `ErrEmptyContent`
3. Если `utf8.RuneCountInString(content) > 10000` → `ErrContentTooLong`
4. Вызов `commentRepo.Create(ctx, &Comment{NewsID: newsID, AuthorID: authorID, ParentID: parentID, Content: content})`
   - Репозиторий в транзакции: проверяет существование `parent_id`, ограничение глубины (depth ≤ 1), заполняет snapshot-поля, выполняет INSERT
   - При `ErrParentNotFound` — пробрасывать наружу без изменений
5. Возврат `id` нового комментария

### Метод `Service.RenderMentions(ctx, text string) (string, error)`

Обрабатывает текст комментария для отображения. Вызывается при формировании `CommentView.ContentHTML`.

1. Извлечь все паттерны `@[a-zA-Z0-9_-]{3,30}` из текста с помощью regex.
   - Паттерн матчится только если `@` стоит в начале строки или после пробельного символа (lookbehind: `(?:^|\s)(@[a-zA-Z0-9_-]{3,30})`)
2. Убрать дубликаты, привести к lower.
3. Если список пуст — вернуть текст с HTML-экранированием.
4. `userRepo.GetByUsernames(ctx, usernames)` — получить существующих пользователей.
5. Для каждого совпадения `@username`:
   - Если пользователь существует → заменить на `<span class="mention">@username</span>` (использовать оригинальный регистр из текста)
   - Если не существует → оставить как текст
6. Весь текст (кроме mention-оборачиваний) HTML-экранируется (`html.EscapeString`) **до** вставки `<span>` тегов.

> **Безопасность:** текст комментария хранится в БД как plain text. HTML-экранирование выполняется в `RenderMentions`. Порядок: сначала экранировать весь текст, затем вставить безопасные `<span>` теги для mentions. Это предотвращает XSS.

> **Рендеринг `ContentHTML`:** Заполняется в хэндлере после получения `[]CommentView` из репозитория. Цикл по комментариям, для каждого вызов `service.RenderMentions(ctx, cv.Content)`.

---

## 9. API-контракты

Все эндпоинты возвращают HTML (server-side rendering). Сервер проверяет заголовок `HX-Request`: если присутствует — возвращает HTML-фрагмент, если нет — полную страницу.

### GET /news/:id

**Описание:** Страница статьи с комментариями.

**Авторизация:** Не требуется. Доступна гостям.

**Параметры:**

| Параметр | Тип | Описание |
|---|---|---|
| id | int64 (path) | ID статьи |

**Данные для шаблона:**

```go
type ArticleData struct {
    User        *user.User            // из контекста (middleware); nil для гостей
    CSRFToken   string                // из middleware
    Article     *news.News            // статья
    Comments    []comment.CommentView // комментарии к статье
    FormError   string                // сообщение об ошибке валидации (422 без HTMX)
    FormContent string                // введённый текст для сохранения при 422
}
```

**Ответы:**

| Код | Условие | Действие |
|---|---|---|
| 200 | Статья найдена и опубликована | HTML-страница статьи с комментариями |
| 404 | Статья не найдена / не опубликована / невалидный ID | Брендированная 404-страница «Статья не найдена» в рамках `base.html` |
| 500 | Ошибка БД | Брендированная 500-страница |

**HTMX:** При `HX-Request: true` — возвращает только блок `content`.

### POST /news/:id/comments

**Описание:** Создание комментария к статье.

**Авторизация:** Требуется. Гость получает HTTP 302 redirect на `/login?next=/news/{id}`.

**Запрос:** `application/x-www-form-urlencoded`

| Параметр | Тип | Обязательно | Описание |
|---|---|---|---|
| content | string | да | Текст комментария |
| parent_id | int64 | нет | ID родительского комментария (для ответов) |
| _csrf | string | да | CSRF-токен |

**Ответы:**

| Код | Условие | Действие |
|---|---|---|
| 201 | Успех | HTMX: обновлённый блок комментариев + заголовок `HX-Trigger: commentAdded`. Без HTMX: redirect 303 на `/news/{id}#comment-{new_id}` |
| 422 | Ошибка валидации (пустой текст, превышение длины, parent не найден) | HTMX: форма с ошибкой + заголовок `HX-Retarget: #comment-form`. Без HTMX: страница статьи с ошибкой в поле `ArticleData.FormError` |
| 404 | Статья не найдена / не опубликована | 404-страница |
| 302 | Гость | Redirect на `/login?next=/news/{id}` |
| 500 | Ошибка БД | Брендированная 500-страница |

> **HTMX 422 и ретаргетинг:** при ошибке валидации сервер возвращает `HX-Retarget: #comment-form` и `HX-Reswap: innerHTML`, чтобы заменить только форму, а не весь `#comments-list`. Это предотвращает потерю уже загруженных комментариев при ошибке.

> **Без HTMX (422):** хэндлер добавляет сообщение об ошибке в `ArticleData.FormError string` и рендерит полную страницу с кодом 422. Форма сохраняет введённый текст через `ArticleData.FormContent string`.

> **`?next=` после логина:** `RequireAuth` middleware перенаправляет гостей на `/login?next=/news/{id}`. После успешного логина middleware `auth` должен выполнить redirect на значение `next` (санитизированное — только `/`-относительные пути).

**HTMX-интеграция:**

Форма комментария отправляет `hx-post="/news/{id}/comments"` с `hx-target="#comments-list"` и `hx-swap="innerHTML"`. Сервер при успешном POST с `HX-Request: true` возвращает обновлённый HTML блока комментариев (включая новый комментарий) с кодом 201. Сервер также устанавливает заголовок `HX-Trigger: commentAdded` для сброса формы на клиенте.

**Форма ответа (reply):** отправляет тот же POST с дополнительным полем `parent_id`. Форма управляется Alpine.js — показ/скрытие через `x-show`.

---

## 10. Шаблон страницы

### Целевая структура HTML

```html
{{define "content"}}
<article class="article-page">
  <header class="article-header">
    <h1>{{.Article.Title}}</h1>
    {{if .Article.PublishedAt}}
    <time class="article-date" datetime="{{.Article.PublishedAt.Format "2006-01-02"}}">
      {{.Article.PublishedAt.Format "02.01.2006"}}
    </time>
    {{end}}
  </header>

  <div class="article-content">
    {{.Article.Content}}
  </div>
  {{/* .Article.Content — тип string, html/template автоматически экранирует.
       Если статья содержит безопасный HTML (разметка), использовать тип template.HTML
       и поле ContentHTML — только после явного решения о доверии источнику контента. */}}
</article>

<section class="comments-section" aria-labelledby="comments-heading">
  <h2 id="comments-heading">Комментарии</h2>

  <div id="comments-list">
    {{if .Comments}}
      {{range .Comments}}
      <div class="comment" id="comment-{{.ID}}">
        {{if .ParentAuthor}}
          {{/* ParentAuthor заполнен из snapshot — значит это ответ, даже если ParentID = nil (родитель удалён) */}}
          {{if .ParentID}}
          <a href="#comment-{{.ParentID}}" class="comment-quote" data-comment-id="{{.ParentID}}">
            <span class="quote-author">{{deref .ParentAuthor}}</span>
            <span class="quote-text">{{if .ParentSnippet}}{{deref .ParentSnippet}}{{end}}</span>
          </a>
          {{else}}
          <div class="comment-quote comment-quote--deleted">
            <span class="quote-author">[удалён]</span>
            <span class="quote-text">[комментарий удалён]</span>
          </div>
          {{end}}
        {{end}}
        <div class="comment-header">
          <span class="comment-author">{{.AuthorUsername}}</span>
          <time class="comment-date" datetime="{{.CreatedAt.Format "2006-01-02T15:04:05Z07:00"}}">
            {{.CreatedAt.Format "02.01.2006 15:04"}}
          </time>
        </div>
        <div class="comment-body">{{.ContentHTML}}</div>
        {{if $.User}}
        <button class="comment-reply-btn"
                type="button"
                aria-label="Ответить {{.AuthorUsername}}"
                @click="replyTo = {{.ID}}; replyAuthor = '{{.AuthorUsername}}'">
          Ответить
        </button>
        {{end}}
      </div>
      {{end}}
    {{else}}
      {{if .User}}
        <p class="empty-state">Пока нет комментариев. Будьте первым!</p>
      {{else}}
        <p class="empty-state">Пока нет комментариев. <a href="/login?next=/news/{{.Article.ID}}">Войдите</a>, чтобы оставить первый!</p>
      {{end}}
    {{end}}
  </div>

  {{if .User}}
  <!-- Форма ответа (reply) — Alpine.js -->
  <template x-if="replyTo !== null">
    <form class="reply-form"
          id="comment-form"
          hx-post="/news/{{.Article.ID}}/comments"
          hx-target="#comments-list"
          hx-swap="innerHTML"
          hx-disabled-elt="find button[type=submit]">
      <input type="hidden" name="_csrf" value="{{.CSRFToken}}">
      <input type="hidden" name="parent_id" :value="replyTo">
      <div class="reply-quote">
        Ответ для <span x-text="replyAuthor"></span>
      </div>
      <textarea name="content" rows="3" required
                maxlength="10000"
                placeholder="Ваш ответ..."
                aria-label="Текст ответа"></textarea>
      <div class="form-actions">
        <button type="submit">Отправить</button>
        <button type="button" @click="replyTo = null">Отмена</button>
      </div>
    </form>
  </template>

  <!-- Основная форма комментария -->
  <form class="comment-form"
        id="comment-form"
        hx-post="/news/{{.Article.ID}}/comments"
        hx-target="#comments-list"
        hx-swap="innerHTML"
        hx-disabled-elt="find button[type=submit]">
    <input type="hidden" name="_csrf" value="{{.CSRFToken}}">
    {{if .FormError}}
    <p class="form-error" role="alert">{{.FormError}}</p>
    {{end}}
    <textarea name="content" rows="4" required
              maxlength="10000"
              placeholder="Ваш комментарий..."
              aria-label="Текст комментария">{{.FormContent}}</textarea>
    <button type="submit">Отправить</button>
  </form>
  {{else}}
  <p class="login-prompt">
    <a href="/login?next=/news/{{.Article.ID}}">Войдите</a>, чтобы оставить комментарий
  </p>
  {{end}}
</section>
{{end}}
```

> **Примечание:** шаблон выше — целевая структура. При реализации Alpine.js директивы (`x-data`, `x-show`, `x-if`, `@click`) размещаются на обёртке секции комментариев: `<section x-data="{ replyTo: null, replyAuthor: '' }" @reset-reply="replyTo = null; replyAuthor = ''">`.

> **`{{.ContentHTML}}`** выводится без экранирования шаблонизатором (`template.HTML` тип). Безопасность обеспечивается на уровне `Service.RenderMentions`: весь текст экранируется через `html.EscapeString`, затем вставляются безопасные `<span>` теги для mentions. XSS невозможен при корректной реализации этой цепочки.

> **`{{.Article.Content}}`** — тип `string`, html/template автоматически экранирует при выводе. Если в будущем потребуется рендерить статьи с HTML-разметкой, необходимо явно использовать тип `template.HTML` и поле `ContentHTML` с чёткой документацией доверенного источника.

### Правила отображения

| Элемент | Формат | Пример |
|---|---|---|
| Дата публикации статьи | `02.01.2006` | `15.03.2026` |
| Дата/время комментария | `02.01.2006 15:04` | `05.04.2026 14:30` |
| `<time datetime>` | ISO 8601 | `2026-04-05T14:30:00+03:00` |
| Цитата (parent snippet) | Первые 100 символов + `…` если обрезано | `Отличная новость! Надо было давно…` |
| Mention | `<span class="mention">@username</span>` | `<span class="mention">@admin</span>` |

### Навигация к родительскому комментарию

Клик по цитате (`<a href="#comment-{parentID}">`) прокручивает к родительскому комментарию. Подсветка через CSS:

```css
.comment:target {
    animation: highlight 2s ease-out;
}
@keyframes highlight {
    from { background-color: #fff3cd; }
    to { background-color: transparent; }
}
```

Если `ParentAuthor != nil` (это ответ), но `ParentID == nil` (родитель удалён через SET NULL) — цитата рендерится без `<a>`:

```html
<div class="comment-quote comment-quote--deleted">
  <span class="quote-author">[удалён]</span>
  <span class="quote-text">[комментарий удалён]</span>
</div>
```

---

## 11. Стилизация

Минимальная стилизация (inline `<style>` в шаблоне `article.html`):

| Элемент | Стили |
|---|---|
| `.article-page` | `max-width: 800px; margin: 0 auto;` |
| `.article-header h1` | `font-size: 1.75rem; margin-bottom: 0.5rem;` |
| `.article-date` | `color: #666; font-size: 0.875rem;` |
| `.article-content` | `line-height: 1.6; margin: 1.5rem 0;` |
| `.comments-section` | `margin-top: 2rem; border-top: 2px solid #c8102e; padding-top: 1rem;` |
| `.comment` | `padding: 0.75rem 0; border-bottom: 1px solid #eee;` |
| `.comment-author` | `font-weight: bold;` |
| `.comment-date` | `color: #666; font-size: 0.875rem; margin-left: 0.5rem;` |
| `.comment-body` | `margin-top: 0.5rem; line-height: 1.5;` |
| `.comment-quote` | `display: block; background: #f5f5f5; border-left: 3px solid #c8102e; padding: 0.5rem 0.75rem; margin-bottom: 0.5rem; font-size: 0.875rem; color: #555; text-decoration: none; cursor: pointer;` |
| `.comment-quote:hover` | `background: #eee;` |
| `.comment-quote--deleted` | `cursor: default; opacity: 0.6;` |
| `.quote-author` | `font-weight: bold; display: block; font-size: 0.8rem;` |
| `.quote-text` | `display: block; font-style: italic;` |
| `.mention` | `color: #c8102e; font-weight: 500;` |
| `.comment-reply-btn` | `background: none; border: none; color: #c8102e; cursor: pointer; font-size: 0.8rem; padding: 0; margin-top: 0.25rem;` |
| `.reply-form` | `background: #f9f9f9; padding: 1rem; border-radius: 4px; margin: 0.5rem 0;` |
| `.reply-quote` | `font-size: 0.875rem; color: #666; margin-bottom: 0.5rem;` |
| `textarea` | `width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; font-family: inherit;` |
| `.comment-form button, .reply-form button[type=submit]` | `background: #c8102e; color: #fff; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-top: 0.5rem;` |
| `.reply-form button[type=button]` | `background: #eee; color: #333; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-top: 0.5rem; margin-left: 0.5rem;` |
| `.login-prompt` | `color: #666; font-style: italic; margin-top: 1rem;` |
| `.empty-state` | `color: #999; font-style: italic;` |

### Мобильная адаптивность (breakpoint < 768px)

```css
@media (max-width: 767px) {
    .article-page {
        padding: 0 1rem;
    }
    .article-header h1 {
        font-size: 1.375rem;
    }
    .comment-reply-btn,
    .comment-form button,
    .reply-form button {
        min-height: 44px;   /* touch target */
        min-width: 44px;
        padding: 0.625rem 1rem;
    }
    textarea {
        font-size: 1rem;    /* предотвращает zoom на iOS */
    }
}
```

---

## 12. Accessibility

| Требование | Реализация |
|---|---|
| Семантическая разметка | `<article>` для статьи, `<section>` для комментариев |
| Заголовки | `<h1>` для статьи, `<h2>` для блока комментариев |
| Машиночитаемые даты | `<time datetime="ISO 8601">` |
| Якоря комментариев | Каждый комментарий имеет `id="comment-{id}"` |
| Формы | `<textarea>` с `aria-label`, `required` |
| Пустые состояния | Текст в `<p>`, видимый для скринридеров |
| Навигация по цитатам | `<a href="#comment-{id}">` — стандартная якорная навигация |

---

## 13. HTMX и Alpine.js

### HTMX

- `GET /news/:id` — при `HX-Request: true` возвращает только `content` блок
- `POST /news/:id/comments` — при `HX-Request: true`:
  - Успех: возвращает обновлённый `#comments-list` (полный HTML блока комментариев). Заголовок `HX-Trigger: commentAdded` для сброса формы
  - Ошибка: возвращает форму с сообщением об ошибке
- Форма: `hx-post`, `hx-target="#comments-list"`, `hx-swap="innerHTML"`

### Alpine.js

Управление UI-стейтом формы ответа:

```html
<section x-data="{ replyTo: null, replyAuthor: '' }">
```

- `replyTo` — ID комментария, на который отвечаем (`null` если форма скрыта)
- `replyAuthor` — username автора комментария (для отображения в цитате формы)
- Кнопка «Ответить»: `@click="replyTo = {id}; replyAuthor = '{username}'"`
- Кнопка «Отмена»: `@click="replyTo = null"`
- Форма ответа: `x-show="replyTo !== null"`

**Сброс формы после успешной отправки:** сервер возвращает заголовок `HX-Trigger: commentAdded`. HTMX диспатчит событие на `#comments-list`. Обёртка секции слушает его через Alpine.js:

```html
<section x-data="{ replyTo: null, replyAuthor: '' }"
         @reset-reply.window="replyTo = null; replyAuthor = ''">
```

```javascript
htmx.on('#comments-list', 'commentAdded', function() {
    window.dispatchEvent(new Event('reset-reply'));
});
```

> **Alpine.js v3:** `__x.$data` — приватное API v2, несовместимо с v3. Вместо него используется `dispatchEvent` + `@reset-reply.window` (или `.self` если обёртка и список — один элемент). Это публичный и стабильный механизм.

> **Загрузка (loading state):** обе формы имеют `hx-disabled-elt="find button[type=submit]"` — кнопка «Отправить» деактивируется на время запроса, предотвращая дублирующие отправки.

> **Разделение ответственности:** HTMX — запросы к серверу и подмена DOM. Alpine.js — только показ/скрытие формы ответа и хранение `replyTo` / `replyAuthor`. На одном элементе `hx-*` и `x-*` не смешиваются: форма имеет `hx-post`, а обёртка секции — `x-data`.

---

## 14. Изменения в существующих файлах

### `internal/news/repo.go`

Добавить метод `GetPublishedByID(ctx, id int64) (*News, error)`.

### `internal/user/repo.go`

Добавить метод `GetByUsernames(ctx, usernames []string) ([]User, error)`.

### `cmd/forum/main.go`

1. Создание `comment.Repo` и `comment.Service`
2. Создание `news.Handler` (новый хэндлер для страницы статьи)
3. Регистрация маршрутов:
   - `GET /news/:id` → `newsHandler.ShowArticle`
   - `POST /news/:id/comments` → `newsHandler.CreateComment` (через `RequireAuth` middleware)
4. Миграция применяется автоматически через `goose.Up`

### `internal/auth/middleware.go`

`RequireAuth` должен поддерживать `?next=` параметр: при редиректе гостя на `/login` добавлять текущий путь как `?next={path}`. После успешного логина `auth.Handler.Login` читает `next` из query string и выполняет redirect (только `/`-относительные пути, без схемы — защита от open redirect).

### `templates/layouts/base.html`

Без изменений.

---

## 15. Структура файлов (целевая)

```
migrations/
└── 006_create_news_comments.sql

internal/
├── news/
│   ├── model.go              # без изменений
│   ├── repo.go               # + GetPublishedByID
│   └── handler.go            # NEW: ShowArticle, CreateComment
├── comment/
│   ├── model.go              # Comment, CommentView
│   ├── repo.go               # CommentRepo: Create, ListByNewsID
│   ├── service.go            # CommentService: Create, RenderMentions
│   └── errors.go             # sentinel-ошибки
└── user/
    ├── model.go              # без изменений
    └── repo.go               # + GetByUsernames

templates/
└── news/
    └── article.html          # NEW: страница статьи с комментариями
```

---

## 16. Тест-план

### Юнит-тесты

| Компонент | Что проверяем |
|---|---|
| `Service.Create` | Пустой текст → `ErrEmptyContent` |
| `Service.Create` | Текст только из пробелов → `ErrEmptyContent` (после TrimSpace) |
| `Service.Create` | Текст > 10 000 символов → `ErrContentTooLong` |
| `Service.Create` | Валидный текст → вызов `repo.Create` |
| `Service.RenderMentions` | `@admin` существующий → `<span class="mention">@admin</span>` |
| `Service.RenderMentions` | `@unknown` несуществующий → текст без обёртки |
| `Service.RenderMentions` | `admin@mail.ru` → не распознаётся как mention |
| `Service.RenderMentions` | `<script>@admin</script>` → HTML экранирован, mention обёрнут |
| `Service.RenderMentions` | Несколько mentions в одном тексте |
| `Service.RenderMentions` | `@admin @admin` → один запрос к БД (дедупликация) |
| `Service.RenderMentions` | Текст без mentions → HTML экранирован, без запросов к БД |
| Валидация `parent_id` | `parent_id` из другой статьи → `ErrParentNotFound` |

### Интеграционные тесты (репозитории, build tag `integration`)

| Компонент | Что проверяем |
|---|---|
| `news.Repo.GetPublishedByID` | Существующая опубликованная статья → `*News` с полными данными |
| `news.Repo.GetPublishedByID` | Несуществующий ID → `nil, nil` |
| `news.Repo.GetPublishedByID` | Неопубликованная статья → `nil, nil` |
| `comment.Repo.Create` | Создание корневого комментария → возвращает ID |
| `comment.Repo.Create` | Создание ответа (parent_id) → возвращает ID |
| `comment.Repo.Create` | parent_id из другой статьи → ошибка |
| `comment.Repo.ListByNewsID` | Нет комментариев → пустой слайс |
| `comment.Repo.ListByNewsID` | Комментарии отсортированы по created_at ASC |
| `comment.Repo.ListByNewsID` | Ответ содержит ParentAuthor и ParentSnippet |
| `comment.Repo.ListByNewsID` | Родительский комментарий удалён → ParentID = nil (SET NULL), ParentAuthorSnapshot заполнен |
| `comment.Repo.Create` | Ответ на ответ (depth > 1) → `ErrParentNotFound` |
| `user.Repo.GetByUsernames` | Существующие пользователи → возвращает совпадения |
| `user.Repo.GetByUsernames` | Пустой список → пустой слайс |
| `user.Repo.GetByUsernames` | Case-insensitive сравнение |

### Интеграционные тесты (HTTP)

| Сценарий | URL | Ожидаемый результат |
|---|---|---|
| Статья существует | GET `/news/1` | HTTP 200, HTML содержит заголовок и текст |
| Статья не существует | GET `/news/999` | HTTP 404 |
| Неопубликованная статья | GET `/news/2` (draft) | HTTP 404 |
| Невалидный ID | GET `/news/abc` | HTTP 404 |
| Гость видит статью | GET `/news/1` (без cookie) | HTTP 200 |
| Комментарии отображаются | GET `/news/1` (есть комментарии) | HTML содержит текст комментариев |
| Нет комментариев (гость) | GET `/news/1` (без комментариев, без cookie) | HTML содержит «Войдите, чтобы оставить первый» |
| Нет комментариев (пользователь) | GET `/news/1` (без комментариев, с cookie) | HTML содержит «Будьте первым» |
| Создание комментария | POST `/news/1/comments` (авторизован) | HTTP 201, комментарий в БД |
| Гость создаёт комментарий | POST `/news/1/comments` (без cookie) | HTTP 302 redirect на `/login?next=/news/1` |
| Пустой комментарий | POST `/news/1/comments` (пустой content) | HTTP 422 |
| Ответ на комментарий (корневой) | POST `/news/1/comments` (parent_id = корневой комментарий) | HTTP 201, parent_id в БД, snapshot-поля заполнены |
| Ответ на ответ (depth > 1) | POST `/news/1/comments` (parent_id = уже ответ) | HTTP 422 (`ErrParentNotFound`) |
| HTMX partial | GET `/news/1` + `HX-Request: true` | Ответ без полного layout |
| Форма ответа (HTMX) | POST `/news/1/comments` + `HX-Request: true` | HTTP 201, обновлённый HTML `#comments-list` |
| HTMX 422 ретаргетинг | POST `/news/1/comments` + `HX-Request: true` (пустой content) | HTTP 422, заголовок `HX-Retarget: #comment-form` |
| HTML: якорь на родителя | GET `/news/1` (есть ответ) | HTML содержит `<a href="#comment-{parentID}">` |
| HTML: удалённый родитель | GET `/news/1` (родитель удалён) | HTML содержит `class="comment-quote comment-quote--deleted"`, нет `<a href>` |

### Визуальная проверка (ручная)

| Сценарий | Ожидаемый результат |
|---|---|
| Открыть статью | Заголовок, дата, полный текст отображаются корректно |
| Создать комментарий | Комментарий появляется без перезагрузки |
| Ответить на комментарий | Форма появляется, ответ сохраняется с цитатой |
| Кликнуть на цитату | Прокрутка к родительскому комментарию, подсветка |
| `@username` в комментарии | Mention выделен цветом |
| Гостевой просмотр | Форма не видна, ссылка на вход |
| Мобильный вид (< 768px) | Контент читаем, формы удобны |

---

## 17. Не входит в scope

- Редактирование и удаление комментариев
- Модерация комментариев (скрытие, бан)
- Нотификации при mention (email, SSE, in-app)
- Автокомплит `@username` при вводе
- Markdown или rich text в комментариях
- Пагинация комментариев (все загружаются одним запросом)
- Вложенные ответы глубже 1 уровня (reply на reply создаёт плоский комментарий с цитатой, не дерево)
- Like / dislike комментариев
- Аватары пользователей
- Счётчик комментариев на главной странице
- SEO-метатеги (`og:title`, `og:description`)
- Кеширование страницы статьи
- Rate-limiting на создание комментариев

---

_Spec v1.0 | 2026-04-05_
