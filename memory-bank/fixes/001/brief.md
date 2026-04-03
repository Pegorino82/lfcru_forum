# Brief: Fix/001 — конфликт `{{define "content"}}` в шаблонах

## Симптом

`GET /login` → 500:
```
template: auth/register.html:24:31: executing "content" at <.Fields>:
can't evaluate field Fields in type auth.loginData
```

## Причина

`internal/tmpl/renderer.go` парсит все `.html` файлы в **одно** `*template.Template`.
`login.html` и `register.html` оба определяют `{{define "content"}}` — последний затирает первый.
При рендеринге `/login` `base.html` вызывает блок `content` из `register.html`, который требует `.Fields`, отсутствующего в `loginData`.

## Объём

Один файл: `internal/tmpl/renderer.go`
