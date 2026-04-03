package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	UserID    int64
	IPAddr    string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
}
