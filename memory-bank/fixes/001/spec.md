# Spec: Fix/001 — изолированные template sets

## Требование

Каждый page-шаблон должен рендериться с **собственным** template set, содержащим только layout-файлы + сам page-файл. Блок `{{define "content"}}` в разных page-шаблонах не должен конфликтовать.

## Разделение файлов

| Тип | Критерий | Пример |
|-----|----------|--------|
| Layout | директория `layouts/` | `templates/layouts/base.html` |
| Page | всё остальное | `templates/auth/login.html` |

## Поведение `Renderer`

- `New(root)` — на каждый page-файл создаёт отдельный `*template.Template`: парсит все layout-файлы + page-файл
- `Render(name, data, ...)` — находит set по `name`, выполняет `ExecuteTemplate(w, name, data)`
- Если `name` не найден — возвращает ошибку `template %q not found`

## Ограничения

- Изменяется только `internal/tmpl/renderer.go`; handler-ы, шаблоны и маршруты — без изменений
- Имена шаблонов в handler-ах остаются прежними (`"templates/auth/login.html"` и т.д.)
