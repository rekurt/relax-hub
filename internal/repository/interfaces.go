package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
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
	UpdateBayesianRating(ctx context.Context, bathhouseID uuid.UUID, bayesianRating float64) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BathhouseStatus) error
	UpdatePhotoVerified(ctx context.Context, id uuid.UUID, verified bool) error
	GetCalendarToken(ctx context.Context, bathhouseID uuid.UUID) (string, error)
	SetCalendarToken(ctx context.Context, bathhouseID uuid.UUID, token string) error
	GetByCalendarToken(ctx context.Context, token string) (*domain.Bathhouse, error)
	SuggestNames(ctx context.Context, filter SuggestionFilter) ([]string, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	UpdateRankingFields(ctx context.Context, id uuid.UUID, conversionRate, occupancyRate float64) error
	UpdateResponseRate(ctx context.Context, id uuid.UUID, responseRate float64, avgResponseMinutes int, lowResponseRateSince *time.Time) error
	ListRequestModeBathhouses(ctx context.Context) ([]domain.Bathhouse, error)
	GetAreaAvgPrice(ctx context.Context, cityID int64, lat, lng float64) (int64, error)
}

// SuggestionFilter specifies parameters for name-based suggestions.
type SuggestionFilter struct {
	Query string
	Limit int
}

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	Update(ctx context.Context, booking *domain.Booking) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Booking], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
	CheckAvailability(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) (bool, error)
	CheckAvailabilityExcluding(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time, excludeBookingID uuid.UUID) (bool, error)
	CreateWithAvailabilityCheck(ctx context.Context, booking *domain.Booking, checkStart, checkEnd time.Time) error
	GetOverlapping(ctx context.Context, bathhouseID uuid.UUID, startTime, endTime time.Time) ([]domain.Booking, error)
	CountActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserBookingStats, error)
	ListTimedOutRequests(ctx context.Context) ([]domain.Booking, error)
	UpdateCheckin(ctx context.Context, bookingID uuid.UUID, checkedInAt *time.Time) error
	UpdateCheckout(ctx context.Context, bookingID uuid.UUID, checkedOutAt *time.Time, status domain.BookingStatus) error
	ListConfirmedWithoutCheckin(ctx context.Context, noShowCutoff time.Time) ([]domain.Booking, error)
	ListUpcoming(ctx context.Context, from, to time.Time) ([]domain.Booking, error)
	UpdateModification(ctx context.Context, bookingID uuid.UUID, startTime, endTime time.Time, guestCount int, totalPrice, addOnTotal, basePrice, longSessionDiscount, extraGuestSurcharge, lastMinuteDiscount, serviceFeeAmount int64, modificationCount int) error
	UpdateEndTime(ctx context.Context, bookingID uuid.UUID, oldEndTime, newEndTime time.Time, newTotalPrice int64) error
	UpdateCancelledByOwner(ctx context.Context, bookingID uuid.UUID) error
	CountOwnerCancellations(ctx context.Context, ownerID uuid.UUID, since time.Time) (int, error)
	GetResponseStats(ctx context.Context, bathhouseID uuid.UUID, since time.Time) (totalRequests int, respondedInTime int, avgResponseMinutes int, err error)
	ListCompletedForReviewRequests(ctx context.Context, checkedOutBefore time.Time) ([]domain.Booking, error)
	GetLastBookingDateByUser(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	ListConfirmedByRegionAndDateRange(ctx context.Context, region string, dateFrom, dateTo time.Time) ([]domain.Booking, error)
	UpdateDeposit(ctx context.Context, bookingID uuid.UUID, depositAmount int64, depositStatus domain.DepositStatus, depositExternalID string) error
	UpdateDepositStatus(ctx context.Context, bookingID uuid.UUID, depositStatus domain.DepositStatus, releasedAt *time.Time) error
	ListHeldDepositsReadyForRelease(ctx context.Context, checkedOutBefore time.Time) ([]domain.Booking, error)
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	ListAll(ctx context.Context, filter domain.AdminBookingFilter) (*domain.PaginatedResult[domain.Booking], error)
}

type BookingModificationRequestRepository interface {
	Create(ctx context.Context, req *domain.BookingModificationRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingModificationRequest, error)
	GetPendingByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.BookingModificationRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ModificationRequestStatus, reason string) error
	ListExpired(ctx context.Context) ([]domain.BookingModificationRequest, error)
	ListByBookingID(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingModificationRequest, error)
}

type ExtensionRequestRepository interface {
	Create(ctx context.Context, req *domain.BookingExtensionRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BookingExtensionRequest, error)
	GetPendingByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.BookingExtensionRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ExtensionRequestStatus, reason string) error
	ListExpired(ctx context.Context) ([]domain.BookingExtensionRequest, error)
	ListByBookingID(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingExtensionRequest, error)
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
	GetCriteriaAverages(ctx context.Context, bathhouseID uuid.UUID) (*domain.ReviewCriteriaAverages, error)
	GetPlatformAverageRating(ctx context.Context) (float64, error)
	ListUnrevealedPastDeadline(ctx context.Context, now time.Time) ([]domain.Review, error)
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
	HasRecentByType(ctx context.Context, userID uuid.UUID, notifType domain.NotificationType, since time.Time) (bool, error)
	GetEventPreferences(ctx context.Context, userID uuid.UUID) ([]domain.NotificationEventPreference, error)
	GetEventPreference(ctx context.Context, userID uuid.UUID, eventType domain.NotificationEventType) (*domain.NotificationEventPreference, error)
	UpsertEventPreference(ctx context.Context, pref *domain.NotificationEventPreference) error
	UpsertEventPreferences(ctx context.Context, prefs []domain.NotificationEventPreference) error
	CreatePushDeliveryLog(ctx context.Context, log *domain.PushDeliveryLog) error
	GetPendingPushDeliveries(ctx context.Context, olderThan time.Time) ([]domain.PushDeliveryLog, error)
	UpdatePushDeliveryStatus(ctx context.Context, id uuid.UUID, status domain.PushDeliveryStatus) error
	MarkPushFallbackSent(ctx context.Context, id uuid.UUID) error
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
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Promotion], error)
	ListAllActive(ctx context.Context) ([]domain.Promotion, error)
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

	// Advanced analytics (FR-147-154)
	GetConversionFunnel(ctx context.Context, from, to time.Time) ([]domain.FunnelStep, error)
	GetCohortAnalysis(ctx context.Context, months int) ([]domain.CohortRow, error)
	GetGeoSupplyDemand(ctx context.Context, from, to time.Time) ([]domain.GeoSupplyDemand, error)
	GetWalletMetrics(ctx context.Context, from, to time.Time) (*domain.WalletMetrics, error)
	GetOwnerPerformance(ctx context.Context, bathhouseID uuid.UUID, from, to time.Time) (*domain.OwnerPerformance, error)

	// Business metrics (FR-148, FR-149)
	GetChurnRate(ctx context.Context, inactiveDays int) (float64, error)
	GetLTV(ctx context.Context) (int64, error)
	GetARPU(ctx context.Context, from, to time.Time) (int64, error)

	// P&L metrics (FR-150)
	GetGMV(ctx context.Context, from, to time.Time) (int64, int64, error) // returns (gmv, bookingCount)
	GetPlatformRevenue(ctx context.Context, from, to time.Time) (serviceFees, subscriptions, promotions int64, err error)
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
	RefundUsage(ctx context.Context, bookingID uuid.UUID) error
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
	UpdateCapture(ctx context.Context, id uuid.UUID, capturedAt time.Time, status domain.PaymentStatus) error
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
	UpdateExternalID(ctx context.Context, id uuid.UUID, externalID string) error
	GetDailyTotal(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error)
	GetMonthlyTotal(ctx context.Context, userID uuid.UUID, year int, month time.Month) (int64, error)
	GetPendingTotal(ctx context.Context, userID uuid.UUID) (int64, error)
	GetAutoPayoutSettings(ctx context.Context, userID uuid.UUID) (*domain.AutoPayoutSettings, error)
	UpsertAutoPayoutSettings(ctx context.Context, settings *domain.AutoPayoutSettings) error
	ListActiveAutoPayoutSettings(ctx context.Context) ([]domain.AutoPayoutSettings, error)
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

type ListingDraftRepository interface {
	Create(ctx context.Context, draft *domain.ListingDraft) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ListingDraft, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.ListingDraft, error)
	UpdateStep(ctx context.Context, id uuid.UUID, step int, data json.RawMessage, currentStep int) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ListingDraftStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type WalletRepository interface {
	Create(ctx context.Context, wallet *domain.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, oldBalance, newBalance int64, oldHeldAmount, newHeldAmount int64) error
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

type AddOnRepository interface {
	Create(ctx context.Context, addon *domain.AddOn) error
	Update(ctx context.Context, addon *domain.AddOn) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AddOn, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error)
	ListActiveByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.AddOn, error)
	CountByBathhouse(ctx context.Context, bathhouseID uuid.UUID) (int64, error)
	CreateBookingAddOn(ctx context.Context, ba *domain.BookingAddOn) error
	ListByBooking(ctx context.Context, bookingID uuid.UUID) ([]domain.BookingAddOn, error)
}

type SavedSearchRepository interface {
	Create(ctx context.Context, search *domain.SavedSearch) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedSearch], error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedSearch, error)
	ListWithNotifications(ctx context.Context) ([]domain.SavedSearch, error)
}

type SavedCardRepository interface {
	Create(ctx context.Context, card *domain.SavedCard) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SavedCard, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.SavedCard], error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetDefault(ctx context.Context, userID uuid.UUID, cardID uuid.UUID) error
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.AuditLog], error)
	List(ctx context.Context, filter domain.AuditLogFilter) (*domain.PaginatedResult[domain.AuditLog], error)
}

type HolidayRepository interface {
	Create(ctx context.Context, holiday *domain.Holiday) error
	Update(ctx context.Context, holiday *domain.Holiday) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Holiday, error)
	ListAll(ctx context.Context) ([]domain.Holiday, error)
	ListByRegion(ctx context.Context, region string) ([]domain.Holiday, error)
	IsHoliday(ctx context.Context, date time.Time, region string) (*domain.Holiday, error)
	GetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID) (float64, error)
	SetBathhouseMultiplier(ctx context.Context, bathhouseID uuid.UUID, multiplier float64) error
}

type SeasonalTariffRepository interface {
	Create(ctx context.Context, tariff *domain.SeasonalTariff) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SeasonalTariff, error)
	Update(ctx context.Context, tariff *domain.SeasonalTariff) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.SeasonalTariff, error)
	GetActiveTariffs(ctx context.Context, bathhouseID uuid.UUID, date time.Time) ([]domain.SeasonalTariff, error)
}

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

type EscrowRepository interface {
	Create(ctx context.Context, escrow *domain.Escrow) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Escrow, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Escrow, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EscrowStatus, releasedAt *time.Time) error
	ListMatured(ctx context.Context) ([]domain.Escrow, error) // status=held AND claim_period_ends_at < now
}

type ServiceFeeRepository interface {
	GetByRegionAndCategory(ctx context.Context, region string, category *string) (*domain.ServiceFeeConfig, error)
	GetByRegion(ctx context.Context, region string) (*domain.ServiceFeeConfig, error)
	GetGlobalDefault(ctx context.Context) (*domain.ServiceFeeConfig, error)
	List(ctx context.Context) ([]domain.ServiceFeeConfig, error)
	Upsert(ctx context.Context, config *domain.ServiceFeeConfig) error
}

type DisputeRepository interface {
	Create(ctx context.Context, dispute *domain.Dispute) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Dispute, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.Dispute, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.DisputeStatus) error
	UpdateResolution(ctx context.Context, id uuid.UUID, resolution domain.DisputeResolution, refundAmount, compensationAmount int64, mediatorNotes string, resolvedAt time.Time) error
	UpdateAppeal(ctx context.Context, id uuid.UUID, status domain.DisputeStatus) error
	Assign(ctx context.Context, id uuid.UUID, mediatorID uuid.UUID) error
	AddEvidence(ctx context.Context, evidence *domain.DisputeEvidence) error
	ListEvidence(ctx context.Context, disputeID uuid.UUID) ([]domain.DisputeEvidence, error)
	ListAll(ctx context.Context, filter domain.DisputeFilter) (*domain.PaginatedResult[domain.Dispute], error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Dispute], error)
	CountOpenByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type FraudFlagRepository interface {
	Create(ctx context.Context, flag *domain.FraudFlag) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.FraudFlag], error)
	ListPending(ctx context.Context, filter domain.FraudFlagFilter) (*domain.PaginatedResult[domain.FraudFlag], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.FraudFlagStatus, reviewedBy uuid.UUID) error
	CountByUserAndRule(ctx context.Context, userID uuid.UUID, rule domain.FraudRuleName, since time.Time) (int64, error)
}

type PlatformSettingsRepository interface {
	Get(ctx context.Context, key string) (*domain.PlatformSetting, error)
	GetAll(ctx context.Context) ([]domain.PlatformSetting, error)
	Set(ctx context.Context, key, value string, updatedBy *uuid.UUID) error
}

type FeatureFlagRepository interface {
	Get(ctx context.Context, key string) (*domain.FeatureFlag, error)
	GetAll(ctx context.Context) ([]domain.FeatureFlag, error)
	Set(ctx context.Context, key string, enabled bool, region *string, updatedBy *uuid.UUID) error
}

type TicketRepository interface {
	Create(ctx context.Context, ticket *domain.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Ticket], error)
	ListAll(ctx context.Context, filter domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TicketStatus) error
	UpdateLevel(ctx context.Context, id uuid.UUID, level domain.TicketLevel) error
	Assign(ctx context.Context, id uuid.UUID, assignedTo uuid.UUID) error
	Resolve(ctx context.Context, id uuid.UUID, resolvedAt time.Time) error
	SubmitCSAT(ctx context.Context, id uuid.UUID, score int) error
	AddMessage(ctx context.Context, msg *domain.TicketMessage) error
	ListMessages(ctx context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error)
	CountByStatus(ctx context.Context) (*domain.TicketStatusCounts, error)
	ListStaleTickets(ctx context.Context, level domain.TicketLevel, olderThan time.Time) ([]domain.Ticket, error)
	ListResolvedForAutoClose(ctx context.Context, resolvedBefore time.Time) ([]domain.Ticket, error)
	GetOperationMetrics(ctx context.Context, filter domain.TicketMetricsFilter) (*domain.TicketOperationMetrics, error)
}

type ForceMajeureRepository interface {
	Create(ctx context.Context, event *domain.ForceMajeureEvent) error
	List(ctx context.Context) ([]domain.ForceMajeureEvent, error)
}

type ClientReviewRepository interface {
	Create(ctx context.Context, review *domain.ClientReview) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientReview, error)
	GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*domain.ClientReview, error)
	Update(ctx context.Context, review *domain.ClientReview) error
	ListByClient(ctx context.Context, clientID uuid.UUID, onlyRevealed bool, page, pageSize int) (*domain.PaginatedResult[domain.ClientReview], error)
	ListUnrevealedPastDeadline(ctx context.Context, now time.Time) ([]domain.ClientReview, error)
	RevealByID(ctx context.Context, id uuid.UUID) error
}

type ReconciliationRepository interface {
	CreateFloatSnapshot(ctx context.Context, snapshot *domain.FloatSnapshot) error
	GetLatestFloatSnapshot(ctx context.Context) (*domain.FloatSnapshot, error)
	ListFloatSnapshots(ctx context.Context, from, to time.Time, page, pageSize int) (*domain.PaginatedResult[domain.FloatSnapshot], error)

	CreateReconciliationReport(ctx context.Context, report *domain.ReconciliationReport) error
	GetLatestReconciliationReport(ctx context.Context) (*domain.ReconciliationReport, error)
	ListReconciliationReports(ctx context.Context, page, pageSize int) (*domain.PaginatedResult[domain.ReconciliationReport], error)

	// Aggregate queries for float calculation
	SumWalletBalancesByRole(ctx context.Context, role string) (total int64, count int, err error)
	SumEscrowHeld(ctx context.Context) (total int64, count int, err error)
	SumActiveWalletHolds(ctx context.Context) (total int64, err error)
	SumPaymentsForPeriod(ctx context.Context, from, to time.Time) (sum int64, count int, err error)
	SumRefundsForPeriod(ctx context.Context, from, to time.Time) (sum int64, count int, err error)
}

// BookingShareRepository manages shareable booking links.
type BookingShareRepository interface {
	Create(ctx context.Context, share *domain.BookingShare) error
	GetByToken(ctx context.Context, token string) (*domain.BookingShare, error)
	DeleteExpired(ctx context.Context) (int64, error)
}

type AmenityRepository interface {
	Create(ctx context.Context, amenity *domain.Amenity) error
	Update(ctx context.Context, amenity *domain.Amenity) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Amenity, error)
	ListAll(ctx context.Context) ([]domain.Amenity, error)
}

type ObjectTypeRepository interface {
	Create(ctx context.Context, objType *domain.ObjectType) error
	Update(ctx context.Context, objType *domain.ObjectType) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ObjectType, error)
	ListAll(ctx context.Context) ([]domain.ObjectType, error)
}

// BankReconciliationRepository manages bank statement entries and uploads.
type BankReconciliationRepository interface {
	CreateUpload(ctx context.Context, upload *domain.BankStatementUpload) error
	UpdateUploadCounts(ctx context.Context, id uuid.UUID, matched, pending, ignored int) error

	CreateEntries(ctx context.Context, entries []domain.BankStatementEntry) error
	GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.BankStatementEntry, error)
	ListEntries(ctx context.Context, filter domain.BankStatementFilter) (*domain.PaginatedResult[domain.BankStatementEntry], error)
	MatchEntry(ctx context.Context, entryID uuid.UUID, txID uuid.UUID, txType string) error

	// FindPaymentsByAmountAndDate finds payments matching amount and date range for auto-matching.
	FindPaymentsByAmountAndDate(ctx context.Context, amount int64, dateFrom, dateTo time.Time) ([]domain.Payment, error)
}

// FAQRepository manages FAQ entries.
type FAQRepository interface {
	Create(ctx context.Context, faq *domain.FAQ) error
	Update(ctx context.Context, faq *domain.FAQ) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.FAQ, error)
	List(ctx context.Context, filter domain.FAQFilter) (*domain.PaginatedResult[domain.FAQ], error)
	SearchByKeywords(ctx context.Context, query string, limit int) ([]domain.FAQMatch, error)
	ListActiveByCategory(ctx context.Context, category domain.FAQCategory) ([]domain.FAQ, error)
}

// AdminNotificationRepository manages admin notifications.
type AdminNotificationRepository interface {
	Create(ctx context.Context, notif *domain.AdminNotification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AdminNotification, error)
	List(ctx context.Context, filter domain.AdminNotificationFilter) (*domain.PaginatedResult[domain.AdminNotification], error)
	MarkAsRead(ctx context.Context, id uuid.UUID, readBy uuid.UUID) error
	MarkAllAsReadByRole(ctx context.Context, role domain.AdminSubRole, readBy uuid.UUID) error
	CountUnreadByRole(ctx context.Context, role domain.AdminSubRole) (int64, error)
	ListUnreadCriticalByRole(ctx context.Context, role domain.AdminSubRole) ([]domain.AdminNotification, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// WebhookRepository manages owner outgoing webhooks.
type WebhookRepository interface {
	Create(ctx context.Context, webhook *domain.Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	Update(ctx context.Context, webhook *domain.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.Webhook], error)
	CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error)
	ListActiveByEvent(ctx context.Context, bathhouseOwnerID uuid.UUID, event domain.WebhookEventType) ([]domain.Webhook, error)
}

// WebhookDeliveryRepository manages webhook delivery attempts.
type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery *domain.WebhookDelivery) error
	Update(ctx context.Context, delivery *domain.WebhookDelivery) error
	ListByWebhook(ctx context.Context, webhookID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.WebhookDelivery], error)
	ListPendingRetries(ctx context.Context, before time.Time) ([]domain.WebhookDelivery, error)
}

// PMSConnectionRepository manages PMS integration connections.
type PMSConnectionRepository interface {
	Create(ctx context.Context, conn *domain.PMSConnection) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PMSConnection, error)
	Update(ctx context.Context, conn *domain.PMSConnection) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSConnection], error)
	GetByBathhouseID(ctx context.Context, bathhouseID uuid.UUID) (*domain.PMSConnection, error)
	ListActive(ctx context.Context) ([]domain.PMSConnection, error)
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, lastSyncAt time.Time, lastSyncError string, status domain.PMSConnectionStatus) error
}

// PMSSyncLogRepository manages PMS sync log entries.
type PMSSyncLogRepository interface {
	Create(ctx context.Context, log *domain.PMSSyncLog) error
	ListByConnection(ctx context.Context, connectionID uuid.UUID, page, pageSize int) (*domain.PaginatedResult[domain.PMSSyncLog], error)
}

// PhotoOrderRepository manages professional photography orders.
type PhotoOrderRepository interface {
	Create(ctx context.Context, order *domain.PhotoOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.PhotoOrder, error)
	Update(ctx context.Context, order *domain.PhotoOrder) error
	List(ctx context.Context, filter domain.PhotoOrderFilter) (*domain.PaginatedResult[domain.PhotoOrder], error)
}

// StoplistRepository manages the antifraud stoplist.
type StoplistRepository interface {
	Create(ctx context.Context, entry *domain.StoplistEntry) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter domain.StoplistFilter) (*domain.PaginatedResult[domain.StoplistEntry], error)
	IsBlocked(ctx context.Context, phone, email, inn, bankCardNumber string) (bool, error)
	CountDuplicateOwners(ctx context.Context, phone, email, inn string, excludeUserID uuid.UUID) (int, error)
}
