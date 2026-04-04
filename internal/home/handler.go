package home

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/forum"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/match"
	"github.com/Pegorino82/lfcru_forum/internal/news"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/labstack/echo/v4"
)

type HomeData struct {
	User      *user.User
	CSRFToken string
	News      []news.News
	NextMatch *match.Match
	Topics    []forum.TopicWithLastAuthor
}

type Handler struct {
	newsRepo  *news.Repo
	matchRepo *match.Repo
	topicRepo *forum.Repo
}

func NewHandler(newsRepo *news.Repo, matchRepo *match.Repo, topicRepo *forum.Repo) *Handler {
	return &Handler{newsRepo: newsRepo, matchRepo: matchRepo, topicRepo: topicRepo}
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

	data := HomeData{
		User:      auth.UserFromContext(c),
		CSRFToken: appMiddleware.CSRFToken(c),
		News:      newsList,
		NextMatch: nextMatch,
		Topics:    topics,
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		r := c.Echo().Renderer.(*tmpl.Renderer)
		return r.RenderPartial(c.Response(), "templates/home/index.html", "content", data)
	}
	return c.Render(http.StatusOK, "templates/home/index.html", data)
}
