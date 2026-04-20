package service

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/payment"
)

func (s *certificateService) CreateOrder(ctx context.Context, amount int64, purchaserID *uuid.UUID, purchaserEmail, recipientEmail, recipientName, message string) (*domain.CertificateOrder, error) {
	if err := validateCertificatePurchaseInput(amount, purchaserEmail, recipientEmail, recipientName, message); err != nil {
		return nil, err
	}
	if s.orderRepo == nil {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now()
	order := &domain.CertificateOrder{
		ID:             uuid.New(),
		PurchaserID:    purchaserID,
		PurchaserEmail: purchaserEmail,
		RecipientEmail: recipientEmail,
		RecipientName:  recipientName,
		Message:        message,
		Amount:         amount,
		Status:         domain.CertificateOrderStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *certificateService) InitiatePayment(ctx context.Context, orderID uuid.UUID, req CertificateOrderPaymentRequest) (string, error) {
	if s.orderRepo == nil || s.paymentProv == nil {
		return "", domain.ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order.Status == domain.CertificateOrderStatusPendingPayment || order.Status == domain.CertificateOrderStatusPaid {
		return "", domain.ErrPaymentAlreadyProcessed
	}

	method := req.PaymentMethod
	if method == "" {
		method = domain.PaymentMethodCard
	}
	if !method.IsValid() {
		return "", domain.ErrInvalidInput
	}
	if method.IsTokenBased() && strings.TrimSpace(req.PaymentToken) == "" {
		return "", domain.ErrInvalidInput
	}

	result, err := s.paymentProv.CreatePayment(ctx, payment.CreatePaymentRequest{
		Amount:       order.Amount,
		Currency:     "RUB",
		Description:  fmt.Sprintf("Оплата сертификата %s", order.ID.String()[:8]),
		ReturnURL:    s.certificateOrderReturnURL(order.ID),
		Metadata:     map[string]string{"certificate_order_id": order.ID.String()},
		Method:       string(method),
		Capture:      true,
		PaymentToken: req.PaymentToken,
	})
	if err != nil {
		return "", err
	}

	if err := s.orderRepo.UpdatePayment(ctx, order.ID, domain.CertificateOrderStatusPendingPayment, method, "yookassa", result.ExternalID); err != nil {
		return "", err
	}
	return result.ConfirmationURL, nil
}

func (s *certificateService) GetOrder(ctx context.Context, orderID uuid.UUID) (*domain.CertificateOrder, error) {
	if s.orderRepo == nil {
		return nil, domain.ErrCertificateOrderNotFound
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.CertificateID != nil && *order.CertificateID != uuid.Nil {
		cert, certErr := s.certRepo.GetByID(ctx, *order.CertificateID)
		if certErr == nil {
			order.Certificate = cert
		}
	}
	return order, nil
}

func (s *certificateService) HandlePaymentWebhook(ctx context.Context, event WebhookEvent) error {
	if s.orderRepo == nil || s.paymentProv == nil {
		return domain.ErrCertificateOrderNotFound
	}

	order, err := s.orderRepo.GetByExternalID(ctx, event.ExternalID)
	if err != nil {
		return err
	}
	if order.Status == domain.CertificateOrderStatusPaid && order.CertificateID != nil {
		return nil
	}

	providerStatus, err := s.paymentProv.GetPaymentStatus(ctx, event.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to verify certificate payment with provider: %w", err)
	}
	if providerStatus != event.Status {
		return fmt.Errorf("%w: webhook status does not match provider", domain.ErrInvalidInput)
	}

	switch event.Status {
	case "succeeded":
		if order.CertificateID != nil && *order.CertificateID != uuid.Nil {
			if err := s.orderRepo.MarkPaid(ctx, order.ID, *order.CertificateID, time.Now()); err != nil {
				return err
			}
			return nil
		}

		cert, err := s.issueCertificate(ctx, order.Amount, order.PurchaserID, order.PurchaserEmail, order.RecipientEmail, order.RecipientName, order.Message)
		if err != nil {
			return err
		}
		if err := s.orderRepo.MarkPaid(ctx, order.ID, cert.ID, time.Now()); err != nil {
			return err
		}
		s.sendCertificateEmail(ctx, cert)
		return nil
	case "canceled":
		return s.orderRepo.UpdateStatus(ctx, order.ID, domain.CertificateOrderStatusCanceled)
	default:
		return s.orderRepo.UpdateStatus(ctx, order.ID, domain.CertificateOrderStatusFailed)
	}
}

func (s *certificateService) certificateOrderReturnURL(orderID uuid.UUID) string {
	base := strings.TrimSpace(s.frontendURL)
	if base == "" {
		return fmt.Sprintf("/certificates?order_id=%s", orderID)
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return fmt.Sprintf("%s/certificates?order_id=%s", strings.TrimRight(base, "/"), orderID)
	}
	parsed.Path = path.Join(parsed.Path, "/certificates")
	query := parsed.Query()
	query.Set("order_id", orderID.String())
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
