package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const (
	referralCodeLength = 8
	defaultBonusAmount = 50000 // 500 рублей в копейках
)

type ReferralService interface {
	GenerateCode(ctx context.Context, userID uuid.UUID) (string, error)
	RegisterReferral(ctx context.Context, referralCode string, newUserID uuid.UUID) error
	CompleteReferral(ctx context.Context, refereeID uuid.UUID) error
	GetBalance(ctx context.Context, userID uuid.UUID) (*domain.ReferralBalance, error)
	UseBalance(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error
	RefundBalance(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error
	GetStats(ctx context.Context, userID uuid.UUID) (*domain.ReferralStats, error)
}

type referralService struct {
	referralRepo repository.ReferralRepository
	userRepo     repository.UserRepository
	logger       *logger.Logger
}

func NewReferralService(
	referralRepo repository.ReferralRepository,
	userRepo repository.UserRepository,
	log *logger.Logger,
) ReferralService {
	return &referralService{
		referralRepo: referralRepo,
		userRepo:     userRepo,
		logger:       log,
	}
}

func (s *referralService) GenerateCode(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if user.ReferralCode != "" {
		return user.ReferralCode, nil
	}

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		code, err := generateReferralCode()
		if err != nil {
			return "", err
		}

		if err := s.userRepo.UpdateReferralCode(ctx, userID, code); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) && i < maxRetries-1 {
				continue // retry with a new code
			}
			return "", err
		}

		s.logger.Info("generated referral code", "user_id", userID)
		return code, nil
	}

	return "", domain.ErrAlreadyExists
}

func (s *referralService) RegisterReferral(ctx context.Context, referralCode string, newUserID uuid.UUID) error {
	if referralCode == "" {
		return domain.ErrInvalidInput
	}

	referrer, err := s.userRepo.GetByReferralCode(ctx, referralCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if referrer.ID == newUserID {
		return domain.ErrSelfReferral
	}

	referral := &domain.Referral{
		ID:           uuid.New(),
		ReferrerID:   referrer.ID,
		RefereeID:    newUserID,
		ReferralCode: referralCode,
		Status:       domain.ReferralStatusPending,
		BonusAmount:  defaultBonusAmount,
		CreatedAt:    time.Now(),
	}

	if err := s.referralRepo.Create(ctx, referral); err != nil {
		return err
	}

	s.logger.Info("referral registered", "referrer_id", referrer.ID, "referee_id", newUserID)
	return nil
}

func (s *referralService) CompleteReferral(ctx context.Context, refereeID uuid.UUID) error {
	referral, err := s.referralRepo.GetByReferee(ctx, refereeID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil // Not a referred user, nothing to do
		}
		return err
	}

	if referral.Status != domain.ReferralStatusPending {
		return nil // Already completed or expired
	}

	// Credit bonuses first, update status last.
	// If the process crashes after crediting but before status update,
	// a retry will re-credit (idempotency issue), but this is safer than
	// marking completed before crediting, which would permanently lose bonuses.

	// Credit bonus to referrer
	if err := s.ensureBalance(ctx, referral.ReferrerID); err != nil {
		return err
	}
	if err := s.referralRepo.UpdateBalance(ctx, referral.ReferrerID, referral.BonusAmount, true); err != nil {
		return err
	}

	// Credit bonus to referee
	if err := s.ensureBalance(ctx, refereeID); err != nil {
		return err
	}
	if err := s.referralRepo.UpdateBalance(ctx, refereeID, referral.BonusAmount, true); err != nil {
		return err
	}

	// Mark as completed only after both bonuses are credited
	now := time.Now()
	if err := s.referralRepo.UpdateStatus(ctx, referral.ID, domain.ReferralStatusCompleted, &now); err != nil {
		return err
	}

	s.logger.Info("referral completed, bonuses credited",
		"referrer_id", referral.ReferrerID,
		"referee_id", refereeID,
		"bonus_amount", referral.BonusAmount,
	)
	return nil
}

func (s *referralService) GetBalance(ctx context.Context, userID uuid.UUID) (*domain.ReferralBalance, error) {
	balance, err := s.referralRepo.GetBalance(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &domain.ReferralBalance{
				UserID:      userID,
				Balance:     0,
				TotalEarned: 0,
			}, nil
		}
		return nil, err
	}
	return balance, nil
}

func (s *referralService) UseBalance(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error {
	if amount <= 0 {
		return domain.ErrInvalidInput
	}

	if err := s.referralRepo.UpdateBalance(ctx, userID, -amount, false); err != nil {
		return err
	}

	s.logger.Info("referral balance used", "user_id", userID, "amount", amount, "booking_id", bookingID)
	return nil
}

func (s *referralService) RefundBalance(ctx context.Context, userID uuid.UUID, amount int64, bookingID uuid.UUID) error {
	if amount <= 0 {
		return nil
	}

	if err := s.ensureBalance(ctx, userID); err != nil {
		return err
	}

	if err := s.referralRepo.UpdateBalance(ctx, userID, amount, false); err != nil {
		return err
	}

	s.logger.Info("referral balance refunded", "user_id", userID, "amount", amount, "booking_id", bookingID)
	return nil
}

func (s *referralService) GetStats(ctx context.Context, userID uuid.UUID) (*domain.ReferralStats, error) {
	totalInvited, totalCompleted, err := s.referralRepo.CountByReferrer(ctx, userID)
	if err != nil {
		return nil, err
	}

	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.ReferralStats{
		TotalInvited:   totalInvited,
		TotalCompleted: totalCompleted,
		TotalEarned:    balance.TotalEarned,
	}, nil
}

func (s *referralService) ensureBalance(ctx context.Context, userID uuid.UUID) error {
	_, err := s.referralRepo.GetBalance(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	err = s.referralRepo.CreateBalance(ctx, &domain.ReferralBalance{
		UserID:      userID,
		Balance:     0,
		TotalEarned: 0,
	})
	if errors.Is(err, domain.ErrAlreadyExists) {
		return nil // concurrent call already created it
	}
	return err
}

func generateReferralCode() (string, error) {
	b := make([]byte, referralCodeLength/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
