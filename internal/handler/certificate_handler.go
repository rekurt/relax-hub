package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type CertificateHandler struct {
	certificateService service.CertificateService
}

func NewCertificateHandler(certificateService service.CertificateService) *CertificateHandler {
	return &CertificateHandler{certificateService: certificateService}
}

type purchaseCertificateRequest struct {
	Amount         int64  `json:"amount"`
	PurchaserEmail string `json:"purchaser_email"`
	RecipientEmail string `json:"recipient_email"`
	RecipientName  string `json:"recipient_name"`
	Message        string `json:"message,omitempty"`
}

type redeemCertificateRequest struct {
	Code string `json:"code"`
}

type createCertificateOrderRequest struct {
	Amount         int64  `json:"amount"`
	PurchaserEmail string `json:"purchaser_email"`
	RecipientEmail string `json:"recipient_email"`
	RecipientName  string `json:"recipient_name"`
	Message        string `json:"message,omitempty"`
}

type certificateOrderPaymentRequest struct {
	PaymentMethod string `json:"payment_method"`
	PaymentToken  string `json:"payment_token,omitempty"`
}

type certificateResponse struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	PurchaserEmail string    `json:"purchaser_email"`
	RecipientEmail string    `json:"recipient_email"`
	RecipientName  string    `json:"recipient_name"`
	Amount         int64     `json:"amount"`
	Balance        int64     `json:"balance"`
	Message        string    `json:"message"`
	Status         string    `json:"status"`
	ValidUntil     time.Time `json:"valid_until"`
	CreatedAt      time.Time `json:"created_at"`
}

type certificateBalanceResponse struct {
	Code       string    `json:"code"`
	Amount     int64     `json:"amount"`
	Balance    int64     `json:"balance"`
	Status     string    `json:"status"`
	ValidUntil time.Time `json:"valid_until"`
}

type certificateOrderResponse struct {
	ID             string               `json:"id"`
	PurchaserEmail string               `json:"purchaser_email"`
	RecipientEmail string               `json:"recipient_email"`
	RecipientName  string               `json:"recipient_name"`
	Message        string               `json:"message"`
	Amount         int64                `json:"amount"`
	Status         string               `json:"status"`
	PaymentMethod  string               `json:"payment_method,omitempty"`
	Provider       string               `json:"provider,omitempty"`
	ExternalID     string               `json:"external_id,omitempty"`
	PaidAt         *time.Time           `json:"paid_at,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Certificate    *certificateResponse `json:"certificate,omitempty"`
}

type initiateCertificateOrderPaymentResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
}

func toCertificateResponse(c *domain.GiftCertificate) certificateResponse {
	return certificateResponse{
		ID:             c.ID.String(),
		Code:           c.Code,
		PurchaserEmail: c.PurchaserEmail,
		RecipientEmail: c.RecipientEmail,
		RecipientName:  c.RecipientName,
		Amount:         c.Amount,
		Balance:        c.Balance,
		Message:        c.Message,
		Status:         string(c.Status),
		ValidUntil:     c.ValidUntil,
		CreatedAt:      c.CreatedAt,
	}
}

func toCertificateBalanceResponse(c *domain.GiftCertificate) certificateBalanceResponse {
	return certificateBalanceResponse{
		Code:       c.Code,
		Amount:     c.Amount,
		Balance:    c.Balance,
		Status:     string(c.Status),
		ValidUntil: c.ValidUntil,
	}
}

func toCertificateListResponse(certs []domain.GiftCertificate) []certificateResponse {
	result := make([]certificateResponse, len(certs))
	for i := range certs {
		result[i] = toCertificateResponse(&certs[i])
	}
	return result
}

func toCertificateOrderResponse(order *domain.CertificateOrder) certificateOrderResponse {
	response := certificateOrderResponse{
		ID:             order.ID.String(),
		PurchaserEmail: order.PurchaserEmail,
		RecipientEmail: order.RecipientEmail,
		RecipientName:  order.RecipientName,
		Message:        order.Message,
		Amount:         order.Amount,
		Status:         string(order.Status),
		PaymentMethod:  string(order.PaymentMethod),
		Provider:       order.Provider,
		ExternalID:     order.ExternalID,
		PaidAt:         order.PaidAt,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
	if order.Certificate != nil {
		cert := toCertificateResponse(order.Certificate)
		response.Certificate = &cert
	}
	return response
}

// Purchase godoc
//
//	@Summary		Purchase gift certificate
//	@Description	Creates a new gift certificate. Can be used without authentication. Amount is in kopecks.
//	@Tags			certificates
//	@Accept			json
//	@Produce		json
//	@Param			body	body		purchaseCertificateRequest	true	"Certificate purchase data"
//	@Success		201		{object}	APIResponse{data=certificateResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Router			/certificates/purchase [post]
func (h *CertificateHandler) Purchase(w http.ResponseWriter, r *http.Request) {
	var req purchaseCertificateRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	// Get purchaser ID from context if authenticated (optional auth)
	var purchaserID *uuid.UUID
	if uid := middleware.GetUserID(r.Context()); uid != uuid.Nil {
		purchaserID = &uid
	}

	cert, err := h.certificateService.Purchase(
		r.Context(),
		req.Amount,
		purchaserID,
		req.PurchaserEmail,
		req.RecipientEmail,
		req.RecipientName,
		req.Message,
	)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toCertificateResponse(cert))
}

// CreateOrder godoc
//
//	@Summary		Create gift certificate order
//	@Description	Creates a draft order for a gift certificate checkout flow. Can be used without authentication.
//	@Tags			certificates
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createCertificateOrderRequest	true	"Certificate order data"
//	@Success		201		{object}	APIResponse{data=certificateOrderResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Router			/certificates/orders [post]
func (h *CertificateHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req createCertificateOrderRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	var purchaserID *uuid.UUID
	if uid := middleware.GetUserID(r.Context()); uid != uuid.Nil {
		purchaserID = &uid
	}

	order, err := h.certificateService.CreateOrder(
		r.Context(),
		req.Amount,
		purchaserID,
		req.PurchaserEmail,
		req.RecipientEmail,
		req.RecipientName,
		req.Message,
	)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toCertificateOrderResponse(order))
}

// InitiateOrderPayment godoc
//
//	@Summary		Initiate gift certificate payment
//	@Description	Starts payment for a previously created gift certificate order.
//	@Tags			certificates
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Certificate order ID"
//	@Param			body	body		certificateOrderPaymentRequest	false	"Payment method and token"
//	@Success		200		{object}	APIResponse{data=initiateCertificateOrderPaymentResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/certificates/orders/{id}/pay [post]
func (h *CertificateHandler) InitiateOrderPayment(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid certificate order id")
		return
	}

	var req certificateOrderPaymentRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := readJSON(w, r, &req); err != nil {
			handleServiceError(w, err)
			return
		}
	}

	confirmationURL, err := h.certificateService.InitiatePayment(r.Context(), orderID, service.CertificateOrderPaymentRequest{
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		PaymentToken:  req.PaymentToken,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, initiateCertificateOrderPaymentResponse{ConfirmationURL: confirmationURL})
}

// GetOrder godoc
//
//	@Summary		Get gift certificate order
//	@Description	Returns the current status of a gift certificate order and the issued certificate after successful payment.
//	@Tags			certificates
//	@Produce		json
//	@Param			id	path		string	true	"Certificate order ID"
//	@Success		200	{object}	APIResponse{data=certificateOrderResponse}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/certificates/orders/{id} [get]
func (h *CertificateHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid certificate order id")
		return
	}

	order, err := h.certificateService.GetOrder(r.Context(), orderID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toCertificateOrderResponse(order))
}

// Redeem godoc
//
//	@Summary		Redeem gift certificate
//	@Description	Binds a gift certificate to the authenticated user's account by its code
//	@Tags			certificates
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		redeemCertificateRequest	true	"Certificate code"
//	@Success		200		{object}	APIResponse{data=certificateResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/certificates/redeem [post]
func (h *CertificateHandler) Redeem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req redeemCertificateRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	cert, err := h.certificateService.Redeem(r.Context(), req.Code, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toCertificateResponse(cert))
}

// GetBalance godoc
//
//	@Summary		Get certificate balance
//	@Description	Returns the balance and status of a gift certificate by its code
//	@Tags			certificates
//	@Produce		json
//	@Param			code	path		string	true	"Certificate code (e.g. BANI-XXXX-XXXX)"
//	@Success		200		{object}	APIResponse{data=certificateBalanceResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/certificates/{code}/balance [get]
func (h *CertificateHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "certificate code is required")
		return
	}

	cert, err := h.certificateService.GetBalance(r.Context(), code)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toCertificateBalanceResponse(cert))
}

// ListMyCertificates godoc
//
//	@Summary		List my certificates
//	@Description	Returns all gift certificates for the authenticated user with pagination
//	@Tags			certificates
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]certificateResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/certificates [get]
func (h *CertificateHandler) ListMyCertificates(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.certificateService.ListByUser(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSONWithMeta(w, http.StatusOK, toCertificateListResponse(result.Items), &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
