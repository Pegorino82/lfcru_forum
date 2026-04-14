package admin

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

// UserSvc is the subset of user.Service the admin users handler needs.
type UserSvc interface {
	ListAll(ctx context.Context) ([]user.User, error)
	BanUser(ctx context.Context, id, requestorID int64) error
	UnbanUser(ctx context.Context, id int64) error
}

// UsersHandler handles admin user management routes.
type UsersHandler struct {
	svc UserSvc
}

// NewUsersHandler creates a new UsersHandler.
func NewUsersHandler(svc UserSvc) *UsersHandler {
	return &UsersHandler{svc: svc}
}

type usersListData struct {
	User          *user.User
	CSRFToken     string
	Users         []user.User
	CurrentUserID int64
}

// List handles GET /admin/users.
func (h *UsersHandler) List(c echo.Context) error {
	users, err := h.svc.ListAll(c.Request().Context())
	if err != nil {
		slog.Error("admin: list users", "error", err)
		return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
	}
	cu := auth.UserFromContext(c)
	return c.Render(http.StatusOK, "templates/admin/users/list.html", usersListData{
		User:          cu,
		CSRFToken:     appMiddleware.CSRFToken(c),
		Users:         users,
		CurrentUserID: cu.ID,
	})
}

// Ban handles POST /admin/users/:id/ban.
func (h *UsersHandler) Ban(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	cu := auth.UserFromContext(c)
	if err := h.svc.BanUser(c.Request().Context(), id, cu.ID); err != nil {
		switch {
		case errors.Is(err, user.ErrCannotBanSelf):
			return c.String(http.StatusBadRequest, "Нельзя заблокировать самого себя")
		case errors.Is(err, user.ErrNotFound):
			return c.String(http.StatusNotFound, "Пользователь не найден")
		default:
			slog.Error("admin: ban user", "error", err)
			return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
		}
	}
	return c.Redirect(http.StatusSeeOther, "/admin/users")
}

// Unban handles POST /admin/users/:id/unban.
func (h *UsersHandler) Unban(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Неверный ID")
	}
	if err := h.svc.UnbanUser(c.Request().Context(), id); err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return c.String(http.StatusNotFound, "Пользователь не найден")
		default:
			slog.Error("admin: unban user", "error", err)
			return c.String(http.StatusInternalServerError, "Внутренняя ошибка сервера")
		}
	}
	return c.Redirect(http.StatusSeeOther, "/admin/users")
}
