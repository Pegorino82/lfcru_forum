package news

import "time"

// ArticleStatus represents the publication status of a news article.
type ArticleStatus string

const (
	StatusDraft     ArticleStatus = "draft"
	StatusInReview  ArticleStatus = "in_review"
	StatusPublished ArticleStatus = "published"
)

type News struct {
	ID          int64
	Title       string
	Content     string
	Status      ArticleStatus
	ReviewerID  *int64
	AuthorID    int64
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ImageView is a read-only projection of article_images for public display.
type ImageView struct {
	ID       int64
	Filename string
}
