package profile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/user"
	_ "golang.org/x/image/webp"
)

var (
	ErrForbidden        = errors.New("forbidden")
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// ForumPostsRepo — минимальный интерфейс для получения статистики постов.
type ForumPostsRepo interface {
	CountByUserID(ctx context.Context, userID int64) (int, error)
	LastPostByUserID(ctx context.Context, userID int64) (*time.Time, error)
}

// CommentsRepo — минимальный интерфейс для получения статистики комментариев.
type CommentsRepo interface {
	CountByUserID(ctx context.Context, userID int64) (int, error)
	LastCommentByUserID(ctx context.Context, userID int64) (*time.Time, error)
}

// UserRepo — минимальный интерфейс для работы с пользователями.
type UserRepo interface {
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	SetAvatarURL(ctx context.Context, userID int64, avatarURL string) error
}

// ProfileData — агрегированные данные профиля пользователя.
type ProfileData struct {
	User           *user.User
	PostCount      int
	CommentCount   int
	LastActivityAt *time.Time
}

type Service struct {
	users    UserRepo
	posts    ForumPostsRepo
	comments CommentsRepo
}

func NewService(users UserRepo, posts ForumPostsRepo, comments CommentsRepo) *Service {
	return &Service{users: users, posts: posts, comments: comments}
}

// GetProfile возвращает агрегированные данные профиля по username.
// Возвращает user.ErrNotFound если пользователь не найден.
func (s *Service) GetProfile(ctx context.Context, username string) (*ProfileData, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	postCount, err := s.posts.CountByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("profile: count posts: %w", err)
	}

	commentCount, err := s.comments.CountByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("profile: count comments: %w", err)
	}

	lastPost, err := s.posts.LastPostByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("profile: last post: %w", err)
	}

	lastComment, err := s.comments.LastCommentByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("profile: last comment: %w", err)
	}

	return &ProfileData{
		User:           u,
		PostCount:      postCount,
		CommentCount:   commentCount,
		LastActivityAt: latestTime(lastPost, lastComment),
	}, nil
}

// SaveAvatar сохраняет аватар пользователя на диск и обновляет avatar_url в БД.
// ownerID — ID аутентифицированного пользователя; targetUserID — ID владельца профиля.
// Возвращает ErrForbidden если ownerID != targetUserID.
// Возвращает ErrUnsupportedFormat для неподдерживаемых форматов (не JPEG/PNG/WebP).
func (s *Service) SaveAvatar(ctx context.Context, ownerID, targetUserID int64, src io.Reader, avatarsDir string) (string, error) {
	if ownerID != targetUserID {
		return "", ErrForbidden
	}

	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("profile: read avatar: %w", err)
	}

	// image.DecodeConfig использует зарегистрированные декодеры (JPEG, PNG, WebP через blank imports)
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", ErrUnsupportedFormat
	}

	switch format {
	case "jpeg", "png", "webp":
		// поддерживаемые форматы
	default:
		return "", ErrUnsupportedFormat
	}

	if err := os.MkdirAll(avatarsDir, 0755); err != nil {
		return "", fmt.Errorf("profile: mkdir avatars: %w", err)
	}

	ext := format
	if ext == "jpeg" {
		ext = "jpg"
	}
	filename := fmt.Sprintf("%d.%s", targetUserID, ext)
	path := filepath.Join(avatarsDir, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("profile: write avatar: %w", err)
	}

	avatarURL := "/storage/avatars/" + filename
	if err := s.users.SetAvatarURL(ctx, targetUserID, avatarURL); err != nil {
		return "", fmt.Errorf("profile: set avatar url: %w", err)
	}

	return avatarURL, nil
}

func latestTime(a, b *time.Time) *time.Time {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.After(*b) {
		return a
	}
	return b
}
