package tmpl

import (
	"testing"
	"time"
)

// extractFuncMap helpers: вызываем функции из funcMap по имени.
func getFuncInitials() func(string) string {
	return funcMap["avatarInitials"].(func(string) string)
}
func getFuncColor() func(string) string {
	return funcMap["avatarColor"].(func(string) string)
}
func getFuncRelTime() func(time.Time) string {
	return funcMap["relativeTime"].(func(time.Time) string)
}

func TestAvatarInitials(t *testing.T) {
	fn := getFuncInitials()
	tests := []struct {
		input string
		want  string
	}{
		{"John Doe", "JD"},
		{"john doe", "JD"},
		{"Иван Петров", "ИП"},
		{"john", "JO"},
		{"x", "X"},
		{"", "?"},
		{"ab", "AB"},
	}
	for _, tc := range tests {
		got := fn(tc.input)
		if got != tc.want {
			t.Errorf("avatarInitials(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestAvatarColorDeterministic(t *testing.T) {
	fn := getFuncColor()
	// Один и тот же username всегда даёт одинаковый цвет (SC-06)
	for i := 0; i < 5; i++ {
		c1 := fn("testuser")
		c2 := fn("testuser")
		if c1 != c2 {
			t.Errorf("avatarColor not deterministic: got %q and %q", c1, c2)
		}
	}
}

func TestAvatarColorInPalette(t *testing.T) {
	fn := getFuncColor()
	palette := make(map[string]bool, len(avatarPalette))
	for _, c := range avatarPalette {
		palette[c] = true
	}
	usernames := []string{"alice", "bob", "charlie", "dave", "eve", "frank", "grace", "heidi"}
	for _, u := range usernames {
		c := fn(u)
		if !palette[c] {
			t.Errorf("avatarColor(%q) = %q, not in palette", u, c)
		}
	}
}

func TestAvatarColorDifferentUsers(t *testing.T) {
	fn := getFuncColor()
	// Разные имена должны давать хотя бы несколько разных цветов (проверяем не все одинаковые)
	colors := make(map[string]bool)
	usernames := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for _, u := range usernames {
		colors[fn(u)] = true
	}
	if len(colors) < 2 {
		t.Errorf("avatarColor: all %d usernames returned same color, expected variety", len(usernames))
	}
}

func TestRelativeTimeJustNow(t *testing.T) {
	fn := getFuncRelTime()
	result := fn(time.Now().Add(-10 * time.Second))
	if result != "только что" {
		t.Errorf("relativeTime(10s ago) = %q, want %q", result, "только что")
	}
}

func TestRelativeTimeMinutes(t *testing.T) {
	fn := getFuncRelTime()
	result := fn(time.Now().Add(-5 * time.Minute))
	expected := "5 минут назад"
	if result != expected {
		t.Errorf("relativeTime(5m ago) = %q, want %q", result, expected)
	}
}

func TestRelativeTimeOneMinute(t *testing.T) {
	fn := getFuncRelTime()
	result := fn(time.Now().Add(-1 * time.Minute))
	expected := "1 минуту назад"
	if result != expected {
		t.Errorf("relativeTime(1m ago) = %q, want %q", result, expected)
	}
}

func TestRelativeTimeHours(t *testing.T) {
	fn := getFuncRelTime()
	result := fn(time.Now().Add(-3 * time.Hour))
	expected := "3 часа назад"
	if result != expected {
		t.Errorf("relativeTime(3h ago) = %q, want %q", result, expected)
	}
}

func TestRelativeTimeYesterday(t *testing.T) {
	fn := getFuncRelTime()
	result := fn(time.Now().Add(-25 * time.Hour))
	expected := "вчера"
	if result != expected {
		t.Errorf("relativeTime(25h ago) = %q, want %q", result, expected)
	}
}

func TestRelativeTimeDays(t *testing.T) {
	fn := getFuncRelTime()
	result := fn(time.Now().Add(-72 * time.Hour))
	expected := "3 дня назад"
	if result != expected {
		t.Errorf("relativeTime(72h ago) = %q, want %q", result, expected)
	}
}

func TestPluralRu(t *testing.T) {
	tests := []struct {
		n            int
		f1, f2, f5   string
		want         string
	}{
		{1, "минуту", "минуты", "минут", "минуту"},
		{2, "минуту", "минуты", "минут", "минуты"},
		{5, "минуту", "минуты", "минут", "минут"},
		{11, "минуту", "минуты", "минут", "минут"},
		{21, "минуту", "минуты", "минут", "минуту"},
		{100, "минуту", "минуты", "минут", "минут"},
	}
	for _, tc := range tests {
		got := pluralRu(tc.n, tc.f1, tc.f2, tc.f5)
		if got != tc.want {
			t.Errorf("pluralRu(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}
