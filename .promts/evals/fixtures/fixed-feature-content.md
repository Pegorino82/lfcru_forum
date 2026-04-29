Прочитал memory-bank/flows/feature-flow.md, trello.md, domain/architecture.md, domain/frontend.md,
domain/problem.md, engineering/testing-policy.md, domain/glossary.md. Создаю feature package.

Обновлена memory-bank/features/README.md — добавлена запись FT-XXX: | [FT-XXX](FT-XXX/) | WYSIWYG-редактор | planned |

## Feature Package FT-XXX

**Change Surface:**
- `templates/admin/articles/edit.html` — форма редактирования статьи
- `static/js/editor.js` — TipTap vanilla JS интеграция
- `static/css/editor.css` — стили редактора (только inline или в шаблоне)
- `internal/article/service.go` — bluemonday-санитизация в Service-слое

**CHK-01:** Playwright E2E — открыть редактор, применить Bold/H2, проверить `<strong>`, `<h2>` в DOM.
**CHK-02:** Playwright E2E — вставить изображение, проверить `<figure><img>` в DOM.

ASM-04: CSRF-токен обеспечен echo middleware для всех POST/PUT — зафиксировано явно (PCON-02).

NS-03: Миграция существующих Markdown-статей вне scope данной фичи.
