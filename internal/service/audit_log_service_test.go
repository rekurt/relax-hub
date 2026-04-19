package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type auditLogTestEnv struct {
	svc service.AuditLogService
}

func newAuditLogTestEnv() *auditLogTestEnv {
	repo := mock.NewAuditLogRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewAuditLogService(repo, log)
	return &auditLogTestEnv{svc: svc}
}

func TestAuditLogService_LogChange(t *testing.T) {
	env := newAuditLogTestEnv()
	entityID := uuid.New()
	userID := uuid.New()

	err := env.svc.LogChange(context.Background(), "bathhouse", entityID, userID, domain.AuditActionUpdate, map[string]interface{}{
		"name": map[string]string{"old": "Old Name", "new": "New Name"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := env.svc.GetHistory(context.Background(), "bathhouse", entityID, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("total_count = %d, want 1", result.TotalCount)
	}
	if result.Items[0].Action != domain.AuditActionUpdate {
		t.Errorf("action = %v, want update", result.Items[0].Action)
	}
	if result.Items[0].EntityType != "bathhouse" {
		t.Errorf("entity_type = %v, want bathhouse", result.Items[0].EntityType)
	}
}

func TestAuditLogService_GetHistory_Empty(t *testing.T) {
	env := newAuditLogTestEnv()

	result, err := env.svc.GetHistory(context.Background(), "bathhouse", uuid.New(), 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("total_count = %d, want 0", result.TotalCount)
	}
}

func TestAuditLogService_ListAll(t *testing.T) {
	env := newAuditLogTestEnv()
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		_ = env.svc.LogChange(context.Background(), "bathhouse", uuid.New(), userID, domain.AuditActionUpdate, nil)
	}

	action := domain.AuditActionUpdate
	result, err := env.svc.ListAll(context.Background(), domain.AuditLogFilter{
		Action:   &action,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
}

func TestAuditLogService_IsSubstantialChange_AddressChange(t *testing.T) {
	env := newAuditLogTestEnv()

	old := &domain.Bathhouse{Address: "Old Address", CityID: 1, Images: []string{"a.jpg"}}
	new := &domain.Bathhouse{Address: "New Address", CityID: 1, Images: []string{"a.jpg"}}

	if !env.svc.IsSubstantialChange(old, new) {
		t.Error("expected address change to be substantial")
	}
}

func TestAuditLogService_IsSubstantialChange_CityChange(t *testing.T) {
	env := newAuditLogTestEnv()

	old := &domain.Bathhouse{Address: "Same", CityID: 1, Images: []string{"a.jpg"}}
	new := &domain.Bathhouse{Address: "Same", CityID: 2, Images: []string{"a.jpg"}}

	if !env.svc.IsSubstantialChange(old, new) {
		t.Error("expected city change to be substantial")
	}
}

func TestAuditLogService_IsSubstantialChange_PhotosReplaced(t *testing.T) {
	env := newAuditLogTestEnv()

	old := &domain.Bathhouse{Address: "Same", CityID: 1, Images: []string{"a.jpg", "b.jpg", "c.jpg", "d.jpg"}}
	// More than 50% replaced (3 of 4 old images gone)
	new := &domain.Bathhouse{Address: "Same", CityID: 1, Images: []string{"a.jpg", "x.jpg", "y.jpg", "z.jpg"}}

	if !env.svc.IsSubstantialChange(old, new) {
		t.Error("expected >50%% photo replacement to be substantial")
	}
}

func TestAuditLogService_IsSubstantialChange_MinorChange(t *testing.T) {
	env := newAuditLogTestEnv()

	old := &domain.Bathhouse{Address: "Same", CityID: 1, Images: []string{"a.jpg", "b.jpg"}}
	new := &domain.Bathhouse{Address: "Same", CityID: 1, Images: []string{"a.jpg", "b.jpg"}}

	if env.svc.IsSubstantialChange(old, new) {
		t.Error("expected no substantial change for identical data")
	}
}

func TestAuditLogService_IsSubstantialChange_PhotosBelowThreshold(t *testing.T) {
	env := newAuditLogTestEnv()

	old := &domain.Bathhouse{Address: "Same", CityID: 1, Images: []string{"a.jpg", "b.jpg", "c.jpg", "d.jpg"}}
	// 1 of 4 replaced (25%) - not substantial
	new := &domain.Bathhouse{Address: "Same", CityID: 1, Images: []string{"a.jpg", "b.jpg", "c.jpg", "x.jpg"}}

	if env.svc.IsSubstantialChange(old, new) {
		t.Error("expected <50%% photo replacement to not be substantial")
	}
}

func TestAuditLogService_ListAdminActions(t *testing.T) {
	env := newAuditLogTestEnv()
	adminID := uuid.New()

	// Create some admin_action logs
	for i := 0; i < 3; i++ {
		_ = env.svc.LogChange(context.Background(), "admin_action", uuid.New(), adminID, domain.AuditActionUpdate, map[string]interface{}{
			"method": "PATCH",
			"path":   "/admin/users/" + uuid.New().String() + "/block",
		})
	}

	// Create a non-admin log
	_ = env.svc.LogChange(context.Background(), "bathhouse", uuid.New(), uuid.New(), domain.AuditActionUpdate, nil)

	// ListAdminActions should only return admin_action entries
	result, err := env.svc.ListAdminActions(context.Background(), domain.AuditLogFilter{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
	for _, item := range result.Items {
		if item.EntityType != "admin_action" {
			t.Errorf("entity_type = %v, want admin_action", item.EntityType)
		}
	}
}

func TestAuditLogService_ListAdminActions_FilterByAdmin(t *testing.T) {
	env := newAuditLogTestEnv()
	admin1 := uuid.New()
	admin2 := uuid.New()

	_ = env.svc.LogChange(context.Background(), "admin_action", uuid.New(), admin1, domain.AuditActionUpdate, nil)
	_ = env.svc.LogChange(context.Background(), "admin_action", uuid.New(), admin1, domain.AuditActionCreate, nil)
	_ = env.svc.LogChange(context.Background(), "admin_action", uuid.New(), admin2, domain.AuditActionDelete, nil)

	result, err := env.svc.ListAdminActions(context.Background(), domain.AuditLogFilter{
		UserID:   &admin1,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("total_count = %d, want 2", result.TotalCount)
	}
}
