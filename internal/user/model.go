package user

import "time"

type User struct {
	ID        int64
	Username  string
	Email     string
	PassHash  []byte
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
