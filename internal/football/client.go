package football

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	liverpoolTeamID   = 64
	defaultAPIBaseURL = "https://api.football-data.org/v4"
	fallbackLastTTL   = 24 * time.Hour
)

// MatchInfo holds the data for the next upcoming match.
type MatchInfo struct {
	Opponent   string
	MatchDate  time.Time
	Stadium    string
	City       string
	Country    string
	IsHome     bool
	Tournament string
}

// LastMatchInfo holds the data for the last finished match.
type LastMatchInfo struct {
	Opponent   string
	MatchDate  time.Time
	Stadium    string
	City       string
	Country    string
	IsHome     bool
	Tournament string
	HomeScore  int
	AwayScore  int
	ForumURL   string
}

// Client fetches Liverpool FC matches from football-data.org
// and caches the results.
type Client struct {
	apiKey     string
	httpClient *http.Client
	ttl        time.Duration
	baseURL    string

	mu          sync.Mutex
	cached      *MatchInfo
	fetchedAt   time.Time
	nextKickoff time.Time

	cachedLast      *LastMatchInfo
	lastFetchedAt   time.Time
}

// NewClient creates a Client with the given API key and cache TTL.
func NewClient(apiKey string, ttl time.Duration) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		ttl:        ttl,
		baseURL:    defaultAPIBaseURL,
	}
}

// NextMatch returns the next scheduled Liverpool FC match.
// Results are cached for the configured TTL.
// Returns nil if the API key is absent, the API is unavailable, or no matches are scheduled.
func (c *Client) NextMatch(ctx context.Context) (*MatchInfo, error) {
	if c.apiKey == "" {
		return nil, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cached != nil && time.Since(c.fetchedAt) < c.ttl {
		return c.cached, nil
	}

	info, err := c.fetch(ctx)
	if err != nil {
		// Return stale cache on error rather than surfacing the error to the page.
		return c.cached, nil
	}

	c.cached = info
	c.fetchedAt = time.Now()
	if info != nil {
		c.nextKickoff = info.MatchDate
	}
	return c.cached, nil
}

// LastMatch returns the last finished Liverpool FC match.
// Results are cached until the kickoff of the next match (or fallbackLastTTL if unknown).
// Returns nil if the API key is absent, the API is unavailable, or no finished matches exist.
func (c *Client) LastMatch(ctx context.Context) (*LastMatchInfo, error) {
	if c.apiKey == "" {
		return nil, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	ttl := c.lastMatchTTL()
	if c.cachedLast != nil && time.Since(c.lastFetchedAt) < ttl {
		return c.cachedLast, nil
	}

	info, err := c.fetchLast(ctx)
	if err != nil {
		return c.cachedLast, nil
	}

	c.cachedLast = info
	c.lastFetchedAt = time.Now()
	return c.cachedLast, nil
}

// lastMatchTTL returns the cache TTL for the last match.
// Uses time until next kickoff, or fallbackLastTTL if kickoff is unknown or in the past.
func (c *Client) lastMatchTTL() time.Duration {
	if c.nextKickoff.IsZero() {
		return fallbackLastTTL
	}
	ttl := time.Until(c.nextKickoff)
	if ttl <= 0 {
		return fallbackLastTTL
	}
	return ttl
}

// apiResponse is the relevant subset of the football-data.org /v4/teams/{id}/matches response.
type apiResponse struct {
	Matches []struct {
		UTCDate     string `json:"utcDate"`
		Competition struct {
			Name string `json:"name"`
		} `json:"competition"`
		HomeTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"homeTeam"`
		AwayTeam struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"awayTeam"`
		Score struct {
			FullTime struct {
				Home *int `json:"home"`
				Away *int `json:"away"`
			} `json:"fullTime"`
		} `json:"score"`
	} `json:"matches"`
}

func (c *Client) fetch(ctx context.Context) (*MatchInfo, error) {
	url := fmt.Sprintf("%s/teams/%d/matches?status=SCHEDULED&limit=1", c.baseURL, liverpoolTeamID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("football-data.org: status %d", resp.StatusCode)
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Matches) == 0 {
		return nil, nil
	}

	m := data.Matches[0]

	matchDate, err := time.Parse(time.RFC3339, m.UTCDate)
	if err != nil {
		return nil, fmt.Errorf("parse utcDate %q: %w", m.UTCDate, err)
	}

	isHome := m.HomeTeam.ID == liverpoolTeamID
	var opponent, homeTeamName string
	if isHome {
		opponent = m.AwayTeam.Name
		homeTeamName = "Liverpool FC"
	} else {
		opponent = m.HomeTeam.Name
		homeTeamName = m.HomeTeam.Name
	}

	v := lookupVenue(homeTeamName)

	return &MatchInfo{
		Opponent:   opponent,
		MatchDate:  matchDate,
		Stadium:    v.Stadium,
		City:       v.City,
		Country:    v.Country,
		IsHome:     isHome,
		Tournament: m.Competition.Name,
	}, nil
}

func (c *Client) fetchLast(ctx context.Context) (*LastMatchInfo, error) {
	url := fmt.Sprintf("%s/teams/%d/matches?status=FINISHED&limit=1", c.baseURL, liverpoolTeamID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("football-data.org: status %d", resp.StatusCode)
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Matches) == 0 {
		return nil, nil
	}

	m := data.Matches[0]

	matchDate, err := time.Parse(time.RFC3339, m.UTCDate)
	if err != nil {
		return nil, fmt.Errorf("parse utcDate %q: %w", m.UTCDate, err)
	}

	isHome := m.HomeTeam.ID == liverpoolTeamID
	var opponent, homeTeamName string
	if isHome {
		opponent = m.AwayTeam.Name
		homeTeamName = "Liverpool FC"
	} else {
		opponent = m.HomeTeam.Name
		homeTeamName = m.HomeTeam.Name
	}

	v := lookupVenue(homeTeamName)

	homeScore := 0
	awayScore := 0
	if m.Score.FullTime.Home != nil {
		homeScore = *m.Score.FullTime.Home
	}
	if m.Score.FullTime.Away != nil {
		awayScore = *m.Score.FullTime.Away
	}

	return &LastMatchInfo{
		Opponent:   opponent,
		MatchDate:  matchDate,
		Stadium:    v.Stadium,
		City:       v.City,
		Country:    v.Country,
		IsHome:     isHome,
		Tournament: m.Competition.Name,
		HomeScore:  homeScore,
		AwayScore:  awayScore,
		ForumURL:   "#",
	}, nil
}
