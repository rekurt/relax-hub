package handler

import (
	"encoding/json"
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
	ExternalID   string            `json:"external_id,omitempty"`
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
		ExternalID:   p.ExternalID,
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

	confirmationURL, err := h.paymentService.InitiatePayment(r.Context(), bookingID)
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
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "failed to read request body")
		return
	}
	defer r.Body.Close()

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
		handleServiceError(w, err)
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

	p, err := h.paymentService.GetPaymentByBooking(r.Context(), bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPaymentResponse(p))
}
