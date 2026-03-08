package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

const (
	certificateCodeChars   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	certificateCodeSegLen  = 4
	certificateCodeRetries = 3
	certificateValidDays   = 365
)

type CertificateService interface {
	Purchase(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error)
	Redeem(ctx context.Context, code string, userID uuid.UUID) (*domain.GiftCertificate, error)
	Apply(ctx context.Context, certificateID, bookingID uuid.UUID, amount int64) error
	GetBalance(ctx context.Context, code string) (*domain.GiftCertificate, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error)
}

type certificateService struct {
	certRepo repository.GiftCertificateRepository
	logger   *logger.Logger
}

func NewCertificateService(
	certRepo repository.GiftCertificateRepository,
	log *logger.Logger,
) CertificateService {
	return &certificateService{
		certRepo: certRepo,
		logger:   log,
	}
}

func (s *certificateService) Purchase(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if purchaserEmail == "" || recipientEmail == "" {
		return nil, domain.ErrInvalidInput
	}

	var cert *domain.GiftCertificate
	for i := 0; i < certificateCodeRetries; i++ {
		code, err := generateCertificateCode()
		if err != nil {
			return nil, fmt.Errorf("generate certificate code: %w", err)
		}

		cert = &domain.GiftCertificate{
			ID:             uuid.New(),
			Code:           code,
			PurchaserID:    purchaserID,
			PurchaserEmail: purchaserEmail,
			RecipientEmail: recipientEmail,
			RecipientName:  recipientName,
			Amount:         amount,
			Balance:        amount,
			Message:        message,
			Status:         domain.CertificateStatusActive,
			ValidUntil:     time.Now().AddDate(0, 0, certificateValidDays),
			CreatedAt:      time.Now(),
		}

		if err := s.certRepo.Create(ctx, cert); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) && i < certificateCodeRetries-1 {
				continue
			}
			return nil, err
		}

		s.logger.Info("certificate purchased", "certificate_id", cert.ID, "amount", amount)
		return cert, nil
	}

	return nil, domain.ErrAlreadyExists
}

func (s *certificateService) Redeem(ctx context.Context, code string, userID uuid.UUID) (*domain.GiftCertificate, error) {
	if code == "" {
		return nil, domain.ErrInvalidInput
	}

	cert, err := s.certRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if cert.IsExpired() {
		return nil, domain.ErrCertificateExpired
	}

	if cert.Status != domain.CertificateStatusActive {
		return nil, domain.ErrCertificateNotFound
	}

	if cert.RedeemedByID != nil {
		if *cert.RedeemedByID == userID {
			return cert, nil // already redeemed by this user
		}
		return nil, domain.ErrCertificateNotFound
	}

	if err := s.certRepo.Redeem(ctx, cert.ID, userID); err != nil {
		return nil, err
	}

	cert.RedeemedByID = &userID
	s.logger.Info("certificate redeemed", "certificate_id", cert.ID, "user_id", userID)
	return cert, nil
}

func (s *certificateService) Apply(ctx context.Context, certificateID, bookingID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidInput
	}

	cert, err := s.certRepo.GetByID(ctx, certificateID)
	if err != nil {
		return err
	}

	if !cert.IsUsable() {
		if cert.IsExpired() {
			return domain.ErrCertificateExpired
		}
		return domain.ErrCertificateInsufficientBalance
	}

	if cert.Balance < amount {
		return domain.ErrCertificateInsufficientBalance
	}

	usage := &domain.CertificateUsage{
		ID:            uuid.New(),
		CertificateID: certificateID,
		BookingID:     bookingID,
		Amount:        amount,
		UsedAt:        time.Now(),
	}
	if err := s.certRepo.ApplyToBooking(ctx, certificateID, usage); err != nil {
		return err
	}

	s.logger.Info("certificate applied", "certificate_id", certificateID, "booking_id", bookingID, "amount", amount)
	return nil
}

func (s *certificateService) GetBalance(ctx context.Context, code string) (*domain.GiftCertificate, error) {
	if code == "" {
		return nil, domain.ErrInvalidInput
	}

	cert, err := s.certRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *certificateService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error) {
	return s.certRepo.ListByUser(ctx, userID, page, pageSize)
}

func generateCertificateCode() (string, error) {
	chars := []byte(certificateCodeChars)
	max := big.NewInt(int64(len(chars)))

	seg1, err := randomSegment(chars, max, certificateCodeSegLen)
	if err != nil {
		return "", err
	}
	seg2, err := randomSegment(chars, max, certificateCodeSegLen)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("BANI-%s-%s", seg1, seg2), nil
}

func randomSegment(chars []byte, max *big.Int, length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = chars[n.Int64()]
	}
	return string(result), nil
}
