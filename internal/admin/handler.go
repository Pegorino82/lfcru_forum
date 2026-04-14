package admin

import (
	"net/http"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

// DashboardData is the template data for the admin dashboard.
type DashboardData struct {
	User      *user.User
	CSRFToken string
}

// Handler handles admin HTTP requests.
type Handler struct{}

// NewHandler creates a new admin Handler.
func NewHandler() *Handler { return &Handler{} }

// Dashboard renders the admin dashboard stub.
func (h *Handler) Dashboard(c echo.Context) error {
	data := DashboardData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
	}
	return c.Render(http.StatusOK, "templates/admin/dashboard.html", data)
}
