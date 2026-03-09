package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/url"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/storage"
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
		Status:       domain.MediaStatusPending,
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
		mainURL, thumbURL, width, height, err := s.processImage(ctx, input.Data, baseID)
		if err != nil {
			return nil, fmt.Errorf("process image: %w", err)
		}
		media.URL = mainURL
		media.ThumbnailURL = thumbURL
		media.Width = width
		media.Height = height
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
		s.cleanupStorageFile(ctx, media.URL)
		if media.ThumbnailURL != "" {
			s.cleanupStorageFile(ctx, media.ThumbnailURL)
		}
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
	s.cleanupStorageFile(ctx, media.URL)
	if media.ThumbnailURL != "" {
		s.cleanupStorageFile(ctx, media.ThumbnailURL)
	}

	return nil
}

func (s *mediaService) ListByReview(ctx context.Context, reviewID uuid.UUID) ([]domain.Media, error) {
	result, err := s.mediaRepo.ListByOwner(ctx, domain.MediaOwnerReview, reviewID, 1, domain.MaxImagesPerReview+domain.MaxVideosPerReview)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *mediaService) ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error) {
	return s.mediaRepo.ListByBathhouseReviews(ctx, bathhouseID, page, pageSize)
}

func (s *mediaService) checkLimits(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, mediaType domain.MediaType) error {
	if ownerType != domain.MediaOwnerReview {
		return nil
	}

	if mediaType == domain.MediaTypeImage {
		imageType := domain.MediaTypeImage
		count, err := s.mediaRepo.CountByOwner(ctx, ownerType, ownerID, &imageType)
		if err != nil {
			return err
		}
		if count >= int64(domain.MaxImagesPerReview) {
			return domain.ErrMediaLimitReached
		}
	} else if mediaType == domain.MediaTypeVideo {
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

func (s *mediaService) processImage(ctx context.Context, data io.Reader, baseID string) (mainURL, thumbURL string, width, height int, err error) {
	src, _, err := image.Decode(data)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("decode image: %w", err)
	}

	bounds := src.Bounds()
	origW := bounds.Dx()

	// Resize if wider than MaxImageWidth, preserving aspect ratio
	resized := src
	if origW > domain.MaxImageWidth {
		resized = imaging.Resize(src, domain.MaxImageWidth, 0, imaging.Lanczos)
	}
	resBounds := resized.Bounds()
	width = resBounds.Dx()
	height = resBounds.Dy()

	// Encode main image as JPEG
	var mainBuf bytes.Buffer
	if err := jpeg.Encode(&mainBuf, resized, &jpeg.Options{Quality: 85}); err != nil {
		return "", "", 0, 0, fmt.Errorf("encode main image: %w", err)
	}

	mainFilename := fmt.Sprintf("media/%s/main.jpg", baseID)
	mainURL, err = s.storage.Upload(ctx, mainFilename, &mainBuf, "image/jpeg")
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("upload main image: %w", err)
	}

	// Create thumbnail
	thumb := imaging.Fill(src, domain.ThumbnailSize, domain.ThumbnailSize, imaging.Center, imaging.Lanczos)
	var thumbBuf bytes.Buffer
	if err := jpeg.Encode(&thumbBuf, thumb, &jpeg.Options{Quality: 80}); err != nil {
		s.log.Warn("failed to create thumbnail", "error", err)
		return mainURL, "", width, height, nil
	}

	thumbFilename := fmt.Sprintf("media/%s/thumb.jpg", baseID)
	thumbURL, err = s.storage.Upload(ctx, thumbFilename, &thumbBuf, "image/jpeg")
	if err != nil {
		s.log.Warn("failed to upload thumbnail", "error", err)
		return mainURL, "", width, height, nil
	}

	return mainURL, thumbURL, width, height, nil
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
