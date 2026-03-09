package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

type paymentResponse struct {
	ID           string            `json:"id"`
	BookingID    string            `json:"booking_id"`
	UserID       string            `json:"user_id"`
	Amount       int64             `json:"amount"`
	Currency     string            `json:"currency"`
	Status       string            `json:"status"`
	Provider     string            `json:"provider"`
	RefundAmount int64             `json:"refund_amount"`
	RefundedAt   *time.Time        `json:"refunded_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type initiatePaymentResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
}

func toPaymentResponse(p *domain.Payment) paymentResponse {
	return paymentResponse{
		ID:           p.ID.String(),
		BookingID:    p.BookingID.String(),
		UserID:       p.UserID.String(),
		Amount:       p.Amount,
		Currency:     p.Currency,
		Status:       string(p.Status),
		Provider:     p.Provider,
		RefundAmount: p.RefundAmount,
		RefundedAt:   p.RefundedAt,
		Metadata:     p.Metadata,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// InitiatePayment handles POST /api/v1/bookings/{id}/pay
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	confirmationURL, err := h.paymentService.InitiatePayment(r.Context(), userID, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, initiatePaymentResponse{
		ConfirmationURL: confirmationURL,
	})
}

// HandleWebhook handles POST /api/v1/webhooks/yookassa
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "failed to read request body")
		return
	}

	var webhook struct {
		Event  string `json:"event"`
		Object struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"object"`
	}

	if err := json.Unmarshal(body, &webhook); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid webhook payload")
		return
	}

	if webhook.Object.ID == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "missing payment id in webhook")
		return
	}

	event := service.WebhookEvent{
		ExternalID: webhook.Object.ID,
		Status:     webhook.Object.Status,
	}

	if err := h.paymentService.HandleWebhook(r.Context(), event); err != nil {
		// For domain errors (not found, invalid input), return 200 to prevent infinite retries
		// since these won't resolve on retry. For transient errors, return 500 so provider retries.
		if errors.Is(err, domain.ErrPaymentNotFound) || errors.Is(err, domain.ErrInvalidInput) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "error"})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "webhook processing failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListUserPayments handles GET /api/v1/my/payments
func (h *PaymentHandler) ListUserPayments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.paymentService.ListUserPayments(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]paymentResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toPaymentResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// GetBookingPayment handles GET /api/v1/bookings/{id}/payment
func (h *PaymentHandler) GetBookingPayment(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	p, err := h.paymentService.GetPaymentByBooking(r.Context(), userID, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPaymentResponse(p))
}
