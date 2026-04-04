package news

import "time"

type News struct {
	ID          int64
	Title       string
	Content     string
	IsPublished bool
	AuthorID    int64
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
