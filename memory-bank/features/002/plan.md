# Implementation Plan — 02_layout_base

**Спецификация:** [rspec.md](rspec.md)
**GitHub Issue:** [#2](https://github.com/Pegorino82/lfcru_forum/issues/2)

---

## Текущее состояние кода

| Элемент | Статус |
|---|---|
| `templates/layouts/base.html` | Есть `<nav>` без `<header>`, `<div class="container">` вместо `<main>`, нет `<footer>`, нет skip-link, нет навигационных ссылок |
| `{{template "nav" .}}` | Использует `.User.Username` (поле в модели), показывает имя без усечения |
| `internal/tmpl/renderer.go` | Нет `FuncMap`, нет `truncate` функции |
| GET `/` | Маршрут не зарегистрирован — хэндлера нет |
| Хэндлеры `/login`, `/register` | Не проверяют `HX-Request`, всегда рендерят полный layout |
| HTTP-тесты layout | Отсутствуют |

---

## Шаги реализации

### Шаг 1 — Обновить `internal/tmpl/renderer.go`

**Файл:** `internal/tmpl/renderer.go`

#### 1а. Добавить `truncate`

Добавить `template.FuncMap` с функцией `truncate(s string, n int) string`:
- Если `len([]rune(s)) <= n` — возвращает `s` без изменений
- Иначе возвращает первые `n` рун + `"…"` (символ `U+2026`, одна руна)

Зарегистрировать FuncMap при создании каждого `template.New("")`:

```go
// до:
t := template.New("")

// после:
t := template.New("").Funcs(funcMap)
```

где `funcMap` определён один раз в пакете как `template.FuncMap{"truncate": truncateFunc}`.

> **Важно:** `Funcs()` должен вызываться до `Parse()`, иначе шаблон не скомпилируется.

#### 1б. Добавить `RenderPartial`

Текущий `Renderer.Render` ищет шаблон по ключу в `r.sets` (ключи вида `"templates/auth/login.html"`). Имя `"content"` там никогда не будет найдено.

Добавить метод `RenderPartial`, который принимает ключ страницы и имя блока:

```go
// RenderPartial рендерит именованный блок из template set страницы.
// pageKey — полный ключ страницы (например "templates/auth/login.html").
// blockName — имя блока (например "content").
func (r *Renderer) RenderPartial(w io.Writer, pageKey, blockName string, data any) error {
    t, ok := r.sets[pageKey]
    if !ok {
        return fmt.Errorf("template %q not found", pageKey)
    }
    return t.ExecuteTemplate(w, blockName, data)
}
```

Хэндлеры вызывают `RenderPartial` явно при HTMX-запросах. Интерфейс `echo.Renderer` не затрагивается.

---

### Шаг 2 — Обновить `templates/layouts/base.html`

**Изменения в CSS (`<style>`):**

```css
/* Заменить строку body на: */
body { font-family: system-ui, sans-serif; background: #f5f5f5; color: #222;
       display: flex; flex-direction: column; min-height: 100vh; }

/* Добавить flex:1 к .container: */
.container { max-width: 960px; margin: 2rem auto; padding: 0 1rem; flex: 1; }

/* Добавить новые правила: */
.main-nav { display: flex; gap: 1rem; }
.nav-username { max-width: 120px; overflow: hidden; text-overflow: ellipsis;
                white-space: nowrap; display: inline-block; }
footer { background: #333; color: #ccc; padding: 1rem 1.5rem;
         text-align: center; font-size: 0.875rem; }
footer a { color: #fff; }
.skip-link { position: absolute; top: -40px; left: 0; background: #c8102e;
             color: #fff; padding: 0.5rem 1rem; z-index: 100; transition: top 0.1s; }
.skip-link:focus { top: 0; }

@media (max-width: 768px) {
  .main-nav { display: none; }
}
```

**Изменения в HTML `<body>`:**

```html
<body>
  <!-- 1. Skip-link (первый элемент) -->
  <a href="#content" class="skip-link">Перейти к содержимому</a>

  <!-- 2. Обернуть nav в header, добавить aria-label -->
  <header>
    <nav aria-label="Основная навигация">
      <a href="/" class="logo">LFC.ru</a>
      <!-- 3. Добавить .main-nav между логотипом и авторизацией -->
      <div class="main-nav">
        <a href="#" class="nav-link">Форум</a>
        <a href="#" class="nav-link">Новости</a>
      </div>
      <div class="nav-links">
        {{template "nav" .}}
      </div>
    </nav>
  </header>

  <!-- 4. div.container → main#content с tabindex="-1" -->
  <main id="content" class="container" tabindex="-1">
    {{template "flash" .}}
    {{block "content" .}}{{end}}
  </main>

  <!-- 5. Добавить footer -->
  <footer>
    <p>© 2026 LFC.ru — фан-сайт болельщиков ФК «Ливерпуль»</p>
    <p>Не является официальным сайтом Liverpool FC</p>
  </footer>

  <!-- script без изменений -->
</body>
```

**Изменения в `{{define "nav"}}`:**

```html
{{define "nav"}}
{{if .User}}
    <span class="nav-username" title="{{.User.Username}}">{{truncate .User.Username 20}}</span>
    <form method="POST" action="/logout" style="display:inline">
        <input type="hidden" name="_csrf" value="{{.CSRFToken}}">
        <button type="submit" ...>Выйти</button>
    </form>
{{else}}
    <a href="/login">Войти</a>
    <a href="/register">Регистрация</a>
{{end}}
{{end}}
```

> Поле модели — `.User.Username` (не `.User.Name` как в spec). Модель `user.User` содержит поле `Username string`.

---

### Шаг 3 — Создать home handler и шаблон

**Новый файл:** `internal/home/handler.go`

```go
package home

import (
    "net/http"

    "github.com/Pegorino82/lfcru_forum/internal/auth"
    appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
    "github.com/Pegorino82/lfcru_forum/internal/tmpl"
    "github.com/labstack/echo/v4"
)

func ShowHome(c echo.Context) error {
    data := map[string]any{
        "User":      auth.UserFromContext(c),
        "CSRFToken": appMiddleware.CSRFToken(c),
    }
    if c.Request().Header.Get("HX-Request") == "true" {
        r := c.Echo().Renderer.(*tmpl.Renderer)
        return r.RenderPartial(c.Response(), "templates/home/index.html", "content", data)
    }
    return c.Render(http.StatusOK, "templates/home/index.html", data)
}
```

> `auth.UserFromContext` — экспортированная функция (`internal/auth/middleware.go:46`), зависимость `home → auth` явная и приемлемая. `RenderPartial` требует type assertion `*tmpl.Renderer` — это нормально, поскольку приложение всегда использует именно этот renderer.

**Новый файл:** `templates/home/index.html`

```html
{{define "templates/home/index.html"}}
{{template "templates/layouts/base.html" .}}
{{end}}

{{define "content"}}
<h1>Добро пожаловать на LFC.ru</h1>
<p>Фан-сайт болельщиков ФК «Ливерпуль»</p>
{{end}}
```

**Регистрация маршрута в `cmd/forum/main.go`:**

```go
import "github.com/Pegorino82/lfcru_forum/internal/home"
// ...
e.GET("/", home.ShowHome)
```

---

### Шаг 4 — HTMX partial render в auth хэндлерах

**Файл:** `internal/auth/handler.go`

`RenderPartial` уже добавлен в Шаге 1б. Обновить `ShowLogin` и `ShowRegister`:

```go
func (h *Handler) ShowLogin(c echo.Context) error {
    if UserFromContext(c) != nil {
        return c.Redirect(http.StatusFound, "/")
    }
    data := loginData{CSRFToken: appMiddleware.CSRFToken(c)}
    if c.Request().Header.Get("HX-Request") == "true" {
        r := c.Echo().Renderer.(*tmpl.Renderer)
        return r.RenderPartial(c.Response(), "templates/auth/login.html", "content", data)
    }
    return c.Render(http.StatusOK, "templates/auth/login.html", data)
}

func (h *Handler) ShowRegister(c echo.Context) error {
    if UserFromContext(c) != nil {
        return c.Redirect(http.StatusFound, "/")
    }
    data := registerData{CSRFToken: appMiddleware.CSRFToken(c), Fields: map[string]string{}}
    if c.Request().Header.Get("HX-Request") == "true" {
        r := c.Echo().Renderer.(*tmpl.Renderer)
        return r.RenderPartial(c.Response(), "templates/auth/register.html", "content", data)
    }
    return c.Render(http.StatusOK, "templates/auth/register.html", data)
}
```

Добавить импорт `"github.com/Pegorino82/lfcru_forum/internal/tmpl"` в `handler.go`.

---

### Шаг 5 — Интеграционные HTTP-тесты layout

**Новый файл:** `internal/layout/layout_test.go` (build tag `//go:build integration`)

Тест-сервер: Echo с реальным renderer (из `templates/`), CSRF middleware, auth middleware.

Сценарии (из spec, раздел 11):

| # | Тест | URL | Проверка |
|---|---|---|---|
| 1 | Header присутствует | GET `/login` | `<header>` в ответе |
| 2 | Nav с aria-label | GET `/login` | `aria-label="Основная навигация"` |
| 3 | Footer присутствует | GET `/login` | `<footer>` в ответе |
| 4 | Копирайт | GET `/login` | `© 2026 LFC.ru` |
| 5 | Дисклеймер | GET `/login` | `Не является официальным сайтом Liverpool FC` |
| 6 | Гостевой блок | GET `/login` | `/login` и `/register` в nav |
| 7 | Авторизованный блок | GET `/` (с сессией) | `action="/logout"` в nav |
| 8 | HTMX partial — login | GET `/login` + `HX-Request: true` | нет `<header>`, нет `<footer>` |
| 9 | HTMX partial — register | GET `/register` + `HX-Request: true` | нет `<header>`, нет `<footer>` |
| 10 | Skip-link target | GET `/login` | `id="content"` в ответе |
| 11 | Skip-link | GET `/login` | `<a href="#content"` в ответе |
| 12 | Семантический main | GET `/login` | `<main` в ответе |

---

## Порядок коммитов

1. `feat(tmpl): add truncate function and RenderPartial method`
2. `feat(layout): update base.html — header, footer, skip-link, nav links`
3. `feat(home): add home handler and template for GET /`
4. `feat(auth): add HTMX partial render to ShowLogin and ShowRegister`
5. `test(layout): add HTTP integration tests for layout structure`

---

## Риски и замечания

| Риск | Митигация |
|---|---|
| Renderer не поддерживает render блока по имени | `RenderPartial` добавлен в Шаге 1б — до всех хэндлеров, которые его используют |
| `home` пакет зависит от `auth` пакета | Принято: `auth.UserFromContext` экспортирована, зависимость однонаправленная и оправдана |
| Поле модели `.User.Username` vs `.User.Name` в spec | В шаблоне использовать `.User.Username` — реальное поле модели |
| Тест авторизованного блока требует реальной сессии в БД | Использовать auth.Service напрямую для создания тестовой сессии (как в auth integration tests) |
| `truncate` на строках с многобайтовыми символами (кириллица) | Работать с `[]rune`, не с байтами — username может содержать кириллицу |

---

_Plan v1.1 | 2026-04-03_
