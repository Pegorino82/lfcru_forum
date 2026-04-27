package football

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_NextMatch_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth-Token") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		resp := map[string]any{
			"matches": []map[string]any{
				{
					"utcDate": "2026-05-03T14:30:00Z",
					"competition": map[string]any{"name": "Premier League"},
					"homeTeam":    map[string]any{"id": 66, "name": "Manchester United FC"},
					"awayTeam":    map[string]any{"id": 64, "name": "Liverpool FC"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("test-key", time.Hour)
	c.httpClient = &http.Client{Timeout: 5 * time.Second}
	c.baseURL = srv.URL

	info, err := c.NextMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected MatchInfo, got nil")
	}
	if info.Opponent != "Manchester United FC" {
		t.Errorf("opponent: got %q, want %q", info.Opponent, "Manchester United FC")
	}
	if info.IsHome {
		t.Errorf("expected away match")
	}
}

func TestClient_NextMatch_EmptyAPIKey(t *testing.T) {
	c := NewClient("", time.Hour)
	info, err := c.NextMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Errorf("expected nil for empty API key, got %+v", info)
	}
}

func TestClient_NextMatch_CacheHit(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := map[string]any{
			"matches": []map[string]any{
				{
					"utcDate":     "2026-05-03T14:30:00Z",
					"competition": map[string]any{"name": "Premier League"},
					"homeTeam":    map[string]any{"id": 66, "name": "Manchester United FC"},
					"awayTeam":    map[string]any{"id": 64, "name": "Liverpool FC"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)

	ctx := context.Background()
	info1, err := c.NextMatch(ctx)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if info1 == nil {
		t.Fatal("expected non-nil on first call")
	}

	info2, err := c.NextMatch(ctx)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected exactly 1 API call, got %d", callCount)
	}
	if info2.Opponent != info1.Opponent {
		t.Errorf("cache returned different data")
	}
}

func TestClient_NextMatch_APIError_ReturnsCachedData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	// Pre-populate cache.
	c.cached = &MatchInfo{Opponent: "Arsenal FC", MatchDate: time.Now().Add(24 * time.Hour)}
	c.fetchedAt = time.Now() // fresh — won't re-fetch

	info, err := c.NextMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil || info.Opponent != "Arsenal FC" {
		t.Errorf("expected cached data, got %v", info)
	}
}

func TestClient_NextMatch_NoMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"matches": []any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	info, err := c.NextMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Errorf("expected nil for empty matches, got %+v", info)
	}
}

func TestLookupVenue_KnownTeam(t *testing.T) {
	v := lookupVenue("Manchester United FC")
	if v.Stadium != "Old Trafford" {
		t.Errorf("expected Old Trafford, got %q", v.Stadium)
	}
	if v.City != "Manchester" {
		t.Errorf("expected Manchester, got %q", v.City)
	}
	if v.Country != "England" {
		t.Errorf("expected England, got %q", v.Country)
	}
}

func TestLookupVenue_UnknownTeam(t *testing.T) {
	v := lookupVenue("Unknown FC")
	if v.Stadium != "" || v.City != "" || v.Country != "" {
		t.Errorf("expected zero value for unknown team, got %+v", v)
	}
}

func TestClient_LastMatch_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth-Token") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		resp := map[string]any{
			"matches": []map[string]any{
				{
					"utcDate":     "2026-04-19T15:30:00Z",
					"competition": map[string]any{"name": "Premier League"},
					"homeTeam":    map[string]any{"id": 64, "name": "Liverpool FC"},
					"awayTeam":    map[string]any{"id": 57, "name": "Arsenal FC"},
					"score": map[string]any{
						"fullTime": map[string]any{"home": 3, "away": 1},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)

	info, err := c.LastMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected LastMatchInfo, got nil")
	}
	if info.Opponent != "Arsenal FC" {
		t.Errorf("opponent: got %q, want %q", info.Opponent, "Arsenal FC")
	}
	if !info.IsHome {
		t.Errorf("expected home match")
	}
	if info.HomeScore != 3 || info.AwayScore != 1 {
		t.Errorf("score: got %d:%d, want 3:1", info.HomeScore, info.AwayScore)
	}
	if info.ForumURL != "#" {
		t.Errorf("forumURL: got %q, want %q", info.ForumURL, "#")
	}
}

func TestClient_LastMatch_CacheHit(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := map[string]any{
			"matches": []map[string]any{
				{
					"utcDate":     "2026-04-19T15:30:00Z",
					"competition": map[string]any{"name": "Premier League"},
					"homeTeam":    map[string]any{"id": 64, "name": "Liverpool FC"},
					"awayTeam":    map[string]any{"id": 57, "name": "Arsenal FC"},
					"score": map[string]any{
						"fullTime": map[string]any{"home": 2, "away": 0},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	// Set nextKickoff in the future so lastMatchTTL returns a positive duration.
	c.nextKickoff = time.Now().Add(2 * time.Hour)

	ctx := context.Background()
	info1, err := c.LastMatch(ctx)
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	if info1 == nil {
		t.Fatal("expected non-nil on first call")
	}

	info2, err := c.LastMatch(ctx)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected exactly 1 API call, got %d", callCount)
	}
	if info2.Opponent != info1.Opponent {
		t.Errorf("cache returned different data")
	}
}

func TestClient_LastMatch_NoFinished(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"matches": []any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	info, err := c.LastMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Errorf("expected nil for empty matches, got %+v", info)
	}
}

func TestClient_LastMatch_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	info, err := c.LastMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No stale cache — should return nil gracefully.
	if info != nil {
		t.Errorf("expected nil on API error with empty cache, got %+v", info)
	}
}

func TestClient_LastMatch_EmptyAPIKey(t *testing.T) {
	c := NewClient("", time.Hour)
	info, err := c.LastMatch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Errorf("expected nil for empty API key, got %+v", info)
	}
}

// newTestClient creates a Client pointing to a test HTTP server instead of the real API.
func newTestClient(apiKey string, ttl time.Duration, serverURL string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		ttl:        ttl,
		baseURL:    serverURL,
	}
}

// standingsResponse builds a minimal standings API response with one row.
func standingsResponse(pos int, name, crest string, played, gf, ga, gd, pts int) map[string]any {
	return map[string]any{
		"standings": []map[string]any{
			{
				"type": "TOTAL",
				"table": []map[string]any{
					{
						"position":       pos,
						"team":           map[string]any{"name": name, "crest": crest},
						"playedGames":    played,
						"goalsFor":       gf,
						"goalsAgainst":   ga,
						"goalDifference": gd,
						"points":         pts,
					},
				},
			},
		},
	}
}

func TestClient_Standings_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth-Token") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(standingsResponse(1, "Liverpool FC", "https://crest/lfc.png", 32, 70, 28, 42, 79))
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	entries, err := c.Standings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Position != 1 {
		t.Errorf("position: got %d, want 1", e.Position)
	}
	if e.TeamName != "Liverpool FC" {
		t.Errorf("team name: got %q, want %q", e.TeamName, "Liverpool FC")
	}
	if e.Points != 79 {
		t.Errorf("points: got %d, want 79", e.Points)
	}
	if e.TeamCrest != "https://crest/lfc.png" {
		t.Errorf("crest: got %q", e.TeamCrest)
	}
}

func TestClient_Standings_EmptyAPIKey(t *testing.T) {
	c := NewClient("", time.Hour)
	entries, err := c.Standings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil for empty API key, got %v", entries)
	}
}

func TestClient_Standings_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	entries, err := c.Standings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil on API error with empty cache, got %v", entries)
	}
}

func TestClient_Standings_CacheHit(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(standingsResponse(1, "Liverpool FC", "", 32, 70, 28, 42, 79))
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	ctx := context.Background()

	if _, err := c.Standings(ctx); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if _, err := c.Standings(ctx); err != nil {
		t.Fatalf("second call: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected exactly 1 API call, got %d", callCount)
	}
}

func TestClient_StandingsTTL_Weekday(t *testing.T) {
	// Monday
	monday := time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)
	got := standingsTTL(monday)
	if got != standingsTTLWeekday {
		t.Errorf("weekday TTL: got %v, want %v", got, standingsTTLWeekday)
	}
}

func TestClient_StandingsTTL_Weekend(t *testing.T) {
	// Saturday
	saturday := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	got := standingsTTL(saturday)
	if got != standingsTTLWeekend {
		t.Errorf("weekend TTL: got %v, want %v", got, standingsTTLWeekend)
	}
	// Sunday
	sunday := time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC)
	got = standingsTTL(sunday)
	if got != standingsTTLWeekend {
		t.Errorf("sunday TTL: got %v, want %v", got, standingsTTLWeekend)
	}
}

func TestClient_Standings_InvalidatedOnNewLastMatch(t *testing.T) {
	oldMatchDate := time.Date(2026, 4, 20, 14, 0, 0, 0, time.UTC)
	newMatchDate := time.Date(2026, 4, 27, 14, 0, 0, 0, time.UTC)

	standingsCallCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/competitions/PL/standings" {
			standingsCallCount++
			json.NewEncoder(w).Encode(standingsResponse(1, "Liverpool FC", "", 32, 70, 28, 42, 79))
			return
		}
		// Last match response
		json.NewEncoder(w).Encode(map[string]any{
			"matches": []map[string]any{
				{
					"utcDate":     newMatchDate.Format(time.RFC3339),
					"competition": map[string]any{"name": "Premier League"},
					"homeTeam":    map[string]any{"id": 64, "name": "Liverpool FC"},
					"awayTeam":    map[string]any{"id": 66, "name": "Manchester United FC"},
					"score": map[string]any{
						"fullTime": map[string]any{"home": 2, "away": 0},
					},
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient("test-key", time.Hour, srv.URL)
	// Seed lastKnownMatchDate so next LastMatch fetch looks like a new match.
	c.lastKnownMatchDate = oldMatchDate

	ctx := context.Background()

	// Prime standings cache.
	if _, err := c.Standings(ctx); err != nil {
		t.Fatalf("standings prime: %v", err)
	}
	if standingsCallCount != 1 {
		t.Fatalf("expected 1 standings call after prime, got %d", standingsCallCount)
	}

	// Trigger LastMatch fetch — detects new match → invalidates standings cache.
	if _, err := c.LastMatch(ctx); err != nil {
		t.Fatalf("last match: %v", err)
	}

	// Standings should now fetch fresh data.
	if _, err := c.Standings(ctx); err != nil {
		t.Fatalf("standings after invalidation: %v", err)
	}
	if standingsCallCount != 2 {
		t.Errorf("expected 2 standings calls after invalidation, got %d", standingsCallCount)
	}
}
