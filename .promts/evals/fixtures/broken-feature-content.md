Прочитал memory-bank/flows/feature-flow.md, trello.md. Создаю feature package.

## Feature Package FT-XXX

**Change Surface:**
- `web/templates/article/edit.html` — форма редактирования
- `web/static/js/editor.js` — TipTap интеграция
- `web/static/css/editor.css` — стили редактора
- `internal/handler/article.go` — bluemonday-санитизация в handler

**CHK-01:** Ручная проверка — открыть редактор, форматирование (<strong>, <h2>) отображается.
**CHK-02:** Ручная проверка — вставить изображение, проверить что появляется.

NS-03: Миграция существующих Markdown-статей обрабатывается отдельным OQ в implementation-plan.md — OQ-01.

FM-03: Если формат тела статьи Markdown — OQ-01.
