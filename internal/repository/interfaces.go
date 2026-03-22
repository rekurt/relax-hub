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
	UpdatePhotoVerified(ctx context.Context, id uuid.UUID, verified bool) error
	GetCalendarToken(ctx context.Context, bathhouseID uuid.UUID) (string, error)
	SetCalendarToken(ctx context.Context, bathhouseID uuid.UUID, token string) error
	GetByCalendarToken(ctx context.Context, token string) (*domain.Bathhouse, error)
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

type BathhousePhotoRepository interface {
	Create(ctx context.Context, photo *domain.BathhousePhoto) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BathhousePhoto, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
	ListVerifiedByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.BathhousePhoto, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PhotoStatus, verifiedByID *uuid.UUID, rejectionReason string) error
	Reorder(ctx context.Context, bathhouseID uuid.UUID, photoIDs []uuid.UUID) error
	ListPending(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.BathhousePhoto], error)
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

type MediaRepository interface {
	Create(ctx context.Context, media *domain.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Media, error)
	ListByOwner(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error)
	ListByOwnerIDs(ctx context.Context, ownerType domain.MediaOwnerType, ownerIDs []uuid.UUID) (map[uuid.UUID][]domain.Media, error)
	ListByBathhouseReviews(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Media], error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MediaStatus) error
	CountByOwner(ctx context.Context, ownerType domain.MediaOwnerType, ownerID uuid.UUID, mediaType *domain.MediaType) (int64, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, externalID string) error
	UpdateRefund(ctx context.Context, id uuid.UUID, refundAmount int64, refundedAt time.Time, status domain.PaymentStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payment], error)
}

type ComplaintRepository interface {
	Create(ctx context.Context, complaint *domain.Complaint) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Complaint, error)
	List(ctx context.Context, filter domain.ComplaintFilter) (*domain.PaginatedResult[domain.Complaint], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ComplaintStatus, resolvedByID *uuid.UUID, resolution string) error
	CountByTarget(ctx context.Context, targetType domain.ComplaintTargetType, targetID uuid.UUID) (int64, error)
	CheckExists(ctx context.Context, reporterID uuid.UUID, targetType domain.ComplaintTargetType, targetID uuid.UUID) (bool, error)
}

type DeviceTokenRepository interface {
	Create(ctx context.Context, token *domain.DeviceToken) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	DeleteByToken(ctx context.Context, token string) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.DeviceToken, error)
}

type ExternalCalendarRepository interface {
	Create(ctx context.Context, cal *domain.ExternalCalendar) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ExternalCalendar, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.ExternalCalendar, error)
	ListAll(ctx context.Context) ([]domain.ExternalCalendar, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, syncedAt time.Time, lastError string) error
}

type SlotBlockRepository interface {
	Create(ctx context.Context, block *domain.SlotBlock) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SlotBlock, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SlotBlock, error)
	GetByExternalID(ctx context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource, externalID string) (*domain.SlotBlock, error)
	DeleteBySource(ctx context.Context, bathhouseID uuid.UUID, source domain.SlotBlockSource) error
	GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.SlotBlock, error)
	HasOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error)
}

type PayoutRepository interface {
	Create(ctx context.Context, payout *domain.Payout) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payout, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Payout], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PayoutStatus, processedAt *time.Time, failureReason string) error
	GetDailyTotal(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error)
	GetMonthlyTotal(ctx context.Context, userID uuid.UUID, year int, month time.Month) (int64, error)
	GetPendingTotal(ctx context.Context, userID uuid.UUID) (int64, error)
	GetAutoPayoutSettings(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error)
	UpsertAutoPayoutSettings(ctx context.Context, settings *domain.AutoPayoutSettings) error
}

type WalletRepository interface {
	Create(ctx context.Context, wallet *domain.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance int64, newHeldAmount int64) error
	UpdateStatus(ctx context.Context, walletID uuid.UUID, status domain.WalletStatus) error
	ListAllIDs(ctx context.Context) ([]uuid.UUID, error)

	CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error
	ListTransactions(ctx context.Context, filter domain.WalletTransactionFilter) (*domain.PaginatedResult[domain.WalletTransaction], error)
	GetExpiringBonuses(ctx context.Context, walletID uuid.UUID, before time.Time) ([]domain.WalletTransaction, error)
	GetBonusTransactionsForSpending(ctx context.Context, walletID uuid.UUID) ([]domain.WalletTransaction, error)
	ExpireBonuses(ctx context.Context, transactionIDs []uuid.UUID) error
	GetExpiringBonusesSoon(ctx context.Context, walletID uuid.UUID, from, to time.Time) ([]domain.WalletTransaction, error)

	CreateHold(ctx context.Context, hold *domain.WalletHold) error
	GetHoldByID(ctx context.Context, holdID uuid.UUID) (*domain.WalletHold, error)
	UpdateHoldStatus(ctx context.Context, holdID uuid.UUID, status domain.WalletHoldStatus, capturedAt, releasedAt *time.Time) error
	GetActiveHolds(ctx context.Context, walletID uuid.UUID) ([]domain.WalletHold, error)
	GetExpiredHolds(ctx context.Context, before time.Time) ([]domain.WalletHold, error)
}
