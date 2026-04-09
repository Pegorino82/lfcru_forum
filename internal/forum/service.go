package forum

import (
	"context"
	"strings"
	"unicode/utf8"
)

type RepoInterface interface {
	CreateSection(context.Context, *Section) (int64, error)
	CreateTopic(context.Context, *Topic) (int64, error)
	CreatePost(context.Context, *Post) (int64, error)
	ListSections(context.Context) ([]SectionView, error)
	GetSection(context.Context, int64) (*Section, error)
	ListTopicsBySection(context.Context, int64) ([]TopicView, error)
	GetTopic(context.Context, int64) (*Topic, error)
	ListPostsByTopic(context.Context, int64) ([]PostView, error)
}

type Service struct {
	repo RepoInterface
}

func NewService(repo RepoInterface) *Service {
	return &Service{repo: repo}
}

// ListSections возвращает все разделы.
func (s *Service) ListSections(ctx context.Context) ([]SectionView, error) {
	return s.repo.ListSections(ctx)
}

// GetSection возвращает раздел по ID (nil, nil если не найдено).
func (s *Service) GetSection(ctx context.Context, id int64) (*Section, error) {
	return s.repo.GetSection(ctx, id)
}

// GetSectionWithTopics возвращает раздел и его темы.
func (s *Service) GetSectionWithTopics(ctx context.Context, id int64) (*Section, []TopicView, error) {
	section, err := s.repo.GetSection(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if section == nil {
		return nil, nil, ErrSectionNotFound
	}

	topics, err := s.repo.ListTopicsBySection(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return section, topics, nil
}

// GetTopicWithPosts возвращает тему и её посты.
func (s *Service) GetTopicWithPosts(ctx context.Context, id int64) (*Topic, []PostView, error) {
	topic, err := s.repo.GetTopic(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if topic == nil {
		return nil, nil, ErrTopicNotFound
	}

	posts, err := s.repo.ListPostsByTopic(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return topic, posts, nil
}

// CreateSection создаёт раздел после валидации.
func (s *Service) CreateSection(ctx context.Context, title, description string, sortOrder int) (int64, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, ErrEmptyTitle
	}
	if len(title) > 255 {
		return 0, ErrTitleTooLong
	}

	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 2000 {
		return 0, ErrDescriptionTooLong
	}

	sec := &Section{
		Title:       title,
		Description: description,
		SortOrder:   sortOrder,
	}

	return s.repo.CreateSection(ctx, sec)
}

// CreateTopic создаёт тему после валидации.
func (s *Service) CreateTopic(ctx context.Context, sectionID, authorID int64, title string) (int64, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, ErrEmptyTitle
	}
	if len(title) > 255 {
		return 0, ErrTitleTooLong
	}

	t := &Topic{
		SectionID: sectionID,
		AuthorID:  authorID,
		Title:     title,
	}

	return s.repo.CreateTopic(ctx, t)
}

// CreatePost создаёт пост после валидации.
func (s *Service) CreatePost(ctx context.Context, topicID, authorID int64, parentID *int64, content string) (int64, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return 0, ErrEmptyContent
	}
	if utf8.RuneCountInString(content) > 20000 {
		return 0, ErrContentTooLong
	}

	p := &Post{
		TopicID:  topicID,
		AuthorID: authorID,
		ParentID: parentID,
		Content:  content,
	}

	return s.repo.CreatePost(ctx, p)
}
