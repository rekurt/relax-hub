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

func newTemplateTestService() (service.TemplateService, *mock.TemplateRepo) {
	repo := mock.NewTemplateRepo().(*mock.TemplateRepo)
	log := logger.New(logger.LevelWarn)
	return service.NewTemplateService(repo, log), repo
}

func TestTemplateService_List_SeedsDefaults(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	templates, err := svc.List(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(templates) != len(domain.DefaultTemplates) {
		t.Fatalf("expected %d default templates, got %d", len(domain.DefaultTemplates), len(templates))
	}

	for i, tmpl := range templates {
		if tmpl.Title != domain.DefaultTemplates[i].Title {
			t.Errorf("template %d: expected title=%q, got %q", i, domain.DefaultTemplates[i].Title, tmpl.Title)
		}
		if !tmpl.IsDefault {
			t.Errorf("template %d: expected is_default=true", i)
		}
		if tmpl.OwnerID != ownerID {
			t.Errorf("template %d: expected owner_id=%s, got %s", i, ownerID, tmpl.OwnerID)
		}
	}
}

func TestTemplateService_List_NoReseedAfterFirstCall(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// First call seeds defaults
	_, _ = svc.List(ctx, ownerID, domain.RoleOwner)

	// Second call should return same templates, not re-seed
	templates, err := svc.List(ctx, ownerID, domain.RoleOwner)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(templates) != len(domain.DefaultTemplates) {
		t.Errorf("expected %d templates on second call, got %d", len(domain.DefaultTemplates), len(templates))
	}
}

func TestTemplateService_Create(t *testing.T) {
	svc, repo := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	template := &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Мой шаблон",
		Body:  "Спасибо за визит!",
	}

	err := svc.Create(ctx, ownerID, domain.RoleOwner, template)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if template.OwnerID != ownerID {
		t.Errorf("expected owner_id=%s, got %s", ownerID, template.OwnerID)
	}

	stored, err := repo.GetByID(ctx, template.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if stored.Title != "Мой шаблон" {
		t.Errorf("expected title='Мой шаблон', got %s", stored.Title)
	}
}

func TestTemplateService_Create_ForbiddenForClient(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()

	template := &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Test",
		Body:  "Test body",
	}

	err := svc.Create(ctx, uuid.New(), domain.RoleClient, template)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTemplateService_Create_InvalidInput(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()

	// Empty title
	err := svc.Create(ctx, uuid.New(), domain.RoleOwner, &domain.ResponseTemplate{
		ID:   uuid.New(),
		Body: "Body",
	})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for empty title, got %v", err)
	}

	// Empty body
	err = svc.Create(ctx, uuid.New(), domain.RoleOwner, &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Title",
	})
	if err != domain.ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput for empty body, got %v", err)
	}
}

func TestTemplateService_Create_LimitReached(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create MaxTemplatesPerOwner templates
	for i := 0; i < domain.MaxTemplatesPerOwner; i++ {
		err := svc.Create(ctx, ownerID, domain.RoleOwner, &domain.ResponseTemplate{
			ID:    uuid.New(),
			Title: "Template",
			Body:  "Body",
		})
		if err != nil {
			t.Fatalf("Create #%d: %v", i+1, err)
		}
	}

	// Next one should fail
	err := svc.Create(ctx, ownerID, domain.RoleOwner, &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Over limit",
		Body:  "Body",
	})
	if err != domain.ErrTemplateLimitReached {
		t.Errorf("expected ErrTemplateLimitReached, got %v", err)
	}
}

func TestTemplateService_Update(t *testing.T) {
	svc, repo := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	template := &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Original",
		Body:  "Original body",
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, template)

	err := svc.Update(ctx, ownerID, domain.RoleOwner, &domain.ResponseTemplate{
		ID:        template.ID,
		Title:     "Updated",
		Body:      "Updated body",
		SortOrder: 5,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	stored, _ := repo.GetByID(ctx, template.ID)
	if stored.Title != "Updated" {
		t.Errorf("expected title='Updated', got %s", stored.Title)
	}
	if stored.Body != "Updated body" {
		t.Errorf("expected body='Updated body', got %s", stored.Body)
	}
	if stored.SortOrder != 5 {
		t.Errorf("expected sort_order=5, got %d", stored.SortOrder)
	}
}

func TestTemplateService_Update_NotFound(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()

	err := svc.Update(ctx, uuid.New(), domain.RoleOwner, &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Test",
		Body:  "Body",
	})
	if err != domain.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestTemplateService_Update_ForbiddenOtherOwner(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	template := &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Test",
		Body:  "Body",
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, template)

	err := svc.Update(ctx, uuid.New(), domain.RoleOwner, &domain.ResponseTemplate{
		ID:    template.ID,
		Title: "Hijacked",
		Body:  "Body",
	})
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTemplateService_Delete(t *testing.T) {
	svc, repo := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	template := &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "To delete",
		Body:  "Body",
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, template)

	err := svc.Delete(ctx, ownerID, domain.RoleOwner, template.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, template.ID)
	if err != domain.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound after delete, got %v", err)
	}
}

func TestTemplateService_Delete_NotFound(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()

	err := svc.Delete(ctx, uuid.New(), domain.RoleOwner, uuid.New())
	if err != domain.ErrTemplateNotFound {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestTemplateService_Delete_ForbiddenOtherOwner(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	template := &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Test",
		Body:  "Body",
	}
	_ = svc.Create(ctx, ownerID, domain.RoleOwner, template)

	err := svc.Delete(ctx, uuid.New(), domain.RoleOwner, template.ID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTemplateService_List_ForbiddenForClient(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()

	_, err := svc.List(ctx, uuid.New(), domain.RoleClient)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTemplateService_RepresentativeCanManage(t *testing.T) {
	svc, _ := newTemplateTestService()
	ctx := context.Background()
	repID := uuid.New()

	template := &domain.ResponseTemplate{
		ID:    uuid.New(),
		Title: "Rep template",
		Body:  "Body",
	}

	err := svc.Create(ctx, repID, domain.RoleRepresentative, template)
	if err != nil {
		t.Fatalf("Create as representative: %v", err)
	}

	templates, err := svc.List(ctx, repID, domain.RoleRepresentative)
	if err != nil {
		t.Fatalf("List as representative: %v", err)
	}
	// Should have 1 created + 0 defaults (because we already have templates)
	// Actually we created 1, so List won't seed defaults
	if len(templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(templates))
	}
}
