package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/notification"
	"github.com/rekurt/relax-hub/internal/payment"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

type recordingCertificateEmailSender struct {
	messages []recordedCertificateEmail
}

type recordedCertificateEmail struct {
	to      string
	subject string
	body    string
}

func (r *recordingCertificateEmailSender) Send(_ context.Context, to, subject, body string) error {
	r.messages = append(r.messages, recordedCertificateEmail{
		to:      to,
		subject: subject,
		body:    body,
	})
	return nil
}

var _ notification.EmailSender = (*recordingCertificateEmailSender)(nil)

type certOrderTestEnv struct {
	svc         service.CertificateService
	certRepo    *mock.CertificateRepo
	orderRepo   *mock.CertificateOrderRepo
	provider    *payment.MockProvider
	emailSender *recordingCertificateEmailSender
}

func newCertOrderTestEnv() *certOrderTestEnv {
	certRepo := mock.NewCertificateRepo().(*mock.CertificateRepo)
	orderRepo := mock.NewCertificateOrderRepo().(*mock.CertificateOrderRepo)
	provider := payment.NewMockProvider()
	emailSender := &recordingCertificateEmailSender{}
	log := logger.New(logger.LevelWarn)
	svc := service.NewCertificateService(
		certRepo,
		orderRepo,
		provider,
		emailSender,
		"http://localhost:5173",
		log,
	)
	return &certOrderTestEnv{
		svc:         svc,
		certRepo:    certRepo,
		orderRepo:   orderRepo,
		provider:    provider,
		emailSender: emailSender,
	}
}

func TestCertificateService_CreateOrder_Success(t *testing.T) {
	env := newCertOrderTestEnv()
	userID := uuid.New()

	order, err := env.svc.CreateOrder(
		context.Background(),
		100000,
		&userID,
		"buyer@example.com",
		"friend@example.com",
		"Иван",
		"Приятного отдыха!",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.ID == uuid.Nil {
		t.Fatal("expected non-empty order ID")
	}
	if order.Status != domain.CertificateOrderStatusDraft {
		t.Fatalf("status = %q, want %q", order.Status, domain.CertificateOrderStatusDraft)
	}
	if order.PurchaserID == nil || *order.PurchaserID != userID {
		t.Fatalf("purchaser ID mismatch")
	}
	if len(env.emailSender.messages) != 0 {
		t.Fatalf("draft order must not send emails")
	}
}

func TestCertificateService_InitiatePayment_SetsPendingPayment(t *testing.T) {
	env := newCertOrderTestEnv()

	order, err := env.svc.CreateOrder(
		context.Background(),
		150000,
		nil,
		"buyer@example.com",
		"friend@example.com",
		"Иван",
		"С днём рождения!",
	)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	confirmationURL, err := env.svc.InitiatePayment(
		context.Background(),
		order.ID,
		service.CertificateOrderPaymentRequest{
			PaymentMethod: domain.PaymentMethodCard,
		},
	)
	if err != nil {
		t.Fatalf("initiate payment: %v", err)
	}
	if confirmationURL == "" {
		t.Fatal("expected confirmation URL")
	}

	updated, err := env.orderRepo.GetByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if updated.Status != domain.CertificateOrderStatusPendingPayment {
		t.Fatalf("status = %q, want %q", updated.Status, domain.CertificateOrderStatusPendingPayment)
	}
	if updated.ExternalID == "" {
		t.Fatal("expected external ID to be stored")
	}
	if updated.PaymentMethod != domain.PaymentMethodCard {
		t.Fatalf("payment method = %q, want %q", updated.PaymentMethod, domain.PaymentMethodCard)
	}
}

func TestCertificateService_HandlePaymentWebhook_IssuesCertificateAndIsIdempotent(t *testing.T) {
	env := newCertOrderTestEnv()

	order, err := env.svc.CreateOrder(
		context.Background(),
		200000,
		nil,
		"buyer@example.com",
		"friend@example.com",
		"Мария",
		"Тёплого пара!",
	)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	_, err = env.svc.InitiatePayment(
		context.Background(),
		order.ID,
		service.CertificateOrderPaymentRequest{
			PaymentMethod: domain.PaymentMethodCard,
		},
	)
	if err != nil {
		t.Fatalf("initiate payment: %v", err)
	}

	pendingOrder, err := env.orderRepo.GetByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("get pending order: %v", err)
	}

	env.provider.SetPaymentStatus(pendingOrder.ExternalID, "succeeded")

	err = env.svc.HandlePaymentWebhook(context.Background(), service.WebhookEvent{
		ExternalID: pendingOrder.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Fatalf("handle webhook: %v", err)
	}

	paidOrder, err := env.orderRepo.GetByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("get paid order: %v", err)
	}
	if paidOrder.Status != domain.CertificateOrderStatusPaid {
		t.Fatalf("status = %q, want %q", paidOrder.Status, domain.CertificateOrderStatusPaid)
	}
	if paidOrder.CertificateID == nil || *paidOrder.CertificateID == uuid.Nil {
		t.Fatal("expected issued certificate ID")
	}

	certificate, err := env.certRepo.GetByID(context.Background(), *paidOrder.CertificateID)
	if err != nil {
		t.Fatalf("get certificate: %v", err)
	}
	if !strings.HasPrefix(certificate.Code, "BANI-") {
		t.Fatalf("certificate code = %q, want BANI-*", certificate.Code)
	}
	if certificate.Amount != order.Amount {
		t.Fatalf("certificate amount = %d, want %d", certificate.Amount, order.Amount)
	}
	if len(env.emailSender.messages) == 0 {
		t.Fatal("expected confirmation email after payment")
	}
	initialEmailCount := len(env.emailSender.messages)
	initialCertificateID := *paidOrder.CertificateID

	err = env.svc.HandlePaymentWebhook(context.Background(), service.WebhookEvent{
		ExternalID: pendingOrder.ExternalID,
		Status:     "succeeded",
	})
	if err != nil {
		t.Fatalf("second webhook should be idempotent: %v", err)
	}
	reloadedOrder, err := env.orderRepo.GetByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("get order after repeated webhook: %v", err)
	}
	if reloadedOrder.CertificateID == nil || *reloadedOrder.CertificateID != initialCertificateID {
		t.Fatal("expected certificate id to stay unchanged after repeated webhook")
	}
	if len(env.emailSender.messages) != initialEmailCount {
		t.Fatalf("expected no duplicate emails on repeated webhook, got %d", len(env.emailSender.messages))
	}
}
