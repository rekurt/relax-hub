package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/notification"
	"github.com/rekurt/relax-hub/internal/payment"
	"github.com/rekurt/relax-hub/internal/repository"
)

const (
	certificateCodeChars   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	certificateCodeSegLen  = 4
	certificateCodeRetries = 3
	certificateValidDays   = 365
	certificateMaxAmount   = 10_000_000 // 100 000 рублей в копейках
)

type CertificateService interface {
	Purchase(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error)
	CreateOrder(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.CertificateOrder, error)
	InitiatePayment(ctx context.Context, orderID uuid.UUID, req CertificateOrderPaymentRequest) (confirmationURL string, err error)
	GetOrder(ctx context.Context, orderID uuid.UUID) (*domain.CertificateOrder, error)
	HandlePaymentWebhook(ctx context.Context, event WebhookEvent) error
	Redeem(ctx context.Context, code string, userID uuid.UUID) (*domain.GiftCertificate, error)
	Apply(ctx context.Context, certificateID, bookingID uuid.UUID, amount int64) error
	RefundUsage(ctx context.Context, bookingID uuid.UUID) error
	GetBalance(ctx context.Context, code string) (*domain.GiftCertificate, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error)
}

type CertificateOrderPaymentRequest struct {
	PaymentMethod domain.PaymentMethod
	PaymentToken  string
}

type certificateService struct {
	certRepo    repository.GiftCertificateRepository
	orderRepo   repository.CertificateOrderRepository
	paymentProv payment.PaymentProvider
	emailSender notification.EmailSender
	frontendURL string
	logger      *logger.Logger
}

func NewCertificateService(
	certRepo repository.GiftCertificateRepository,
	orderRepo repository.CertificateOrderRepository,
	paymentProv payment.PaymentProvider,
	emailSender notification.EmailSender,
	frontendURL string,
	log *logger.Logger,
) CertificateService {
	return &certificateService{
		certRepo:    certRepo,
		orderRepo:   orderRepo,
		paymentProv: paymentProv,
		emailSender: emailSender,
		frontendURL: frontendURL,
		logger:      log,
	}
}

func (s *certificateService) Purchase(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error) {
	if err := validateCertificatePurchaseInput(amount, purchaserEmail, recipientEmail, recipientName, message); err != nil {
		return nil, err
	}

	cert, err := s.issueCertificate(ctx, amount, purchaserID, purchaserEmail, recipientEmail, recipientName, message)
	if err != nil {
		return nil, err
	}

	s.logger.Info("certificate purchased", "certificate_id", cert.ID, "amount", amount)
	s.sendCertificateEmail(ctx, cert)
	return cert, nil
}

func validateCertificatePurchaseInput(amount int64, purchaserEmail, recipientEmail, recipientName, message string) error {
	if amount <= 0 || amount > certificateMaxAmount {
		return domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(purchaserEmail); err != nil {
		return domain.ErrInvalidInput
	}
	if recipientEmail != "" {
		if _, err := mail.ParseAddress(recipientEmail); err != nil {
			return domain.ErrInvalidInput
		}
	}
	if len(recipientName) > 255 {
		return domain.ErrInvalidInput
	}
	if len(message) > 1000 {
		return domain.ErrInvalidInput
	}
	return nil
}

func (s *certificateService) issueCertificate(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.GiftCertificate, error) {
	var cert *domain.GiftCertificate
	for i := 0; i < certificateCodeRetries; i++ {
		code, err := generateCertificateCode()
		if err != nil {
			return nil, fmt.Errorf("generate certificate code: %w", err)
		}

		now := time.Now()
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
			ValidUntil:     now.AddDate(0, 0, certificateValidDays),
			CreatedAt:      now,
		}

		if err := s.certRepo.Create(ctx, cert); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) && i < certificateCodeRetries-1 {
				continue
			}
			return nil, err
		}
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

func (s *certificateService) RefundUsage(ctx context.Context, bookingID uuid.UUID) error {
	return s.certRepo.RefundUsage(ctx, bookingID)
}

func (s *certificateService) GetBalance(ctx context.Context, code string) (*domain.GiftCertificate, error) {
	if code == "" {
		return nil, domain.ErrInvalidInput
	}

	cert, err := s.certRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if cert.Status == domain.CertificateStatusActive && cert.IsExpired() {
		cert.Status = domain.CertificateStatusExpired
	}

	return cert, nil
}

func (s *certificateService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error) {
	result, err := s.certRepo.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	for i := range result.Items {
		if result.Items[i].Status == domain.CertificateStatusActive && result.Items[i].IsExpired() {
			result.Items[i].Status = domain.CertificateStatusExpired
		}
	}

	return result, nil
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

func (s *certificateService) sendCertificateEmail(ctx context.Context, cert *domain.GiftCertificate) {
	if s.emailSender == nil {
		return
	}
	amountRub := float64(cert.Amount) / 100
	// Email to recipient (only if email provided)
	if cert.RecipientEmail != "" {
		recipientSubject := "Вам подарили сертификат Bani!"
		recipientBody := fmt.Sprintf(
			"Вам подарен сертификат на %.0f руб.\n\nКод сертификата: %s\n\nСрок действия до: %s",
			amountRub, cert.Code, cert.ValidUntil.Format("02.01.2006"),
		)
		if cert.Message != "" {
			recipientBody += fmt.Sprintf("\n\nСообщение: %s", cert.Message)
		}
		if err := s.emailSender.Send(ctx, cert.RecipientEmail, recipientSubject, recipientBody); err != nil {
			s.logger.Warn("failed to send certificate email to recipient",
				"certificate_id", cert.ID, "recipient_email", cert.RecipientEmail, "error", err)
		}
	}

	// Email to purchaser
	purchaserSubject := "Подтверждение покупки сертификата"
	purchaserBody := fmt.Sprintf(
		"Вы приобрели подарочный сертификат на %.0f руб.\n\nКод: %s\nПолучатель: %s (%s)\nДействителен до: %s",
		amountRub, cert.Code, cert.RecipientName, cert.RecipientEmail, cert.ValidUntil.Format("02.01.2006"),
	)
	if err := s.emailSender.Send(ctx, cert.PurchaserEmail, purchaserSubject, purchaserBody); err != nil {
		s.logger.Warn("failed to send certificate confirmation to purchaser",
			"certificate_id", cert.ID, "purchaser_email", cert.PurchaserEmail, "error", err)
	}
}
