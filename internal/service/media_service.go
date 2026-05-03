package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"strings"

	_ "golang.org/x/image/webp"

	blurhash "github.com/buckket/go-blurhash"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/rekurt/relax-hub/internal/storage"
)

type UploadMediaInput struct {
	OwnerType    domain.MediaOwnerType
	OwnerID      uuid.UUID
	Data         io.Reader
	OriginalName string
	Size         int64
	MimeType     string
}

type MediaService interface {
	Upload(ctx context.Context, userID uuid.UUID, input UploadMediaInput) (*domain.Media, error)
	Delete(ctx context.Context, mediaID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) error
	ListByReview(ctx context.Context, reviewID uuid.UUID) ([]domain.Media, error)
	ListByReviewIDs(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID][]domain.Media, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error)
}

type mediaService struct {
	mediaRepo repository.MediaRepository
	storage   storage.FileStorage
	log       *logger.Logger
}

func NewMediaService(
	mediaRepo repository.MediaRepository,
	fileStorage storage.FileStorage,
	log *logger.Logger,
) MediaService {
	return &mediaService{
		mediaRepo: mediaRepo,
		storage:   fileStorage,
		log:       log,
	}
}

func (s *mediaService) Upload(ctx context.Context, userID uuid.UUID, input UploadMediaInput) (*domain.Media, error) {
	mediaType := detectMediaType(input.MimeType)

	media := &domain.Media{
		ID:           uuid.New(),
		OwnerType:    input.OwnerType,
		OwnerID:      input.OwnerID,
		UserID:       userID,
		Type:         mediaType,
		OriginalName: input.OriginalName,
		Size:         input.Size,
		MimeType:     input.MimeType,
		Status:       domain.MediaStatusApproved,
	}

	if err := media.ValidateMimeType(); err != nil {
		return nil, err
	}
	if err := media.ValidateFileSize(); err != nil {
		return nil, err
	}

	// Check upload limits
	if err := s.checkLimits(ctx, input.OwnerType, input.OwnerID, mediaType); err != nil {
		return nil, err
	}

	baseID := media.ID.String()

	if mediaType == domain.MediaTypeImage {
		result, err := s.processImage(ctx, input.Data, baseID)
		if err != nil {
			return nil, fmt.Errorf("process image: %w", err)
		}
		media.URL = result.fullURL
		media.ThumbnailURL = result.thumbURL
		media.MediumURL = result.mediumURL
		media.LargeURL = result.largeURL
		media.BlurHash = result.blurHash
		media.Width = result.width
		media.Height = result.height
		media.MimeType = "image/jpeg"
	} else {
		// Video: upload as-is, no processing
		filename := fmt.Sprintf("media/%s/original", baseID)
		videoURL, err := s.storage.Upload(ctx, filename, input.Data, input.MimeType)
		if err != nil {
			return nil, fmt.Errorf("upload video: %w", err)
		}
		media.URL = videoURL
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		// Clean up uploaded files on DB failure
		s.cleanupMediaFiles(ctx, media)
		return nil, err
	}

	return media, nil
}

func (s *mediaService) Delete(ctx context.Context, mediaID uuid.UUID, userID uuid.UUID, userRole domain.UserRole) error {
	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return err
	}

	// Only the author or admin can delete
	if media.UserID != userID && userRole != domain.RoleAdmin {
		return domain.ErrForbidden
	}

	if err := s.mediaRepo.Delete(ctx, mediaID); err != nil {
		return err
	}

	// Clean up storage after successful DB delete
	s.cleanupMediaFiles(ctx, media)

	return nil
}

func (s *mediaService) ListByReview(ctx context.Context, reviewID uuid.UUID) ([]domain.Media, error) {
	result, err := s.mediaRepo.ListByOwner(ctx, domain.MediaOwnerReview, reviewID, 1, domain.MaxImagesPerReview+domain.MaxVideosPerReview)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *mediaService) ListByReviewIDs(ctx context.Context, reviewIDs []uuid.UUID) (map[uuid.UUID][]domain.Media, error) {
	return s.mediaRepo.ListByOwnerIDs(ctx, domain.MediaOwnerReview, reviewIDs)
}

func (s *mediaService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error) {
	return s.mediaRepo.ListByBathhouseReviews(ctx, bathhouseID, page, pageSize)
}

func (s *mediaService) checkLimits(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, mediaType domain.MediaType) error {
	if ownerType != domain.MediaOwnerReview {
		return nil
	}

	switch mediaType {
	case domain.MediaTypeImage:
		imageType := domain.MediaTypeImage
		count, err := s.mediaRepo.CountByOwner(ctx, ownerType, ownerID, &imageType)
		if err != nil {
			return err
		}
		if count >= int64(domain.MaxImagesPerReview) {
			return domain.ErrMediaLimitReached
		}
	case domain.MediaTypeVideo:
		videoType := domain.MediaTypeVideo
		count, err := s.mediaRepo.CountByOwner(ctx, ownerType, ownerID, &videoType)
		if err != nil {
			return err
		}
		if count >= int64(domain.MaxVideosPerReview) {
			return domain.ErrMediaLimitReached
		}
	}

	return nil
}

type imageProcessResult struct {
	fullURL   string
	thumbURL  string
	mediumURL string
	largeURL  string
	blurHash  string
	width     int
	height    int
}

func (s *mediaService) processImage(ctx context.Context, data io.Reader, baseID string) (*imageProcessResult, error) {
	src, _, err := image.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := src.Bounds()
	origW := bounds.Dx()

	// Resize to full size if wider than MaxImageWidth
	fullImg := src
	if origW > domain.MaxImageWidth {
		fullImg = imaging.Resize(src, domain.MaxImageWidth, 0, imaging.Lanczos)
	}
	fullBounds := fullImg.Bounds()

	result := &imageProcessResult{
		width:  fullBounds.Dx(),
		height: fullBounds.Dy(),
	}

	// Generate 4 sizes: full (1920), large (1200), medium (800), thumbnail (300x300 crop)
	sizes := []struct {
		name    string
		img     image.Image
		quality int
	}{
		{"full", fullImg, 85},
		{"large", s.resizeIfWider(src, domain.LargeImageSize), 85},
		{"medium", s.resizeIfWider(src, domain.MediumImageSize), 80},
		{"thumb", imaging.Fill(src, domain.ThumbnailSize, domain.ThumbnailSize, imaging.Center, imaging.Lanczos), 80},
	}

	for _, sz := range sizes {
		// Upload JPEG
		jpegURL, err := s.uploadJPEG(ctx, sz.img, fmt.Sprintf("media/%s/%s.jpg", baseID, sz.name), sz.quality)
		if err != nil {
			return nil, fmt.Errorf("upload %s jpeg: %w", sz.name, err)
		}

		// Upload WebP alongside
		s.uploadWebP(ctx, sz.img, fmt.Sprintf("media/%s/%s.webp", baseID, sz.name), sz.quality)

		switch sz.name {
		case "full":
			result.fullURL = jpegURL
		case "large":
			result.largeURL = jpegURL
		case "medium":
			result.mediumURL = jpegURL
		case "thumb":
			result.thumbURL = jpegURL
		}
	}

	// Generate blur-hash from thumbnail (small image = fast computation)
	result.blurHash = s.generateBlurHash(sizes[3].img)

	return result, nil
}

func (s *mediaService) resizeIfWider(src image.Image, maxWidth int) image.Image {
	if src.Bounds().Dx() > maxWidth {
		return imaging.Resize(src, maxWidth, 0, imaging.Lanczos)
	}
	return src
}

func (s *mediaService) uploadJPEG(ctx context.Context, img image.Image, filename string, quality int) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return "", fmt.Errorf("encode jpeg: %w", err)
	}
	return s.storage.Upload(ctx, filename, &buf, "image/jpeg")
}

func (s *mediaService) uploadWebP(ctx context.Context, img image.Image, filename string, quality int) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: float32(quality)}); err != nil {
		s.log.Warn("failed to encode webp", "filename", filename, "error", err)
		return
	}
	if _, err := s.storage.Upload(ctx, filename, &buf, "image/webp"); err != nil {
		s.log.Warn("failed to upload webp", "filename", filename, "error", err)
	}
}

func (s *mediaService) generateBlurHash(img image.Image) string {
	hash, err := blurhash.Encode(4, 3, img)
	if err != nil {
		s.log.Warn("failed to generate blur hash", "error", err)
		return ""
	}
	return hash
}

func (s *mediaService) cleanupMediaFiles(ctx context.Context, media *domain.Media) {
	for _, u := range []string{media.URL, media.ThumbnailURL, media.MediumURL, media.LargeURL} {
		if u != "" {
			s.cleanupStorageFile(ctx, u)
			// Also try to clean up corresponding WebP variant
			if strings.HasSuffix(u, ".jpg") {
				s.cleanupStorageFile(ctx, strings.TrimSuffix(u, ".jpg")+".webp")
			}
		}
	}
}

func (s *mediaService) cleanupStorageFile(ctx context.Context, fileURL string) {
	key := extractStorageKey(fileURL)
	if err := s.storage.Delete(ctx, key); err != nil {
		s.log.Warn("failed to delete media from storage", "key", key, "error", err)
	}
}

func extractStorageKey(fileURL string) string {
	u, err := url.Parse(fileURL)
	if err != nil {
		return fileURL
	}
	parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return fileURL
}

func detectMediaType(mimeType string) domain.MediaType {
	if domain.AllowedVideoMimeTypes[mimeType] {
		return domain.MediaTypeVideo
	}
	return domain.MediaTypeImage
}
