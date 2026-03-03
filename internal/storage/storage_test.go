package storage

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestMockStorage_Upload(t *testing.T) {
	s := NewMockStorage()
	ctx := context.Background()

	data := bytes.NewReader([]byte("file content"))
	url, err := s.Upload(ctx, "test/file.jpg", data, "image/jpeg")
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if url != "http://mock-storage/test/file.jpg" {
		t.Errorf("Upload() url = %q, want %q", url, "http://mock-storage/test/file.jpg")
	}
	if !s.Has("test/file.jpg") {
		t.Error("file should exist after upload")
	}
	if s.Len() != 1 {
		t.Errorf("Len() = %d, want 1", s.Len())
	}
}

func TestMockStorage_Delete(t *testing.T) {
	s := NewMockStorage()
	ctx := context.Background()

	data := bytes.NewReader([]byte("file content"))
	_, _ = s.Upload(ctx, "test/file.jpg", data, "image/jpeg")

	err := s.Delete(ctx, "test/file.jpg")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if s.Has("test/file.jpg") {
		t.Error("file should not exist after delete")
	}
}

func TestMockStorage_Get(t *testing.T) {
	s := NewMockStorage()
	ctx := context.Background()

	content := []byte("hello world")
	_, _ = s.Upload(ctx, "file.txt", bytes.NewReader(content), "text/plain")

	got, ok := s.Get("file.txt")
	if !ok {
		t.Fatal("Get() returned not found")
	}
	if !bytes.Equal(got, content) {
		t.Errorf("Get() = %q, want %q", got, content)
	}

	_, ok = s.Get("nonexistent.txt")
	if ok {
		t.Error("Get() should return false for nonexistent file")
	}
}

func TestMockStorage_ImplementsInterface(t *testing.T) {
	var _ FileStorage = (*MockStorage)(nil)
}

func TestAvatarPath(t *testing.T) {
	tests := []struct {
		userID string
		suffix string
		ext    string
		want   string
	}{
		{"abc-123", "full", ".jpg", "avatars/abc-123/full.jpg"},
		{"abc-123", "thumb", ".jpg", "avatars/abc-123/thumb.jpg"},
	}

	for _, tt := range tests {
		got := AvatarPath(tt.userID, tt.suffix, tt.ext)
		if got != tt.want {
			t.Errorf("AvatarPath(%q, %q, %q) = %q, want %q", tt.userID, tt.suffix, tt.ext, got, tt.want)
		}
	}
}

func createTestJPEG(t *testing.T, width, height int) *bytes.Buffer {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func createTestPNG(t *testing.T, width, height int) *bytes.Buffer {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func TestResizeAvatar_JPEG(t *testing.T) {
	src := createTestJPEG(t, 400, 400)

	results, err := ResizeAvatar(src)
	if err != nil {
		t.Fatalf("ResizeAvatar() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("ResizeAvatar() returned %d results, want 2", len(results))
	}

	for _, r := range results {
		img, err := jpeg.Decode(bytes.NewReader(r.Data.Bytes()))
		if err != nil {
			t.Fatalf("failed to decode %s result: %v", r.Size.Suffix, err)
		}
		bounds := img.Bounds()
		if bounds.Dx() != r.Size.Width || bounds.Dy() != r.Size.Height {
			t.Errorf("%s: got %dx%d, want %dx%d", r.Size.Suffix, bounds.Dx(), bounds.Dy(), r.Size.Width, r.Size.Height)
		}
	}

	// Verify sizes
	if results[0].Size.Suffix != "full" {
		t.Errorf("first result suffix = %q, want %q", results[0].Size.Suffix, "full")
	}
	if results[1].Size.Suffix != "thumb" {
		t.Errorf("second result suffix = %q, want %q", results[1].Size.Suffix, "thumb")
	}
}

func TestResizeAvatar_PNG(t *testing.T) {
	src := createTestPNG(t, 500, 300)

	results, err := ResizeAvatar(src)
	if err != nil {
		t.Fatalf("ResizeAvatar() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("ResizeAvatar() returned %d results, want 2", len(results))
	}

	// All outputs are JPEG regardless of input format
	for _, r := range results {
		img, err := jpeg.Decode(bytes.NewReader(r.Data.Bytes()))
		if err != nil {
			t.Fatalf("failed to decode %s result as JPEG: %v", r.Size.Suffix, err)
		}
		bounds := img.Bounds()
		if bounds.Dx() != r.Size.Width || bounds.Dy() != r.Size.Height {
			t.Errorf("%s: got %dx%d, want %dx%d", r.Size.Suffix, bounds.Dx(), bounds.Dy(), r.Size.Width, r.Size.Height)
		}
	}
}

func TestResizeAvatar_NonSquareInput(t *testing.T) {
	src := createTestJPEG(t, 800, 200)

	results, err := ResizeAvatar(src)
	if err != nil {
		t.Fatalf("ResizeAvatar() error = %v", err)
	}

	// imaging.Fill should crop to exact dimensions
	for _, r := range results {
		img, _ := jpeg.Decode(bytes.NewReader(r.Data.Bytes()))
		bounds := img.Bounds()
		if bounds.Dx() != r.Size.Width || bounds.Dy() != r.Size.Height {
			t.Errorf("%s: got %dx%d, want %dx%d", r.Size.Suffix, bounds.Dx(), bounds.Dy(), r.Size.Width, r.Size.Height)
		}
	}
}

func TestResizeAvatar_InvalidInput(t *testing.T) {
	_, err := ResizeAvatar(bytes.NewReader([]byte("not an image")))
	if err == nil {
		t.Error("ResizeAvatar() expected error for invalid input, got nil")
	}
}

func TestResizeAvatar_SmallInput(t *testing.T) {
	src := createTestJPEG(t, 30, 30)

	results, err := ResizeAvatar(src)
	if err != nil {
		t.Fatalf("ResizeAvatar() error = %v", err)
	}

	// Even images smaller than target should be resized (upscaled)
	for _, r := range results {
		img, _ := jpeg.Decode(bytes.NewReader(r.Data.Bytes()))
		bounds := img.Bounds()
		if bounds.Dx() != r.Size.Width || bounds.Dy() != r.Size.Height {
			t.Errorf("%s: got %dx%d, want %dx%d", r.Size.Suffix, bounds.Dx(), bounds.Dy(), r.Size.Width, r.Size.Height)
		}
	}
}
