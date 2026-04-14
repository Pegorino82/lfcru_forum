package admin

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // register PNG decoder
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register WebP decoder
)

const (
	maxUploadBytes = 10 << 20 // 10 MB
	maxWidth       = 1200
	jpegQuality    = 85
)

// ErrFileTooLarge is returned when the uploaded file exceeds the size limit.
var ErrFileTooLarge = errors.New("file exceeds 10 MB limit")

// ErrUnsupportedType is returned for non-JPEG/PNG/WebP input.
var ErrUnsupportedType = errors.New("unsupported image format; supported: JPEG, PNG, WebP")

// ImageService handles image normalization and disk storage.
type ImageService struct {
	uploadsDir string
}

// NewImageService creates a new ImageService that stores files under uploadsDir.
func NewImageService(uploadsDir string) *ImageService {
	return &ImageService{uploadsDir: uploadsDir}
}

// Save reads src, validates size and format, resizes if needed, encodes as JPEG,
// saves to uploadsDir/{articleID}/{uuid}.jpg and returns the relative filename.
func (s *ImageService) Save(articleID int64, src io.Reader) (string, error) {
	// Buffer the entire file with a hard cap
	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, io.LimitReader(src, maxUploadBytes+1))
	if err != nil {
		return "", fmt.Errorf("read upload: %w", err)
	}
	if n > maxUploadBytes {
		return "", ErrFileTooLarge
	}

	img, format, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", ErrUnsupportedType
	}
	switch format {
	case "jpeg", "png", "webp":
		// ok
	default:
		return "", ErrUnsupportedType
	}

	img = resizeToMaxWidth(img, maxWidth)

	dir := filepath.Join(s.uploadsDir, fmt.Sprintf("%d", articleID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}

	filename := uuid.New().String() + ".jpg"
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("encode jpeg: %w", err)
	}

	return fmt.Sprintf("%d/%s", articleID, filename), nil
}

// Delete removes the file at uploadsDir/filename. Missing files are silently ignored.
func (s *ImageService) Delete(filename string) error {
	path := filepath.Join(s.uploadsDir, filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// resizeToMaxWidth returns img scaled down so its width equals maxW,
// preserving aspect ratio. Returns img unchanged if it is already narrow enough.
func resizeToMaxWidth(img image.Image, maxW int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= maxW {
		return img
	}
	newH := h * maxW / w
	dst := image.NewRGBA(image.Rect(0, 0, maxW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}
