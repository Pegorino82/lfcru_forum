package tmpl

import (
	"fmt"
	"hash/fnv"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var ruMonths = [12]string{
	"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря",
}

// avatarPalette — фиксированная палитра цветов для fallback-аватаров.
// Порядок не меняется: изменение ломает SC-06 (стабильность цвета).
var avatarPalette = [...]string{
	"#e53935", "#d81b60", "#8e24aa", "#5e35b1",
	"#3949ab", "#1e88e5", "#00897b", "#43a047",
	"#f4511e", "#fb8c00",
}

func pluralRu(n int, form1, form2, form5 string) string {
	n = n % 100
	if n >= 11 && n <= 19 {
		return form5
	}
	switch n % 10 {
	case 1:
		return form1
	case 2, 3, 4:
		return form2
	default:
		return form5
	}
}

var funcMap = template.FuncMap{
	// ruDate formats a time.Time as "02 апреля 2006" with Russian month name.
	// Use instead of .Format "02 января 2006" — month in Go format strings is a literal.
	"ruDate": func(t time.Time) string {
		return fmt.Sprintf("%02d %s %d", t.Day(), ruMonths[t.Month()-1], t.Year())
	},
	"truncate": func(s string, n int) string {
		r := []rune(s)
		if len(r) <= n {
			return s
		}
		return string(r[:n]) + "…"
	},
	"contains": strings.Contains,
	"deref": func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	},
	"derefInt64": func(i *int64) int64 {
		if i == nil {
			return 0
		}
		return *i
	},
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	// avatarInitials возвращает 1–2 буквы для fallback-аватара.
	// "John Doe" → "JD"; "john" → "JO"; "x" → "X".
	"avatarInitials": func(username string) string {
		words := strings.Fields(username)
		if len(words) >= 2 {
			r1 := []rune(words[0])
			r2 := []rune(words[1])
			if len(r1) > 0 && len(r2) > 0 {
				return strings.ToUpper(string(r1[0]) + string(r2[0]))
			}
		}
		r := []rune(username)
		if len(r) >= 2 {
			return strings.ToUpper(string(r[:2]))
		}
		if len(r) == 1 {
			return strings.ToUpper(string(r))
		}
		return "?"
	},
	// avatarColor детерминированно выбирает цвет из палитры по имени пользователя.
	"avatarColor": func(username string) string {
		h := fnv.New32a()
		h.Write([]byte(username))
		return avatarPalette[h.Sum32()%uint32(len(avatarPalette))]
	},
	// relativeTime возвращает относительное время на русском языке.
	"relativeTime": func(t time.Time) string {
		d := time.Since(t)
		switch {
		case d < time.Minute:
			return "только что"
		case d < time.Hour:
			mins := int(d.Minutes())
			return fmt.Sprintf("%d %s назад", mins, pluralRu(mins, "минуту", "минуты", "минут"))
		case d < 24*time.Hour:
			hours := int(d.Hours())
			return fmt.Sprintf("%d %s назад", hours, pluralRu(hours, "час", "часа", "часов"))
		case d < 48*time.Hour:
			return "вчера"
		default:
			days := int(d.Hours() / 24)
			return fmt.Sprintf("%d %s назад", days, pluralRu(days, "день", "дня", "дней"))
		}
	},
	// paginate returns a compact slice of page numbers for navigation.
	// -1 represents an ellipsis gap.
	"paginate": func(current, total int) []int {
		if total <= 1 {
			return []int{1}
		}
		set := make(map[int]bool)
		set[1] = true
		set[total] = true
		for p := current - 2; p <= current+2; p++ {
			if p >= 1 && p <= total {
				set[p] = true
			}
		}
		pages := make([]int, 0, len(set))
		prev := 0
		for p := 1; p <= total; p++ {
			if set[p] {
				if prev > 0 && p-prev > 1 {
					pages = append(pages, -1)
				}
				pages = append(pages, p)
				prev = p
			}
		}
		return pages
	},
}

// Renderer holds an isolated template set per page file.
type Renderer struct {
	sets map[string]*template.Template
}

// New builds a separate *template.Template for each page file.
// prefix is prepended to all keys (e.g. "templates/") so that handler names match.
// Layout files (those inside a "layouts" directory) are included in every set.
// Partial files (those inside a "partials" directory) are included in every page
// set AND get their own minimal set for use with RenderPartial.
func New(root fs.FS, prefix string) (*Renderer, error) {
	var layoutPaths []string
	var partialPaths []string
	var pagePaths []string

	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		dir := filepath.ToSlash(filepath.Dir(path))
		if dir == "layouts" || strings.HasSuffix(dir, "/layouts") {
			layoutPaths = append(layoutPaths, path)
		} else if dir == "partials" || strings.HasSuffix(dir, "/partials") {
			partialPaths = append(partialPaths, path)
		} else {
			pagePaths = append(pagePaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Pre-read layout contents.
	layoutContents := make(map[string]string, len(layoutPaths))
	for _, lp := range layoutPaths {
		data, err := fs.ReadFile(root, lp)
		if err != nil {
			return nil, err
		}
		layoutContents[prefix+lp] = string(data)
	}

	// Pre-read partial contents.
	partialContents := make(map[string]string, len(partialPaths))
	for _, pp := range partialPaths {
		data, err := fs.ReadFile(root, pp)
		if err != nil {
			return nil, err
		}
		partialContents[prefix+pp] = string(data)
	}

	sets := make(map[string]*template.Template, len(pagePaths)+len(partialPaths))

	// Build minimal sets for partials (no layout chain — just the partial itself).
	// These are used by RenderPartial in Go code.
	for pp, pc := range partialContents {
		t := template.New("").Funcs(funcMap)
		if _, err := t.New(pp).Parse(pc); err != nil {
			return nil, fmt.Errorf("parse partial %q: %w", pp, err)
		}
		sets[pp] = t
	}

	// Build page sets: layouts + partials (for {{template}} calls) + page itself.
	for _, p := range pagePaths {
		key := prefix + p
		t := template.New("").Funcs(funcMap)
		for lp, lc := range layoutContents {
			if _, err := t.New(lp).Parse(lc); err != nil {
				return nil, fmt.Errorf("parse layout %q: %w", lp, err)
			}
		}
		for pp, pc := range partialContents {
			if _, err := t.New(pp).Parse(pc); err != nil {
				return nil, fmt.Errorf("parse partial %q in set %q: %w", pp, key, err)
			}
		}
		content, err := fs.ReadFile(root, p)
		if err != nil {
			return nil, err
		}
		if _, err := t.New(key).Parse(string(content)); err != nil {
			return nil, fmt.Errorf("parse page %q: %w", p, err)
		}
		sets[key] = t
	}

	return &Renderer{sets: sets}, nil
}

// Render implements echo.Renderer.
// name может содержать "#blockName" суффикс для рендера именованного блока из page set:
// "templates/forum/topic.html#posts-list" → pageKey="templates/forum/topic.html", block="posts-list".
func (r *Renderer) Render(w io.Writer, name string, data any, _ echo.Context) error {
	pageKey, blockName, hasBlock := strings.Cut(name, "#")
	t, ok := r.sets[pageKey]
	if !ok {
		return fmt.Errorf("template %q not found", pageKey)
	}
	if hasBlock {
		return t.ExecuteTemplate(w, blockName, data)
	}
	return t.ExecuteTemplate(w, pageKey, data)
}

// RenderPartial рендерит именованный блок из template set страницы.
func (r *Renderer) RenderPartial(w io.Writer, pageKey, blockName string, data any) error {
	t, ok := r.sets[pageKey]
	if !ok {
		return fmt.Errorf("template %q not found", pageKey)
	}
	return t.ExecuteTemplate(w, blockName, data)
}
