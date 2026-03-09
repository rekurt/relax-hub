package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func TestPaymentRepo_Create(t *testing.T) {
	repo := mock.NewPaymentRepo()

	payment := &domain.Payment{
		BookingID: uuid.New(),
		UserID:    uuid.New(),
		Amount:    500000,
		Currency:  "RUB",
		Status:    domain.PaymentPending,
		Provider:  "yookassa",
	}

	err := repo.Create(context.Background(), payment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if payment.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPaymentRepo_GetByID(t *testing.T) {
	repo := mock.NewPaymentRepo()

	payment := &domain.Payment{
		BookingID: uuid.New(),
		UserID:    uuid.New(),
		Amount:    500000,
		Currency:  "RUB",
		Status:    domain.PaymentPending,
		Provider:  "yookassa",
	}
	_ = repo.Create(context.Background(), payment)

	found, err := repo.GetByID(context.Background(), payment.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Amount != payment.Amount {
		t.Errorf("amount = %d, want %d", found.Amount, payment.Amount)
	}
}

func TestPaymentRepo_GetByID_NotFound(t *testing.T) {
	repo := mock.NewPaymentRepo()

	_, err := repo.GetByID(context.Background(), uuid.New())
	if err != domain.ErrPaymentNotFound {
		t.Errorf("err = %v, want ErrPaymentNotFound", err)
	}
}

func TestPaymentRepo_GetByBookingID(t *testing.T) {
	repo := mock.NewPaymentRepo()

	bookingID := uuid.New()
	payment := &domain.Payment{
		BookingID: bookingID,
		UserID:    uuid.New(),
		Amount:    300000,
		Currency:  "RUB",
		Status:    domain.PaymentPending,
		Provider:  "yookassa",
	}
	_ = repo.Create(context.Background(), payment)

	found, err := repo.GetByBookingID(context.Background(), bookingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != payment.ID {
		t.Errorf("id = %v, want %v", found.ID, payment.ID)
	}
}

func TestPaymentRepo_GetByBookingID_NotFound(t *testing.T) {
	repo := mock.NewPaymentRepo()

	_, err := repo.GetByBookingID(context.Background(), uuid.New())
	if err != domain.ErrPaymentNotFound {
		t.Errorf("err = %v, want ErrPaymentNotFound", err)
	}
}

func TestPaymentRepo_GetByExternalID(t *testing.T) {
	repo := mock.NewPaymentRepo()

	payment := &domain.Payment{
		BookingID:  uuid.New(),
		UserID:     uuid.New(),
		Amount:     300000,
		Currency:   "RUB",
		Status:     domain.PaymentProcessing,
		Provider:   "yookassa",
		ExternalID: "ext-123",
	}
	_ = repo.Create(context.Background(), payment)

	found, err := repo.GetByExternalID(context.Background(), "ext-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != payment.ID {
		t.Errorf("id = %v, want %v", found.ID, payment.ID)
	}
}

func TestPaymentRepo_GetByExternalID_NotFound(t *testing.T) {
	repo := mock.NewPaymentRepo()

	_, err := repo.GetByExternalID(context.Background(), "nonexistent")
	if err != domain.ErrPaymentNotFound {
		t.Errorf("err = %v, want ErrPaymentNotFound", err)
	}
}

func TestPaymentRepo_UpdateStatus(t *testing.T) {
	repo := mock.NewPaymentRepo()

	payment := &domain.Payment{
		BookingID: uuid.New(),
		UserID:    uuid.New(),
		Amount:    500000,
		Currency:  "RUB",
		Status:    domain.PaymentPending,
		Provider:  "yookassa",
	}
	_ = repo.Create(context.Background(), payment)

	err := repo.UpdateStatus(context.Background(), payment.ID, domain.PaymentSucceeded, "ext-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetByID(context.Background(), payment.ID)
	if found.Status != domain.PaymentSucceeded {
		t.Errorf("status = %v, want %v", found.Status, domain.PaymentSucceeded)
	}
	if found.ExternalID != "ext-456" {
		t.Errorf("external_id = %v, want ext-456", found.ExternalID)
	}
}

func TestPaymentRepo_UpdateStatus_NotFound(t *testing.T) {
	repo := mock.NewPaymentRepo()

	err := repo.UpdateStatus(context.Background(), uuid.New(), domain.PaymentSucceeded, "ext-456")
	if err != domain.ErrPaymentNotFound {
		t.Errorf("err = %v, want ErrPaymentNotFound", err)
	}
}

func TestPaymentRepo_UpdateRefund(t *testing.T) {
	repo := mock.NewPaymentRepo()

	payment := &domain.Payment{
		BookingID: uuid.New(),
		UserID:    uuid.New(),
		Amount:    500000,
		Currency:  "RUB",
		Status:    domain.PaymentSucceeded,
		Provider:  "yookassa",
	}
	_ = repo.Create(context.Background(), payment)

	refundedAt := time.Now()
	err := repo.UpdateRefund(context.Background(), payment.ID, 500000, refundedAt, domain.PaymentRefunded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.GetByID(context.Background(), payment.ID)
	if found.RefundAmount != 500000 {
		t.Errorf("refund_amount = %d, want 500000", found.RefundAmount)
	}
	if found.Status != domain.PaymentRefunded {
		t.Errorf("status = %v, want %v", found.Status, domain.PaymentRefunded)
	}
	if found.RefundedAt == nil {
		t.Error("expected RefundedAt to be set")
	}
}

func TestPaymentRepo_UpdateRefund_NotFound(t *testing.T) {
	repo := mock.NewPaymentRepo()

	err := repo.UpdateRefund(context.Background(), uuid.New(), 500000, time.Now(), domain.PaymentRefunded)
	if err != domain.ErrPaymentNotFound {
		t.Errorf("err = %v, want ErrPaymentNotFound", err)
	}
}

func TestPaymentRepo_ListByUser(t *testing.T) {
	repo := mock.NewPaymentRepo()
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		_ = repo.Create(context.Background(), &domain.Payment{
			BookingID: uuid.New(),
			UserID:    userID,
			Amount:    int64((i + 1) * 100000),
			Currency:  "RUB",
			Status:    domain.PaymentPending,
			Provider:  "yookassa",
		})
	}

	// Unrelated payment
	_ = repo.Create(context.Background(), &domain.Payment{
		BookingID: uuid.New(),
		UserID:    uuid.New(),
		Amount:    999999,
		Currency:  "RUB",
		Status:    domain.PaymentPending,
		Provider:  "yookassa",
	})

	result, err := repo.ListByUser(context.Background(), userID, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 3 {
		t.Errorf("total_count = %d, want 3", result.TotalCount)
	}
	if len(result.Items) != 3 {
		t.Errorf("items count = %d, want 3", len(result.Items))
	}
}

func TestPaymentRepo_ListByUser_Pagination(t *testing.T) {
	repo := mock.NewPaymentRepo()
	userID := uuid.New()

	for i := 0; i < 5; i++ {
		_ = repo.Create(context.Background(), &domain.Payment{
			BookingID: uuid.New(),
			UserID:    userID,
			Amount:    int64((i + 1) * 100000),
			Currency:  "RUB",
			Status:    domain.PaymentPending,
			Provider:  "yookassa",
		})
	}

	result, err := repo.ListByUser(context.Background(), userID, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 5 {
		t.Errorf("total_count = %d, want 5", result.TotalCount)
	}
	if len(result.Items) != 2 {
		t.Errorf("items count = %d, want 2", len(result.Items))
	}
	if result.TotalPages != 3 {
		t.Errorf("total_pages = %d, want 3", result.TotalPages)
	}
}

func TestPaymentRepo_IsolatesData(t *testing.T) {
	repo := mock.NewPaymentRepo()

	payment := &domain.Payment{
		BookingID: uuid.New(),
		UserID:    uuid.New(),
		Amount:    500000,
		Currency:  "RUB",
		Status:    domain.PaymentPending,
		Provider:  "yookassa",
		Metadata:  map[string]string{"key": "value"},
	}
	_ = repo.Create(context.Background(), payment)

	found, _ := repo.GetByID(context.Background(), payment.ID)
	found.Metadata["key"] = "modified"

	found2, _ := repo.GetByID(context.Background(), payment.ID)
	if found2.Metadata["key"] != "value" {
		t.Errorf("metadata was mutated: got %v, want value", found2.Metadata["key"])
	}
}
