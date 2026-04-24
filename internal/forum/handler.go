package forum

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc      *Service
	renderer *tmpl.Renderer
	hub      *Hub
}

func NewHandler(svc *Service, renderer *tmpl.Renderer, hub *Hub) *Handler {
	return &Handler{svc: svc, renderer: renderer, hub: hub}
}

// Index renders GET /forum — list of sections
func (h *Handler) Index(c echo.Context) error {
	ctx := c.Request().Context()

	sections, err := h.svc.ListSections(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	u := auth.UserFromContext(c)
	canManage := u != nil && (u.Role == "moderator" || u.Role == "admin")

	data := map[string]interface{}{
		"User":      u,
		"Sections":  sections,
		"CanManage": canManage,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/index.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/index.html", data)
}

// ShowSection renders GET /forum/sections/:id
func (h *Handler) ShowSection(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	ctx := c.Request().Context()
	section, topics, err := h.svc.GetSectionWithTopics(ctx, id)
	if err != nil {
		if errors.Is(err, ErrSectionNotFound) {
			return c.String(http.StatusNotFound, "Not found")
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	u := auth.UserFromContext(c)
	canManage := u != nil && (u.Role == "moderator" || u.Role == "admin")

	data := map[string]interface{}{
		"User":      u,
		"Section":   section,
		"Topics":    topics,
		"CanManage": canManage,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/section.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/section.html", data)
}

// ShowTopic renders GET /forum/topics/:id
func (h *Handler) ShowTopic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	ctx := c.Request().Context()
	topic, posts, err := h.svc.GetTopicWithPosts(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTopicNotFound) {
			return c.String(http.StatusNotFound, "Not found")
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}
	if topic == nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	// Fetch section for breadcrumbs
	section, err := h.svc.GetSection(ctx, topic.SectionID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	u := auth.UserFromContext(c)
	canReply := u != nil

	data := map[string]interface{}{
		"User":      u,
		"Topic":     topic,
		"Section":   section,
		"Posts":     posts,
		"CanReply":  canReply,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/topic.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/topic.html", data)
}

// NewSection renders GET /forum/sections/new — form
func (h *Handler) NewSection(c echo.Context) error {
	data := map[string]interface{}{
		"User":      auth.UserFromContext(c),
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/new_section.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/new_section.html", data)
}

// CreateSection handles POST /forum/sections
func (h *Handler) CreateSection(c echo.Context) error {
	title := c.FormValue("title")
	description := c.FormValue("description")
	sortOrderStr := c.FormValue("sort_order")

	sortOrder := 0
	if sortOrderStr != "" {
		so, err := strconv.Atoi(sortOrderStr)
		if err == nil {
			sortOrder = so
		}
	}

	ctx := c.Request().Context()
	_, err := h.svc.CreateSection(ctx, title, description, sortOrder)

	if err != nil {
		// Validation errors -> 422
		if errors.Is(err, ErrEmptyTitle) || errors.Is(err, ErrTitleTooLong) || errors.Is(err, ErrDescriptionTooLong) {
			errMsg := mapErrorMessage(err)
			data := map[string]interface{}{
				"User":      auth.UserFromContext(c),
				"FormError": errMsg,
				"FormContent": map[string]string{
					"title":       title,
					"description": description,
				},
				"CSRFToken": appMiddleware.CSRFToken(c),
			}
			return c.Render(http.StatusUnprocessableEntity, "templates/forum/new_section.html", data)
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	// Success -> redirect
	return c.Redirect(http.StatusSeeOther, "/forum")
}

// NewTopic renders GET /forum/sections/:id/topics/new
func (h *Handler) NewTopic(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	ctx := c.Request().Context()
	section, err := h.svc.GetSection(ctx, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}
	if section == nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	data := map[string]interface{}{
		"User":      auth.UserFromContext(c),
		"Section":   section,
		"CSRFToken": appMiddleware.CSRFToken(c),
	}

	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		return c.Render(http.StatusOK, "templates/forum/new_topic.html#content", data)
	}
	return c.Render(http.StatusOK, "templates/forum/new_topic.html", data)
}

// CreateTopic handles POST /forum/sections/:id/topics
func (h *Handler) CreateTopic(c echo.Context) error {
	sectionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	title := c.FormValue("title")

	u := auth.UserFromContext(c)
	ctx := c.Request().Context()

	topicID, err := h.svc.CreateTopic(ctx, sectionID, u.ID, title)

	if err != nil {
		// 404 if section not found
		if errors.Is(err, ErrSectionNotFound) {
			return c.String(http.StatusNotFound, "Not found")
		}
		// Validation errors -> 422
		if errors.Is(err, ErrEmptyTitle) || errors.Is(err, ErrTitleTooLong) {
			errMsg := mapErrorMessage(err)
			section, _ := h.svc.GetSection(ctx, sectionID)
			data := map[string]interface{}{
				"User":      u,
				"FormError": errMsg,
				"FormContent": map[string]string{
					"title": title,
				},
				"Section":   section,
				"CSRFToken": appMiddleware.CSRFToken(c),
			}
			return c.Render(http.StatusUnprocessableEntity, "templates/forum/new_topic.html", data)
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	// Success -> redirect
	return c.Redirect(http.StatusSeeOther, "/forum/topics/"+strconv.FormatInt(topicID, 10))
}

// CreatePost handles POST /forum/topics/:id/posts
func (h *Handler) CreatePost(c echo.Context) error {
	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	content := c.FormValue("content")
	parentIDStr := c.FormValue("parent_id")

	var parentID *int64
	if parentIDStr != "" {
		pID, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err == nil {
			parentID = &pID
		}
	}

	u := auth.UserFromContext(c)
	ctx := c.Request().Context()

	postID, err := h.svc.CreatePost(ctx, topicID, u.ID, parentID, content)

	if err != nil {
		// 404 if topic not found
		if errors.Is(err, ErrTopicNotFound) {
			return c.String(http.StatusNotFound, "Not found")
		}
		// Validation/logic errors -> 422
		if errors.Is(err, ErrEmptyContent) || errors.Is(err, ErrContentTooLong) ||
			errors.Is(err, ErrParentNotFound) || errors.Is(err, ErrReplyToReply) {
			errMsg := mapErrorMessage(err)

			isHTMX := c.Request().Header.Get("HX-Request") == "true"
			if isHTMX {
				// HTMX: return 422 with form + HX-Retarget
				c.Response().Header().Set("HX-Retarget", "#post-form")
				c.Response().Header().Set("HX-Reswap", "innerHTML")
				return c.String(http.StatusUnprocessableEntity, "Error: "+errMsg)
			}

			// Non-HTMX: return full page with error
			topic, posts, _ := h.svc.GetTopicWithPosts(ctx, topicID)
			section, _ := h.svc.GetSection(ctx, topic.SectionID)
			data := map[string]interface{}{
				"User":      u,
				"Topic":     topic,
				"Section":   section,
				"Posts":     posts,
				"FormError": errMsg,
				"FormContent": map[string]string{
					"content":   content,
					"parent_id": parentIDStr,
				},
				"CanReply":  true,
				"CSRFToken": appMiddleware.CSRFToken(c),
			}
			return c.Render(http.StatusUnprocessableEntity, "templates/forum/topic.html", data)
		}
		return c.String(http.StatusInternalServerError, "Internal error")
	}

	// Broadcast to SSE subscribers (best-effort; errors are logged and ignored)
	topic, posts, err := h.svc.GetTopicWithPosts(ctx, topicID)
	if err == nil {
		for _, pv := range posts {
			if pv.ID == postID {
				if fragment, ferr := h.renderPostFragment(pv); ferr == nil {
					h.hub.Broadcast(topicID, u.ID, fragment)
				}
				break
			}
		}
	}

	// Success
	isHTMX := c.Request().Header.Get("HX-Request") == "true"
	if isHTMX {
		// HTMX: return updated posts list
		data := map[string]interface{}{
			"User":      u,
			"Posts":     posts,
			"Topic":     topic,
			"CanReply":  true,
			"CSRFToken": appMiddleware.CSRFToken(c),
		}
		c.Response().Header().Set("HX-Trigger", "postAdded")
		return c.Render(http.StatusCreated, "templates/forum/topic.html#posts-list", data)
	}

	// Non-HTMX: redirect with anchor
	return c.Redirect(http.StatusSeeOther, "/forum/topics/"+strconv.FormatInt(topicID, 10)+"#post-"+strconv.FormatInt(postID, 10))
}

// StreamEvents handles GET /forum/topics/:id/events — SSE endpoint.
// Available to all users including unauthenticated (CON-03).
func (h *Handler) StreamEvents(c echo.Context) error {
	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	ctx := c.Request().Context()

	// Verify topic exists
	topic, err := h.svc.GetTopic(ctx, topicID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal error")
	}
	if topic == nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	// Determine userID (0 for anonymous — never matches authorUserID in broadcast)
	var userID int64
	if u := auth.UserFromContext(c); u != nil {
		userID = u.ID
	}

	ch, err := h.hub.Subscribe(ctx, topicID, userID)
	if err != nil {
		return c.String(http.StatusServiceUnavailable, "Too many subscribers")
	}

	w := c.Response().Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		return c.String(http.StatusInternalServerError, "Streaming not supported")
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)
	flusher.Flush()

	// Catch-up: deliver posts missed since Last-Event-ID (CTR-02)
	if lastIDStr := c.Request().Header.Get("Last-Event-ID"); lastIDStr != "" {
		if lastID, parseErr := strconv.ParseInt(lastIDStr, 10, 64); parseErr == nil && lastID >= 0 {
			missedPosts, dbErr := h.svc.ListPostsAfter(ctx, topicID, lastID)
			if dbErr == nil {
				for _, pv := range missedPosts {
					if fragment, renderErr := h.renderPostFragment(pv); renderErr == nil {
						fmt.Fprintf(w, "id: %d\nevent: post-added\ndata: %s\n\n", pv.ID, fragment)
					}
				}
				if len(missedPosts) == 50 {
					fmt.Fprintf(w, "event: catch-up-overflow\ndata:\n\n")
				}
				flusher.Flush()
			}
		}
	}

	// Stream live events
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			fmt.Fprintf(w, "event: post-added\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			return nil
		}
	}
}

// renderPostFragment renders a PostView to a single-line HTML string for SSE delivery.
func (h *Handler) renderPostFragment(pv PostView) (string, error) {
	var buf bytes.Buffer
	if err := h.renderer.RenderPartial(&buf, "templates/forum/partials/post.html", "post", pv); err != nil {
		return "", err
	}
	return strings.NewReplacer("\n", " ", "\r", " ").Replace(buf.String()), nil
}

func mapErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrEmptyTitle):
		return "Заголовок не может быть пустым"
	case errors.Is(err, ErrTitleTooLong):
		return "Заголовок не может быть длиннее 255 символов"
	case errors.Is(err, ErrDescriptionTooLong):
		return "Описание не может быть длиннее 2000 символов"
	case errors.Is(err, ErrEmptyContent):
		return "Сообщение не может быть пустым"
	case errors.Is(err, ErrContentTooLong):
		return "Сообщение не может быть длиннее 20000 символов"
	case errors.Is(err, ErrParentNotFound):
		return "Цитируемое сообщение не найдено"
	case errors.Is(err, ErrReplyToReply):
		return "Нельзя цитировать ответ на ответ"
	default:
		return "Ошибка при обработке запроса"
	}
}
