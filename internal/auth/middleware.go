package auth

import (
	"net/http"
	"net/url"

	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const userContextKey = "user"

// LoadSession reads the session cookie and, if valid, stores the user in the Echo context.
// It never returns an error — missing/invalid sessions result in a nil user (guest).
func LoadSession(svc *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				c.Set(userContextKey, (*user.User)(nil))
				return next(c)
			}

			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				clearSessionCookie(c)
				c.Set(userContextKey, (*user.User)(nil))
				return next(c)
			}

			u, _, err := svc.GetSession(c.Request().Context(), sessionID)
			if err != nil {
				clearSessionCookie(c)
				c.Set(userContextKey, (*user.User)(nil))
				return next(c)
			}

			c.Set(userContextKey, u)
			return next(c)
		}
	}
}

// UserFromContext returns the authenticated user from context, or nil for guests.
func UserFromContext(c echo.Context) *user.User {
	v := c.Get(userContextKey)
	if v == nil {
		return nil
	}
	u, _ := v.(*user.User)
	return u
}

// RequireAuth redirects unauthenticated users to /login?next=<current_path>.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if UserFromContext(c) == nil {
			next_ := c.Request().URL.RequestURI()
			return c.Redirect(http.StatusFound, "/login?next="+url.QueryEscape(next_))
		}
		return next(c)
	}
}

func clearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
