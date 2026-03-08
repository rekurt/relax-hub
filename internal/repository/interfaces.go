package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	List(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.User], error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	GetByReferralCode(ctx context.Context, code string) (*domain.User, error)
	UpdateReferralCode(ctx context.Context, userID uuid.UUID, code string) error
}

type CityRepository interface {
	Create(ctx context.Context, city *domain.City) error
	GetAll(ctx context.Context) ([]domain.City, error)
	GetBySlug(ctx context.Context, slug string) (*domain.City, error)
	GetByID(ctx context.Context, id int64) (*domain.City, error)
	Update(ctx context.Context, city *domain.City) error
	Delete(ctx context.Context, id int64) error
}

type BathhouseRepository interface {
	Create(ctx context.Context, bh *domain.Bathhouse) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bathhouse, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Bathhouse, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*domain.Bathhouse, error)
	Update(ctx context.Context, bh *domain.Bathhouse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Bathhouse], error)
	ListIDsByOwner(ctx context.Context, ownerID uuid.UUID) ([]uuid.UUID, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	UpdateRating(ctx context.Context, bathhouseID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BathhouseStatus) error
}

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
	CheckAvailability(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error)
	GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.Booking, error)
	CountActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserBookingStats, error)
}

type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	Update(ctx context.Context, review *domain.Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Review], error)
	ListByBathhouseFiltered(ctx context.Context, filter domain.ReviewFilter) (*domain.PaginatedResult[domain.Review], error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Review, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReviewStatus) error
	UpdateStatusWithReasons(ctx context.Context, id uuid.UUID, status domain.ReviewStatus, reasons []string) error
	AddOwnerResponse(ctx context.Context, id uuid.UUID, response string, respondedAt time.Time) error
	GetUserReviewStats(ctx context.Context, userID uuid.UUID) (*domain.UserReviewStats, error)
	CountPendingReviews(ctx context.Context) (int64, error)
	ListAllReviews(ctx context.Context, filter domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error)
}

type FavoriteRepository interface {
	Add(ctx context.Context, favorite *domain.Favorite) error
	Remove(ctx context.Context, userID, bathhouseID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Favorite], error)
	IsFavorite(ctx context.Context, userID, bathhouseID uuid.UUID) (bool, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Notification], error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *domain.NotificationPreferences) error
}

type SocialAccountRepository interface {
	Create(ctx context.Context, account *domain.SocialAccount) error
	GetByProviderAndID(ctx context.Context, provider domain.OAuthProvider, providerID string) (*domain.SocialAccount, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.SocialAccount, error)
	Delete(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error
}

type RepresentativeRepository interface {
	Create(ctx context.Context, rep *domain.Representative) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Representative, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserAndBathhouse(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error)
	ListBathhouseIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type RecommendationRepository interface {
	// User preferences
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error)
	SaveUserPreferences(ctx context.Context, prefs *domain.UserPreferences) error

	// Activity tracking
	RecordActivity(ctx context.Context, activity *domain.UserActivity) error

	// Booking history
	GetUserBookedBathhouses(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error)
	// GetUserBookedBathhousesWithDates returns booked bathhouses with their booking dates for recency calculation
	GetUserBookedBathhousesWithDates(ctx context.Context, userID uuid.UUID, limit int) ([]domain.BookedBathhouseWithDate, error)

	// Collaborative filtering
	GetSimilarUsers(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error)

	// Popularity
	GetPopularBathhouses(ctx context.Context, cityID int64, limit int) ([]uuid.UUID, error)

	// Similarity based on amenities and location
	GetSimilarBathhouses(ctx context.Context, bathhouseID uuid.UUID, limit int) ([]uuid.UUID, error)
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
	RecordImpression(ctx context.Context, promotionID uuid.UUID) error
	RecordClick(ctx context.Context, promotionID uuid.UUID) error
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

type PricingRuleRepository interface {
	Create(ctx context.Context, rule *domain.PricingRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PricingRule, error)
	Update(ctx context.Context, rule *domain.PricingRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
	GetActiveRules(ctx context.Context, bathhouseID uuid.UUID) ([]domain.PricingRule, error)
}

type ConversationRepository interface {
	Create(ctx context.Context, conv *domain.Conversation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error)
	GetByParticipants(ctx context.Context, bathhouseID, clientID uuid.UUID) (*domain.Conversation, error)
	ListByUser(ctx context.Context, userID uuid.UUID, bathhouseIDs []uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error)
	ListAll(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.Conversation], error)
	GetOrCreate(ctx context.Context, conv *domain.Conversation) (*domain.Conversation, error)
	UpdateLastMessageAt(ctx context.Context, id uuid.UUID, t time.Time) error
}

type MessageRepository interface {
	Create(ctx context.Context, msg *domain.Message) error
	ListByConversation(ctx context.Context, conversationID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Message], error)
	MarkAsRead(ctx context.Context, conversationID, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID, conversationIDs []uuid.UUID) (int64, error)
	CountUnreadByUser(ctx context.Context, userID uuid.UUID, bathhouseIDs []uuid.UUID) (int64, error)
}

type AnalyticsRepository interface {
	RecordView(ctx context.Context, view *domain.BathhouseView) error
	GetBathhouseStats(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) (*domain.AnalyticsSnapshot, error)
	GetDailyStats(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) ([]domain.AnalyticsSnapshot, error)
	GetPlatformStats(ctx context.Context, from, to time.Time) (*domain.AnalyticsSnapshot, error)
	GetTopBathhouses(ctx context.Context, metric domain.TopMetric, limit int) ([]uuid.UUID, error)
	AggregateRawData(ctx context.Context, bathhouseID uuid.UUID, date time.Time) (*domain.AnalyticsSnapshot, error)
	CreateSnapshot(ctx context.Context, snapshot *domain.AnalyticsSnapshot) error
	DeleteOldViews(ctx context.Context, before time.Time) (int64, error)
}

type TelegramLinkRepository interface {
	Create(ctx context.Context, link *domain.TelegramLink) error
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.TelegramLink, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TelegramLink, error)
	Delete(ctx context.Context, userID uuid.UUID) error
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
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.GiftCertificate], error)
	Redeem(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type ComplaintRepository interface {
	Create(ctx context.Context, complaint *domain.Complaint) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error)
	List(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ComplaintStatus, resolvedByID *uuid.UUID, resolution string) error
	CountByTarget(ctx context.Context, targetType domain.ComplaintTargetType, targetID uuid.UUID) (int64, error)
	CheckExists(ctx context.Context, reporterID uuid.UUID, targetType domain.ComplaintTargetType, targetID uuid.UUID) (bool, error)
}
