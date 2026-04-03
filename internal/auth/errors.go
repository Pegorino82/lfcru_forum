package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRateLimited        = errors.New("rate limited")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrDuplicateUsername  = errors.New("duplicate username")
	ErrSessionNotFound    = errors.New("session not found")
)

// ValidationErrors maps field names to error messages.
type ValidationErrors map[string]string

func (e ValidationErrors) Error() string {
	return "validation failed"
}
