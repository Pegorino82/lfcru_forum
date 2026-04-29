# Ревью FT-023/feature.md на соответствие memory-bank

## CRITICAL

**1. Нарушение Layer Stack — санитизация в Handler вместо Service**

`Change Surface` и `Solution` помещают bluemonday-санитизацию в `internal/handler/article.go`. Это нарушает `architecture.md`:

> Handler — парсит запрос, вызывает Service, рендерит шаблон.
> Service — содержит **всю domain-логику**: валидацию, rate-limit политику...

XSS-санитизация — бизнес-логика безопасности, canonical место — Service-слой. `Change Surface` должна указывать на `internal/<content|article>/service.go` (или аналог), а не на handler.

---

## HIGH

**2. PCON-02 (CSRF) не упомянут**

`domain/problem.md` → `PCON-02`: CSRF-токен обязателен для всех POST/PUT/DELETE. FT-023 отправляет `TipTap.getHTML()` через POST/PUT-эндпоинт сохранения статьи. Ни `CON-*`, ни `ASM-*` не фиксирует, что CSRF уже обеспечен middleware — это молчаливое допущение, которое нужно зафиксировать явно (например, `ASM-04`).

**3. Producer в Evidence contract не соответствует Checks**

| Check | How to check | Producer в EVID-* |
|---|---|---|
| CHK-01 | Playwright E2E | human reviewer |
| CHK-02 | Playwright E2E | human reviewer |

Playwright — автоматизированный инструмент. `testing-policy.md`: "Browser-специфика и HTMX/Alpine.js-взаимодействия покрываются Playwright, не являются основанием для manual-only." Producer должен быть `automated` / `playwright runner`, не `human reviewer`.

---

## MEDIUM

**4. CON-02 и ASM-02 лучше оформить как DEC-***

По taxonomy из `feature-flow.md`:
- `CON-*` — ограничения
- `DEC-*` — blocking decisions (решение ещё не принято, что именно оно блокирует)

- **CON-02** ("ADR-007 имеет статус proposed; используется как hypothesis") — это blocking decision: если ADR отклонён, scope пересматривается. → `DEC-01`
- **ASM-02** ("способ сборки TipTap определяется в implementation-plan.md") — нерешённый вопрос, влияющий на Change Surface (`static/js/`). → `DEC-02`

**5. Не указано обновление `domain/architecture.md`**

`ADR-007` в разделе Follow-up явно требует:
> `memory-bank/domain/architecture.md` необходимо обновить: зафиксировать, что `articles.body` хранит HTML.

Этого обновления нет в `Change Surface` и нет среди артефактов фичи. Для `Design Ready` это gap: Change Surface должна отражать docs-изменения.

**6. UC-001 не упомянут**

`feature-flow.md` правило 11: если фича materially changes существующий project-level scenario, UC должен быть обновлён до closure. FT-023 меняет WF-01 (чтение статей) и workflow редактирования. `UC-001-article-publishing.md` не упомянут ни в `derived_from`, ни в `Change Surface`.

---

## INFO

**7. `static/css/` в Change Surface — неточная формулировка**

`frontend.md`: "Локальные стили допустимы только в рамках конкретной страницы." Запись `static/css/ или inline` в Change Surface неоднозначна — лучше уточнить: инлайн-стили внутри шаблона или конкретный файл.

---

## Резюме

| # | Severity | Что не так | Где исправить |
|---|---|---|---|
| 1 | CRITICAL | bluemonday в Handler вместо Service | `Change Surface`, `Solution`, `Flow` |
| 2 | HIGH | CSRF не зафиксирован | добавить `ASM-04` |
| 3 | HIGH | Producer CHK-01/02 = human вместо automated | `Evidence contract` |
| 4 | MEDIUM | CON-02, ASM-02 → DEC-* | `Constraints / Assumptions` |
| 5 | MEDIUM | architecture.md update не в Change Surface | `Change Surface` |
| 6 | MEDIUM | UC-001 не упомянут | `derived_from` или `Change Surface` |
| 7 | INFO | `static/css/` — неточно | `Change Surface` |

Для перехода в **Design Ready** блокерами являются пункты **1, 2, 3** — они нарушают core rules архитектуры и testing-policy. Пункты 4–6 — gap'ы, которые по правилам `feature-flow.md` также требуются до Design Ready.
