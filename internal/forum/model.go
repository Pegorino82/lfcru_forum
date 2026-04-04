package forum

import "time"

type Topic struct {
	ID         int64
	SectionID  int64
	Title      string
	AuthorID   int64
	PostCount  int
	LastPostAt *time.Time
	LastPostBy *int64
	CreatedAt  time.Time
}

// TopicWithLastAuthor — проекция для главной страницы.
type TopicWithLastAuthor struct {
	ID             int64
	Title          string
	LastPostAt     time.Time
	LastPostByName string
}
