package admin

import (
	"net/http"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/labstack/echo/v4"
)

// RequireAdminOrMod checks that the authenticated user has admin or moderator role.
// Guests and inactive users → redirect /login. Insufficient role → 403.
func RequireAdminOrMod(renderer echo.Renderer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u := auth.UserFromContext(c)
			if u == nil || !u.IsActive {
				return c.Redirect(http.StatusFound, "/login")
			}
			if u.Role != "admin" && u.Role != "moderator" {
				return c.Render(http.StatusForbidden, "templates/errors/403.html", map[string]interface{}{})
			}
			return next(c)
		}
	}
}
