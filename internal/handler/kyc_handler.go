package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type KYCHandler struct {
	kycService service.KYCService
}

func NewKYCHandler(kycService service.KYCService) *KYCHandler {
	return &KYCHandler{kycService: kycService}
}

type submitKYCRequest struct {
	EntityType   string   `json:"entity_type"`
	FullName     string   `json:"full_name"`
	INN          string   `json:"inn"`
	OGRNIP       string   `json:"ogrnip"`
	CompanyName  string   `json:"company_name"`
	DocumentURLs []string `json:"document_urls"`
}

type kycResponse struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Status          string     `json:"status"`
	EntityType      string     `json:"entity_type"`
	FullName        string     `json:"full_name"`
	INN             string     `json:"inn"`
	OGRNIP          string     `json:"ogrnip,omitempty"`
	CompanyName     string     `json:"company_name,omitempty"`
	DocumentURLs    []string   `json:"document_urls"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type rejectKYCRequest struct {
	Reason string `json:"reason"`
}

func toKYCResponse(k *domain.KYCApplication) kycResponse {
	return kycResponse{
		ID:              k.ID.String(),
		UserID:          k.UserID.String(),
		Status:          string(k.Status),
		EntityType:      string(k.EntityType),
		FullName:        k.FullName,
		INN:             k.INN,
		OGRNIP:          k.OGRNIP,
		CompanyName:     k.CompanyName,
		DocumentURLs:    k.DocumentURLs,
		RejectionReason: k.RejectionReason,
		SubmittedAt:     k.SubmittedAt,
		ReviewedAt:      k.ReviewedAt,
		ExpiresAt:       k.ExpiresAt,
		CreatedAt:       k.CreatedAt,
	}
}

// SubmitKYC godoc
//
//	@Summary		Submit KYC application
//	@Description	Submit KYC documents for verification (owner only)
//	@Tags			kyc
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		submitKYCRequest	true	"KYC application data"
//	@Success		201		{object}	APIResponse{data=kycResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/kyc [post]
func (h *KYCHandler) SubmitKYC(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req submitKYCRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	app, err := h.kycService.Submit(r.Context(), userID, service.SubmitKYCInput{
		EntityType:   domain.KYCEntityType(req.EntityType),
		FullName:     req.FullName,
		INN:          req.INN,
		OGRNIP:       req.OGRNIP,
		CompanyName:  req.CompanyName,
		DocumentURLs: req.DocumentURLs,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toKYCResponse(app))
}

// GetKYCStatus godoc
//
//	@Summary		Get KYC status
//	@Description	Get current user's KYC verification status (owner only)
//	@Tags			kyc
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=kycResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/kyc [get]
func (h *KYCHandler) GetKYCStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	app, err := h.kycService.GetStatus(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toKYCResponse(app))
}

// ListPendingKYC godoc
//
//	@Summary		List pending KYC applications
//	@Description	Get list of pending KYC applications for admin review
//	@Tags			kyc,admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]kycResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/kyc/pending [get]
func (h *KYCHandler) ListPendingKYC(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := h.kycService.ListPending(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]kycResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toKYCResponse(&result.Items[i]))
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// ApproveKYC godoc
//
//	@Summary		Approve KYC application
//	@Description	Approve a pending KYC application (admin only)
//	@Tags			kyc,admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"KYC Application ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/kyc/{id}/approve [patch]
func (h *KYCHandler) ApproveKYC(w http.ResponseWriter, r *http.Request) {
	kycID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid KYC ID")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.kycService.Approve(r.Context(), kycID, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "KYC application approved"})
}

// RejectKYC godoc
//
//	@Summary		Reject KYC application
//	@Description	Reject a pending KYC application with reason (admin only)
//	@Tags			kyc,admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"KYC Application ID (UUID)"
//	@Param			body	body		rejectKYCRequest	true	"Rejection reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/kyc/{id}/reject [patch]
func (h *KYCHandler) RejectKYC(w http.ResponseWriter, r *http.Request) {
	kycID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid KYC ID")
		return
	}

	var req rejectKYCRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.kycService.Reject(r.Context(), kycID, adminID, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "KYC application rejected"})
}
