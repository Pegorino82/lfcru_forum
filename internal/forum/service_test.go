package forum

import (
	"context"
	"testing"
)

// Mock репозиторий для юнит-тестов
type mockRepo struct {
	createSectionFunc func(context.Context, *Section) (int64, error)
	createTopicFunc   func(context.Context, *Topic) (int64, error)
	createPostFunc    func(context.Context, *Post) (int64, error)
}

func (m *mockRepo) CreateSection(ctx context.Context, s *Section) (int64, error) {
	if m.createSectionFunc != nil {
		return m.createSectionFunc(ctx, s)
	}
	return 1, nil
}

func (m *mockRepo) CreateTopic(ctx context.Context, t *Topic) (int64, error) {
	if m.createTopicFunc != nil {
		return m.createTopicFunc(ctx, t)
	}
	return 1, nil
}

func (m *mockRepo) CreatePost(ctx context.Context, p *Post) (int64, error) {
	if m.createPostFunc != nil {
		return m.createPostFunc(ctx, p)
	}
	return 1, nil
}

func (m *mockRepo) UpdateSection(_ context.Context, _ int64, _, _ string) error { return nil }
func (m *mockRepo) UpdateTopic(_ context.Context, _ int64, _ string) error      { return nil }
func (m *mockRepo) ListSections(context.Context) ([]SectionView, error)         { return nil, nil }
func (m *mockRepo) GetSection(context.Context, int64) (*Section, error)         { return nil, nil }
func (m *mockRepo) ListTopicsBySection(context.Context, int64) ([]TopicView, error) {
	return nil, nil
}
func (m *mockRepo) GetTopic(context.Context, int64) (*Topic, error) { return nil, nil }
func (m *mockRepo) ListPostsByTopic(context.Context, int64) ([]PostView, error) {
	return nil, nil
}
func (m *mockRepo) LatestActive(context.Context, int) ([]TopicWithLastAuthor, error) {
	return nil, nil
}
func (m *mockRepo) ListPostsAfter(_ context.Context, _, _ int64) ([]PostView, error) {
	return nil, nil
}

// Tests for CreateSection

func TestCreateSection_EmptyTitle(t *testing.T) {
	svc := NewService(&mockRepo{})
	_, err := svc.CreateSection(context.Background(), "", "desc", 0)
	if err != ErrEmptyTitle {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestCreateSection_OnlyWhitespace(t *testing.T) {
	svc := NewService(&mockRepo{})
	_, err := svc.CreateSection(context.Background(), "   ", "desc", 0)
	if err != ErrEmptyTitle {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestCreateSection_TitleTooLong(t *testing.T) {
	svc := NewService(&mockRepo{})
	longTitle := string(make([]rune, 256)) // 256 symbols
	_, err := svc.CreateSection(context.Background(), longTitle, "desc", 0)
	if err != ErrTitleTooLong {
		t.Errorf("expected ErrTitleTooLong, got %v", err)
	}
}

func TestCreateSection_DescriptionTooLong(t *testing.T) {
	svc := NewService(&mockRepo{})
	longDesc := string(make([]rune, 2001)) // 2001 symbols
	_, err := svc.CreateSection(context.Background(), "title", longDesc, 0)
	if err != ErrDescriptionTooLong {
		t.Errorf("expected ErrDescriptionTooLong, got %v", err)
	}
}

func TestCreateSection_Valid(t *testing.T) {
	mock := &mockRepo{
		createSectionFunc: func(ctx context.Context, s *Section) (int64, error) {
			if s.Title != "test" || s.Description != "desc" || s.SortOrder != 5 {
				t.Errorf("incorrect values passed to repo: %+v", s)
			}
			return 123, nil
		},
	}
	svc := NewService(mock)
	id, err := svc.CreateSection(context.Background(), " test ", "desc", 5)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id != 123 {
		t.Errorf("expected id 123, got %d", id)
	}
}

// Tests for CreateTopic

func TestCreateTopic_EmptyTitle(t *testing.T) {
	svc := NewService(&mockRepo{})
	_, err := svc.CreateTopic(context.Background(), 1, 1, "")
	if err != ErrEmptyTitle {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestCreateTopic_TitleTooLong(t *testing.T) {
	svc := NewService(&mockRepo{})
	longTitle := string(make([]rune, 256))
	_, err := svc.CreateTopic(context.Background(), 1, 1, longTitle)
	if err != ErrTitleTooLong {
		t.Errorf("expected ErrTitleTooLong, got %v", err)
	}
}

func TestCreateTopic_Valid(t *testing.T) {
	mock := &mockRepo{
		createTopicFunc: func(ctx context.Context, topic *Topic) (int64, error) {
			if topic.SectionID != 10 || topic.AuthorID != 20 || topic.Title != "topic" {
				t.Errorf("incorrect values: %+v", topic)
			}
			return 456, nil
		},
	}
	svc := NewService(mock)
	id, err := svc.CreateTopic(context.Background(), 10, 20, "topic")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id != 456 {
		t.Errorf("expected id 456, got %d", id)
	}
}

// Tests for CreatePost

func TestCreatePost_EmptyContent(t *testing.T) {
	svc := NewService(&mockRepo{})
	_, err := svc.CreatePost(context.Background(), 1, 1, nil, "")
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestCreatePost_OnlyWhitespace(t *testing.T) {
	svc := NewService(&mockRepo{})
	_, err := svc.CreatePost(context.Background(), 1, 1, nil, "   ")
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestCreatePost_ContentTooLong(t *testing.T) {
	svc := NewService(&mockRepo{})
	longContent := string(make([]rune, 20001))
	_, err := svc.CreatePost(context.Background(), 1, 1, nil, longContent)
	if err != ErrContentTooLong {
		t.Errorf("expected ErrContentTooLong, got %v", err)
	}
}

func TestCreatePost_Valid(t *testing.T) {
	mock := &mockRepo{
		createPostFunc: func(ctx context.Context, p *Post) (int64, error) {
			if p.TopicID != 100 || p.AuthorID != 200 || p.Content != "hello" || p.ParentID != nil {
				t.Errorf("incorrect values: %+v", p)
			}
			return 789, nil
		},
	}
	svc := NewService(mock)
	id, err := svc.CreatePost(context.Background(), 100, 200, nil, "hello")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id != 789 {
		t.Errorf("expected id 789, got %d", id)
	}
}

func TestCreatePost_WithParentID(t *testing.T) {
	parentID := int64(50)
	mock := &mockRepo{
		createPostFunc: func(ctx context.Context, p *Post) (int64, error) {
			if p.ParentID == nil || *p.ParentID != 50 {
				t.Errorf("expected parentID 50, got %v", p.ParentID)
			}
			return 999, nil
		},
	}
	svc := NewService(mock)
	id, err := svc.CreatePost(context.Background(), 100, 200, &parentID, "reply")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if id != 999 {
		t.Errorf("expected id 999, got %d", id)
	}
}

func TestCreatePost_RepoReturnsErrParentNotFound(t *testing.T) {
	parentID := int64(50)
	mock := &mockRepo{
		createPostFunc: func(ctx context.Context, p *Post) (int64, error) {
			return 0, ErrParentNotFound
		},
	}
	svc := NewService(mock)
	_, err := svc.CreatePost(context.Background(), 100, 200, &parentID, "reply")
	if err != ErrParentNotFound {
		t.Errorf("expected ErrParentNotFound, got %v", err)
	}
}

func TestCreatePost_RepoReturnsErrReplyToReply(t *testing.T) {
	parentID := int64(50)
	mock := &mockRepo{
		createPostFunc: func(ctx context.Context, p *Post) (int64, error) {
			return 0, ErrReplyToReply
		},
	}
	svc := NewService(mock)
	_, err := svc.CreatePost(context.Background(), 100, 200, &parentID, "reply")
	if err != ErrReplyToReply {
		t.Errorf("expected ErrReplyToReply, got %v", err)
	}
}
