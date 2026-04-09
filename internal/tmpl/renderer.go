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
}

// Renderer holds an isolated template set per page file.
type Renderer struct {
	sets map[string]*template.Template
}

// New builds a separate *template.Template for each page file.
// prefix is prepended to all keys (e.g. "templates/") so that handler names match.
// Layout files (those inside a "layouts" directory) are included in every set.
func New(root fs.FS, prefix string) (*Renderer, error) {
	var layoutPaths []string
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

	sets := make(map[string]*template.Template, len(pagePaths))
	for _, p := range pagePaths {
		key := prefix + p
		t := template.New("").Funcs(funcMap)
		for lp, lc := range layoutContents {
			if _, err := t.New(lp).Parse(lc); err != nil {
				return nil, fmt.Errorf("parse layout %q: %w", lp, err)
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
