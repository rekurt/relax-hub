package mock

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

func TestWebhookRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()
	repo := NewWebhookRepo()
	ownerID := uuid.New()

	t.Run("CreateGetUpdateDelete", func(t *testing.T) {
		hook := &domain.Webhook{
			OwnerID:  ownerID,
			URL:      "https://example.com/wh",
			Secret:   "s",
			Events:   []domain.WebhookEventType{domain.WebhookEventBookingCreated},
			IsActive: true,
		}
		if err := repo.Create(ctx, hook); err != nil {
			t.Fatal(err)
		}
		got, err := repo.GetByID(ctx, hook.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.URL != hook.URL {
			t.Errorf("URL: got %q", got.URL)
		}

		if _, err := repo.GetByID(ctx, uuid.New()); err != domain.ErrWebhookNotFound {
			t.Errorf("GetByID missing: %v", err)
		}

		got.URL = "https://example.com/v2"
		got.Events = append(got.Events, domain.WebhookEventBookingCancelled)
		if err := repo.Update(ctx, got); err != nil {
			t.Fatal(err)
		}
		got2, _ := repo.GetByID(ctx, hook.ID)
		if got2.URL != "https://example.com/v2" {
			t.Errorf("Update URL: %q", got2.URL)
		}
		if len(got2.Events) != 2 {
			t.Errorf("Update Events: %d", len(got2.Events))
		}

		ghost := &domain.Webhook{ID: uuid.New()}
		if err := repo.Update(ctx, ghost); err != domain.ErrWebhookNotFound {
			t.Errorf("Update missing: %v", err)
		}

		if err := repo.Delete(ctx, hook.ID); err != nil {
			t.Fatal(err)
		}
		if err := repo.Delete(ctx, hook.ID); err != domain.ErrWebhookNotFound {
			t.Errorf("Delete twice: %v", err)
		}
	})

	t.Run("ListActiveByEvent", func(t *testing.T) {
		fresh := NewWebhookRepo()
		oID := uuid.New()
		_ = fresh.Create(ctx, &domain.Webhook{
			OwnerID: oID, URL: "https://example.com/a",
			Events: []domain.WebhookEventType{domain.WebhookEventBookingCreated},
			IsActive: true, Secret: "s",
		})
		_ = fresh.Create(ctx, &domain.Webhook{
			OwnerID: oID, URL: "https://example.com/b",
			Events: []domain.WebhookEventType{domain.WebhookEventPaymentReceived},
			IsActive: false, Secret: "s",
		})
		_ = fresh.Create(ctx, &domain.Webhook{
			OwnerID: uuid.New(), URL: "https://example.com/c",
			Events: []domain.WebhookEventType{domain.WebhookEventBookingCreated},
			IsActive: true, Secret: "s",
		})

		active, err := fresh.ListActiveByEvent(ctx, oID, domain.WebhookEventBookingCreated)
		if err != nil {
			t.Fatal(err)
		}
		if len(active) != 1 {
			t.Errorf("Active by event: %d, want 1", len(active))
		}

		// inactive event
		none, _ := fresh.ListActiveByEvent(ctx, oID, domain.WebhookEventBookingCancelled)
		if len(none) != 0 {
			t.Errorf("Inactive event: %d", len(none))
		}
	})

	t.Run("ListByOwnerAndCount", func(t *testing.T) {
		fresh := NewWebhookRepo()
		oID := uuid.New()
		for i := 0; i < 3; i++ {
			_ = fresh.Create(ctx, &domain.Webhook{
				OwnerID: oID, URL: "https://example.com/x",
				Events:   []domain.WebhookEventType{domain.WebhookEventBookingCreated},
				IsActive: true, Secret: "s",
			})
		}
		_ = fresh.Create(ctx, &domain.Webhook{
			OwnerID: uuid.New(), URL: "https://example.com/o",
			Events:   []domain.WebhookEventType{domain.WebhookEventBookingCreated},
			IsActive: true, Secret: "s",
		})

		page, _ := fresh.ListByOwner(ctx, oID, 1, 2)
		if page.TotalCount != 3 || len(page.Items) != 2 {
			t.Errorf("ListByOwner page1: got total=%d items=%d", page.TotalCount, len(page.Items))
		}
		page, _ = fresh.ListByOwner(ctx, oID, 2, 2)
		if len(page.Items) != 1 {
			t.Errorf("ListByOwner page2: %d", len(page.Items))
		}
		// out of range
		page, _ = fresh.ListByOwner(ctx, oID, 5, 10)
		if len(page.Items) != 0 {
			t.Errorf("ListByOwner out of range: %d", len(page.Items))
		}

		count, err := fresh.CountByOwner(ctx, oID)
		if err != nil {
			t.Fatal(err)
		}
		if count != 3 {
			t.Errorf("CountByOwner: %d", count)
		}
	})
}

func TestWebhookDeliveryRepo_FullCoverage(t *testing.T) {
	ctx := context.Background()
	repo := NewWebhookDeliveryRepo()
	hookID := uuid.New()

	t.Run("CreateAndUpdate", func(t *testing.T) {
		next := time.Now().Add(time.Minute)
		d := &domain.WebhookDelivery{
			WebhookID:    hookID,
			EventType:    domain.WebhookEventBookingCreated,
			Payload:      []byte(`{"k":"v"}`),
			Status:       domain.WebhookDeliveryPending,
			AttemptCount: 0,
			NextRetryAt:  &next,
		}
		if err := repo.Create(ctx, d); err != nil {
			t.Fatal(err)
		}

		d.Status = domain.WebhookDeliverySuccess
		d.HTTPStatus = 200
		d.AttemptCount = 1
		if err := repo.Update(ctx, d); err != nil {
			t.Fatal(err)
		}

		ghost := &domain.WebhookDelivery{ID: uuid.New()}
		if err := repo.Update(ctx, ghost); err != domain.ErrNotFound {
			t.Errorf("Update missing: %v", err)
		}
	})

	t.Run("ListByWebhookAndPendingRetries", func(t *testing.T) {
		fresh := NewWebhookDeliveryRepo()
		past := time.Now().Add(-time.Minute)
		future := time.Now().Add(time.Hour)
		_ = fresh.Create(ctx, &domain.WebhookDelivery{
			WebhookID: hookID, EventType: domain.WebhookEventBookingCreated,
			Status: domain.WebhookDeliveryPending, NextRetryAt: &past,
		})
		_ = fresh.Create(ctx, &domain.WebhookDelivery{
			WebhookID: hookID, EventType: domain.WebhookEventBookingCreated,
			Status: domain.WebhookDeliveryFailed, NextRetryAt: &future,
		})
		_ = fresh.Create(ctx, &domain.WebhookDelivery{
			WebhookID: uuid.New(), EventType: domain.WebhookEventBookingCreated,
			Status: domain.WebhookDeliveryPending, NextRetryAt: &past,
		})

		page, _ := fresh.ListByWebhook(ctx, hookID, 1, 10)
		if page.TotalCount != 2 {
			t.Errorf("ListByWebhook: %d", page.TotalCount)
		}
		// out of range
		page, _ = fresh.ListByWebhook(ctx, hookID, 5, 10)
		if len(page.Items) != 0 {
			t.Errorf("ListByWebhook out of range: %d", len(page.Items))
		}

		ready, _ := fresh.ListPendingRetries(ctx, time.Now())
		// 2 across both webhooks (past + past for diff webhook); other-webhook is owner-agnostic.
		if len(ready) != 2 {
			t.Errorf("ListPendingRetries: %d, want 2", len(ready))
		}

		// future scope
		none, _ := fresh.ListPendingRetries(ctx, time.Now().Add(-2*time.Hour))
		if len(none) != 0 {
			t.Errorf("ListPendingRetries past: %d", len(none))
		}
	})
}
