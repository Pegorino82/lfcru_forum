package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CSRFMiddleware returns Echo's built-in CSRF middleware configured for this app.
// Token is stored in a cookie (_csrf) and checked in form field "_csrf" for POST/PUT/DELETE.
func CSRFMiddleware() echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:_csrf,header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookieHTTPOnly: false,
		CookieSameSite: 3, // http.SameSiteStrictMode
		CookiePath:     "/",
		ContextKey:     "csrf",
	})
}

// CSRFToken extracts the CSRF token from the Echo context (set by CSRFMiddleware).
func CSRFToken(c echo.Context) string {
	v := c.Get("csrf")
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
