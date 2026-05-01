package admin

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/labstack/echo/v4"
)

// ImageUploadData is the template data for an uploaded image partial.
type ImageUploadData struct {
	Image     ArticleImage
	ArticleID int64
	CSRFToken string
}

// ImagesHandler handles image upload and deletion for articles.
type ImagesHandler struct {
	imagesRepo *ImagesRepo
	imgSvc     *ImageService
}

// NewImagesHandler creates a new ImagesHandler.
func NewImagesHandler(imagesRepo *ImagesRepo, imgSvc *ImageService) *ImagesHandler {
	return &ImagesHandler{imagesRepo: imagesRepo, imgSvc: imgSvc}
}

// Upload handles POST /admin/articles/:id/images.
// Accepts multipart/form-data with field "image".
// Returns a partial HTML snippet for HTMX swap.
func (h *ImagesHandler) Upload(c echo.Context) error {
	ctx := c.Request().Context()

	articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || articleID <= 0 {
		return c.String(http.StatusBadRequest, "Некорректный ID статьи")
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.String(http.StatusBadRequest, "Файл не найден в запросе")
	}

	if fileHeader.Size > maxUploadBytes {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Файл слишком большой (максимум %d MB)", maxUploadBytes>>20))
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Не удалось открыть файл")
	}
	defer src.Close()

	filename, err := h.imgSvc.Save(articleID, src)
	if err != nil {
		switch {
		case errors.Is(err, ErrFileTooLarge):
			return c.String(http.StatusBadRequest, fmt.Sprintf("Файл слишком большой (максимум %d MB)", maxUploadBytes>>20))
		case errors.Is(err, ErrUnsupportedType):
			return c.String(http.StatusBadRequest, "Неподдерживаемый формат. Допустимы: JPEG, PNG, WebP")
		default:
			slog.Error("image upload: save", "article_id", articleID, "err", err)
			return c.String(http.StatusInternalServerError, "Ошибка сохранения файла")
		}
	}

	// If the upload is from Tiptap, return JSON with the URL and skip DB insert.
	if c.QueryParam("from") == "tiptap" {
		return c.JSON(http.StatusOK, echo.Map{
			"url": "/storage/news/" + filename,
		})
	}

	img := &ArticleImage{
		ArticleID:        articleID,
		Filename:         filename,
		OriginalFilename: fileHeader.Filename,
	}
	if err := h.imagesRepo.Create(ctx, img); err != nil {
		// Best-effort cleanup of the saved file
		h.imgSvc.Delete(filename) //nolint:errcheck
		slog.Error("image upload: db insert", "article_id", articleID, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка сохранения записи")
	}

	data := ImageUploadData{Image: *img, ArticleID: articleID, CSRFToken: appMiddleware.CSRFToken(c)}
	return c.Render(http.StatusOK, "templates/admin/articles/image_item.html#image-item", data)
}

// Delete handles DELETE /admin/articles/:id/images/:image_id.
func (h *ImagesHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || articleID <= 0 {
		return c.String(http.StatusBadRequest, "Некорректный ID статьи")
	}
	imageID, err := strconv.ParseInt(c.Param("image_id"), 10, 64)
	if err != nil || imageID <= 0 {
		return c.String(http.StatusBadRequest, "Некорректный ID изображения")
	}

	img, err := h.imagesRepo.GetByID(ctx, imageID)
	if err != nil {
		slog.Error("image delete: get by id", "image_id", imageID, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка запроса к базе данных")
	}
	if img == nil || img.ArticleID != articleID {
		return c.String(http.StatusNotFound, "Изображение не найдено")
	}

	if err := h.imgSvc.Delete(img.Filename); err != nil {
		slog.Error("image delete: remove file", "filename", img.Filename, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка удаления файла")
	}

	if err := h.imagesRepo.Delete(ctx, imageID); err != nil {
		slog.Error("image delete: db delete", "image_id", imageID, "err", err)
		return c.String(http.StatusInternalServerError, "Ошибка удаления записи")
	}

	return c.NoContent(http.StatusOK)
}
