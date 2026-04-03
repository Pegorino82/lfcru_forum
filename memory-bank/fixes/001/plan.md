# Plan: Fix/001 — реализация

## Файл: `internal/tmpl/renderer.go`

### Структура данных

```go
type Renderer struct {
    sets map[string]*template.Template // page-path → template set
}
```

### `New(root fs.FS)`

```
1. Walk root:
   - dir == "layouts" или HasSuffix(dir, "/layouts") → layoutPaths
   - иначе → pagePaths

2. Прочитать содержимое всех layout-файлов в map[path]string

3. Для каждого page-файла:
   t := template.New("")
   for lp, lc := range layoutContents { t.New(lp).Parse(lc) }
   content, _ := fs.ReadFile(root, p)
   t.New(p).Parse(string(content))
   sets[p] = t

4. Вернуть &Renderer{sets: sets}, nil
```

### `Render`

```go
func (r *Renderer) Render(w io.Writer, name string, data any, _ echo.Context) error {
    t, ok := r.sets[name]
    if !ok {
        return fmt.Errorf("template %q not found", name)
    }
    return t.ExecuteTemplate(w, name, data)
}
```

### Импорты

Добавить `"fmt"` и `"strings"`, убрать `"path/filepath"` если не используется.

## Проверка

```bash
docker compose -f docker-compose.dev.yml up --build
# Открыть http://localhost:8080/login  → 200, форма входа
# Открыть http://localhost:8080/register → 200, форма регистрации
docker exec <app> go test ./...
```
