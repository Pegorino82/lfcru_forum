package profile

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

const maxAvatarBytes = 5 << 20 // 5 MB

// PageData — данные для страницы профиля.
type PageData struct {
	User      *user.User // аутентифицированный пользователь (nil для гостей)
	CSRFToken string
	Profile   *ProfileData
	IsOwner   bool // текущий пользователь — владелец профиля
}

// ModalData — данные для модального окна профиля.
type ModalData struct {
	User    *user.User
	Profile *ProfileData
	IsOwner bool
}

type Handler struct {
	svc        *Service
	avatarsDir string
}

func NewHandler(svc *Service, avatarsDir string) *Handler {
	return &Handler{svc: svc, avatarsDir: avatarsDir}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/profile/:username", h.ShowPage)
	e.GET("/profile/:username/modal", h.ShowModal)

	authGroup := e.Group("", auth.RequireAuth)
	authGroup.POST("/profile/:username/avatar", h.UploadAvatar)
}

// ShowPage рендерит страницу профиля.
func (h *Handler) ShowPage(c echo.Context) error {
	username := c.Param("username")
	profile, err := h.svc.GetProfile(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.String(http.StatusNotFound, "Пользователь не найден")
		}
		slog.Error("profile: show page", "username", username, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка сервера")
	}

	currentUser := auth.UserFromContext(c)
	isOwner := currentUser != nil && currentUser.ID == profile.User.ID

	return c.Render(http.StatusOK, "templates/profile/page.html", PageData{
		User:      currentUser,
		CSRFToken: appMiddleware.CSRFToken(c),
		Profile:   profile,
		IsOwner:   isOwner,
	})
}

// ShowModal возвращает HTML модального окна для HTMX.
func (h *Handler) ShowModal(c echo.Context) error {
	username := c.Param("username")
	profile, err := h.svc.GetProfile(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.String(http.StatusNotFound, "Пользователь не найден")
		}
		slog.Error("profile: show modal", "username", username, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка сервера")
	}

	currentUser := auth.UserFromContext(c)
	isOwner := currentUser != nil && currentUser.ID == profile.User.ID

	return c.Render(http.StatusOK, "templates/profile/modal.html", ModalData{
		User:    currentUser,
		Profile: profile,
		IsOwner: isOwner,
	})
}

// UploadAvatar обрабатывает загрузку аватара. Требует авторизации.
func (h *Handler) UploadAvatar(c echo.Context) error {
	username := c.Param("username")
	currentUser := auth.UserFromContext(c)

	// Найти целевого пользователя
	targetProfile, err := h.svc.GetProfile(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.String(http.StatusNotFound, "Пользователь не найден")
		}
		slog.Error("profile: upload avatar get profile", "username", username, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка сервера")
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return c.String(http.StatusBadRequest, "Файл не найден в запросе")
	}

	if fileHeader.Size > maxAvatarBytes {
		return c.String(http.StatusRequestEntityTooLarge, "Файл слишком большой (максимум 5 МБ)")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Не удалось открыть файл")
	}
	defer src.Close()

	_, err = h.svc.SaveAvatar(
		c.Request().Context(),
		currentUser.ID,
		targetProfile.User.ID,
		src,
		h.avatarsDir,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			return c.String(http.StatusForbidden, "Нет доступа")
		case errors.Is(err, ErrUnsupportedFormat):
			return c.String(http.StatusUnprocessableEntity, "Неподдерживаемый формат. Допустимы: JPEG, PNG, WebP")
		default:
			slog.Error("profile: save avatar", "username", username, "err", err)
			return c.String(http.StatusInternalServerError, "Ошибка сохранения файла")
		}
	}

	// Перечитываем профиль чтобы получить свежие данные (AvatarURL обновлён)
	freshProfile, err := h.svc.GetProfile(c.Request().Context(), username)
	if err != nil {
		slog.Error("profile: reload after upload", "username", username, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка обновления профиля")
	}

	// HTMX: возвращаем обновлённый avatar-block
	return c.Render(http.StatusOK, "templates/profile/page.html#avatar-block", PageData{
		User:      currentUser,
		CSRFToken: appMiddleware.CSRFToken(c),
		Profile:   freshProfile,
		IsOwner:   true,
	})
}
