package handler

import (
	"net/http"
	"time"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

type walletResponse struct {
	Balance        int64   `json:"balance"`
	HeldAmount     int64   `json:"held_amount"`
	Available      int64   `json:"available"`
	Currency       string  `json:"currency"`
	ExpiringSoon   int64   `json:"expiring_soon"`
	EarliestExpiry *string `json:"earliest_expiry,omitempty"`
}

type walletTransactionResponse struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Amount        int64   `json:"amount"`
	BalanceAfter  int64   `json:"balance_after"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
	ReferenceType string  `json:"reference_type,omitempty"`
	ReferenceID   *string `json:"reference_id,omitempty"`
	IsBonus       bool    `json:"is_bonus"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

type walletHoldResponse struct {
	ID            string  `json:"id"`
	Amount        int64   `json:"amount"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
	ReferenceType string  `json:"reference_type,omitempty"`
	ReferenceID   *string `json:"reference_id,omitempty"`
	ExpiresAt     string  `json:"expires_at"`
	CreatedAt     string  `json:"created_at"`
}

type topUpRequest struct {
	Amount int64 `json:"amount"`
}

type topUpResponse struct {
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	BalanceAfter  int64  `json:"balance_after"`
}

func toWalletResponse(s *service.WalletBalanceSummary) walletResponse {
	resp := walletResponse{
		Balance:      s.Balance,
		HeldAmount:   s.HeldAmount,
		Available:    s.Available,
		Currency:     string(s.Currency),
		ExpiringSoon: s.ExpiringSoon,
	}
	if s.EarliestExpiry != nil {
		t := s.EarliestExpiry.Format(time.RFC3339)
		resp.EarliestExpiry = &t
	}
	return resp
}

func toWalletTransactionResponse(tx *domain.WalletTransaction) walletTransactionResponse {
	resp := walletTransactionResponse{
		ID:            tx.ID.String(),
		Type:          string(tx.Type),
		Amount:        tx.Amount,
		BalanceAfter:  tx.BalanceAfter,
		Status:        string(tx.Status),
		Description:   tx.Description,
		ReferenceType: tx.ReferenceType,
		IsBonus:       tx.IsBonus,
		CreatedAt:     tx.CreatedAt.Format(time.RFC3339),
	}
	if tx.ReferenceID != nil {
		s := tx.ReferenceID.String()
		resp.ReferenceID = &s
	}
	if tx.ExpiresAt != nil {
		s := tx.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &s
	}
	return resp
}

func toWalletHoldResponse(h *domain.WalletHold) walletHoldResponse {
	resp := walletHoldResponse{
		ID:            h.ID.String(),
		Amount:        h.Amount,
		Status:        string(h.Status),
		Description:   h.Description,
		ReferenceType: h.ReferenceType,
		ExpiresAt:     h.ExpiresAt.Format(time.RFC3339),
		CreatedAt:     h.CreatedAt.Format(time.RFC3339),
	}
	if h.ReferenceID != nil {
		s := h.ReferenceID.String()
		resp.ReferenceID = &s
	}
	return resp
}

// GetWallet godoc
// @Summary      Get wallet balance
// @Description  Returns the wallet balance summary for the authenticated user, including held amount, available balance, and expiring bonuses
// @Tags         wallet
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  APIResponse{data=walletResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/wallet [get]
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	summary, err := h.walletService.GetBalance(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toWalletResponse(summary))
}

// ListTransactions godoc
// @Summary      List wallet transactions
// @Description  Returns paginated transaction history for the authenticated user's wallet
// @Tags         wallet
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int     false  "Page number"          default(1)
// @Param        page_size  query     int     false  "Page size"            default(20)
// @Param        type       query     string  false  "Transaction type filter (topup, spend, refund, bonus, etc.)"
// @Param        is_bonus   query     bool    false  "Filter bonus transactions only"
// @Success      200  {object}  APIResponse{data=[]walletTransactionResponse,meta=Meta}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/wallet/transactions [get]
func (h *WalletHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	filter := domain.WalletTransactionFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if txType := r.URL.Query().Get("type"); txType != "" {
		t := domain.WalletTransactionType(txType)
		if t.IsValid() {
			filter.Type = &t
		}
	}

	if isBonusStr := r.URL.Query().Get("is_bonus"); isBonusStr == "true" {
		b := true
		filter.IsBonus = &b
	} else if isBonusStr == "false" {
		b := false
		filter.IsBonus = &b
	}

	result, err := h.walletService.ListTransactions(r.Context(), userID, filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]walletTransactionResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toWalletTransactionResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// TopUp godoc
// @Summary      Top up wallet
// @Description  Initiates a wallet top-up. Amount is in kopecks (min 50000 = 500 RUB, max 3000000 = 30000 RUB)
// @Tags         wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      topUpRequest  true  "Top-up amount in kopecks"
// @Success      200   {object}  APIResponse{data=topUpResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /my/wallet/topup [post]
func (h *WalletHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req topUpRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "amount must be positive")
		return
	}

	tx, err := h.walletService.TopUp(r.Context(), userID, req.Amount)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, topUpResponse{
		TransactionID: tx.ID.String(),
		Amount:        tx.Amount,
		BalanceAfter:  tx.BalanceAfter,
	})
}

// ListHolds godoc
// @Summary      List active wallet holds
// @Description  Returns active holds (frozen amounts) on the authenticated user's wallet
// @Tags         wallet
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  APIResponse{data=[]walletHoldResponse}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /my/wallet/holds [get]
func (h *WalletHandler) ListHolds(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	wallet, err := h.walletService.GetWallet(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	holds, err := h.walletService.GetActiveHolds(r.Context(), wallet.ID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]walletHoldResponse, len(holds))
	for i := range holds {
		items[i] = toWalletHoldResponse(&holds[i])
	}

	writeJSON(w, http.StatusOK, items)
}
