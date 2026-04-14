package admin

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeJPEG returns a minimal valid JPEG of the given dimensions.
func makeJPEG(w, h int) []byte {
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

// makePNG returns a minimal valid PNG of the given dimensions.
func makePNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// TestResizeToMaxWidth verifies that wide images are scaled down and narrow ones left unchanged.
func TestImageResize(t *testing.T) {
	t.Run("wide image is resized", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 2400, 1200))
		result := resizeToMaxWidth(img, 1200)
		b := result.Bounds()
		if b.Dx() != 1200 {
			t.Fatalf("want width 1200, got %d", b.Dx())
		}
		if b.Dy() != 600 {
			t.Fatalf("want height 600, got %d", b.Dy())
		}
	})

	t.Run("narrow image is unchanged", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 800, 600))
		result := resizeToMaxWidth(img, 1200)
		if result != img {
			t.Fatal("expected original image to be returned unchanged")
		}
	})

	t.Run("exact max width is unchanged", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 1200, 900))
		result := resizeToMaxWidth(img, 1200)
		if result != img {
			t.Fatal("expected original image to be returned unchanged")
		}
	})
}

// TestImageService_Save tests the Save method with various inputs.
func TestImageService_Save(t *testing.T) {
	dir := t.TempDir()
	svc := NewImageService(dir)
	const articleID int64 = 42

	t.Run("saves JPEG and returns relative filename", func(t *testing.T) {
		data := makeJPEG(100, 100)
		filename, err := svc.Save(articleID, bytes.NewReader(data))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(filename, "42/") || !strings.HasSuffix(filename, ".jpg") {
			t.Fatalf("unexpected filename: %q", filename)
		}
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("file not found at %s: %v", path, err)
		}
	})

	t.Run("saves PNG and returns relative filename", func(t *testing.T) {
		data := makePNG(50, 50)
		filename, err := svc.Save(articleID, bytes.NewReader(data))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasSuffix(filename, ".jpg") {
			t.Fatalf("expected .jpg output, got %q", filename)
		}
	})

	t.Run("wide image is resized to max width", func(t *testing.T) {
		data := makeJPEG(2400, 1200)
		filename, err := svc.Save(articleID, bytes.NewReader(data))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		path := filepath.Join(dir, filename)
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open saved file: %v", err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Fatalf("decode saved file: %v", err)
		}
		if img.Bounds().Dx() > maxWidth {
			t.Fatalf("saved image width %d exceeds max %d", img.Bounds().Dx(), maxWidth)
		}
	})

	t.Run("unsupported format returns error", func(t *testing.T) {
		// GIF magic bytes
		gifData := []byte("GIF89a\x01\x00\x01\x00\x00\xff\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x00;")
		_, err := svc.Save(articleID, bytes.NewReader(gifData))
		if err == nil {
			t.Fatal("expected error for GIF, got nil")
		}
	})

	t.Run("file too large returns ErrFileTooLarge", func(t *testing.T) {
		// Create a reader that reports > 10MB
		big := bytes.NewReader(make([]byte, maxUploadBytes+1))
		_, err := svc.Save(articleID, big)
		if err != ErrFileTooLarge {
			t.Fatalf("want ErrFileTooLarge, got %v", err)
		}
	})

	t.Run("garbage data returns ErrUnsupportedType", func(t *testing.T) {
		_, err := svc.Save(articleID, bytes.NewReader([]byte("not an image")))
		if err != ErrUnsupportedType {
			t.Fatalf("want ErrUnsupportedType, got %v", err)
		}
	})
}

// TestImageService_Delete tests that Delete removes the file and ignores missing files.
func TestImageService_Delete(t *testing.T) {
	dir := t.TempDir()
	svc := NewImageService(dir)

	t.Run("deletes existing file", func(t *testing.T) {
		path := filepath.Join(dir, "42", "test.jpg")
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("data"), 0644)

		if err := svc.Delete("42/test.jpg"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("file still exists after delete")
		}
	})

	t.Run("missing file is silently ignored", func(t *testing.T) {
		if err := svc.Delete("42/nonexistent.jpg"); err != nil {
			t.Fatalf("unexpected error for missing file: %v", err)
		}
	})
}
