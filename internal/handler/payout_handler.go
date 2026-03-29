package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type PayoutHandler struct {
	payoutService service.PayoutService
}

func NewPayoutHandler(payoutService service.PayoutService) *PayoutHandler {
	return &PayoutHandler{payoutService: payoutService}
}

type requestPayoutRequest struct {
	Amount      int64           `json:"amount"`
	BankDetails json.RawMessage `json:"bank_details"`
}

type payoutResponse struct {
	ID            string          `json:"id"`
	Amount        int64           `json:"amount"`
	Status        string          `json:"status"`
	BankDetails   json.RawMessage `json:"bank_details,omitempty"`
	RequestedAt   string          `json:"requested_at"`
	ProcessedAt   *string         `json:"processed_at,omitempty"`
	FailureReason string          `json:"failure_reason,omitempty"`
}

type autoPayoutRequest struct {
	Threshold int64 `json:"threshold"`
}

type autoPayoutResponse struct {
	Threshold int64  `json:"threshold"`
	UpdatedAt string `json:"updated_at"`
}

func toPayoutResponse(p *domain.Payout) payoutResponse {
	resp := payoutResponse{
		ID:            p.ID.String(),
		Amount:        p.Amount,
		Status:        string(p.Status),
		BankDetails:   p.BankDetails,
		RequestedAt:   p.RequestedAt.Format(time.RFC3339),
		FailureReason: p.FailureReason,
	}
	if p.ProcessedAt != nil {
		t := p.ProcessedAt.Format(time.RFC3339)
		resp.ProcessedAt = &t
	}
	return resp
}

// RequestPayout godoc
//
//	@Summary		Request payout
//	@Description	Request withdrawal of funds from wallet. Amount in kopecks (min 50000 = 500 RUB). Daily limit 500,000 RUB, monthly 3,000,000 RUB.
//	@Tags			wallet
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		requestPayoutRequest	true	"Payout request"
//	@Success		201		{object}	APIResponse{data=payoutResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/wallet/payout [post]
func (h *PayoutHandler) RequestPayout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req requestPayoutRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "amount must be positive")
		return
	}

	payout, err := h.payoutService.RequestPayout(r.Context(), userID, req.Amount, req.BankDetails)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPayoutResponse(payout))
}

// SetAutoPayoutThreshold godoc
//
//	@Summary		Set auto-payout threshold
//	@Description	Configure automatic payout when wallet balance exceeds threshold. Set to 0 to disable. Threshold in kopecks.
//	@Tags			wallet
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		autoPayoutRequest	true	"Auto-payout settings"
//	@Success		200		{object}	APIResponse{data=autoPayoutResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/wallet/auto-payout [put]
func (h *PayoutHandler) SetAutoPayoutThreshold(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req autoPayoutRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if err := h.payoutService.SetAutoPayoutThreshold(r.Context(), userID, req.Threshold); err != nil {
		handleServiceError(w, err)
		return
	}

	settings, err := h.payoutService.GetAutoPayoutSettings(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, autoPayoutResponse{
		Threshold: settings.Threshold,
		UpdatedAt: settings.UpdatedAt.Format(time.RFC3339),
	})
}

// ListPayouts godoc
//
//	@Summary		List payout history
//	@Description	Returns paginated payout history for the authenticated owner
//	@Tags			wallet
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]payoutResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/my/wallet/payouts [get]
func (h *PayoutHandler) ListPayouts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.payoutService.GetPayoutHistory(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]payoutResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toPayoutResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
