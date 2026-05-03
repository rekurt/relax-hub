package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	GetByReferralCode(ctx context.Context, code string) (*domain.User, error)
	UpdateReferralCode(ctx context.Context, userID uuid.UUID, code string) error
	SetDeletionSchedule(ctx context.Context, userID uuid.UUID, requestedAt, scheduledAt *time.Time) error
	ListPendingDeletions(ctx context.Context, before time.Time) ([]domain.User, error)
	AnonymizeUser(ctx context.Context, userID uuid.UUID, anonEmail string) error
	CountByCreatedAtRange(ctx context.Context, from, to time.Time) (int64, error)
}

type SocialAccountRepository interface {
	Create(ctx context.Context, account *domain.SocialAccount) error
	GetByProviderAndID(ctx context.Context, provider domain.OAuthProvider, providerID string) (*domain.SocialAccount, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error)
	Delete(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllExcept(ctx context.Context, userID uuid.UUID, exceptID uuid.UUID) error
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
	UpdateLastActive(ctx context.Context, id uuid.UUID, lastActiveAt time.Time) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type KYCRepository interface {
	Create(ctx context.Context, kyc *domain.KYCApplication) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.KYCApplication, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.KYCApplication, error)
	Update(ctx context.Context, kyc *domain.KYCApplication) error
	ListPending(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.KYCApplication], error)
	Approve(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, expiresAt time.Time) error
	Reject(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, reason string) error
	ListExpiredApproved(ctx context.Context, before time.Time) ([]domain.KYCApplication, error)
}

type OfferRepository interface {
	Create(ctx context.Context, acceptance *domain.OfferAcceptance) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.OfferAcceptance, error)
	GetByUserAndVersion(ctx context.Context, userID uuid.UUID, version string) (*domain.OfferAcceptance, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.OfferAcceptance, error)
}

type PaymentDetailsRepository interface {
	Upsert(ctx context.Context, details *domain.PaymentDetails) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PaymentDetails, error)
}

type DeviceTokenRepository interface {
	Create(ctx context.Context, token *domain.DeviceToken) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	DeleteByToken(ctx context.Context, token string) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.DeviceToken, error)
}
