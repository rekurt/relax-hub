package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/fiscal"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/payment"
	"github.com/rekurt/relax-hub/internal/repository"
)

// WebhookEvent represents a payment webhook notification from the provider.
type WebhookEvent struct {
	ExternalID string
	Status     string // "succeeded", "canceled", etc.
}

// ComboPaymentRequest describes how a booking payment should be split between wallet and card/SBP.
type ComboPaymentRequest struct {
	WalletAmount  int64                // amount to pay from wallet (0 = card only)
	CardAmount    int64                // amount to pay by card/SBP (0 = wallet only)
	PaymentMethod domain.PaymentMethod // "card" or "sbp" (for card portion)
}

// TokenPaymentRequest extends a standard payment with a client-side payment token (Apple Pay / Google Pay).
type TokenPaymentRequest struct {
	PaymentMethod domain.PaymentMethod
	PaymentToken  string // token from Apple Pay JS or Google Pay API
}

type PaymentService interface {
	InitiatePayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, paymentMethod domain.PaymentMethod) (confirmationURL string, err error)
	InitiateTokenPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req TokenPaymentRequest) (confirmationURL string, err error)
	InitiateComboPayment(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, req ComboPaymentRequest) (confirmationURL string, err error)
	HandleWebhook(ctx context.Context, event WebhookEvent) error
	RefundPayment(ctx context.Context, bookingID uuid.UUID, forceFullRefund bool, refundTo string, policy domain.CancellationPolicy) error
	AdminRefund(ctx context.Context, adminUserID uuid.UUID, bookingID uuid.UUID, amount int64, reason string, refundTo string) error
	CaptureHoldPayment(ctx context.Context, bookingID uuid.UUID) error
	ReleaseHoldPayment(ctx context.Context, bookingID uuid.UUID) error
	GetPaymentByBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Payment, error)
	ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
}

const defaultWalletRefundBonusPercent = 5

type paymentService struct {
	paymentRepo              repository.PaymentRepository
	bookingRepo              repository.BookingRepository
	bathhouseRepo            repository.BathhouseRepository
	kycRepo                  repository.KYCRepository
	auditLogRepo             repository.AuditLogRepository
	provider                 payment.PaymentProvider
	fiscalProvider           fiscal.FiscalProvider
	walletSvc                WalletService
	notifSvc                 NotificationService
	returnURL                string
	walletRefundBonusPercent int
	logger                   *logger.Logger
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	bathhouseRepo repository.BathhouseRepository,
	kycRepo repository.KYCRepository,
	auditLogRepo repository.AuditLogRepository,
	provider payment.PaymentProvider,
	fiscalProvider fiscal.FiscalProvider,
	walletSvc WalletService,
	notifSvc NotificationService,
	returnURL string,
	walletRefundBonusPercent int,
	log *logger.Logger,
) PaymentService {
	if walletRefundBonusPercent < 0 || walletRefundBonusPercent > 15 {
		walletRefundBonusPercent = defaultWalletRefundBonusPercent
	}
	return &paymentService{
		paymentRepo:              paymentRepo,
		bookingRepo:              bookingRepo,
		bathhouseRepo:            bathhouseRepo,
		kycRepo:                  kycRepo,
		auditLogRepo:             auditLogRepo,
		provider:                 provider,
		fiscalProvider:           fiscalProvider,
		walletSvc:                walletSvc,
		notifSvc:                 notifSvc,
		returnURL:                returnURL,
		walletRefundBonusPercent: walletRefundBonusPercent,
		logger:                   log,
	}
}

func (s *paymentService) GetPaymentByBooking(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*domain.Payment, error) {
	p, err := s.paymentRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return p, nil
}

func (s *paymentService) ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error) {
	return s.paymentRepo.ListByUser(ctx, userID, page, pageSize)
}
