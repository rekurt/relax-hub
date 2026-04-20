package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type GuestCardRepository interface {
	Upsert(ctx context.Context, card *domain.GuestCard) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GuestCard, error)
	GetByOwnerAndClient(ctx context.Context, ownerID, clientID, bathhouseID uuid.UUID) (*domain.GuestCard, error)
	ListByOwner(ctx context.Context, filter domain.GuestCardFilter) (*domain.PaginatedResult[domain.GuestCard], error)
	UpdateNotes(ctx context.Context, id uuid.UUID, notes string, tags []string) error
	GetStats(ctx context.Context, filter domain.GuestCardFilter) (*domain.GuestCardStats, error)
	CountBySegment(ctx context.Context, filter domain.GuestCardFilter, segment domain.GuestSegmentSlug) (int64, error)
	GetRFMScores(ctx context.Context, filter domain.GuestCardFilter) (*domain.RFMResult, error)
}

type CustomSegmentRepository interface {
	Create(ctx context.Context, segment *domain.CustomSegment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomSegment, error)
	Update(ctx context.Context, segment *domain.CustomSegment) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOwner(ctx context.Context, filter domain.CustomSegmentFilter) ([]domain.CustomSegment, error)
	EvaluateSegment(ctx context.Context, segment *domain.CustomSegment, ownerFilter domain.GuestCardFilter, page, pageSize int) (*domain.PaginatedResult[domain.GuestCard], error)
	CountSegmentGuests(ctx context.Context, segment *domain.CustomSegment, ownerFilter domain.GuestCardFilter) (int64, error)
}

type BroadcastRepository interface {
	Create(ctx context.Context, broadcast *domain.Broadcast) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Broadcast, error)
	ListByOwner(ctx context.Context, filter domain.BroadcastFilter) (*domain.PaginatedResult[domain.Broadcast], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BroadcastStatus) error
	UpdateStats(ctx context.Context, id uuid.UUID, delivered, read, clicked int64) error
	CountRecentByOwner(ctx context.Context, ownerID uuid.UUID, since time.Time) (int64, error)
}

type AutoScenarioRepository interface {
	Upsert(ctx context.Context, scenario *domain.AutoScenario) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.AutoScenario, error)
	GetByOwnerAndType(ctx context.Context, ownerID uuid.UUID, scenarioType domain.AutoScenarioType) (*domain.AutoScenario, error)
	ListEnabled(ctx context.Context) ([]domain.AutoScenario, error)
	RecordExecution(ctx context.Context, scenarioID, guestCardID uuid.UUID) error
	HasBeenExecuted(ctx context.Context, scenarioID, guestCardID uuid.UUID) (bool, error)
}

type ResponseTemplateRepository interface {
	Create(ctx context.Context, template *domain.ResponseTemplate) error
	Update(ctx context.Context, template *domain.ResponseTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.ResponseTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ResponseTemplate, error)
	CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error)
}

type LoyaltyRepository interface {
	GetAccount(ctx context.Context, userID uuid.UUID) (*domain.LoyaltyAccount, error)
	CreateAccount(ctx context.Context, account *domain.LoyaltyAccount) error
	AddPoints(ctx context.Context, userID uuid.UUID, amount int64) error
	SpendPoints(ctx context.Context, userID uuid.UUID, amount int64) error
	RefundPoints(ctx context.Context, userID uuid.UUID, amount int64) error
	IncrementVisitCount(ctx context.Context, userID uuid.UUID) error
	UpdateLevel(ctx context.Context, userID uuid.UUID, level domain.LoyaltyLevel) error
	ListTransactions(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.LoyaltyTransaction], error)
	CreateTransaction(ctx context.Context, tx *domain.LoyaltyTransaction) error
}

type ReferralRepository interface {
	Create(ctx context.Context, referral *domain.Referral) error
	GetByReferee(ctx context.Context, refereeID uuid.UUID) (*domain.Referral, error)
	ListByReferrer(ctx context.Context, referrerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Referral], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReferralStatus, completedAt *time.Time) error
	RevertToPending(ctx context.Context, id uuid.UUID) error
	GetBalance(ctx context.Context, userID uuid.UUID) (*domain.ReferralBalance, error)
	CreateBalance(ctx context.Context, balance *domain.ReferralBalance) error
	UpdateBalance(ctx context.Context, userID uuid.UUID, delta int64, trackEarnings bool) error
	CountByReferrer(ctx context.Context, referrerID uuid.UUID) (int, int, error) // totalInvited, totalCompleted
}

type GiftCertificateRepository interface {
	Create(ctx context.Context, cert *domain.GiftCertificate) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GiftCertificate, error)
	GetByCode(ctx context.Context, code string) (*domain.GiftCertificate, error)
	ApplyToBooking(ctx context.Context, id uuid.UUID, usage *domain.CertificateUsage) error
	RefundUsage(ctx context.Context, bookingID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error)
	Redeem(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int, error)
}

type CertificateOrderRepository interface {
	Create(ctx context.Context, order *domain.CertificateOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CertificateOrder, error)
	GetByExternalID(ctx context.Context, externalID string) (*domain.CertificateOrder, error)
	UpdatePayment(ctx context.Context, id uuid.UUID, status domain.CertificateOrderStatus, paymentMethod domain.PaymentMethod, provider, externalID string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CertificateOrderStatus) error
	MarkPaid(ctx context.Context, id uuid.UUID, certificateID uuid.UUID, paidAt time.Time) error
}

type PromoCodeRepository interface {
	Create(ctx context.Context, promo *domain.PromoCode) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PromoCode, error)
	GetByCode(ctx context.Context, code string) (*domain.PromoCode, error)
	Update(ctx context.Context, promo *domain.PromoCode) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error)
	ListByCreator(ctx context.Context, creatorID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PromoCode], error)
	IncrementUses(ctx context.Context, id uuid.UUID) error
	RecordUsage(ctx context.Context, usage *domain.PromoUsage) error
	ApplyUsage(ctx context.Context, id uuid.UUID, usage *domain.PromoUsage) error
	GetUsageByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.PromoUsage, error)
	DecrementUses(ctx context.Context, id uuid.UUID) error
	DeleteUsage(ctx context.Context, usageID uuid.UUID) error
	RefundUsage(ctx context.Context, promoCodeID uuid.UUID, usageID uuid.UUID) error
	DeactivateExpired(ctx context.Context, before time.Time) (int64, error)
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Subscription], error)
	GetExpiring(ctx context.Context, before time.Time) ([]domain.Subscription, error)
}

type PromotionRepository interface {
	Create(ctx context.Context, promo *domain.Promotion) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Promotion, error)
	GetActiveBybathhouse(ctx context.Context, bathhouseID uuid.UUID) (*domain.Promotion, error)
	Update(ctx context.Context, promo *domain.Promotion) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error)
	ListAllActive(ctx context.Context) ([]domain.Promotion, error)
	RecordImpression(ctx context.Context, promotionID uuid.UUID) error
	RecordClick(ctx context.Context, promotionID uuid.UUID) error
}

type SeasonalTariffRepository interface {
	Create(ctx context.Context, tariff *domain.SeasonalTariff) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SeasonalTariff, error)
	Update(ctx context.Context, tariff *domain.SeasonalTariff) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SeasonalTariff, error)
	GetActiveTariffs(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]domain.SeasonalTariff, error)
}

type PricingRuleRepository interface {
	Create(ctx context.Context, rule *domain.PricingRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PricingRule, error)
	Update(ctx context.Context, rule *domain.PricingRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
	GetActiveRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
}
