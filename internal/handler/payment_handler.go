package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
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
	WalletAmount  int64             `json:"wallet_amount"`
	CardAmount    int64             `json:"card_amount"`
	IsHold        bool              `json:"is_hold"`
	CapturedAt    *time.Time        `json:"captured_at,omitempty"`
	RefundAmount  int64             `json:"refund_amount"`
	RefundedAt    *time.Time        `json:"refunded_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type initiatePaymentRequest struct {
	PaymentMethod     string `json:"payment_method"`                // "card", "sbp", "wallet", "combo", "apple_pay", "google_pay"
	WalletAmount      int64  `json:"wallet_amount,omitempty"`       // for combo/wallet
	CardAmount        int64  `json:"card_amount,omitempty"`         // for combo
	CardPaymentMethod string `json:"card_payment_method,omitempty"` // for combo: "card" (default) or "sbp"
	PaymentToken      string `json:"payment_token,omitempty"`       // token from Apple Pay JS or Google Pay API
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
		WalletAmount:  p.WalletAmount,
		CardAmount:    p.CardAmount,
		IsHold:        p.IsHold,
		CapturedAt:    p.CapturedAt,
		RefundAmount:  p.RefundAmount,
		RefundedAt:    p.RefundedAt,
		Metadata:      p.Metadata,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// InitiatePayment godoc
//
//	@Summary		Initiate payment
//	@Description	Creates a payment for a booking. Supports card, SBP, wallet, combo (wallet+card/SBP), apple_pay, and google_pay payment methods. For wallet/combo, provide wallet_amount and card_amount fields. For Apple Pay/Google Pay, provide payment_token from the client-side SDK.
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		initiatePaymentRequest	false	"Payment options"
//	@Success		200		{object}	APIResponse{data=initiatePaymentResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/pay [post]
//
// InitiatePayment handles POST /api/v1/bookings/{id}/pay
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req initiatePaymentRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := readJSON(w, r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
			return
		}
	}

	paymentMethod := domain.PaymentMethod(req.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = domain.PaymentMethodCard
	}

	if !paymentMethod.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid payment_method")
		return
	}

	userID := middleware.GetUserID(r.Context())

	// Token-based payments (Apple Pay, Google Pay)
	if paymentMethod.IsTokenBased() {
		if req.PaymentToken == "" {
			writeError(w, http.StatusBadRequest, "invalid_input", "payment_token is required for Apple Pay / Google Pay")
			return
		}
		tokenReq := service.TokenPaymentRequest{
			PaymentMethod: paymentMethod,
			PaymentToken:  req.PaymentToken,
		}
		confirmationURL, err := h.paymentService.InitiateTokenPayment(r.Context(), userID, bookingID, tokenReq)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, initiatePaymentResponse{
			ConfirmationURL: confirmationURL,
		})
		return
	}

	switch paymentMethod {
	case domain.PaymentMethodWallet:
		comboReq := service.ComboPaymentRequest{
			WalletAmount:  req.WalletAmount,
			CardAmount:    0,
			PaymentMethod: domain.PaymentMethodWallet,
		}
		_, err := h.paymentService.InitiateComboPayment(r.Context(), userID, bookingID, comboReq)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, initiatePaymentResponse{})

	case domain.PaymentMethodCombo:
		cardMethod := domain.PaymentMethodCard
		if req.CardPaymentMethod == string(domain.PaymentMethodSBP) {
			cardMethod = domain.PaymentMethodSBP
		}
		comboReq := service.ComboPaymentRequest{
			WalletAmount:  req.WalletAmount,
			CardAmount:    req.CardAmount,
			PaymentMethod: cardMethod,
		}
		confirmationURL, err := h.paymentService.InitiateComboPayment(r.Context(), userID, bookingID, comboReq)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, initiatePaymentResponse{
			ConfirmationURL: confirmationURL,
		})

	default:
		// card or sbp — existing flow
		confirmationURL, err := h.paymentService.InitiatePayment(r.Context(), userID, bookingID, paymentMethod)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, initiatePaymentResponse{
			ConfirmationURL: confirmationURL,
		})
	}
}

// HandleWebhook godoc
//
//	@Summary		YooKassa webhook
//	@Description	Processes payment status updates from YooKassa. Returns 200 even for domain errors to prevent infinite retries.
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Param			body	body		object	true	"YooKassa webhook payload"
//	@Success		200		{object}	APIResponse{data=statusResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (30/min)"
//	@Failure		500		{object}	APIResponse{error=APIError}
//	@Router			/webhooks/yookassa [post]
//
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

// HandleBePaidWebhook godoc
//
//	@Summary		bePaid webhook
//	@Description	Processes payment status updates from bePaid (Belarus). Returns 200 even for domain errors to prevent infinite retries.
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Param			body	body		object	true	"bePaid webhook payload"
//	@Success		200		{object}	APIResponse{data=statusResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (30/min)"
//	@Failure		500		{object}	APIResponse{error=APIError}
//	@Router			/webhooks/bepaid [post]
//
// HandleBePaidWebhook handles POST /api/v1/webhooks/bepaid
func (h *PaymentHandler) HandleBePaidWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "failed to read request body")
		return
	}

	var webhook struct {
		Transaction struct {
			UID    string `json:"uid"`
			Status string `json:"status"`
		} `json:"transaction"`
	}

	if err := json.Unmarshal(body, &webhook); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid webhook payload")
		return
	}

	if webhook.Transaction.UID == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "missing transaction uid in webhook")
		return
	}

	// Map bePaid status to internal status
	internalStatus := mapBePaidWebhookStatus(webhook.Transaction.Status)

	event := service.WebhookEvent{
		ExternalID: webhook.Transaction.UID,
		Status:     internalStatus,
	}

	if err := h.paymentService.HandleWebhook(r.Context(), event); err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) || errors.Is(err, domain.ErrInvalidInput) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "error"})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "webhook processing failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// mapBePaidWebhookStatus maps bePaid webhook statuses to our internal status strings.
func mapBePaidWebhookStatus(bepaidStatus string) string {
	switch bepaidStatus {
	case "successful":
		return "succeeded"
	case "failed", "expired":
		return "canceled"
	case "authorized":
		return "waiting_for_capture"
	default:
		return bepaidStatus
	}
}

// ListUserPayments godoc
//
//	@Summary		List my payments
//	@Description	Returns a paginated list of payments for the authenticated user
//	@Tags			payments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]paymentResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/payments [get]
//
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
//
//	@Summary		Get booking payment
//	@Description	Returns the payment associated with a specific booking
//	@Tags			payments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=paymentResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/payment [get]
//
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

// ValidateApplePayMerchant godoc
//
//	@Summary		Validate Apple Pay merchant
//	@Description	Performs Apple Pay merchant validation. In production requires server-side Apple certificates.
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Param			body	body		applePayValidationRequest	true	"Validation URL from Apple Pay session"
//	@Success		200		{object}	APIResponse{data=interface{}}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Router			/apple-pay/validate-merchant [post]
func (h *PaymentHandler) ValidateApplePayMerchant(w http.ResponseWriter, r *http.Request) {
	var req applePayValidationRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}
	if req.ValidationURL == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "validation_url is required")
		return
	}
	// Production: call Apple's validation_url with merchant identity certificate.
	// Dev: return stub so frontend completes the session flow.
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"displayName": "RelaxHub",
	})
}

type applePayValidationRequest struct {
	ValidationURL string `json:"validation_url"`
}

type adminRefundRequest struct {
	Amount   int64  `json:"amount"`
	Reason   string `json:"reason"`
	RefundTo string `json:"refund_to"` // "wallet" or "card", default "card"
}

// AdminRefund godoc
//
//	@Summary		Admin manual refund
//	@Description	Allows admin to refund a booking payment with a specified amount and reason. Creates an audit log entry.
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Booking ID (UUID)"
//	@Param			body	body		adminRefundRequest	true	"Refund details"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/bookings/{id}/refund [post]
func (h *PaymentHandler) AdminRefund(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req adminRefundRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "amount must be positive")
		return
	}

	if req.RefundTo != "" && req.RefundTo != "wallet" && req.RefundTo != "card" {
		writeError(w, http.StatusBadRequest, "invalid_input", "refund_to must be 'wallet' or 'card'")
		return
	}

	adminUserID := middleware.GetUserID(r.Context())
	if err := h.paymentService.AdminRefund(r.Context(), adminUserID, bookingID, req.Amount, req.Reason, req.RefundTo); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "refund_processed"})
}
