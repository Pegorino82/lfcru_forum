//go:build integration

package admin_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pegorino82/lfcru_forum/internal/admin"
	"github.com/labstack/echo/v4"
)

// makeTestJPEG returns a minimal valid JPEG of the given dimensions.
func makeTestJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 16, B: 46, A: 255})
		}
	}
	buf := &bytes.Buffer{}
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 80}); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// buildMultipartRequest creates a POST multipart request for image upload.
// csrfToken is added as a form field and cookie.
func buildMultipartUploadRequest(t *testing.T, path, fieldName, filename, sessID, csrfToken string, data []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if csrfToken != "" {
		writer.WriteField("_csrf", csrfToken)
	}
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write part: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var cookies []string
	if sessID != "" {
		cookies = append(cookies, "session_id="+sessID)
	}
	if csrfToken != "" {
		cookies = append(cookies, "_csrf="+csrfToken)
	}
	if len(cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(cookies, "; "))
	}
	return req
}

// buildDeleteRequest creates a DELETE request with CSRF via X-CSRF-Token header.
func buildDeleteRequest(path, sessID, csrfToken string) *http.Request {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	var cookies []string
	if sessID != "" {
		cookies = append(cookies, "session_id="+sessID)
	}
	if csrfToken != "" {
		cookies = append(cookies, "_csrf="+csrfToken)
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	if len(cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(cookies, "; "))
	}
	return req
}

// newImagesServer builds an Echo server with image upload/delete routes.
func newImagesServer(t *testing.T, uploadsDir string) (*echo.Echo, *admin.ImagesRepo) {
	t.Helper()
	pool := testDB(t)
	e := newTestServer(t, pool)

	imagesRepo := admin.NewImagesRepo(pool)
	imgSvc := admin.NewImageService(uploadsDir)
	imagesHandler := admin.NewImagesHandler(imagesRepo, imgSvc)

	adminGroup := e.Group("", admin.RequireAdminOrMod(e.Renderer))
	adminGroup.POST("/admin/articles/:id/images", imagesHandler.Upload)
	adminGroup.DELETE("/admin/articles/:id/images/:image_id", imagesHandler.Delete)

	return e, imagesRepo
}

// insertTestNewsArticle inserts a minimal news article for image tests.
func insertTestNewsArticle(t *testing.T, authorID int64) int64 {
	t.Helper()
	ctx := context.Background()
	pool := testDB(t)
	var id int64
	err := pool.QueryRow(ctx, `
		INSERT INTO news (title, content, status, author_id)
		VALUES ('img-test article', 'body', 'draft', $1)
		RETURNING id
	`, authorID).Scan(&id)
	if err != nil {
		t.Fatalf("insert test article: %v", err)
	}
	return id
}

// cleanImageTestData removes image test fixtures.
func cleanImageTestData(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	pool := testDB(t)
	pool.Exec(ctx, `DELETE FROM article_images WHERE article_id IN (SELECT id FROM news WHERE title = 'img-test article')`)
	pool.Exec(ctx, `DELETE FROM news WHERE title = 'img-test article'`)
}

// getCsrfFromRecorder extracts the _csrf cookie value from a recorder.
func getCsrfFromRecorder(rec *httptest.ResponseRecorder) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == "_csrf" {
			return c.Value
		}
	}
	return ""
}

// SC-01 / EC-01: Upload JPEG → 200, file saved on disk, DB record created.
func TestImageUpload_JPEG(t *testing.T) {
	pool := testDB(t)
	cleanAdminData(t, pool)
	cleanImageTestData(t)

	userID, sessID := createUserWithRole(t, pool, "admintest-imgadmin@test.com", "admintest_imgadmin", "admin")
	articleID := insertTestNewsArticle(t, userID)

	uploadsDir := t.TempDir()
	e, imagesRepo := newImagesServer(t, uploadsDir)

	// GET /admin to obtain CSRF token
	csrfRec := doGet(t, e, "/admin", sessID)
	csrfToken := getCsrfFromRecorder(csrfRec)

	path := fmt.Sprintf("/admin/articles/%d/images", articleID)
	req := buildMultipartUploadRequest(t, path, "image", "photo.jpg", sessID, csrfToken, makeTestJPEG(100, 100))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	ctx := context.Background()
	images, err := imagesRepo.ListByArticleID(ctx, articleID)
	if err != nil {
		t.Fatalf("list images: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("want 1 image in DB, got %d", len(images))
	}
	if images[0].OriginalFilename != "photo.jpg" {
		t.Fatalf("want original filename 'photo.jpg', got %q", images[0].OriginalFilename)
	}

	path2 := filepath.Join(uploadsDir, images[0].Filename)
	if _, statErr := os.Stat(path2); statErr != nil {
		t.Fatalf("file not found on disk at %s: %v", path2, statErr)
	}
}

// EC-03: Delete image → file deleted from disk, DB record deleted.
func TestImageDelete(t *testing.T) {
	pool := testDB(t)
	cleanAdminData(t, pool)
	cleanImageTestData(t)

	userID, sessID := createUserWithRole(t, pool, "admintest-imgdel@test.com", "admintest_imgdel", "admin")
	articleID := insertTestNewsArticle(t, userID)

	uploadsDir := t.TempDir()
	e, imagesRepo := newImagesServer(t, uploadsDir)

	// Obtain CSRF token
	csrfRec := doGet(t, e, "/admin", sessID)
	csrfToken := getCsrfFromRecorder(csrfRec)

	// Upload
	uploadPath := fmt.Sprintf("/admin/articles/%d/images", articleID)
	req := buildMultipartUploadRequest(t, uploadPath, "image", "del.jpg", sessID, csrfToken, makeTestJPEG(100, 100))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload: want 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	ctx := context.Background()
	images, _ := imagesRepo.ListByArticleID(ctx, articleID)
	if len(images) != 1 {
		t.Fatalf("want 1 image after upload, got %d", len(images))
	}
	savedFilename := images[0].Filename
	imageID := images[0].ID

	// Delete
	delPath := fmt.Sprintf("/admin/articles/%d/images/%d", articleID, imageID)
	delReq := buildDeleteRequest(delPath, sessID, csrfToken)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("delete: want 200, got %d; body: %s", delRec.Code, delRec.Body.String())
	}

	remaining, _ := imagesRepo.ListByArticleID(ctx, articleID)
	if len(remaining) != 0 {
		t.Fatalf("want 0 images after delete, got %d", len(remaining))
	}

	diskPath := filepath.Join(uploadsDir, savedFilename)
	if _, statErr := os.Stat(diskPath); !os.IsNotExist(statErr) {
		t.Fatal("file still exists on disk after delete")
	}
}

// EC-04: File too large → 400.
func TestImageUpload_TooLarge(t *testing.T) {
	pool := testDB(t)
	cleanAdminData(t, pool)

	_, sessID := createUserWithRole(t, pool, "admintest-imglarge@test.com", "admintest_imglarge", "admin")

	uploadsDir := t.TempDir()
	e, _ := newImagesServer(t, uploadsDir)

	csrfRec := doGet(t, e, "/admin", sessID)
	csrfToken := getCsrfFromRecorder(csrfRec)

	// Build multipart with > 10MB of raw bytes (will trigger size check in service)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("_csrf", csrfToken)
	part, _ := writer.CreateFormFile("image", "big.jpg")
	part.Write(make([]byte, (10<<20)+1))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/admin/articles/1/images", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "session_id="+sessID+"; _csrf="+csrfToken)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

// EC-05: Unsupported format (GIF bytes) → 400.
func TestImageUpload_UnsupportedFormat(t *testing.T) {
	pool := testDB(t)
	cleanAdminData(t, pool)

	userID, sessID := createUserWithRole(t, pool, "admintest-imgfmt@test.com", "admintest_imgfmt", "admin")
	articleID := insertTestNewsArticle(t, userID)

	uploadsDir := t.TempDir()
	e, _ := newImagesServer(t, uploadsDir)

	csrfRec := doGet(t, e, "/admin", sessID)
	csrfToken := getCsrfFromRecorder(csrfRec)

	gifData := []byte("GIF89a\x01\x00\x01\x00\x00\xff\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x00;")
	path := fmt.Sprintf("/admin/articles/%d/images", articleID)
	req := buildMultipartUploadRequest(t, path, "image", "anim.gif", sessID, csrfToken, gifData)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

