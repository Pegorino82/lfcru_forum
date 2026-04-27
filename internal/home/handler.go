package home

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/football"
	"github.com/Pegorino82/lfcru_forum/internal/forum"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/match"
	"github.com/Pegorino82/lfcru_forum/internal/news"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

// FootballSource is the interface for fetching Liverpool FC match data.
type FootballSource interface {
	NextMatch(ctx context.Context) (*football.MatchInfo, error)
	LastMatch(ctx context.Context) (*football.LastMatchInfo, error)
	Standings(ctx context.Context) ([]football.StandingsEntry, error)
}

type HomeData struct {
	User              *user.User
	CSRFToken         string
	News              []news.News
	NextMatch         *match.Match
	NextFootballMatch *football.MatchInfo
	LastFootballMatch *football.LastMatchInfo
	Topics            []forum.TopicWithLastAuthor
	Standings         []football.StandingsEntry
}

type Handler struct {
	newsRepo       *news.Repo
	matchRepo      *match.Repo
	topicRepo      *forum.Repo
	footballClient FootballSource
}

func NewHandler(newsRepo *news.Repo, matchRepo *match.Repo, topicRepo *forum.Repo, footballClient FootballSource) *Handler {
	return &Handler{
		newsRepo:       newsRepo,
		matchRepo:      matchRepo,
		topicRepo:      topicRepo,
		footballClient: footballClient,
	}
}

func (h *Handler) ShowHome(c echo.Context) error {
	ctx := c.Request().Context()

	newsList, err := h.newsRepo.LatestPublished(ctx, 5)
	if err != nil {
		slog.Error("home: failed to load news", "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}

	nextMatch, err := h.matchRepo.NextUpcoming(ctx, time.Now())
	if err != nil {
		slog.Error("home: failed to load match", "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}

	topics, err := h.topicRepo.LatestActive(ctx, 5)
	if err != nil {
		slog.Error("home: failed to load topics", "err", err)
		return c.String(http.StatusInternalServerError, "Что-то пошло не так. Попробуйте обновить страницу.")
	}

	var nextFootballMatch *football.MatchInfo
	var lastFootballMatch *football.LastMatchInfo
	var standings []football.StandingsEntry
	if h.footballClient != nil {
		nextFootballMatch, err = h.footballClient.NextMatch(ctx)
		if err != nil {
			slog.Warn("home: failed to load next football match", "err", err)
		}
		lastFootballMatch, err = h.footballClient.LastMatch(ctx)
		if err != nil {
			slog.Warn("home: failed to load last football match", "err", err)
		}
		standings, err = h.footballClient.Standings(ctx)
		if err != nil {
			slog.Warn("home: failed to load standings", "err", err)
		}
	}

	data := HomeData{
		User:              auth.UserFromContext(c),
		CSRFToken:         appMiddleware.CSRFToken(c),
		News:              newsList,
		NextMatch:         nextMatch,
		NextFootballMatch: nextFootballMatch,
		LastFootballMatch: lastFootballMatch,
		Topics:            topics,
		Standings:         standings,
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		r := c.Echo().Renderer.(*tmpl.Renderer)
		return r.RenderPartial(c.Response(), "templates/home/index.html", "content", data)
	}
	return c.Render(http.StatusOK, "templates/home/index.html", data)
}
