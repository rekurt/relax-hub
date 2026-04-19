package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/repository"
)

type FavoriteService interface {
	Toggle(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error)
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error)
	IsFavorite(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error)
}

type favoriteService struct {
	favoriteRepo repository.FavoriteRepository
	bhRepo       repository.BathhouseRepository
}

func NewFavoriteService(
	favoriteRepo repository.FavoriteRepository,
	bhRepo repository.BathhouseRepository,
) FavoriteService {
	return &favoriteService{
		favoriteRepo: favoriteRepo,
		bhRepo:       bhRepo,
	}
}

func (s *favoriteService) Toggle(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error) {
	if _, err := s.bhRepo.GetByID(ctx, bathhouseID); err != nil {
		return false, err
	}

	isFav, err := s.favoriteRepo.IsFavorite(ctx, userID, bathhouseID)
	if err != nil {
		return false, err
	}

	if isFav {
		if err := s.favoriteRepo.Remove(ctx, userID, bathhouseID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return false, nil
			}
			return false, err
		}
		return false, nil
	}

	fav := &domain.Favorite{
		ID:          uuid.New(),
		UserID:      userID,
		BathhouseID: bathhouseID,
		CreatedAt:   time.Now(),
	}
	if err := s.favoriteRepo.Add(ctx, fav); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

func (s *favoriteService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error) {
	return s.favoriteRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *favoriteService) IsFavorite(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error) {
	return s.favoriteRepo.IsFavorite(ctx, userID, bathhouseID)
}
