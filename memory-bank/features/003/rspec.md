# [APPROVED] 03_homepage — Спецификация (SDD)

## 1. Контекст

Главная страница — первый экран сайта LFC.ru. Сейчас `/` отображает только заголовок «Добро пожаловать на LFC.ru» без полезного контента. Задача — наполнить главную страницу тремя информационными блоками: новости клуба, ближайший матч, последняя активность форума. Это позволит посетителю сразу оценить актуальность ресурса.

**Ссылка на бриф:** [brief.md](brief.md)
**GitHub Issue:** [#3](https://github.com/Pegorino82/lfcru_forum/issues/3)

---

## 2. Акторы

| Актор | Описание |
|---|---|
| Гость | Незарегистрированный посетитель. Видит главную без ограничений |
| Пользователь | Зарегистрированный аккаунт. Видит ту же главную страницу |

> Для данной фичи (homepage) поведение одинаково для всех ролей. Авторизация влияет только на хэдер (уже реализовано в фиче 002).

---

## 3. Текущее состояние

- `GET /` обслуживается хэндлером `internal/home/handler.go` → рендерит `templates/home/index.html`
- Шаблон `index.html` содержит только `<h1>Добро пожаловать на LFC.ru</h1>` и `<p>`
- Хэндлер поддерживает HTMX partial render (`HX-Request`)
- Таблиц `news`, `matches`, `forum_topics`, `forum_posts` в БД нет — необходимо создать

---

## 4. Сценарии использования (Use Cases)

### UC-1: Просмотр последних новостей

**Актор:** Гость / Пользователь

**Предусловие:** Пользователь открывает `/`.

**Основной поток:**

1. Система загружает из БД до 5 последних опубликованных новостей, отсортированных по `published_at DESC`
2. Система отображает блок «Последние новости» со списком записей
3. Каждая запись содержит: заголовок (ссылка на `/news/{id}`), дата публикации в формате `02.01.2006`

**Альтернативный поток:**

- **1a.** В таблице `news` нет записей с `is_published = true` → блок отображает текст «На сайте еще не добавлены новости»

**Acceptance Criteria:**

- **Given** в БД есть 7 опубликованных новостей, **When** пользователь открывает `/`, **Then** блок «Последние новости» содержит ровно 5 записей, отсортированных по дате (новые первыми)
- **Given** в БД есть 3 опубликованные новости, **When** пользователь открывает `/`, **Then** блок содержит 3 записи
- **Given** в БД нет опубликованных новостей, **When** пользователь открывает `/`, **Then** блок содержит текст «На сайте еще не добавлены новости»
- **Given** в БД есть опубликованные и неопубликованные (draft) новости, **When** пользователь открывает `/`, **Then** блок содержит только опубликованные
- Каждый заголовок новости является ссылкой `<a href="/news/{id}">`. Страница `/news/{id}` пока не реализована — ссылка ведёт на 404

### UC-2: Просмотр ближайшего матча

**Актор:** Гость / Пользователь

**Предусловие:** Пользователь открывает `/`.

**Основной поток:**

1. Система загружает из БД ближайший матч: запись из `matches` с `match_date >= now()`, отсортированная по `match_date ASC`, `LIMIT 1`
2. Система отображает блок «Ближайший матч» с информацией: соперник, дата матча в формате `02.01.2006 15:04`, турнир

**Альтернативный поток:**

- **1a.** В таблице `matches` нет записей с `match_date >= now()` → блок отображает текст «Ближайших матчей нет»

**Acceptance Criteria:**

- **Given** в БД есть 3 будущих матча, **When** пользователь открывает `/`, **Then** блок «Ближайший матч» отображает матч с самой ранней датой из будущих
- **Given** в БД нет будущих матчей (все в прошлом), **When** пользователь открывает `/`, **Then** блок содержит текст «Ближайших матчей нет»
- **Given** в БД нет записей в таблице `matches`, **When** пользователь открывает `/`, **Then** блок содержит текст «Ближайших матчей нет»
- Блок содержит имя соперника, дату/время матча и название турнира

### UC-3: Просмотр последней активности форума

**Актор:** Гость / Пользователь

**Предусловие:** Пользователь открывает `/`.

**Основной поток:**

1. Система загружает из БД до 5 тем форума с самой свежей активностью: темы, отсортированные по `last_post_at DESC`, `LIMIT 5`
2. Система отображает блок «Последнее на форуме» со списком тем
3. Каждая запись содержит: название темы (ссылка на `/forum/topics/{id}`), автор последнего сообщения (username), дата последнего сообщения в формате `02.01.2006 15:04`

**Альтернативный поток:**

- **1a.** В таблице `forum_topics` нет записей **ИЛИ** все темы без сообщений (`last_post_at IS NULL`) → блок отображает текст «В форуме пока нет активных обсуждений»

**Acceptance Criteria:**

- **Given** в БД есть 8 тем с сообщениями, **When** пользователь открывает `/`, **Then** блок «Последнее на форуме» содержит ровно 5 тем, отсортированных по дате последнего сообщения (новые первыми)
- **Given** в БД есть 2 темы с сообщениями, **When** пользователь открывает `/`, **Then** блок содержит 2 записи
- **Given** в БД нет тем, **When** пользователь открывает `/`, **Then** блок содержит текст «В форуме пока нет активных обсуждений»
- **Given** в БД есть темы, но ни одна не имеет сообщений (`last_post_at IS NULL`), **When** пользователь открывает `/`, **Then** блок содержит текст «В форуме пока нет активных обсуждений»
- Название каждой темы является ссылкой `<a href="/forum/topics/{id}">`. Страница `/forum/topics/{id}` пока не реализована — ссылка ведёт на 404
- Отображается username автора последнего сообщения, а не автора темы (если они различаются)

---

## 5. Модель данных

### Таблица `news`

```sql
CREATE TABLE news (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title        VARCHAR(255)  NOT NULL,
    content      TEXT          NOT NULL DEFAULT '',
    is_published BOOLEAN       NOT NULL DEFAULT false,
    author_id    BIGINT        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    published_at TIMESTAMPTZ,
    CONSTRAINT chk_news_published CHECK (is_published = false OR published_at IS NOT NULL),
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_news_published ON news (published_at DESC)
    WHERE is_published = true;
```

> **Примечания:**
> - `content` — тело новости; для главной страницы не используется, но необходим для будущей страницы `/news/{id}`.
> - `is_published` — черновик/опубликовано. На главной отображаются только опубликованные (`is_published = true`).
> - `published_at` — дата публикации; устанавливается при первой публикации. Используется для сортировки. Может быть `NULL` для черновиков.
> - Partial index `idx_news_published` покрывает запрос главной (только опубликованные, сортировка по дате).
> - `author_id` — автор новости; на главной не отображается, но нужен для администрирования.

### Таблица `matches`

```sql
CREATE TABLE matches (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    opponent     VARCHAR(255)  NOT NULL,
    match_date   TIMESTAMPTZ   NOT NULL,
    tournament   VARCHAR(255)  NOT NULL,
    is_home      BOOLEAN       NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_matches_date ON matches (match_date ASC);
```

> **Примечания:**
> - `opponent` — название команды-соперника (например, «Манчестер Юнайтед»).
> - `match_date` — дата и время матча в UTC. Для отображения конвертируется в `Europe/Moscow` (UTC+3). Итоговый формат: `«02.01.2006 15:04 МСК»`. Timezone конфигурируется через переменную окружения `DISPLAY_TIMEZONE` (дефолт `Europe/Moscow`).
> - `tournament` — название турнира (например, «Премьер-лига», «Лига Чемпионов»).
> - `is_home` — домашний / гостевой матч. На главной не отображается в данной фиче, но поле необходимо для будущих задач (отображение «Ливерпуль — Соперник» vs «Соперник — Ливерпуль»).

### Таблица `forum_sections`

```sql
CREATE TABLE forum_sections (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

> **Примечание:** Минимальная таблица разделов — нужна как FK для тем. Полная реализация форума — отдельная задача.

### Таблица `forum_topics`

```sql
CREATE TABLE forum_topics (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    section_id     BIGINT        NOT NULL REFERENCES forum_sections(id) ON DELETE RESTRICT,
    title          VARCHAR(255)  NOT NULL,
    author_id      BIGINT        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    post_count     INT           NOT NULL DEFAULT 0,
    last_post_at   TIMESTAMPTZ,
    last_post_by   BIGINT        REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_topics_last_post ON forum_topics (last_post_at DESC NULLS LAST);
```

> **Примечания:**
> - `last_post_at` и `last_post_by` — денормализованные поля, обновляемые при создании нового сообщения. Это избавляет от JOIN при выборке для главной страницы.
> - `last_post_at = NULL` для тем без сообщений (только что созданная тема без ответов). Такие темы попадают в конец списка (`NULLS LAST`).
> - `post_count` — денормализованный счётчик; на главной не отображается, но необходим для будущей страницы форума.
> - `section_id` — FK на раздел; на главной не используется, но обязателен для структуры форума.
> - **Механизм синхронизации:** `last_post_at`, `last_post_by`, `post_count` обновляются триггером PostgreSQL `AFTER INSERT OR DELETE ON forum_posts`. Триггер выполняет `UPDATE forum_topics SET post_count = ..., last_post_at = ..., last_post_by = ... WHERE id = NEW.topic_id`. При удалении последнего сообщения в теме: `last_post_at = NULL`, `last_post_by = NULL`, `post_count = 0`. **Реализация триггера входит в scope данной фичи** — без него блок «Последнее на форуме» всегда будет пустым.

### Таблица `forum_posts`

```sql
CREATE TABLE forum_posts (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic_id    BIGINT       NOT NULL REFERENCES forum_topics(id) ON DELETE CASCADE,
    author_id   BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content     TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_forum_posts_topic ON forum_posts (topic_id, created_at ASC);
```

> **Примечание:** Таблица сообщений необходима для полноты модели данных. На главной не запрашивается напрямую — данные о последнем сообщении денормализованы в `forum_topics`.

### Миграции

Создать два файла миграций (goose, формат `NNNN_description.sql`):

| Файл | Содержимое |
|---|---|
| `004_create_news.sql` | Таблица `news` + индекс |
| `005_create_forum_and_matches.sql` | Таблицы `matches`, `forum_sections`, `forum_topics`, `forum_posts` + индексы |

> Разделение на два файла: `news` — независимая сущность; `matches` + `forum_*` — могут быть в одной миграции, так как создаются одновременно и взаимосвязаны (forum_* между собой).

---

## 6. Структура Go-пакетов

### Новые пакеты

```
internal/
├── news/
│   ├── model.go           # struct News
│   └── repo.go            # NewsRepo: LatestPublished(ctx, limit int) ([]News, error)
├── match/
│   ├── model.go           # struct Match
│   └── repo.go            # MatchRepo: NextUpcoming(ctx) (*Match, error)
└── forum/
    ├── model.go           # struct Section, Topic, Post, TopicWithLastAuthor
    └── repo.go            # TopicRepo: LatestActive(ctx, limit int) ([]TopicWithLastAuthor, error)
```

### Модели

```go
// internal/news/model.go
type News struct {
    ID          int64
    Title       string
    Content     string
    IsPublished bool
    AuthorID    int64
    PublishedAt *time.Time // nil для черновиков
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

```go
// internal/match/model.go
type Match struct {
    ID         int64
    Opponent   string
    MatchDate  time.Time
    Tournament string
    IsHome     bool
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

```go
// internal/forum/model.go
type Topic struct {
    ID         int64
    SectionID  int64
    Title      string
    AuthorID   int64
    PostCount  int
    LastPostAt *time.Time
    LastPostBy *int64
    CreatedAt  time.Time
}

// TopicWithLastAuthor — проекция для главной страницы.
// Содержит username автора последнего сообщения (JOIN с users).
type TopicWithLastAuthor struct {
    ID             int64
    Title          string
    LastPostAt     time.Time
    LastPostByName string // username из таблицы users
}
```

### Репозитории — SQL-запросы

**`NewsRepo.LatestPublished(ctx, limit)`:**

```sql
SELECT id, title, published_at
FROM news
WHERE is_published = true
ORDER BY published_at DESC
LIMIT $1
```

Возвращает `[]News` (только `ID`, `Title`, `PublishedAt` заполнены; остальные поля — нулевые значения Go). При пустом результате возвращает пустой слайс `[]News{}`, а не `nil` и не ошибку.

> **Частичное заполнение модели — намеренно.** SQL выбирает только три поля, достаточных для главной страницы. Если потребуется строгая типизация — создать `NewsPreview{ID int64; Title string; PublishedAt *time.Time}`.

**`MatchRepo.NextUpcoming(ctx, asOf time.Time)`:**

```sql
SELECT id, opponent, match_date, tournament, is_home
FROM matches
WHERE match_date >= $1
ORDER BY match_date ASC
LIMIT 1
```

Возвращает `*Match` или `nil` если будущих матчей нет. Ошибка — только при проблемах с БД.

> **Параметр `asOf`:** передаётся явно для детерминированности тестов. В хэндлере передаётся `time.Now()`. В тестах — фиксированное время. Семантика: матч с `match_date = asOf` (точная граница) считается ближайшим (`>=`).

**`TopicRepo.LatestActive(ctx, limit)`:**

```sql
SELECT t.id, t.title, t.last_post_at, COALESCE(u.username, '[удалён]') AS last_post_by_name
FROM forum_topics t
LEFT JOIN users u ON u.id = t.last_post_by
WHERE t.last_post_at IS NOT NULL
ORDER BY t.last_post_at DESC
LIMIT $1
```

Возвращает `[]TopicWithLastAuthor`. Темы без сообщений (`last_post_at IS NULL`) не включаются — они ещё не имели активности. При пустом результате возвращает пустой слайс.

> **`LEFT JOIN` вместо `INNER JOIN`:** если `last_post_by` ссылается на удалённого пользователя (или временно равен NULL из-за гонки), тема не должна пропадать из выборки. `COALESCE(u.username, '[удалён]')` обеспечивает корректное отображение в этих случаях.

---

## 7. Изменения в существующих файлах

### `internal/home/handler.go`

Хэндлер `ShowHome` расширяется:

1. Принимает зависимости: `NewsRepo`, `MatchRepo`, `TopicRepo` (через структуру или функциональные параметры)
2. Вызывает три метода репозиториев для получения данных
3. Передаёт данные в шаблон

**Данные для шаблона (TemplateData):**

```go
type HomeData struct {
    User       *user.User            // из контекста (middleware)
    CSRFToken  string                // из middleware
    News       []news.News           // до 5 записей или пустой слайс
    NextMatch  *match.Match          // nil если нет будущих матчей
    Topics     []forum.TopicWithLastAuthor // до 5 записей или пустой слайс
}
```

**Порядок вызовов:** три запроса к БД независимы — могут выполняться последовательно (простота) или параллельно (через `errgroup`). Рекомендация: последовательно, так как суммарное время трёх `LIMIT 1/5` запросов пренебрежимо мало.

**Обработка ошибок:** стратегия «всё или ничего». При первой ошибке любого репозитория — немедленно прервать (fail-fast), вернуть HTTP 500. `pgx.ErrNoRows` / `pgconn.ErrNoRows` **не являются ошибкой** — возвращают `nil` / пустой слайс. Логировать: `log.Error("home: failed to load data", "err", err)`.

> **Обоснование:** простота на текущем этапе. При необходимости graceful degradation (показывать доступные блоки при отказе одного репозитория) — переработать Handler на независимые вызовы с индивидуальными fallback-значениями.

> **HTMX + ошибка:** при `HX-Request: true` и HTTP 500 полный HTML-layout **не возвращается**. Клиентская обработка через `htmx.on('htmx:responseError', handler)` в `base.html`.

### `templates/home/index.html`

Полная замена содержимого шаблона. Три блока на странице (см. [§9 Шаблон](#9-шаблон-страницы)).

### `cmd/forum/main.go`

Добавить:
1. Создание `NewsRepo`, `MatchRepo`, `TopicRepo` (передать `pgxpool.Pool`)
2. Передачу репозиториев в `home.Handler` (или в `ShowHome`)
3. Миграции уже применяются автоматически через `goose.Up`

---

## 8. API-контракт

### GET /

**Описание:** Главная страница сайта.

**Авторизация:** Не требуется. Доступна гостям.

**Ответ:** HTML (полный layout через `base.html`).

**HTMX:** При `HX-Request: true` — возвращает только блок `content` (без `<header>`, `<footer>`).

**Данные:**

| Блок | Источник | Пустое состояние |
|---|---|---|
| Последние новости | `NewsRepo.LatestPublished(ctx, 5)` | «На сайте еще не добавлены новости» |
| Ближайший матч | `MatchRepo.NextUpcoming(ctx)` | «Ближайших матчей нет» |
| Последнее на форуме | `TopicRepo.LatestActive(ctx, 5)` | «В форуме пока нет активных обсуждений» |

**Ответы:**

| Код | Условие | Действие |
|---|---|---|
| 200 | Успех | HTML с тремя блоками данных |
| 500 | Ошибка БД | Брендированная страница ошибки в рамках `base.html`: «Что-то пошло не так. Попробуйте обновить страницу.» |

---

## 9. Шаблон страницы

### Целевая структура HTML

```html
{{define "content"}}
<div class="home-grid">

  <!-- Блок новостей -->
  <section class="home-news" aria-labelledby="news-heading">
    <h2 id="news-heading">Последние новости</h2>
    {{if .News}}
      <ul class="news-list">
        {{range .News}}
        <li>
          <a href="/news/{{.ID}}">{{.Title}}</a>
          {{if .PublishedAt}}
          <time datetime="{{.PublishedAt.Format "2006-01-02"}}">
            {{.PublishedAt.Format "02.01.2006"}}
          </time>
          {{end}}
        </li>
        {{end}}
      </ul>
    {{else}}
      <p class="empty-state">На сайте еще не добавлены новости</p>
    {{end}}
  </section>

  <!-- Блок ближайшего матча -->
  <section class="home-match" aria-labelledby="match-heading">
    <h2 id="match-heading">Ближайший матч</h2>
    {{if .NextMatch}}
      <div class="match-card">
        <p class="match-opponent">{{.NextMatch.Opponent}}</p>
        <p class="match-date">
          <time datetime="{{.NextMatch.MatchDate.Format "2006-01-02T15:04:05Z07:00"}}">
            {{.NextMatch.MatchDate.Format "02.01.2006 15:04"}} МСК
          </time>
        </p>
        <p class="match-tournament">{{.NextMatch.Tournament}}</p>
      </div>
    {{else}}
      <p class="empty-state">Ближайших матчей нет</p>
    {{end}}
  </section>

  <!-- Блок активности форума -->
  <section class="home-forum" aria-labelledby="forum-heading">
    <h2 id="forum-heading">Последнее на форуме</h2>
    {{if .Topics}}
      <ul class="forum-list">
        {{range .Topics}}
        <li>
          <a href="/forum/topics/{{.ID}}">{{.Title}}</a>
          <span class="forum-meta">
            {{.LastPostByName}} —
            <time datetime="{{.LastPostAt.Format "2006-01-02T15:04:05Z07:00"}}">
              {{.LastPostAt.Format "02.01.2006 15:04"}}
            </time>
          </span>
        </li>
        {{end}}
      </ul>
    {{else}}
      <p class="empty-state">В форуме пока нет активных обсуждений</p>
    {{end}}
  </section>

</div>
{{end}}
```

### Правила отображения

| Элемент | Формат | Пример |
|---|---|---|
| Дата новости | `02.01.2006` | `15.03.2026` |
| Дата/время матча | `02.01.2006 15:04 МСК` | `20.04.2026 21:00 МСК` |
| Дата/время форума | `02.01.2006 15:04` | `04.04.2026 14:30` |
| `<time datetime>` | ISO 8601 | `2026-04-20T21:00:00+03:00` |

> Все `<time>` элементы имеют атрибут `datetime` в формате ISO 8601 для машинной читаемости и accessibility.

---

## 10. Стилизация

Минимальная стилизация (inline `<style>` в `base.html` или в шаблоне `index.html`, как принято в проекте):

| Элемент | Стили |
|---|---|
| `.home-grid` | `display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; padding: 1rem 0;` |
| `.home-news` | `grid-column: 1 / 2; grid-row: 1 / 2;` |
| `.home-match` | `grid-column: 2 / 3; grid-row: 1 / 2;` |
| `.home-forum` | `grid-column: 1 / -1; grid-row: 2 / 3;` |
| `.news-list, .forum-list` | `list-style: none; padding: 0; margin: 0;` |
| `.news-list li, .forum-list li` | `padding: 0.5rem 0; border-bottom: 1px solid #eee;` |
| `.news-list time, .forum-meta` | `color: #666; font-size: 0.875rem;` |
| `.match-card` | `background: #f9f9f9; border: 1px solid #eee; border-radius: 8px; padding: 1rem; text-align: center;` |
| `.match-opponent` | `font-size: 1.25rem; font-weight: bold;` |
| `.match-tournament` | `color: #666; font-size: 0.875rem;` |
| `.empty-state` | `color: #999; font-style: italic;` |
| `section h2` | `font-size: 1.25rem; margin-bottom: 0.75rem; border-bottom: 2px solid #c8102e; padding-bottom: 0.5rem;` |

### Responsive (< 768px)

```css
@media (max-width: 768px) {
  .home-grid {
    grid-template-columns: 1fr;
  }
  .home-news, .home-match, .home-forum {
    grid-column: 1 / -1;
  }
}
```

### Требование «все три блока видны без прокрутки» (≥ 1280×900px)

Grid-layout с двумя колонками (`news` + `match` в первом ряду, `forum` во втором) гарантирует, что при viewport ≥ 1280×900px все три блока помещаются без прокрутки. Максимальная высота контента: header (~60px) + grid (~700px) + footer (~60px) = ~820px < 900px.

> Если при 5 новостях и 5 темах контент превышает viewport — уменьшить padding. Точная подгонка — ответственность визуальной проверки.

---

## 11. Accessibility

| Требование | Реализация |
|---|---|
| Семантические секции | Каждый блок в `<section>` с `aria-labelledby` |
| Заголовки блоков | `<h2>` с уникальными `id` |
| Машиночитаемые даты | `<time datetime="ISO 8601">` |
| Пустые состояния | Текст в `<p>`, видимый для скринридеров |
| Списки | `<ul>/<li>` для новостей и тем форума |

---

## 12. HTMX-совместимость

Поведение аналогично фиче 002:

- При `HX-Request: true` — сервер возвращает только блок `content` (без `<header>`, `<footer>`, `<html>`, `<body>`)
- При обычном запросе — полный layout

Хэндлер `ShowHome` уже поддерживает эту логику — она не требует изменений.

**Сценарий инициации HTMX-запроса:** навигация через `hx-boost="true"` на элементе `<body>` или `<main>` в `base.html`. При клике по ссылке с `hx-boost` браузер делает HTMX-запрос с заголовком `HX-Request: true` — сервер возвращает только содержимое `{{block "content"}}`, которое HTMX подставляет в целевой контейнер. Это стандартный механизм SPA-навигации без full page reload.

---

## 13. Тест-план

### Юнит-тесты

| Компонент | Что проверяем |
|---|---|
| `NewsRepo.LatestPublished` | Возвращает пустой слайс при отсутствии данных |
| `NewsRepo.LatestPublished` | Возвращает не более `limit` записей |
| `NewsRepo.LatestPublished` | Не включает неопубликованные (`is_published = false`) |
| `NewsRepo.LatestPublished` | Сортировка по `published_at DESC` |
| `MatchRepo.NextUpcoming` | Возвращает `nil` при отсутствии будущих матчей |
| `MatchRepo.NextUpcoming` | Возвращает ближайший будущий матч |
| `MatchRepo.NextUpcoming` | Не возвращает прошедшие матчи |
| `TopicRepo.LatestActive` | Возвращает пустой слайс при отсутствии данных |
| `TopicRepo.LatestActive` | Возвращает не более `limit` записей |
| `TopicRepo.LatestActive` | Сортировка по `last_post_at DESC` |
| `TopicRepo.LatestActive` | Не включает темы без сообщений (`last_post_at IS NULL`) |
| `TopicRepo.LatestActive` | Возвращает корректный `LastPostByName` (username из users) |
| `TopicRepo.LatestActive` | При `last_post_by` удалённого пользователя возвращает `«[удалён]»` (LEFT JOIN) |
| `MatchRepo.NextUpcoming` | Детерминировано при передаче явного `asOf` — тест проходит в любое время суток |

> Тесты репозиториев — интеграционные (build tag `integration`), так как проверяют SQL-запросы.

### Интеграционные тесты (HTTP)

| Сценарий | URL | Ожидаемый результат |
|---|---|---|
| Главная без данных — новости | GET `/` | HTML содержит «На сайте еще не добавлены новости» |
| Главная без данных — матч | GET `/` | HTML содержит «Ближайших матчей нет» |
| Главная без данных — форум | GET `/` | HTML содержит «В форуме пока нет тем» |
| Главная с данными — новости | GET `/` (после INSERT в news) | HTML содержит заголовки новостей |
| Главная с данными — матч | GET `/` (после INSERT в matches) | HTML содержит имя соперника |
| Главная с данными — форум | GET `/` (после INSERT в topics + posts) | HTML содержит название темы и username |
| HTMX partial | GET `/` + `HX-Request: true` | Ответ не содержит `<html>`, `<head>`, `<body>` полного layout |
| Ошибка БД | GET `/` (mock/repo возвращает ошибку) | HTTP 500 |
| Доступ гостя | GET `/` (без cookie) | HTTP 200 |
| Темы без сообщений | GET `/` (темы есть, `last_post_at IS NULL`) | HTML содержит «В форуме пока нет активных обсуждений» |
| Секции HTML | GET `/` | HTML содержит `<section` с `aria-labelledby` |
| Заголовки блоков | GET `/` | HTML содержит `id="news-heading"`, `id="match-heading"`, `id="forum-heading"` |

### Визуальная проверка (ручная)

| Сценарий | Ожидаемый результат |
|---|---|
| Открыть `/` с пустой БД | Три блока с пустыми состояниями видны |
| Открыть `/` с данными | Три блока с данными видны без прокрутки (≥ 1280×900px) |
| Ширина < 768px | Блоки выстраиваются в одну колонку |
| Клик по заголовку новости | Переход на `/news/{id}` (404 на данном этапе) |
| Клик по названию темы | Переход на `/forum/topics/{id}` (404 на данном этапе) |

---

## 14. Seed-данные для тестирования и демонстрации

CRUD администрирования новостей и матчей вынесен за scope данной фичи. Для первоначального наполнения и тестирования использовать прямые SQL-вставки:

```sql
-- Тестовые новости
INSERT INTO news (title, is_published, author_id, published_at)
VALUES
  ('Ливерпуль победил в дерби', true, 1, now() - interval '1 day'),
  ('Трансферное окно: последние новости', true, 1, now() - interval '3 days'),
  ('Черновик: не публиковать', false, 1, NULL);

-- Тестовые матчи
INSERT INTO matches (opponent, match_date, tournament, is_home)
VALUES
  ('Манчестер Юнайтед', now() + interval '7 days', 'Премьер-лига', true),
  ('Реал Мадрид', now() + interval '14 days', 'Лига Чемпионов', false);
```

> Seed-данные для форума появятся в рамках фичи полного форума. Для тестирования блока «Последнее на форуме» необходимо также создать запись в `forum_sections`.

---

## 15. Не входит в scope

- Страницы `/news/{id}`, `/forum/topics/{id}` — ссылки создаются, но страницы реализуются в отдельных задачах
- Администрирование новостей (CRUD) — отдельная задача
- Администрирование матчей (CRUD) — отдельная задача
- Полная реализация форума (разделы, создание тем, сообщения) — отдельная задача
- Пагинация блоков — на главной фиксированный лимит (5 записей, 1 матч)
- Кеширование данных главной — при текущей нагрузке не требуется
- Real-time обновление блоков (SSE/WebSocket) — отдельная задача
- Timezone пользователя — время отображается в серверной timezone
- Счётчик сообщений в теме на главной
- Аватары пользователей

---

## 15. Структура файлов (целевая)

```
migrations/
├── 004_create_news.sql
└── 005_create_forum_and_matches.sql

internal/
├── news/
│   ├── model.go
│   └── repo.go
├── match/
│   ├── model.go
│   └── repo.go
├── forum/
│   ├── model.go
│   └── repo.go
└── home/
    ├── handler.go              # расширенный (новые зависимости)
    └── handler_test.go         # интеграционные тесты

templates/
└── home/
    └── index.html              # полная замена содержимого
```

---

> **[Approved by @Pegorino82 2026-04-04]**
> Спецификация прошла архитектурное и бизнес-ревью.
> Итераций: 1. Исправлено проблем: 32 (10 критичных, 22 высоких).

---
_Spec v1.1 | 2026-04-04 | Обновлено по результатам spec-review (итерация 1)_

---
_Spec Review v1.11.0 | 2026-04-04_
