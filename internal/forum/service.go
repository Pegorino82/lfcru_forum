package forum

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/Pegorino82/lfcru_forum/internal/football"
)

type RepoInterface interface {
	CreateSection(context.Context, *Section) (int64, error)
	UpdateSection(ctx context.Context, id int64, title, description string) error
	CreateTopic(context.Context, *Topic) (int64, error)
	UpdateTopic(ctx context.Context, id int64, title string) error
	CreatePost(context.Context, *Post) (int64, error)
	ListSections(context.Context) ([]SectionView, error)
	GetSection(context.Context, int64) (*Section, error)
	ListTopicsBySection(context.Context, int64) ([]TopicView, error)
	GetTopic(context.Context, int64) (*Topic, error)
	ListPostsByTopic(context.Context, int64) ([]PostView, error)
	ListPostsAfter(ctx context.Context, topicID, afterID int64) ([]PostView, error)
	FindSectionByTitle(ctx context.Context, title string) (*Section, error)
	TopicExistsByTitle(ctx context.Context, sectionID int64, title string) (bool, error)
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

// GetTopic возвращает тему по ID (nil, nil если не найдено).
func (s *Service) GetTopic(ctx context.Context, id int64) (*Topic, error) {
	return s.repo.GetTopic(ctx, id)
}

// ListTopicsBySection возвращает темы раздела.
func (s *Service) ListTopicsBySection(ctx context.Context, sectionID int64) ([]TopicView, error) {
	return s.repo.ListTopicsBySection(ctx, sectionID)
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

// UpdateSection обновляет раздел после валидации.
func (s *Service) UpdateSection(ctx context.Context, id int64, title, description string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrEmptyTitle
	}
	if len(title) > 255 {
		return ErrTitleTooLong
	}
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 2000 {
		return ErrDescriptionTooLong
	}
	return s.repo.UpdateSection(ctx, id, title, description)
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

// UpdateTopic обновляет тему после валидации.
func (s *Service) UpdateTopic(ctx context.Context, id int64, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrEmptyTitle
	}
	if len(title) > 255 {
		return ErrTitleTooLong
	}
	return s.repo.UpdateTopic(ctx, id, title)
}

// ListPostsAfter возвращает до 50 постов темы с id > afterID (для SSE catch-up).
func (s *Service) ListPostsAfter(ctx context.Context, topicID, afterID int64) ([]PostView, error) {
	return s.repo.ListPostsAfter(ctx, topicID, afterID)
}

// GenerateTeamTopics создаёт раздел «Команда» и темы по игрокам из football-data.org.
// Idempotent: пропускает уже существующие раздел и темы.
func (s *Service) GenerateTeamTopics(ctx context.Context, players []football.Player, adminUserID int64) error {
	if len(players) == 0 {
		return nil
	}

	// Найти или создать раздел «Команда».
	const sectionTitle = "Команда"
	section, err := s.repo.FindSectionByTitle(ctx, sectionTitle)
	if err != nil {
		return fmt.Errorf("find section: %w", err)
	}

	var sectionID int64
	if section != nil {
		sectionID = section.ID
	} else {
		sectionID, err = s.repo.CreateSection(ctx, &Section{
			Title:       sectionTitle,
			Description: "Игроки ФК Ливерпуль — обсуждение, статистика, карточки.",
			SortOrder:   100,
		})
		if err != nil {
			return fmt.Errorf("create section: %w", err)
		}
		slog.Info("created section", "title", sectionTitle, "id", sectionID)
	}

	// Создать тему + первый пост для каждого игрока.
	var created, skipped int
	for _, p := range players {
		topicTitle := p.Name

		exists, err := s.repo.TopicExistsByTitle(ctx, sectionID, topicTitle)
		if err != nil {
			return fmt.Errorf("check topic %q: %w", topicTitle, err)
		}
		if exists {
			skipped++
			continue
		}

		topicID, err := s.repo.CreateTopic(ctx, &Topic{
			SectionID: sectionID,
			AuthorID:  adminUserID,
			Title:     topicTitle,
		})
		if err != nil {
			return fmt.Errorf("create topic %q: %w", topicTitle, err)
		}

		content := formatPlayerCard(p)
		_, err = s.repo.CreatePost(ctx, &Post{
			TopicID:  topicID,
			AuthorID: adminUserID,
			Content:  content,
		})
		if err != nil {
			return fmt.Errorf("create post for %q: %w", topicTitle, err)
		}

		created++
	}

	slog.Info("generate team topics done", "created", created, "skipped", skipped)
	return nil
}

// formatPlayerCard формирует текст карточки игрока для первого поста.
func formatPlayerCard(p football.Player) string {
	position := translatePosition(p.Position)

	dob := p.DateOfBirth
	if dob == "" {
		dob = "—"
	}

	nationality := p.Nationality
	if nationality == "" {
		nationality = "—"
	}

	return fmt.Sprintf("Имя: %s\nПозиция: %s\nДата рождения: %s\nНациональность: %s\nНомер: —",
		p.Name, position, dob, nationality)
}

// translatePosition переводит позицию из API в русский.
func translatePosition(pos string) string {
	switch strings.ToUpper(pos) {
	case "GOALKEEPER":
		return "Вратарь"
	case "DEFENCE":
		return "Защитник"
	case "MIDFIELD":
		return "Полузащитник"
	case "OFFENCE":
		return "Нападающий"
	default:
		return pos
	}
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
