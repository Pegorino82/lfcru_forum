package comment

import (
	"context"
	"strings"
	"testing"

	"github.com/Pegorino82/lfcru_forum/internal/user"
)

// mockRepo implements repoCreator.
type mockRepo struct {
	created []*Comment
}

func (m *mockRepo) Create(_ context.Context, c *Comment) (int64, error) {
	m.created = append(m.created, c)
	return int64(len(m.created)), nil
}

// mockUserRepo implements UserRepo.
type mockUserRepo struct {
	users map[string]bool // lowercase username → exists
}

func (m *mockUserRepo) GetByUsernames(_ context.Context, usernames []string) ([]user.User, error) {
	var result []user.User
	for _, u := range usernames {
		if m.users[strings.ToLower(u)] {
			result = append(result, user.User{Username: u})
		}
	}
	return result, nil
}

func newTestSvc(existingUsers ...string) (*Service, *mockRepo, *mockUserRepo) {
	mRepo := &mockRepo{}
	mUsers := &mockUserRepo{users: make(map[string]bool)}
	for _, u := range existingUsers {
		mUsers.users[strings.ToLower(u)] = true
	}
	svc := &Service{repo: mRepo, userRepo: mUsers}
	return svc, mRepo, mUsers
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestService_Create_EmptyContent(t *testing.T) {
	svc, _, _ := newTestSvc()
	_, err := svc.Create(context.Background(), &Comment{Content: ""})
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestService_Create_OnlySpaces(t *testing.T) {
	svc, _, _ := newTestSvc()
	_, err := svc.Create(context.Background(), &Comment{Content: "   \t\n"})
	if err != ErrEmptyContent {
		t.Errorf("expected ErrEmptyContent, got %v", err)
	}
}

func TestService_Create_TooLong(t *testing.T) {
	svc, _, _ := newTestSvc()
	content := strings.Repeat("а", 10001)
	_, err := svc.Create(context.Background(), &Comment{Content: content})
	if err != ErrContentTooLong {
		t.Errorf("expected ErrContentTooLong, got %v", err)
	}
}

func TestService_Create_Valid(t *testing.T) {
	svc, mRepo, _ := newTestSvc()
	_, err := svc.Create(context.Background(), &Comment{Content: "  Привет мир  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mRepo.created) != 1 {
		t.Fatal("expected one comment to be created")
	}
	if mRepo.created[0].Content != "Привет мир" {
		t.Errorf("expected trimmed content %q, got %q", "Привет мир", mRepo.created[0].Content)
	}
}

// ─── RenderMentions ───────────────────────────────────────────────────────────

func TestService_RenderMentions_ExistingUser(t *testing.T) {
	svc, _, _ := newTestSvc("admin")
	got, err := svc.RenderMentions(context.Background(), "@admin привет")
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	if !strings.Contains(out, `class="mention"`) {
		t.Errorf("expected mention span for existing user, got: %s", out)
	}
	if !strings.Contains(out, "@admin") {
		t.Errorf("expected @admin in output, got: %s", out)
	}
}

func TestService_RenderMentions_NonExistingUser(t *testing.T) {
	svc, _, _ := newTestSvc() // no users
	got, err := svc.RenderMentions(context.Background(), "@ghost hello")
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	if strings.Contains(out, `class="mention"`) {
		t.Errorf("unknown user should not get mention span, got: %s", out)
	}
	if !strings.Contains(out, "@ghost") {
		t.Errorf("expected @ghost in plain text, got: %s", out)
	}
}

func TestService_RenderMentions_EmailNotMention(t *testing.T) {
	svc, _, _ := newTestSvc("mail")
	got, err := svc.RenderMentions(context.Background(), "user@mail.ru отличный email")
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	if strings.Contains(out, `class="mention"`) {
		t.Errorf("email address should not be treated as mention, got: %s", out)
	}
}

func TestService_RenderMentions_XSSEscaped(t *testing.T) {
	svc, _, _ := newTestSvc("admin")
	got, err := svc.RenderMentions(context.Background(), `<script>alert("xss")</script> @admin`)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	if strings.Contains(out, "<script>") {
		t.Errorf("XSS not escaped, got: %s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Errorf("expected escaped script tag, got: %s", out)
	}
	if !strings.Contains(out, `class="mention"`) {
		t.Errorf("@admin should still be a mention after XSS content, got: %s", out)
	}
}

func TestService_RenderMentions_Dedup(t *testing.T) {
	svc, _, _ := newTestSvc("admin")
	got, err := svc.RenderMentions(context.Background(), "@admin hello @admin world")
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	count := strings.Count(out, `class="mention"`)
	if count != 2 {
		t.Errorf("expected 2 mention spans for duplicate @admin, got %d in: %s", count, out)
	}
}
