package service_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/rekurt/relax-hub/internal/storage"
)

type mediaTestEnv struct {
	svc       service.MediaService
	mediaRepo *mock.MediaRepo
	store     *storage.MockStorage
}

func newMediaTestEnv() *mediaTestEnv {
	mediaRepo := mock.NewMediaRepo()
	store := storage.NewMockStorage()
	log := logger.New(logger.LevelWarn)
	svc := service.NewMediaService(mediaRepo, store, log)
	return &mediaTestEnv{
		svc:       svc,
		mediaRepo: mediaRepo,
		store:     store,
	}
}

func createTestJPEGData(t *testing.T, width, height int) *bytes.Buffer {
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

func TestMediaService_Upload_Image_Success(t *testing.T) {
	env := newMediaTestEnv()
	userID := uuid.New()
	reviewID := uuid.New()
	imgData := createTestJPEGData(t, 800, 600)

	media, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      reviewID,
		Data:         imgData,
		OriginalName: "photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.UserID != userID {
		t.Errorf("user_id = %v, want %v", media.UserID, userID)
	}
	if media.OwnerType != domain.MediaOwnerReview {
		t.Errorf("owner_type = %v, want review", media.OwnerType)
	}
	if media.Type != domain.MediaTypeImage {
		t.Errorf("type = %v, want image", media.Type)
	}
	if media.URL == "" {
		t.Error("url should not be empty")
	}
	if media.ThumbnailURL == "" {
		t.Error("thumbnail_url should not be empty")
	}
	if media.Width == 0 || media.Height == 0 {
		t.Errorf("dimensions should be set, got %dx%d", media.Width, media.Height)
	}
	if media.Status != domain.MediaStatusApproved {
		t.Errorf("status = %v, want approved", media.Status)
	}
}

func TestMediaService_Upload_Image_MultiSize(t *testing.T) {
	env := newMediaTestEnv()
	imgData := createTestJPEGData(t, 2500, 1800)

	media, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "large-photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All 4 size URLs should be populated
	if media.URL == "" {
		t.Error("full URL should not be empty")
	}
	if media.ThumbnailURL == "" {
		t.Error("thumbnail URL should not be empty")
	}
	if media.MediumURL == "" {
		t.Error("medium URL should not be empty")
	}
	if media.LargeURL == "" {
		t.Error("large URL should not be empty")
	}

	// URLs should be distinct
	urls := map[string]bool{media.URL: true, media.ThumbnailURL: true, media.MediumURL: true, media.LargeURL: true}
	if len(urls) != 4 {
		t.Errorf("expected 4 distinct URLs, got %d", len(urls))
	}

	// Full size should be capped at MaxImageWidth
	if media.Width > domain.MaxImageWidth {
		t.Errorf("width = %d, should be <= %d after resize", media.Width, domain.MaxImageWidth)
	}
}

func TestMediaService_Upload_Image_WebPGenerated(t *testing.T) {
	env := newMediaTestEnv()
	imgData := createTestJPEGData(t, 1000, 800)

	media, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WebP variants should exist in storage alongside JPEG
	// The JPEG URL contains the media ID, derive the WebP path
	baseID := media.ID.String()
	webpFiles := []string{
		"media/" + baseID + "/full.webp",
		"media/" + baseID + "/large.webp",
		"media/" + baseID + "/medium.webp",
		"media/" + baseID + "/thumb.webp",
	}

	for _, f := range webpFiles {
		if !env.store.Has(f) {
			t.Errorf("WebP file %q should exist in storage", f)
		}
	}
}

func TestMediaService_Upload_Image_BlurHash(t *testing.T) {
	env := newMediaTestEnv()
	imgData := createTestJPEGData(t, 600, 400)

	media, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if media.BlurHash == "" {
		t.Error("blur_hash should not be empty")
	}
	// BlurHash strings are typically 20-30+ characters
	if len(media.BlurHash) < 6 {
		t.Errorf("blur_hash too short: %q", media.BlurHash)
	}
}

func TestMediaService_Upload_Image_SmallImageNoUpscale(t *testing.T) {
	env := newMediaTestEnv()
	// Image smaller than all size targets except thumbnail
	imgData := createTestJPEGData(t, 500, 400)

	media, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "small.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Width should not exceed original (no upscaling)
	if media.Width > 500 {
		t.Errorf("width = %d, should not be upscaled beyond 500", media.Width)
	}

	// All URLs should still be set (using original size for larger targets)
	if media.URL == "" || media.MediumURL == "" || media.LargeURL == "" || media.ThumbnailURL == "" {
		t.Error("all size URLs should be set even for small images")
	}
}

func TestMediaService_Upload_Image_ResizeLarge(t *testing.T) {
	env := newMediaTestEnv()
	imgData := createTestJPEGData(t, 3000, 2000)

	media, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "big.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.Width > domain.MaxImageWidth {
		t.Errorf("width = %d, should be <= %d after resize", media.Width, domain.MaxImageWidth)
	}
}

func TestMediaService_Upload_Video_Success(t *testing.T) {
	env := newMediaTestEnv()
	userID := uuid.New()
	reviewID := uuid.New()
	videoData := bytes.NewReader([]byte("fake video content"))

	media, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      reviewID,
		Data:         videoData,
		OriginalName: "video.mp4",
		Size:         1000,
		MimeType:     "video/mp4",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.Type != domain.MediaTypeVideo {
		t.Errorf("type = %v, want video", media.Type)
	}
	if media.URL == "" {
		t.Error("url should not be empty")
	}
	// Videos should not have multi-size URLs or blur hash
	if media.MediumURL != "" || media.LargeURL != "" || media.BlurHash != "" {
		t.Error("videos should not have medium/large/blurhash")
	}
}

func TestMediaService_Upload_InvalidMimeType(t *testing.T) {
	env := newMediaTestEnv()

	_, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         strings.NewReader("data"),
		OriginalName: "file.exe",
		Size:         100,
		MimeType:     "application/octet-stream",
	})

	if err != domain.ErrMediaInvalidType {
		t.Errorf("err = %v, want ErrMediaInvalidType", err)
	}
}

func TestMediaService_Upload_FileTooLarge(t *testing.T) {
	env := newMediaTestEnv()

	_, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         strings.NewReader("data"),
		OriginalName: "huge.jpg",
		Size:         domain.MaxImageSizeBytes + 1,
		MimeType:     "image/jpeg",
	})

	if err != domain.ErrMediaFileTooLarge {
		t.Errorf("err = %v, want ErrMediaFileTooLarge", err)
	}
}

func TestMediaService_Upload_ImageLimitReached(t *testing.T) {
	env := newMediaTestEnv()
	userID := uuid.New()
	reviewID := uuid.New()

	// Upload max images
	for i := 0; i < domain.MaxImagesPerReview; i++ {
		imgData := createTestJPEGData(t, 100, 100)
		_, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
			OwnerType:    domain.MediaOwnerReview,
			OwnerID:      reviewID,
			Data:         imgData,
			OriginalName: "photo.jpg",
			Size:         int64(imgData.Len()),
			MimeType:     "image/jpeg",
		})
		if err != nil {
			t.Fatalf("upload %d: %v", i, err)
		}
	}

	// The next one should fail
	imgData := createTestJPEGData(t, 100, 100)
	_, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      reviewID,
		Data:         imgData,
		OriginalName: "one-too-many.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})

	if err != domain.ErrMediaLimitReached {
		t.Errorf("err = %v, want ErrMediaLimitReached", err)
	}
}

func TestMediaService_Upload_VideoLimitReached(t *testing.T) {
	env := newMediaTestEnv()
	userID := uuid.New()
	reviewID := uuid.New()

	// Upload max videos (1)
	_, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      reviewID,
		Data:         bytes.NewReader([]byte("video")),
		OriginalName: "vid.mp4",
		Size:         100,
		MimeType:     "video/mp4",
	})
	if err != nil {
		t.Fatalf("first video: %v", err)
	}

	// Second video should fail
	_, err = env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      reviewID,
		Data:         bytes.NewReader([]byte("video2")),
		OriginalName: "vid2.mp4",
		Size:         100,
		MimeType:     "video/mp4",
	})
	if err != domain.ErrMediaLimitReached {
		t.Errorf("err = %v, want ErrMediaLimitReached", err)
	}
}

func TestMediaService_Delete_Success(t *testing.T) {
	env := newMediaTestEnv()
	userID := uuid.New()
	imgData := createTestJPEGData(t, 200, 200)

	media, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	err = env.svc.Delete(context.Background(), media.ID, userID, domain.RoleClient)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Verify it's gone from repo
	_, err = env.mediaRepo.GetByID(context.Background(), media.ID)
	if err != domain.ErrMediaNotFound {
		t.Errorf("err = %v, want ErrMediaNotFound after delete", err)
	}
}

func TestMediaService_Delete_Forbidden(t *testing.T) {
	env := newMediaTestEnv()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	imgData := createTestJPEGData(t, 200, 200)

	media, _ := env.svc.Upload(context.Background(), ownerID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})

	err := env.svc.Delete(context.Background(), media.ID, otherUserID, domain.RoleClient)
	if err != domain.ErrForbidden {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestMediaService_Delete_AdminCanDelete(t *testing.T) {
	env := newMediaTestEnv()
	ownerID := uuid.New()
	adminID := uuid.New()
	imgData := createTestJPEGData(t, 200, 200)

	media, _ := env.svc.Upload(context.Background(), ownerID, service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})

	err := env.svc.Delete(context.Background(), media.ID, adminID, domain.RoleAdmin)
	if err != nil {
		t.Fatalf("admin delete: %v", err)
	}
}

func TestMediaService_Delete_NotFound(t *testing.T) {
	env := newMediaTestEnv()

	err := env.svc.Delete(context.Background(), uuid.New(), uuid.New(), domain.RoleAdmin)
	if err != domain.ErrMediaNotFound {
		t.Errorf("err = %v, want ErrMediaNotFound", err)
	}
}

func TestMediaService_ListByReview(t *testing.T) {
	env := newMediaTestEnv()
	userID := uuid.New()
	reviewID := uuid.New()

	for i := 0; i < 3; i++ {
		imgData := createTestJPEGData(t, 100, 100)
		_, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
			OwnerType:    domain.MediaOwnerReview,
			OwnerID:      reviewID,
			Data:         imgData,
			OriginalName: "photo.jpg",
			Size:         int64(imgData.Len()),
			MimeType:     "image/jpeg",
		})
		if err != nil {
			t.Fatalf("upload %d: %v", i, err)
		}
	}

	items, err := env.svc.ListByReview(context.Background(), reviewID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("len = %d, want 3", len(items))
	}
}

func TestMediaService_ListByBathhouse(t *testing.T) {
	env := newMediaTestEnv()
	userID := uuid.New()
	bathhouseID := uuid.New()
	reviewID := uuid.New()

	// Map the review to the bathhouse in mock
	env.mediaRepo.SetReviewBathhouse(reviewID, bathhouseID)

	// Upload 2 images for the review
	for i := 0; i < 2; i++ {
		imgData := createTestJPEGData(t, 100, 100)
		_, err := env.svc.Upload(context.Background(), userID, service.UploadMediaInput{
			OwnerType:    domain.MediaOwnerReview,
			OwnerID:      reviewID,
			Data:         imgData,
			OriginalName: "photo.jpg",
			Size:         int64(imgData.Len()),
			MimeType:     "image/jpeg",
		})
		if err != nil {
			t.Fatalf("upload %d: %v", i, err)
		}
	}

	result, err := env.svc.ListByBathhouse(context.Background(), bathhouseID, 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}

func TestMediaService_Upload_WebP_Accepted(t *testing.T) {
	// Verify that image/webp is accepted as input mime type
	env := newMediaTestEnv()
	// WebP input validation should pass (actual decode may fail with fake data,
	// but the mime type should be accepted)
	imgData := createTestJPEGData(t, 400, 300)

	// Upload as JPEG first, then verify WebP mime type is in allowed list
	if !domain.AllowedImageMimeTypes["image/webp"] {
		t.Error("image/webp should be in AllowedImageMimeTypes")
	}

	// Actual upload with JPEG data but verifying the flow works
	media, err := env.svc.Upload(context.Background(), uuid.New(), service.UploadMediaInput{
		OwnerType:    domain.MediaOwnerReview,
		OwnerID:      uuid.New(),
		Data:         imgData,
		OriginalName: "photo.jpg",
		Size:         int64(imgData.Len()),
		MimeType:     "image/jpeg",
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if media.MediumURL == "" || media.LargeURL == "" {
		t.Error("medium and large URLs should be populated")
	}
}
