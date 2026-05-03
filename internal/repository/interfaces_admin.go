package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

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

	// Heatmap (FR-153)
	GetHeatmapData(ctx context.Context, from, to time.Time, cellSize float64) ([]domain.HeatmapCell, error)

	// User activity metrics
	CountDistinctActiveUsers(ctx context.Context, from, to time.Time) (int64, error)
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

type RepresentativeRepository interface {
	Create(ctx context.Context, rep *domain.Representative) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Representative, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserAndBathhouse(ctx context.Context, userID, bathhouseID uuid.UUID) (*domain.Representative, error)
	ListByBathhouse(ctx context.Context, bathhouseID uuid.UUID) ([]domain.Representative, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Representative, error)
	ListBathhouseIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
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

type ListingDraftRepository interface {
	Create(ctx context.Context, draft *domain.ListingDraft) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ListingDraft, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.ListingDraft, error)
	UpdateStep(ctx context.Context, id uuid.UUID, step int, data json.RawMessage, currentStep int) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ListingDraftStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}
