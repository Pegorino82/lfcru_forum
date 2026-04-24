package tmpl

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

var funcMap = template.FuncMap{
	"truncate": func(s string, n int) string {
		r := []rune(s)
		if len(r) <= n {
			return s
		}
		return string(r[:n]) + "…"
	},
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
