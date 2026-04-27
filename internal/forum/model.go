package forum

import "time"

type Section struct {
	ID          int64
	Title       string
	Description string
	SortOrder   int
	TopicCount  int
	CreatedAt   time.Time
}

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

type Post struct {
	ID                     int64
	TopicID                int64
	AuthorID               int64
	ParentID               *int64
	ParentAuthorSnapshot   *string
	ParentContentSnapshot  *string
	Content                string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type SectionView struct {
	ID          int64
	Title       string
	Description string
	TopicCount  int
}

type TopicView struct {
	ID             int64
	Title          string
	AuthorID       int64
	AuthorUsername string
	PostCount      int
	LastPostAt     *time.Time
	CreatedAt      time.Time
}

type PostView struct {
	ID                     int64
	TopicID                int64
	AuthorID               int64
	AuthorUsername         string
	ParentID               *int64
	ParentAuthorSnapshot   *string
	ParentContentSnapshot  *string
	Content                string
	CreatedAt              time.Time
}

// TopicWithLastAuthor — проекция для главной страницы.
type TopicWithLastAuthor struct {
	ID             int64
	Title          string
	LastPostAt     time.Time
	LastPostByName string
	SectionName    string
	PostCount      int
}
