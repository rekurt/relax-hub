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
	ID            string            `json:"id"`
	BookingID     string            `json:"booking_id"`
	UserID        string            `json:"user_id"`
	Amount        int64             `json:"amount"`
	Currency      string            `json:"currency"`
	Status        string            `json:"status"`
	Provider      string            `json:"provider"`
	PaymentMethod string            `json:"payment_method"`
	RefundAmount  int64             `json:"refund_amount"`
	RefundedAt    *time.Time        `json:"refunded_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type initiatePaymentRequest struct {
	PaymentMethod string `json:"payment_method"` // "card" or "sbp", default "card"
}

type initiatePaymentResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
}

func toPaymentResponse(p *domain.Payment) paymentResponse {
	return paymentResponse{
		ID:            p.ID.String(),
		BookingID:     p.BookingID.String(),
		UserID:        p.UserID.String(),
		Amount:        p.Amount,
		Currency:      p.Currency,
		Status:        string(p.Status),
		Provider:      p.Provider,
		PaymentMethod: string(p.PaymentMethod),
		RefundAmount:  p.RefundAmount,
		RefundedAt:    p.RefundedAt,
		Metadata:      p.Metadata,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// InitiatePayment godoc
// @Summary      Initiate payment
// @Description  Creates a payment for a booking via YooKassa and returns the confirmation URL for redirect. Supports card and SBP payment methods.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                   true  "Booking ID (UUID)"
// @Param        body  body      initiatePaymentRequest   false "Payment options"
// @Success      200   {object}  APIResponse{data=initiatePaymentResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /bookings/{id}/pay [post]
// InitiatePayment handles POST /api/v1/bookings/{id}/pay
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req initiatePaymentRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
			return
		}
	}

	paymentMethod := domain.PaymentMethod(req.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = domain.PaymentMethodCard
	}
	if paymentMethod != domain.PaymentMethodCard && paymentMethod != domain.PaymentMethodSBP {
		writeError(w, http.StatusBadRequest, "invalid_input", "payment_method must be 'card' or 'sbp'")
		return
	}

	userID := middleware.GetUserID(r.Context())
	confirmationURL, err := h.paymentService.InitiatePayment(r.Context(), userID, bookingID, paymentMethod)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, initiatePaymentResponse{
		ConfirmationURL: confirmationURL,
	})
}

// HandleWebhook godoc
// @Summary      YooKassa webhook
// @Description  Processes payment status updates from YooKassa. Returns 200 even for domain errors to prevent infinite retries.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "YooKassa webhook payload"
// @Success      200   {object}  APIResponse{data=statusResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      429   {object}  APIResponse{error=APIError}  "Rate limited (30/min)"
// @Failure      500   {object}  APIResponse{error=APIError}
// @Router       /webhooks/yookassa [post]
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

// ListUserPayments godoc
// @Summary      List my payments
// @Description  Returns a paginated list of payments for the authenticated user
// @Tags         payments
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int  false  "Page number"  default(1)
// @Param        page_size  query     int  false  "Page size"    default(20)
// @Success      200        {object}  APIResponse{data=[]paymentResponse,meta=Meta}
// @Failure      401        {object}  APIResponse{error=APIError}
// @Router       /my/payments [get]
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

// GetBookingPayment godoc
// @Summary      Get booking payment
// @Description  Returns the payment associated with a specific booking
// @Tags         payments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Booking ID (UUID)"
// @Success      200  {object}  APIResponse{data=paymentResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /bookings/{id}/payment [get]
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
