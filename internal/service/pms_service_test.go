package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/pms"
	"github.com/rekurt/relax-hub/internal/repository"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPMSProvider implements pms.PMSProvider for testing.
type mockPMSProvider struct {
	mu           sync.Mutex
	name         domain.PMSProvider
	testErr      error
	pullBookings []domain.PMSBooking
	pullErr      error
	pushCalls    []domain.PMSBooking
	pushErr      error
}

func (m *mockPMSProvider) Name() domain.PMSProvider { return m.name }

func (m *mockPMSProvider) TestConnection(_ context.Context, _ string) error {
	return m.testErr
}

func (m *mockPMSProvider) PullBookings(_ context.Context, _ string, _ string, _, _ time.Time) ([]domain.PMSBooking, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pullErr != nil {
		return nil, m.pullErr
	}
	return m.pullBookings, nil
}

func (m *mockPMSProvider) PushBooking(_ context.Context, _ string, _ string, booking domain.PMSBooking) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pushCalls = append(m.pushCalls, booking)
	return m.pushErr
}

func (m *mockPMSProvider) SyncSchedule(_ context.Context, _ string, _ string, _, _ time.Time) ([]domain.PMSScheduleSlot, error) {
	return nil, nil
}

func (m *mockPMSProvider) CancelBooking(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

type pmsTestEnv struct {
	svc         service.PMSService
	connRepo    *mock.PMSConnectionRepo
	syncLogRepo *mock.PMSSyncLogRepo
	bhRepo      repository.BathhouseRepository
}

func newPMSTestEnv(provider *mockPMSProvider) *pmsTestEnv {
	connRepo := mock.NewPMSConnectionRepo().(*mock.PMSConnectionRepo)
	syncLogRepo := mock.NewPMSSyncLogRepo().(*mock.PMSSyncLogRepo)
	bhRepo := mock.NewBathhouseRepo()
	repRepo := mock.NewRepresentativeRepo()
	access := service.NewAccessChecker(repRepo, bhRepo)
	registry := pms.NewProviderRegistry([]pms.PMSProvider{provider})
	log := logger.New(logger.LevelWarn)
	svc := service.NewPMSService(connRepo, syncLogRepo, access, registry, log)
	return &pmsTestEnv{svc: svc, connRepo: connRepo, syncLogRepo: syncLogRepo, bhRepo: bhRepo}
}

func (e *pmsTestEnv) seedBathhouse(ctx context.Context, ownerID, bathhouseID uuid.UUID) {
	_ = e.bhRepo.Create(ctx, &domain.Bathhouse{
		ID:      bathhouseID,
		OwnerID: ownerID,
		Name:    "Test Bathhouse",
		Status:  domain.BathhouseStatusActive,
	})
}

func validPMSConnection(ownerID, bathhouseID uuid.UUID) *domain.PMSConnection {
	return &domain.PMSConnection{
		ID:                   uuid.New(),
		OwnerID:              ownerID,
		BathhouseID:          bathhouseID,
		Provider:             domain.PMSProviderYclients,
		CredentialsEncrypted: `{"partner_token":"test","user_token":"test"}`,
		SyncDirection:        domain.PMSSyncDirectionBoth,
		SyncIntervalMinutes:  15,
		ExternalID:           "ext-123",
	}
}

func TestPMSService_Create(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	err := env.svc.Create(ctx, ownerID, domain.RoleOwner, conn)
	require.NoError(t, err)
	assert.Equal(t, domain.PMSConnectionActive, conn.Status)
	assert.Equal(t, ownerID, conn.OwnerID)

	stored, err := env.connRepo.GetByID(ctx, conn.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PMSProviderYclients, stored.Provider)
}

func TestPMSService_Create_DuplicateBathhouse(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	conn2 := validPMSConnection(ownerID, bathhouseID)
	err := env.svc.Create(ctx, ownerID, domain.RoleOwner, conn2)
	assert.ErrorIs(t, err, domain.ErrPMSConnectionAlreadyExists)
}

func TestPMSService_Create_ForbiddenForClient(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	conn := validPMSConnection(uuid.New(), uuid.New())
	err := env.svc.Create(ctx, uuid.New(), domain.RoleClient, conn)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestPMSService_Create_InvalidProvider(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	conn.Provider = "unknown_pms"
	err := env.svc.Create(ctx, ownerID, domain.RoleOwner, conn)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestPMSService_GetByID(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	got, err := env.svc.GetByID(ctx, ownerID, domain.RoleOwner, conn.ID)
	require.NoError(t, err)
	assert.Equal(t, conn.ID, got.ID)
}

func TestPMSService_GetByID_NotFound(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	_, err := env.svc.GetByID(ctx, uuid.New(), domain.RoleOwner, uuid.New())
	assert.ErrorIs(t, err, domain.ErrPMSConnectionNotFound)
}

func TestPMSService_Update(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	conn.SyncIntervalMinutes = 30
	conn.SyncDirection = domain.PMSSyncDirectionInbound
	err := env.svc.Update(ctx, ownerID, domain.RoleOwner, conn)
	require.NoError(t, err)

	updated, err := env.connRepo.GetByID(ctx, conn.ID)
	require.NoError(t, err)
	assert.Equal(t, 30, updated.SyncIntervalMinutes)
	assert.Equal(t, domain.PMSSyncDirectionInbound, updated.SyncDirection)
}

func TestPMSService_Delete(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	err := env.svc.Delete(ctx, ownerID, domain.RoleOwner, conn.ID)
	require.NoError(t, err)

	_, err = env.connRepo.GetByID(ctx, conn.ID)
	assert.ErrorIs(t, err, domain.ErrPMSConnectionNotFound)
}

func TestPMSService_List(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bh1 := uuid.New()
	bh2 := uuid.New()
	env.seedBathhouse(ctx, ownerID, bh1)
	env.seedBathhouse(ctx, ownerID, bh2)

	conn1 := validPMSConnection(ownerID, bh1)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn1))

	conn2 := validPMSConnection(ownerID, bh2)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn2))

	result, err := env.svc.List(ctx, ownerID, domain.RoleOwner, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.TotalCount)
}

func TestPMSService_List_ForbiddenForClient(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	_, err := env.svc.List(ctx, uuid.New(), domain.RoleClient, 1, 10)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestPMSService_TestConnection_Success(t *testing.T) {
	provider := &mockPMSProvider{name: domain.PMSProviderYclients, testErr: nil}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	err := env.svc.TestConnection(ctx, ownerID, domain.RoleOwner, conn.ID)
	assert.NoError(t, err)
}

func TestPMSService_TestConnection_Failure(t *testing.T) {
	provider := &mockPMSProvider{
		name:    domain.PMSProviderYclients,
		testErr: errors.New("invalid credentials"),
	}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	err := env.svc.TestConnection(ctx, ownerID, domain.RoleOwner, conn.ID)
	assert.ErrorIs(t, err, domain.ErrPMSSyncFailed)
}

func TestPMSService_TriggerSync_Success(t *testing.T) {
	provider := &mockPMSProvider{
		name: domain.PMSProviderYclients,
		pullBookings: []domain.PMSBooking{
			{ExternalID: "b1", GuestName: "Иван", StartTime: time.Now(), EndTime: time.Now().Add(2 * time.Hour)},
			{ExternalID: "b2", GuestName: "Мария", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)},
		},
	}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	err := env.svc.TriggerSync(ctx, ownerID, domain.RoleOwner, conn.ID)
	require.NoError(t, err)

	updated, _ := env.connRepo.GetByID(ctx, conn.ID)
	assert.NotNil(t, updated.LastSyncAt)
	assert.Equal(t, domain.PMSConnectionActive, updated.Status)
	assert.Empty(t, updated.LastSyncError)

	logs, err := env.syncLogRepo.ListByConnection(ctx, conn.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), logs.TotalCount)
	assert.Equal(t, "success", logs.Items[0].Status)
	assert.Equal(t, 2, logs.Items[0].ItemsSynced)
}

func TestPMSService_TriggerSync_PullError(t *testing.T) {
	provider := &mockPMSProvider{
		name:    domain.PMSProviderYclients,
		pullErr: errors.New("api timeout"),
	}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	err := env.svc.TriggerSync(ctx, ownerID, domain.RoleOwner, conn.ID)
	assert.Error(t, err)

	updated, _ := env.connRepo.GetByID(ctx, conn.ID)
	assert.Equal(t, domain.PMSConnectionError, updated.Status)
	assert.Contains(t, updated.LastSyncError, "api timeout")

	logs, _ := env.syncLogRepo.ListByConnection(ctx, conn.ID, 1, 10)
	assert.Equal(t, int64(1), logs.TotalCount)
	assert.Equal(t, "error", logs.Items[0].Status)
}

func TestPMSService_SyncAllActive(t *testing.T) {
	provider := &mockPMSProvider{
		name:         domain.PMSProviderYclients,
		pullBookings: []domain.PMSBooking{{ExternalID: "b1"}},
	}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bh1 := uuid.New()
	bh2 := uuid.New()
	env.seedBathhouse(ctx, ownerID, bh1)
	env.seedBathhouse(ctx, ownerID, bh2)

	conn1 := validPMSConnection(ownerID, bh1)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn1))

	conn2 := validPMSConnection(ownerID, bh2)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn2))

	err := env.svc.SyncAllActive(ctx)
	require.NoError(t, err)

	logs1, _ := env.syncLogRepo.ListByConnection(ctx, conn1.ID, 1, 10)
	logs2, _ := env.syncLogRepo.ListByConnection(ctx, conn2.ID, 1, 10)
	assert.Equal(t, int64(1), logs1.TotalCount)
	assert.Equal(t, int64(1), logs2.TotalCount)
}

func TestPMSService_ListSyncLogs(t *testing.T) {
	provider := &mockPMSProvider{
		name:         domain.PMSProviderYclients,
		pullBookings: []domain.PMSBooking{{ExternalID: "b1"}},
	}
	env := newPMSTestEnv(provider)
	ctx := context.Background()

	ownerID := uuid.New()
	bathhouseID := uuid.New()
	env.seedBathhouse(ctx, ownerID, bathhouseID)

	conn := validPMSConnection(ownerID, bathhouseID)
	require.NoError(t, env.svc.Create(ctx, ownerID, domain.RoleOwner, conn))

	require.NoError(t, env.svc.TriggerSync(ctx, ownerID, domain.RoleOwner, conn.ID))
	require.NoError(t, env.svc.TriggerSync(ctx, ownerID, domain.RoleOwner, conn.ID))

	logs, err := env.svc.ListSyncLogs(ctx, ownerID, domain.RoleOwner, conn.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), logs.TotalCount)
}
