package forum

import (
	"context"
	"strings"
	"testing"

	"github.com/Pegorino82/lfcru_forum/internal/football"
)

// Mock репозиторий для юнит-тестов
type mockRepo struct {
	createSectionFunc      func(context.Context, *Section) (int64, error)
	createTopicFunc        func(context.Context, *Topic) (int64, error)
	createPostFunc         func(context.Context, *Post) (int64, error)
	findSectionByTitleFunc func(context.Context, string) (*Section, error)
	topicExistsByTitleFunc func(context.Context, int64, string) (bool, error)
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

func (m *mockRepo) FindSectionByTitle(ctx context.Context, title string) (*Section, error) {
	if m.findSectionByTitleFunc != nil {
		return m.findSectionByTitleFunc(ctx, title)
	}
	return nil, nil
}

func (m *mockRepo) TopicExistsByTitle(ctx context.Context, sectionID int64, title string) (bool, error) {
	if m.topicExistsByTitleFunc != nil {
		return m.topicExistsByTitleFunc(ctx, sectionID, title)
	}
	return false, nil
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

// --- GenerateTeamTopics tests ---

func testPlayers() []football.Player {
	return []football.Player{
		{ID: 1, Name: "Mohamed Salah", Position: "Offence", DateOfBirth: "1992-06-15", Nationality: "Egypt"},
		{ID: 2, Name: "Virgil van Dijk", Position: "Defence", DateOfBirth: "1991-07-08", Nationality: "Netherlands"},
	}
}

func TestGenerateTeamTopics_EmptyPlayers(t *testing.T) {
	svc := NewService(&mockRepo{})
	err := svc.GenerateTeamTopics(context.Background(), nil, 1)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestGenerateTeamTopics_CreatesSectionAndTopics(t *testing.T) {
	var createdTopics []string
	var createdPosts []string

	mock := &mockRepo{
		findSectionByTitleFunc: func(_ context.Context, title string) (*Section, error) {
			if title != "Команда" {
				t.Errorf("expected section title 'Команда', got %q", title)
			}
			return nil, nil // section not found
		},
		createSectionFunc: func(_ context.Context, s *Section) (int64, error) {
			if s.Title != "Команда" {
				t.Errorf("expected section title 'Команда', got %q", s.Title)
			}
			return 10, nil
		},
		topicExistsByTitleFunc: func(_ context.Context, sectionID int64, title string) (bool, error) {
			if sectionID != 10 {
				t.Errorf("expected sectionID 10, got %d", sectionID)
			}
			return false, nil
		},
		createTopicFunc: func(_ context.Context, topic *Topic) (int64, error) {
			createdTopics = append(createdTopics, topic.Title)
			if topic.SectionID != 10 {
				t.Errorf("expected sectionID 10, got %d", topic.SectionID)
			}
			if topic.AuthorID != 1 {
				t.Errorf("expected authorID 1, got %d", topic.AuthorID)
			}
			return int64(100 + len(createdTopics)), nil
		},
		createPostFunc: func(_ context.Context, p *Post) (int64, error) {
			createdPosts = append(createdPosts, p.Content)
			if p.AuthorID != 1 {
				t.Errorf("expected authorID 1, got %d", p.AuthorID)
			}
			return 1, nil
		},
	}

	svc := NewService(mock)
	err := svc.GenerateTeamTopics(context.Background(), testPlayers(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(createdTopics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(createdTopics))
	}
	if createdTopics[0] != "Mohamed Salah" {
		t.Errorf("expected topic 'Mohamed Salah', got %q", createdTopics[0])
	}
	if createdTopics[1] != "Virgil van Dijk" {
		t.Errorf("expected topic 'Virgil van Dijk', got %q", createdTopics[1])
	}

	if len(createdPosts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(createdPosts))
	}
	if !strings.Contains(createdPosts[0], "Нападающий") {
		t.Errorf("expected post to contain 'Нападающий', got %q", createdPosts[0])
	}
}

func TestGenerateTeamTopics_ExistingSectionReused(t *testing.T) {
	sectionCreated := false

	mock := &mockRepo{
		findSectionByTitleFunc: func(_ context.Context, _ string) (*Section, error) {
			return &Section{ID: 42, Title: "Команда"}, nil
		},
		createSectionFunc: func(_ context.Context, _ *Section) (int64, error) {
			sectionCreated = true
			return 0, nil
		},
		topicExistsByTitleFunc: func(_ context.Context, sectionID int64, _ string) (bool, error) {
			if sectionID != 42 {
				t.Errorf("expected sectionID 42, got %d", sectionID)
			}
			return false, nil
		},
		createTopicFunc: func(_ context.Context, topic *Topic) (int64, error) {
			if topic.SectionID != 42 {
				t.Errorf("expected sectionID 42, got %d", topic.SectionID)
			}
			return 100, nil
		},
	}

	svc := NewService(mock)
	err := svc.GenerateTeamTopics(context.Background(), testPlayers(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sectionCreated {
		t.Error("should not create section when it already exists")
	}
}

func TestGenerateTeamTopics_IdempotentSkipsExisting(t *testing.T) {
	var createdTopics int

	mock := &mockRepo{
		findSectionByTitleFunc: func(_ context.Context, _ string) (*Section, error) {
			return &Section{ID: 10, Title: "Команда"}, nil
		},
		topicExistsByTitleFunc: func(_ context.Context, _ int64, title string) (bool, error) {
			// Salah already exists, van Dijk doesn't
			return title == "Mohamed Salah", nil
		},
		createTopicFunc: func(_ context.Context, topic *Topic) (int64, error) {
			createdTopics++
			if topic.Title == "Mohamed Salah" {
				t.Error("should not create topic for Salah — already exists")
			}
			return 100, nil
		},
	}

	svc := NewService(mock)
	err := svc.GenerateTeamTopics(context.Background(), testPlayers(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdTopics != 1 {
		t.Errorf("expected 1 topic created, got %d", createdTopics)
	}
}

// --- formatPlayerCard / translatePosition tests ---

func TestFormatPlayerCard(t *testing.T) {
	p := football.Player{
		Name:        "Mohamed Salah",
		Position:    "Offence",
		DateOfBirth: "1992-06-15",
		Nationality: "Egypt",
	}
	card := formatPlayerCard(p)

	expected := []string{
		"Имя: Mohamed Salah",
		"Позиция: Нападающий",
		"Дата рождения: 1992-06-15",
		"Национальность: Egypt",
		"Номер: —",
	}
	for _, s := range expected {
		if !strings.Contains(card, s) {
			t.Errorf("card missing %q:\n%s", s, card)
		}
	}
}

func TestFormatPlayerCard_EmptyFields(t *testing.T) {
	p := football.Player{Name: "Test Player"}
	card := formatPlayerCard(p)

	if !strings.Contains(card, "Дата рождения: —") {
		t.Errorf("expected dash for empty dob:\n%s", card)
	}
	if !strings.Contains(card, "Национальность: —") {
		t.Errorf("expected dash for empty nationality:\n%s", card)
	}
}

func TestTranslatePosition(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Goalkeeper", "Вратарь"},
		{"DEFENCE", "Защитник"},
		{"Midfield", "Полузащитник"},
		{"Offence", "Нападающий"},
		{"Unknown", "Unknown"},
		{"", ""},
	}
	for _, tt := range tests {
		got := translatePosition(tt.input)
		if got != tt.want {
			t.Errorf("translatePosition(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
