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
	liverpoolTeamID    = 64
	defaultAPIBaseURL  = "https://api.football-data.org/v4"
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

// Client fetches the next Liverpool FC match from football-data.org
// and caches the result for the configured TTL.
type Client struct {
	apiKey     string
	httpClient *http.Client
	ttl        time.Duration
	baseURL    string

	mu        sync.Mutex
	cached    *MatchInfo
	fetchedAt time.Time
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
	return c.cached, nil
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
