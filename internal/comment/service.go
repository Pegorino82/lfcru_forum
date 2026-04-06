package comment

import (
	"context"
	"html"
	"html/template"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Pegorino82/lfcru_forum/internal/user"
)

// mentionRe matches @username at the start of a string or after whitespace.
// Does NOT match email addresses (e.g. user@example.com is not a mention).
var mentionRe = regexp.MustCompile(`(^|\s)@([a-zA-Z0-9_-]{3,30})`)

// UserRepo is the subset of user.Repo used by the comment service.
type UserRepo interface {
	GetByUsernames(ctx context.Context, usernames []string) ([]user.User, error)
}

// repoCreator allows mocking the repo in unit tests.
type repoCreator interface {
	Create(ctx context.Context, c *Comment) (int64, error)
}

// Service provides business logic for comments.
type Service struct {
	repo     repoCreator
	userRepo UserRepo
}

// NewService creates a new comment Service.
func NewService(repo *Repo, userRepo UserRepo) *Service {
	return &Service{repo: repo, userRepo: userRepo}
}

// Create trims, validates, and persists a comment.
func (s *Service) Create(ctx context.Context, c *Comment) (int64, error) {
	c.Content = strings.TrimSpace(c.Content)
	if c.Content == "" {
		return 0, ErrEmptyContent
	}
	if utf8.RuneCountInString(c.Content) > 10000 {
		return 0, ErrContentTooLong
	}
	return s.repo.Create(ctx, c)
}

// RenderMentions converts @username mentions to HTML spans for known users.
// Unknown mentions are left as plain text. All content is HTML-escaped.
func (s *Service) RenderMentions(ctx context.Context, content string) (template.HTML, error) {
	// 1. Collect unique mention usernames.
	seen := make(map[string]bool)
	var usernames []string
	for _, m := range mentionRe.FindAllStringSubmatch(content, -1) {
		lower := strings.ToLower(m[2])
		if !seen[lower] {
			seen[lower] = true
			usernames = append(usernames, m[2])
		}
	}

	// 2. Batch-lookup existing users.
	existing := make(map[string]bool)
	if len(usernames) > 0 {
		users, err := s.userRepo.GetByUsernames(ctx, usernames)
		if err != nil {
			return "", err
		}
		for _, u := range users {
			existing[strings.ToLower(u.Username)] = true
		}
	}

	// 3. Build safe HTML output segment by segment.
	locs := mentionRe.FindAllStringSubmatchIndex(content, -1)
	var buf strings.Builder
	prev := 0
	for _, loc := range locs {
		// loc[0..1] = full match, loc[2..3] = prefix (space/""), loc[4..5] = username
		buf.WriteString(html.EscapeString(content[prev:loc[0]]))
		buf.WriteString(html.EscapeString(content[loc[2]:loc[3]])) // space prefix or ""
		username := content[loc[4]:loc[5]]
		if existing[strings.ToLower(username)] {
			buf.WriteString(`<span class="mention">@`)
			buf.WriteString(html.EscapeString(username))
			buf.WriteString(`</span>`)
		} else {
			buf.WriteString("@")
			buf.WriteString(html.EscapeString(username))
		}
		prev = loc[1]
	}
	buf.WriteString(html.EscapeString(content[prev:]))
	return template.HTML(buf.String()), nil
}
