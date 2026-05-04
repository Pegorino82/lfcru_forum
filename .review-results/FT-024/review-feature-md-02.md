---
gate: "Draft → Design Ready"
artifact: feature.md
ft_id: FT-024
date: 2026-05-03
iteration: 2
outcome: revise
---

# Review: FT-024 feature.md — Gate Draft → Design Ready — Итерация 2

## Сводка

**Outcome: revise**

Документ значительно улучшен по сравнению с итерацией 1. Все блокеры первой итерации устранены. Найдено одно замечание уровня HIGH: неверный путь к шаблону в секции Change Surface.

---

## Результаты по пунктам

### A. Идентификаторы и структура

| Пункт | Статус | Комментарий |
|---|---|---|
| A-1 | OK | Plan IDs (`OQ-*`, `PRE-*`, `STEP-*`, `WS-*`, `AG-*`, `PAR-*`, `CP-*`, `ER-*`, `STOP-*`) в `feature.md` отсутствуют. |
| A-2 | OK | Все `REQ-01..07`, `NS-01..06`, `SC-01..07`, `CHK-01..02`, `EVID-01..02` присутствуют. |
| A-3 | OK | Каждый объявленный идентификатор (`MET-*`, `CON-*`, `ASM-*`, `CTR-*`, `FM-*`, `EC-*`, `NEG-*`) явно определён. |
| A-4 | OK | Фича содержит 7 `REQ-*`, 3 `ASM-*`, 4 `CTR-*`, 6 `FM-*`, 7 `SC-*`, 2 `CHK-*` — все условия short.md нарушены, выбор large.md обоснован. |
| A-5 | OK | `features/README.md`: `planned`; `feature.md`: `delivery_status: planned` — совпадает. |

### B. Traceability

| Пункт | Статус | Комментарий |
|---|---|---|
| B-1 | OK | Все `REQ-01..07` прослеживаются к ≥1 `SC-*` через traceability matrix. |
| B-2 | OK | Все `SC-*` покрыты `CHK-01`; `NEG-01..02` покрыты `CHK-02`. |
| B-3 | OK | `CHK-01` → `EVID-01`, `CHK-02` → `EVID-02`. |
| B-4 | OK | Критичные failure modes (`FM-01`..`FM-05`) покрыты `NEG-01..05`. `FM-03` и `FM-06` — серверные/UI fallback, не требуют отдельных `NEG-*`. |

### C. Непротиворечивость

| Пункт | Статус | Комментарий |
|---|---|---|
| C-1 | OK | `CON-03` расширяет ADR-005 (субдиректорий `avatars/`) без противоречий. |
| C-2 | OK | `feature.md` → `problem.md`, `ADR-005` → `FT-009/feature.md`. Нет цикла. |
| C-3 | OK | ADR-005: `decision_status: accepted`. В ADR Dependencies: «Canonical input — следовать без альтернатив». |
| C-4 | OK | `NS-06` ↔ `CON-03` согласованы. `NS-04` ↔ `CON-03` согласованы. |

### D. Соответствие архитектуре

| Пункт | Статус | Комментарий |
|---|---|---|
| D-1 | HIGH | **Неверный путь в Change Surface.** Цитата из feature.md: `templates/news/show.html`. Реальный файл в репозитории: `templates/news/article.html`. Изменение должно применяться к существующему файлу. Исправление: заменить `templates/news/show.html` на `templates/news/article.html`. |
| D-2 | OK | Новые шаблоны `templates/profile/page.html`, `templates/profile/modal.html` соответствуют правилу `templates/<domain>/name.html` из `frontend.md`. |
| D-3 | OK | «inline стили в шаблонах» — явное осознанное решение, согласованное с frontend.md («Локальные стили допустимы только в рамках конкретной страницы»). |
| D-4 | OK | Логика профиля (посты, комментарии, relative time) — в `internal/user/service.go`. Handler — только HTTP-парсинг и рендер. |

### E. Тестовая политика

| Пункт | Статус | Комментарий |
|---|---|---|
| E-1 | OK | Нет manual-only пометок для UI-изменений. Все UI-сценарии покрыты Playwright. |
| E-2 | OK | HTMX/Alpine.js взаимодействия (модалка, закрытие) — Playwright E2E в `CHK-01`. |
| E-3 | OK | `EVID-01` producer: CI/Playwright. `EVID-02` producer: Playwright. Соответствует методам проверки. |
| E-4 | OK | Manual-only гэпов нет. |

### F. Системные ограничения

| Пункт | Статус | Комментарий |
|---|---|---|
| F-1 | OK | `CON-01` фиксирует CSRF для `POST /profile/avatar` со ссылкой на PCON-02. `CTR-03` дублирует требование. |
| F-2 | OK | `CON-02` (проверка владельца), `NEG-03` (без авторизации → /login), `NEG-05` (чужой профиль → 403) — явные security-предикаты. |

### G. Glossary и Use Case

| Пункт | Статус | Комментарий |
|---|---|---|
| G-1 | OK | `Quick-view` и `Relative time` добавлены в `memory-bank/domain/glossary.md` с правильными определениями. |
| G-2 | OK | UC-* требуется к Done gate, не к Design Ready (feature-flow.md п.11). Для текущего gate — OK. |

---

## Замечания для исправления

### 1. HIGH — Неверный путь шаблона в Change Surface

**Цитата из feature.md (строка 81):**
```
| `templates/news/show.html` | code | Аватар + кликабельное имя в комментариях |
```

**Норма:** D-1 требует, чтобы пути в Change Surface реально существовали в репозитории. Реальный файл в репозитории: `templates/news/article.html` (подтверждено Glob по `templates/news/**`).

**Исправление:** Заменить `templates/news/show.html` на `templates/news/article.html` в таблице Change Surface.

---

## Итог

**Outcome: revise**

Одно замечание HIGH: неверный путь `templates/news/show.html` в Change Surface (реальный файл: `templates/news/article.html`). После исправления этого пути документ готов к переходу в Design Ready.
