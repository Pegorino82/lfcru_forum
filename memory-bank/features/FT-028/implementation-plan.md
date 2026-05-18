---
title: "FT-028: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-028. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical feature-фактов."
derived_from:
  - feature.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_028_scope
  - ft_028_architecture
  - ft_028_acceptance_criteria
  - ft_028_blocker_state
---

# План имплементации

## Цель текущего плана

Реализовать форумный раздел «Команда» с автоматической генерацией тем по игрокам из football-data.org API. После выполнения: раздел виден на форуме, каждый игрок имеет тему с карточкой, повторный запуск idempotent, graceful degradation при недоступном API.

## Discovery Context / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `internal/football/client.go` | API-клиент football-data.org (NextMatch, LastMatch, Standings) | Добавляем метод Squad() — должен следовать тому же паттерну кеширования и graceful degradation | Паттерн: mutex lock → check TTL → fetch or return cache → nil on error |
| `internal/football/client_test.go` | Unit-тесты клиента с httptest.NewServer | Тесты Squad() должны следовать тому же паттерну | Mock HTTP server, newTestClient helper |
| `internal/forum/repo.go` | CreateSection (L143-155), CreateTopic (L158-175), CreatePost (L178-283) | Используем существующие методы для создания раздела и тем | CreateSection → id; CreateTopic → id (FK validation); CreatePost → id |
| `internal/forum/service.go` | Валидация (trim, length checks), CreateTopic (L128-144) | Добавляем метод GenerateTeamTopics — использует существующие CreateTopic/CreatePost | Validation: trim → check empty → check length → repo call |
| `internal/admin/forum_handler.go` | ForumHandler с interface ForumSvc (L29-37) | Расширяем ForumSvc interface методом GenerateTeamTopics; добавляем admin endpoint | Interface-based DI; error → slog.Error + render |
| `cmd/forum/main.go` | Регистрация routes (L142-174), footballClient init (L112) | Нужно пробросить footballClient в admin ForumHandler и зарегистрировать новый route | adminGroup.POST pattern; NewForumHandler injection |
| `templates/forum/partials/post.html` | Рендеринг поста — plain text Content, no safeHTML | Карточка игрока как HTML в Content требует template.HTML для статической разметки | Content рендерится как `{{.Content}}` — plain text; для HTML нужен отдельный подход |
| `internal/forum/model.go` | PostView (L54-65) — Content string | Content — plain text; для player card нужно хранить HTML и рендерить через safeHTML | Два варианта: хранить HTML в content + safeHTML рендер, или специальное поле |
| `e2e/forum/reply.spec.ts` | E2E тест форума — login, navigate, assert | Паттерн для team-section.spec.ts | page.goto, fill, click, expect assertions |
| `migrations/005_create_forum_and_matches.sql` | Схема forum_sections, forum_topics, forum_posts | Раздел создаётся через INSERT в forum_sections | No schema changes needed — existing tables sufficient |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites | Required CI suites | Manual-only gap | Approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `football.Client.Squad()` | `REQ-01`, `SC-01`, `SC-03`, `NEG-01`, `CHK-01`, `CHK-03` | Нет | Unit: happy path, API error (500), empty squad, empty player name; mock HTTP server | `go test ./internal/football/...` | Go Tests job | none | — |
| `forum.Service.GenerateTeamTopics()` | `REQ-03`, `REQ-04`, `SC-01`, `SC-02`, `CHK-01` | Нет | Unit: mock repo + mock footballClient → topics created, idempotent on re-run | `go test ./internal/forum/...` | Go Tests job | none | — |
| Admin endpoint POST | `REQ-04`, `CON-02`, `SC-01` | Нет | Integration: full HTTP test with CSRF, auth, mock API | `go test -tags integration -p 1 ./internal/admin/...` | Go Tests job | none | — |
| UI: раздел «Команда» + карточка игрока | `REQ-02`, `REQ-03`, `SC-01`, `SC-02`, `NEG-02`, `CHK-02` | Нет | Playwright: section visible, topic with card, idempotent, XSS escape | `npx playwright test e2e/forum/team-section.spec.ts` | E2E job | none | — |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Как хранить HTML-карточку в content поста? Стандартный Content — plain text. | Текущий post.Content рендерится через Go template escaping; HTML будет экранирован | `STEP-03` | **Default:** рендерить карточку на стороне сервера в Go-шаблоне, а в content поста хранить structured text (имя, позиция, и т.д.); шаблон форума определяет, является ли пост "player-card" и рендерит по-другому. **Alt:** хранить HTML в content и маркировать пост флагом `is_html`. Выбираем default — без schema changes. |
| `OQ-02` | Фильтровать youth/reserve (~43 записи) или показывать всех? | DEC-03 из feature.md — отложено на план | `STEP-02` | **Default:** показывать всех, кого возвращает API. Фильтрация по position — в будущем, если потребуется. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | `docker compose -f docker-compose.dev.yml up -d` — postgres + app запущены | All STEP | DB unavailable, app not responding |
| test (Go) | `docker run --rm -v $(pwd):/app -w /app -v lfcru_gomod:/root/go/pkg/mod --network lfcru_forum_default golang:1.23-alpine go test ./...` | `CHK-01`, `CHK-03` | Tests fail with "no such host" or module download timeout |
| test (E2E) | `docker compose -f docker-compose.e2e.yml up -d` + `npx playwright test` | `CHK-02` | Playwright can't reach app-e2e:8081 |
| API | `FOOTBALL_DATA_API_KEY` env var set and valid | `STEP-02` | Squad() returns nil (graceful degradation) |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-01` | football-data.org API доступен, `/v4/teams/64` возвращает squad | `STEP-02` | no (graceful degradation) |
| `PRE-02` | `CON-01` | API key set in `.env.local` | `STEP-02` | no (Squad returns nil) |
| `PRE-03` | — | Worktree `../lfcru_forum-FT-028` создан, PR #19 открыт | All | yes (already done) |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` API Client | `REQ-01`, `CTR-01`, `CTR-02` | `Squad()` метод + unit-тесты | agent | — |
| `WS-2` Forum Logic | `REQ-02`, `REQ-03`, `REQ-04`, `REQ-05`, `CTR-03` | Service method + repo queries + player card template | agent | `WS-1` |
| `WS-3` Admin Endpoint | `REQ-04`, `CON-02` | POST route + handler + wiring in main.go | agent | `WS-2` |
| `WS-4` Tests | `CHK-01`, `CHK-02`, `CHK-03` | Unit + integration + E2E tests | agent | `WS-3` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | PR ready for review | Closure | Человеческая приёмка перед merge | User / PR review |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `CTR-01`, `CTR-02` | Добавить структуры Player, SquadResponse и метод Squad() в football.Client с кешированием 24h | `internal/football/client.go`, `internal/football/client_test.go` | Метод Squad() возвращает []Player; unit-тесты (happy, error, empty, null fields) | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` | `go test ./internal/football/...` | — | none | API response format изменился |
| `STEP-02` | agent | `REQ-02`, `REQ-03`, `REQ-04`, `REQ-05` | Добавить метод GenerateTeamTopics в forum.Service: создать раздел «Команда» + темы с карточками; idempotent | `internal/forum/service.go`, `internal/forum/repo.go`, `templates/forum/player-card.html` | Метод создаёт раздел и темы; карточка рендерится в шаблоне; unit-тесты | `CHK-01` | `EVID-01` | `go test ./internal/forum/...` | `STEP-01` | none | OQ-01 требует schema change |
| `STEP-03` | agent | `CON-02`, `REQ-04` | Добавить admin endpoint POST /admin/forum/generate-team и зарегистрировать в main.go | `internal/admin/forum_handler.go`, `cmd/forum/main.go` | Admin может запустить генерацию; CSRF + auth protection | — | — | Manual: curl или браузер | `STEP-02` | none | — |
| `STEP-04` | agent | `CHK-01`, `CHK-03` | Добавить/дополнить Go-тесты: unit для Squad(), unit для GenerateTeamTopics (idempotent, graceful degradation) | `internal/football/client_test.go`, `internal/forum/service_test.go` | Тесты зелёные локально | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` | Docker go test command | `STEP-03` | none | — |
| `STEP-05` | agent | `CHK-02`, `NEG-02` | Добавить Playwright E2E: раздел виден, карточка игрока, idempotent, XSS escape | `e2e/forum/team-section.spec.ts` | E2E тесты зелёные | `CHK-02` | `EVID-02` | `npx playwright test e2e/forum/team-section.spec.ts` | `STEP-04` | none | — |

## Parallelizable Work

- `PAR-01` STEP-01 (API client) может выполняться параллельно с чтением существующих forum patterns (research only).
- `PAR-02` STEP-04 и STEP-05 (Go tests и E2E) могут выполняться параллельно после STEP-03.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02` | Squad() работает + GenerateTeamTopics создаёт раздел и темы; unit-тесты зелёные | `EVID-01` |
| `CP-02` | `STEP-03`, `STEP-04`, `STEP-05` | Admin endpoint работает + все тесты (Go + E2E) зелёные | `EVID-01`, `EVID-02`, `EVID-03` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | football-data.org API format changed | Squad() возвращает пустые данные | Тест с реальным API response в fixture; graceful degradation | Unit-тесты красные при обновлении fixture |
| `ER-02` | Хранение HTML в post content конфликтует с existing escaping | Карточка отображается как plain text | OQ-01 resolved: рендерить в шаблоне, не хранить HTML в content | Карточка экранирована в UI |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `CON-01` | API rate limit exceeded during development | Подождать 1 min; использовать cached fixture | Тесты работают с mock; реальный API — только для smoke |
| `STOP-02` | `OQ-01` | Player card requires schema migration | Escalate к человеку | Раздел создан без карточек; plain text fallback |

## Готово для приемки

1. Все `STEP-*` выполнены, `CP-01` и `CP-02` пройдены.
2. Go-тесты зелёные локально (unit + integration).
3. Playwright E2E зелёные.
4. CI jobs (Lint, Go Tests, E2E) зелёные после push.
5. PR переведён из draft в ready for review.
