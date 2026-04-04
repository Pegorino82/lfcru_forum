package auth

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts auth routes on the Echo instance.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/register", h.ShowRegister)
	e.POST("/register", h.Register)
	e.GET("/login", h.ShowLogin)
	e.POST("/login", h.Login)
	e.POST("/logout", h.Logout)
}

// --- template data structs ---

type registerData struct {
	CSRFToken string
	Errors    ValidationErrors
	Fields    map[string]string
	User      interface{}
}

type loginData struct {
	CSRFToken string
	Error     string
	Email     string
	User      interface{}
}

// --- handlers ---

func (h *Handler) ShowRegister(c echo.Context) error {
	if UserFromContext(c) != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	data := registerData{CSRFToken: appMiddleware.CSRFToken(c), Fields: map[string]string{}}
	if c.Request().Header.Get("HX-Request") == "true" {
		r := c.Echo().Renderer.(*tmpl.Renderer)
		return r.RenderPartial(c.Response(), "templates/auth/register.html", "content", data)
	}
	return c.Render(http.StatusOK, "templates/auth/register.html", data)
}

func (h *Handler) Register(c echo.Context) error {
	if UserFromContext(c) != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	in := RegisterInput{
		Username:        c.FormValue("username"),
		Email:           c.FormValue("email"),
		Password:        c.FormValue("password"),
		PasswordConfirm: c.FormValue("password_confirm"),
		IPAddr:          realIP(c),
		UserAgent:       c.Request().UserAgent(),
	}

	_, sess, err := h.svc.Register(c.Request().Context(), in)
	if err != nil {
		var verr ValidationErrors
		switch {
		case errors.As(err, &verr):
			return c.Render(http.StatusUnprocessableEntity,
				"templates/auth/register.html", registerData{
					CSRFToken: appMiddleware.CSRFToken(c),
					Errors:    verr,
					Fields: map[string]string{
						"username": in.Username,
						"email":    in.Email,
					},
				})
		case errors.Is(err, ErrDuplicateEmail):
			return c.Render(http.StatusConflict,
				"templates/auth/register.html", registerData{
					CSRFToken: appMiddleware.CSRFToken(c),
					Errors:    ValidationErrors{"email": "Пользователь с таким email уже зарегистрирован"},
					Fields:    map[string]string{"username": in.Username, "email": in.Email},
				})
		case errors.Is(err, ErrDuplicateUsername):
			return c.Render(http.StatusConflict,
				"templates/auth/register.html", registerData{
					CSRFToken: appMiddleware.CSRFToken(c),
					Errors:    ValidationErrors{"username": "Это имя уже занято"},
					Fields:    map[string]string{"username": in.Username, "email": in.Email},
				})
		case errors.Is(err, ErrRateLimited):
			return c.String(http.StatusTooManyRequests, "Слишком много попыток. Попробуйте позже.")
		default:
			return err
		}
	}

	setSessionCookie(c, sess.ID.String(), h.svc.cfg.CookieSecure)
	setFlash(c, "Регистрация прошла успешно!")
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) ShowLogin(c echo.Context) error {
	if UserFromContext(c) != nil {
		return c.Redirect(http.StatusFound, "/")
	}
	data := loginData{CSRFToken: appMiddleware.CSRFToken(c)}
	if c.Request().Header.Get("HX-Request") == "true" {
		r := c.Echo().Renderer.(*tmpl.Renderer)
		return r.RenderPartial(c.Response(), "templates/auth/login.html", "content", data)
	}
	return c.Render(http.StatusOK, "templates/auth/login.html", data)
}

func (h *Handler) Login(c echo.Context) error {
	if UserFromContext(c) != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	email := c.FormValue("email")
	in := LoginInput{
		Email:     email,
		Password:  c.FormValue("password"),
		IPAddr:    realIP(c),
		UserAgent: c.Request().UserAgent(),
	}

	_, sess, err := h.svc.Login(c.Request().Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, ErrRateLimited):
			return c.String(http.StatusTooManyRequests, "Слишком много попыток. Попробуйте позже.")
		case errors.Is(err, ErrInvalidCredentials):
			return c.Render(http.StatusUnprocessableEntity,
				"templates/auth/login.html", loginData{
					CSRFToken: appMiddleware.CSRFToken(c),
					Error:     "Неверный email или пароль",
					Email:     email,
				})
		default:
			return err
		}
	}

	setSessionCookie(c, sess.ID.String(), h.svc.cfg.CookieSecure)
	setFlash(c, "Вы вошли в систему!")
	return c.Redirect(http.StatusSeeOther, safeRedirect(c.QueryParam("next")))
}

func (h *Handler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		if id, err2 := uuid.Parse(cookie.Value); err2 == nil {
			_ = h.svc.Logout(c.Request().Context(), id)
		}
	}
	clearSessionCookie(c)
	return c.Redirect(http.StatusSeeOther, "/")
}

// --- helpers ---

func safeRedirect(next string) string {
	if next == "" {
		return "/"
	}
	u, err := url.Parse(next)
	if err != nil || u.Host != "" || u.Scheme != "" {
		return "/"
	}
	p := u.Path
	if p == "/login" || p == "/register" || p == "/logout" {
		return "/"
	}
	return next
}

func setSessionCookie(c echo.Context, sessionID string, secure bool) {
	c.SetCookie(&http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   2592000, // 30 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func setFlash(c echo.Context, msg string) {
	c.SetCookie(&http.Cookie{
		Name:     "flash",
		Value:    url.QueryEscape(msg),
		Path:     "/",
		MaxAge:   60,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func realIP(c echo.Context) string {
	if ip := c.Request().Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.Request().Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return c.RealIP()
}
