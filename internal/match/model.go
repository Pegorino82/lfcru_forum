package match

import "time"

type Match struct {
	ID         int64
	Opponent   string
	MatchDate  time.Time
	Tournament string
	IsHome     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
