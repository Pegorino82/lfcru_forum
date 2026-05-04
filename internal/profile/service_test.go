package profile

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/user"
)

// --- Mock implementations ---

type mockUserRepo struct {
	u            *user.User
	err          error
	setAvatarURL string
	setAvatarErr error
}

func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*user.User, error) {
	return m.u, m.err
}
func (m *mockUserRepo) SetAvatarURL(_ context.Context, _ int64, url string) error {
	m.setAvatarURL = url
	return m.setAvatarErr
}

type mockPostsRepo struct {
	count    int
	lastTime *time.Time
	err      error
}

func (m *mockPostsRepo) CountByUserID(_ context.Context, _ int64) (int, error) {
	return m.count, m.err
}
func (m *mockPostsRepo) LastPostByUserID(_ context.Context, _ int64) (*time.Time, error) {
	return m.lastTime, m.err
}

type mockCommentsRepo struct {
	count    int
	lastTime *time.Time
	err      error
}

func (m *mockCommentsRepo) CountByUserID(_ context.Context, _ int64) (int, error) {
	return m.count, m.err
}
func (m *mockCommentsRepo) LastCommentByUserID(_ context.Context, _ int64) (*time.Time, error) {
	return m.lastTime, m.err
}

// --- Tests ---

func TestGetProfile_Success(t *testing.T) {
	now := time.Now()
	u := &user.User{ID: 1, Username: "alice"}
	svc := NewService(
		&mockUserRepo{u: u},
		&mockPostsRepo{count: 5, lastTime: &now},
		&mockCommentsRepo{count: 3},
	)

	p, err := svc.GetProfile(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.User.Username != "alice" {
		t.Errorf("Username = %q, want %q", p.User.Username, "alice")
	}
	if p.PostCount != 5 {
		t.Errorf("PostCount = %d, want 5", p.PostCount)
	}
	if p.CommentCount != 3 {
		t.Errorf("CommentCount = %d, want 3", p.CommentCount)
	}
	if p.LastActivityAt == nil || !p.LastActivityAt.Equal(now) {
		t.Errorf("LastActivityAt mismatch")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	svc := NewService(
		&mockUserRepo{err: user.ErrNotFound},
		&mockPostsRepo{},
		&mockCommentsRepo{},
	)
	_, err := svc.GetProfile(context.Background(), "nobody")
	if !errors.Is(err, user.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetProfile_LastActivityLatest(t *testing.T) {
	earlier := time.Now().Add(-2 * time.Hour)
	later := time.Now().Add(-1 * time.Hour)
	svc := NewService(
		&mockUserRepo{u: &user.User{ID: 1, Username: "bob"}},
		&mockPostsRepo{count: 1, lastTime: &earlier},
		&mockCommentsRepo{count: 1, lastTime: &later},
	)
	p, err := svc.GetProfile(context.Background(), "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.LastActivityAt == nil || !p.LastActivityAt.Equal(later) {
		t.Errorf("expected LastActivityAt = later, got %v", p.LastActivityAt)
	}
}

func TestSaveAvatar_ForbiddenIfDifferentUser(t *testing.T) {
	svc := NewService(&mockUserRepo{}, &mockPostsRepo{}, &mockCommentsRepo{})
	_, err := svc.SaveAvatar(context.Background(), 1, 2, bytes.NewReader(nil), t.TempDir())
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestSaveAvatar_UnsupportedFormat(t *testing.T) {
	svc := NewService(&mockUserRepo{}, &mockPostsRepo{}, &mockCommentsRepo{})
	// Передаём случайные байты (не валидное изображение)
	_, err := svc.SaveAvatar(context.Background(), 1, 1, bytes.NewReader([]byte("not an image")), t.TempDir())
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestSaveAvatar_ValidPNG(t *testing.T) {
	dir := t.TempDir()
	avatarURL := ""
	mu := &mockUserRepo{u: &user.User{ID: 42, Username: "tester"}}
	svc := NewService(mu, &mockPostsRepo{}, &mockCommentsRepo{})

	// Создаём валидный PNG в памяти
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	var err error
	avatarURL, err = svc.SaveAvatar(context.Background(), 42, 42, &buf, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avatarURL == "" {
		t.Error("avatarURL should not be empty")
	}

	// Файл должен существовать на диске
	filename := filepath.Join(dir, "42.png")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("avatar file not found at %s", filename)
	}

	// SetAvatarURL должен был быть вызван с правильным URL
	if mu.setAvatarURL != "/storage/avatars/42.png" {
		t.Errorf("SetAvatarURL called with %q, want /storage/avatars/42.png", mu.setAvatarURL)
	}
}
