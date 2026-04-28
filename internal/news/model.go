package news

import (
	"regexp"
	"strings"
	"time"
)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

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

// HomeNewsItem is a read-only projection used for the home page news feed.
type HomeNewsItem struct {
	ID            int64
	Title         string
	Content       string
	CoverImageURL string
	CommentCount  int
	PublishedAt   *time.Time
}

// ExcerptText strips HTML tags and returns up to 200 characters of plain text.
func (n HomeNewsItem) ExcerptText() string {
	plain := strings.TrimSpace(htmlTagRe.ReplaceAllString(n.Content, ""))
	runes := []rune(plain)
	if len(runes) <= 200 {
		return plain
	}
	return string(runes[:200]) + "…"
}
