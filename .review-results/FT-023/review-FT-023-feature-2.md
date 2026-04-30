# Ревью FT-023/feature.md x memory-bank

## Blocker 1 — Неверные пути файлов в Change Surface

feature.md, строки 64–70, указывает несуществующий префикс `web/`:

| feature.md (неверно) | Реальный путь в проекте |
|---|---|
| `web/templates/article/edit.html` | `templates/admin/articles/edit.html` |
| `web/templates/article/view.html` | `templates/news/article.html` |
| `web/static/js/editor.js` | `static/js/editor.js` |
| `web/static/css/` | `static/css/` |

Директория `web/` в проекте отсутствует. `frontend.md` явно фиксирует `templates/<domain>/` и `static/`. Ошибка системная — agent/reviewer, читая feature.md, будет работать не с теми файлами.

## Blocker 2 — CHK-01/CHK-02 помечены manual-only, нарушая testing-policy.md

CHK-01 и CHK-02 обозначены как «Ручная» проверка. `testing-policy.md` (строки 107–108):

> UI-изменения **обязаны** пройти Playwright-верификацию: отсутствие JS-ошибок в консоли + assertions на наличие элементов, текст и видимость. **Падение Playwright — блокер closure gate.**
> Browser-специфика и HTMX/Alpine.js-взаимодействия — покрываются Playwright, **не являются основанием для manual-only.**

- CHK-01 (проверка форматирования в DOM) — автоматизируем Playwright-ом: check наличия `<strong>`, `<h2>`, `style="text-align:..."` и т.д.
- CHK-02 (вставка изображения) — автоматизируем: check `<figure>`, `<img>`, `<figcaption>` в DOM.

Единственное исключение по policy — «пиксельный layout, анимации, шрифты» — здесь неприменимо.

Также MET-01 (`measurement method: ручная проверка каждого пункта`) конфликтует с той же нормой.

**Необходимо**: добавить CHK-04 или расширить CHK-01/CHK-02 под Playwright E2E; manual остаётся только как дополнительная визуальная проверка.

## High — UC-001 не обновлён и не упомянут

`features/README.md` (строка 22):
> Если feature существенно меняет устойчивый сценарий проекта, UC-* должен быть создан или обновлён, **а feature.md должна ссылаться на него**.

`UC-001 "Публикация статьи"` (`active`, implemented by FT-008, FT-009) — FT-023 принципиально меняет этот flow: редактор заменяется, формат хранения меняется. feature.md не упоминает UC-001 ни в scope, ни в contracts, ни в dependencies.

**Необходимо**: в feature.md добавить ссылку на UC-001 + зафиксировать что UC-001 будет обновлён (добавить FT-023 в `Implemented by`).

## Low — Follow-up ADR-007 не зафиксирован в feature.md

ADR-007 (строка 70):
> `memory-bank/domain/architecture.md` необходимо обновить: зафиксировать, что `articles.body` хранит HTML.

Этот deliverable нигде не отражён в feature.md (ни в Change Surface, ни в Exit Criteria). Не является блокером сейчас (ADR ещё `proposed`), но риск упустить при переводе в `accepted`.

## Что соответствует корректно

- ADR-007 → `proposed`, CON-02 это правильно фиксирует
- ADR-005 → `accepted`, ASM-01 ссылается напрямую
- Трассировка REQ → CTR → FM → EC → SC → CHK → EVID полная
- CHK-03 — `Авто`, соответствует testing-policy
- Статус в features/README.md: `planned` — совпадает с feature.md `delivery_status: planned`

## Итог

| Приоритет | Проблема | Файл, строки |
|---|---|---|
| Блокер | Неверный префикс `web/` в Change Surface | feature.md:64–70 |
| Блокер | CHK-01/CHK-02 manual вместо Playwright | feature.md:127–129, MET-01:28 |
| High | UC-001 не обновлён, не упомянут | feature.md — нет ссылки |
| Low | Follow-up architecture.md из ADR-007 не в scope | feature.md — нет deliverable |
