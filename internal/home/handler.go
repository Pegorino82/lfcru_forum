package home

import (
	"net/http"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/labstack/echo/v4"
)

func ShowHome(c echo.Context) error {
	data := map[string]any{
		"User":      auth.UserFromContext(c),
		"CSRFToken": appMiddleware.CSRFToken(c),
	}
	if c.Request().Header.Get("HX-Request") == "true" {
		r := c.Echo().Renderer.(*tmpl.Renderer)
		return r.RenderPartial(c.Response(), "templates/home/index.html", "content", data)
	}
	return c.Render(http.StatusOK, "templates/home/index.html", data)
}
