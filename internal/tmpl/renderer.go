package tmpl

import (
	"html/template"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// Renderer wraps html/template for use with Echo.
type Renderer struct {
	templates *template.Template
}

// New parses all .html files found under the given root directory.
func New(root fs.FS) (*Renderer, error) {
	t := template.New("")

	err := fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		content, err := fs.ReadFile(root, path)
		if err != nil {
			return err
		}
		if _, err := t.New(path).Parse(string(content)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Renderer{templates: t}, nil
}

// Render implements echo.Renderer.
func (r *Renderer) Render(w io.Writer, name string, data any, _ echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}
