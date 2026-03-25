package cron

import (
	"context"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/service"
)

// mockTicketServiceForCron implements service.TicketService for cron tests.
type mockTicketServiceForCron struct {
	autoEscalateCalled bool
	autoCloseCalled    bool
}

func (m *mockTicketServiceForCron) CreateTicket(_ context.Context, _ uuid.UUID, _ *domain.Ticket, _ string) error {
	return nil
}
func (m *mockTicketServiceForCron) GetTicket(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) (*domain.Ticket, error) {
	return nil, nil
}
func (m *mockTicketServiceForCron) ListUserTickets(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Ticket], error) {
	return nil, nil
}
func (m *mockTicketServiceForCron) ListAllTickets(_ context.Context, _ domain.TicketFilter) (*domain.PaginatedResult[domain.Ticket], error) {
	return nil, nil
}
func (m *mockTicketServiceForCron) AddMessage(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string, _ []string) (*domain.TicketMessage, error) {
	return nil, nil
}
func (m *mockTicketServiceForCron) ListMessages(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) ([]domain.TicketMessage, error) {
	return nil, nil
}
func (m *mockTicketServiceForCron) AssignTicket(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockTicketServiceForCron) EscalateTicket(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockTicketServiceForCron) ResolveTicket(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockTicketServiceForCron) CloseTicket(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockTicketServiceForCron) SubmitCSAT(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int) error {
	return nil
}
func (m *mockTicketServiceForCron) GetStats(_ context.Context) (*domain.TicketStatusCounts, error) {
	return nil, nil
}
func (m *mockTicketServiceForCron) AutoEscalateStaleTickets(_ context.Context) error {
	m.autoEscalateCalled = true
	return nil
}
func (m *mockTicketServiceForCron) AutoCloseResolvedTickets(_ context.Context) error {
	m.autoCloseCalled = true
	return nil
}

// mockEscrowServiceForCron implements service.EscrowService for cron tests.
type mockEscrowServiceForCron struct {
	released            int
	processMaturedCalled bool
}

func (m *mockEscrowServiceForCron) CreateEscrow(_ context.Context, _ uuid.UUID, _, _ int64) (*domain.Escrow, error) {
	return nil, nil
}
func (m *mockEscrowServiceForCron) ReleaseToOwner(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockEscrowServiceForCron) MarkDisputed(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockEscrowServiceForCron) MarkDisputedByBookingID(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockEscrowServiceForCron) ProcessRefund(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}
func (m *mockEscrowServiceForCron) ProcessRefundByBookingID(_ context.Context, _ uuid.UUID, _ int64) error {
	return nil
}
func (m *mockEscrowServiceForCron) ProcessMaturedEscrows(_ context.Context) (int, error) {
	m.processMaturedCalled = true
	return m.released, nil
}

// mockReviewServiceForCron implements service.ReviewService for cron tests.
type mockReviewServiceForCron struct {
	refreshPlatformCalled    bool
	sendReviewRequestsCalled bool
	sendReviewRequestsDelay  int
}

func (m *mockReviewServiceForCron) Create(_ context.Context, _ uuid.UUID, _ service.CreateReviewInput) (*domain.Review, error) {
	return nil, nil
}
func (m *mockReviewServiceForCron) GetByID(_ context.Context, _ uuid.UUID) (*domain.Review, error) {
	return nil, nil
}
func (m *mockReviewServiceForCron) Update(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ service.UpdateReviewInput) (*domain.Review, error) {
	return nil, nil
}
func (m *mockReviewServiceForCron) Delete(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID) error {
	return nil
}
func (m *mockReviewServiceForCron) ListByBathhouse(_ context.Context, _ uuid.UUID, _, _ int) (*domain.PaginatedResult[domain.Review], error) {
	return nil, nil
}
func (m *mockReviewServiceForCron) AddOwnerResponse(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ uuid.UUID, _ string) (*domain.Review, error) {
	return nil, nil
}
func (m *mockReviewServiceForCron) GetCriteriaAverages(_ context.Context, _ uuid.UUID) (*domain.ReviewCriteriaAverages, error) {
	return nil, nil
}
func (m *mockReviewServiceForCron) RecalculateBayesianRating(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockReviewServiceForCron) RefreshPlatformAverage(_ context.Context) error {
	m.refreshPlatformCalled = true
	return nil
}
func (m *mockReviewServiceForCron) SendReviewRequests(_ context.Context, delayHours int) (int, error) {
	m.sendReviewRequestsCalled = true
	m.sendReviewRequestsDelay = delayHours
	return 0, nil
}
func (m *mockReviewServiceForCron) ListAllReviews(_ context.Context, _ domain.AdminReviewFilter) (*domain.PaginatedResult[domain.Review], error) {
	return nil, nil
}
func (m *mockReviewServiceForCron) CountPendingReviews(_ context.Context) (int64, error) {
	return 0, nil
}
func (m *mockReviewServiceForCron) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.ReviewStatus) error {
	return nil
}
func (m *mockReviewServiceForCron) UpdateStatusWithReasons(_ context.Context, _ uuid.UUID, _ domain.ReviewStatus, _ []string) error {
	return nil
}

// mockAutoScenarioServiceForCron implements service.AutoScenarioService for cron tests.
type mockAutoScenarioServiceForCron struct {
	executeCalled bool
}

func (m *mockAutoScenarioServiceForCron) ListScenarios(_ context.Context, _ uuid.UUID, _ domain.UserRole) ([]domain.AutoScenario, error) {
	return nil, nil
}
func (m *mockAutoScenarioServiceForCron) UpdateScenario(_ context.Context, _ uuid.UUID, _ domain.UserRole, _ *domain.AutoScenario) error {
	return nil
}
func (m *mockAutoScenarioServiceForCron) ExecuteScenarios(_ context.Context) (int, error) {
	m.executeCalled = true
	return 0, nil
}
