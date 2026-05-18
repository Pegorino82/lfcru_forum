# Review: FT-028 Spec — Iteration 01

**Date:** 2026-05-18
**Reviewer:** evaluator agent
**Artifact:** `memory-bank/features/FT-028/feature.md` (## How, ## Verify)
**Loop:** Spec Improve Loop

---

## Outcome: `revise`

---

## Checks

### A. Solution

| # | Verdict | Details |
|---|---------|---------|
| A-1 | OK | Solution описывает конкретный технический подход: расширение `football.Client` методом `Squad()`, admin endpoint, bulk-создание тем, кеширование TTL 24h. Не повторяет REQ-*. |
| A-2 | MEDIUM | Trade-off упомянут неявно (HTML-фрагмент в теле поста vs отдельный UI-компонент через `DEC-01`), но в Solution нет явной формулировки "главный trade-off" или "рассмотренная альтернатива". |
| A-3 | OK | Business-логика генерации тем отнесена к `forum.Service` (Change Surface строка `internal/forum/service.go`). |

### B. Change Surface

| # | Verdict | Details |
|---|---------|---------|
| B-1 | HIGH | `internal/football/models.go (новый)` — файл помечен как новый, это допустимо. Однако `internal/football/client.go` существует — OK. `internal/forum/handler.go`, `internal/forum/service.go`, `internal/forum/repo.go` — все существуют — OK. `templates/forum/player-card.html (новый)` — помечен как новый, допустимо. `migrations/` — директория существует — OK. **Проблема:** `templates/forum/player-card.html` расположен по пути `templates/forum/` что корректно для `frontend.md` (`templates/<domain>/`). Однако в существующем каталоге `internal/football/` уже есть `venues.go`, а не `models.go`. Сам по себе новый файл допустим, но: admin handler указан как `internal/forum/handler.go или новый admin handler` — расплывчато, неизвестно какой именно файл будет затронут. |
| B-2 | OK | `templates/forum/player-card.html` соответствует конвенции `templates/<domain>/`. |
| B-3 | OK | Нет изменений в static/js/ или static/css/ — для данной фичи это допустимо (HTML-карточка в теле поста). |
| B-4 | HIGH | **Отсутствует `internal/config/config.go`** в Change Surface. REQ-01 и CTR-01 требуют API key для football-data.org. Согласно `architecture.md` § Configuration Ownership: "При добавлении новой переменной: сначала обновить `internal/config/config.go`". Если football-data.org API key уже сконфигурирован (клиент существует), это допустимо, но требует явного подтверждения. Также отсутствует `cmd/forum/main.go` — DI для нового admin endpoint потребует изменения маршрутов. |

### C. Contracts и Failure Modes

| # | Verdict | Details |
|---|---------|---------|
| C-1 | OK | CTR-01, CTR-02, CTR-03 корректно описывают API contract, internal contract и UI contract. |
| C-2 | BLOCKER | **XSS-вектор не покрыт.** CTR-03 указывает: "Карточка рендерится как safeHTML в теле поста". Данные из внешнего API (имя игрока, национальность) вставляются как HTML без упоминания sanitization. `safeHTML` в Go templates отключает escaping — если API вернёт `<script>` в поле name, это прямой XSS. Failure Modes не содержат FM для XSS/injection через данные API. |
| C-3 | OK | ADR Dependencies — "Нет" — корректно, используются существующие паттерны. |

### D. Traceability

| # | Verdict | Details |
|---|---------|---------|
| D-1 | OK | Каждый REQ-01..REQ-05 связан с минимум одним SC-* через traceability matrix. |
| D-2 | OK | SC-01, SC-02, SC-03 описывают наблюдаемые результаты. |
| D-3 | OK | Каждый SC-* связан с минимум одним CHK-*. |
| D-4 | OK | Каждый CHK-* связан с минимум одним EVID-*. |

### E. Checks и Evidence

| # | Verdict | Details |
|---|---------|---------|
| E-1 | MEDIUM | CHK-01: "Unit-тест `Squad()` + integration-тест генерации тем" — нет конкретной команды запуска. CHK-02: "Playwright: открыть список разделов..." — описана процедура, но нет команды. CHK-03: "Unit-тест `Squad()` с mock HTTP 500" — нет команды. Согласно spec-improve-loop exit criteria: "каждый CHK-* имеет команду или ручную процедуру (не 'проверить вручную' без инструкции)". Процедуры описаны, но без executable команд (docker run go test ..., npx playwright test ...). |
| E-2 | HIGH | EVID-01: "Go test pass" — path contract: "CI run log". EVID-03: "Go test pass" — path contract: "CI run log". Это не конкретный path contract. Согласно spec-improve-loop: "каждый EVID-* имеет конкретный path contract (не 'где-нибудь')". EVID-02 имеет `e2e/test-report/` — это конкретный path, OK. Но EVID-01 и EVID-03 указывают только "CI run log" без path. |
| E-3 | OK | UI-изменения (CHK-02) проверяются через Playwright, не помечены manual-only. |
| E-4 | OK | HTMX/Alpine.js взаимодействия не используются для обоснования manual-only. |
| E-5 | OK | Нет manual-only gaps. |

### F. Системные ограничения

| # | Verdict | Details |
|---|---------|---------|
| F-1 | MEDIUM | CSRF зафиксирован в CON-02, но admin endpoint (POST для генерации тем) не упоминается явно в FM-* как требующий CSRF-проверки. CON-02 лишь констатирует факт, но Flow (шаг 1: "Администратор запускает генерацию (admin endpoint или CLI)") допускает CLI-вариант, где CSRF не применим. Нет ясности, применяется ли CSRF к admin endpoint. |
| F-2 | HIGH | **Отсутствует NEG-*.** Feature зависит от внешнего API, имеет failure modes (FM-01, FM-02, FM-03). Согласно feature-flow.md gate DR: "если deliverable нельзя принять без negative/edge coverage -> >= 1 NEG-*". SC-03 описывает graceful degradation, но это acceptance scenario, а не negative test case с NEG-* идентификатором. |

---

## Summary of Findings

### BLOCKER (1)

1. **C-2: XSS-вектор через данные внешнего API не покрыт.**
   - **Цитата:** `CTR-03`: "Карточка рендерится как safeHTML в теле поста"
   - **Норма:** spec-improve-loop exit criteria: "покрыты критичные failure modes (auth, data loss, XSS)"; problem.md PCON-01 запрещает unsanitized input.
   - **Исправление:** Добавить FM-04 для XSS-вектора: данные из внешнего API должны проходить через Go template escaping (не `safeHTML`) или явную sanitization. Уточнить в CTR-03, что карточка рендерится через стандартный Go template escaping, а `safeHTML` применяется только к статической HTML-разметке карточки, не к данным игрока.

### HIGH (3)

2. **B-4: Неполная Change Surface — отсутствуют `cmd/forum/main.go` и возможно `internal/config/config.go`.**
   - **Цитата:** Change Surface не содержит `cmd/forum/main.go`.
   - **Норма:** architecture.md: "DI и инициализация в `cmd/forum/main.go`" — новый admin endpoint потребует регистрации маршрута.
   - **Исправление:** Добавить `cmd/forum/main.go` в Change Surface (регистрация admin route). Если API key для football-data.org уже сконфигурирован — явно указать это; если нет — добавить `internal/config/config.go`.

3. **E-2: EVID-01 и EVID-03 не имеют конкретного path contract.**
   - **Цитата:** EVID-01 path contract: "CI run log"; EVID-03 path contract: "CI run log".
   - **Норма:** spec-improve-loop exit criteria: "каждый EVID-* имеет конкретный path contract (не 'где-нибудь')".
   - **Исправление:** Указать конкретный path contract для Go test evidence, например `docker run ... go test ./internal/football/... ./internal/forum/... | tee` с выходным файлом, или ссылку на CI job artifact path.

4. **F-2: Отсутствует NEG-* для negative/edge coverage.**
   - **Цитата:** Нет ни одного `NEG-*` идентификатора в документе.
   - **Норма:** feature-flow.md gate DR: "если deliverable нельзя принять без negative/edge coverage -> >= 1 NEG-*"; фича зависит от внешнего API с failure modes.
   - **Исправление:** Добавить минимум один `NEG-*`, например: `NEG-01` API возвращает HTTP 500 -> генерация пропускается, раздел пустой, нет panic. Связать с CHK-03 и EVID-03 в traceability matrix.

### MEDIUM (3)

5. **A-2: Trade-off не сформулирован явно в Solution.**
   - **Цитата:** Solution не содержит слов "trade-off" или "альтернатива".
   - **Норма:** spec-improve-loop exit criteria: "Solution описывает конкретный технический подход и главный trade-off".
   - **Исправление:** Добавить в Solution одно предложение: например, "Trade-off: HTML-карточка встраивается в тело поста (простота, нет отдельного UI-компонента) за счёт невозможности обновить карточки при изменении данных API без перегенерации постов."

6. **E-1: CHK-* не содержат executable команд.**
   - **Цитата:** CHK-01: "Unit-тест `Squad()` + integration-тест генерации тем" — без команды.
   - **Норма:** spec-improve-loop: "каждый CHK-* имеет команду или ручную процедуру".
   - **Исправление:** Добавить команды запуска в колонку "How to check", например: `docker run ... go test ./internal/football/...` для CHK-01, `npx playwright test e2e/forum/team-section.spec.ts` для CHK-02.

7. **F-1: Применимость CSRF к admin endpoint неоднозначна.**
   - **Цитата:** CON-02: "CSRF-токен обязателен для POST-запросов создания тем"; Flow шаг 1: "admin endpoint или CLI".
   - **Норма:** problem.md PCON-02: "CSRF-токен обязателен для всех POST/PUT/DELETE".
   - **Исправление:** Уточнить в Flow или Contracts: если admin endpoint — HTTP POST, CSRF обязателен через middleware. Если CLI — CSRF не применим. Зафиксировать выбранный вариант.

### B-1 Additional Note

8. **B-1: Change Surface entry для admin handler расплывчатый.**
   - **Цитата:** "`internal/forum/handler.go` или новый admin handler"
   - **Норма:** Change Surface должен содержать конкретные пути.
   - **Исправление:** Определиться: handler добавляется в `internal/forum/handler.go` или создаётся отдельный файл (например `internal/forum/admin_handler.go`). Указать конкретный путь.
