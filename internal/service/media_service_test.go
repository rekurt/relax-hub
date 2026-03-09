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
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/nikitaaldaev/bani/internal/service"
	"github.com/nikitaaldaev/bani/internal/storage"
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
