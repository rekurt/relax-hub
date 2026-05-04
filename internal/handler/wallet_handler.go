package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/config"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/payment"
	"github.com/rekurt/relax-hub/internal/service"
)

type WalletHandler struct {
	walletService   service.WalletService
	paymentProvider payment.PaymentProvider
	paymentCfg      config.PaymentConfig
}

func NewWalletHandler(walletService service.WalletService, paymentProvider payment.PaymentProvider, cfg *config.Config) *WalletHandler {
	return &WalletHandler{
		walletService:   walletService,
		paymentProvider: paymentProvider,
		paymentCfg:      cfg.Payment,
	}
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
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method,omitempty"` // "card" (default), "sbp"
}

type topUpResponse struct {
	PaymentID       string `json:"payment_id"`
	Amount          int64  `json:"amount"`
	ConfirmationURL string `json:"confirmation_url"`
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
//
//	@Summary		Get wallet balance
//	@Description	Returns the wallet balance summary for the authenticated user, including held amount, available balance, and expiring bonuses
//	@Tags			wallet
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=walletResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/wallet [get]
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	summary, err := h.walletService.GetBalance(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			if _, ensureErr := h.walletService.EnsureWallet(r.Context(), userID); ensureErr != nil {
				handleServiceError(w, ensureErr)
				return
			}
			summary, err = h.walletService.GetBalance(r.Context(), userID)
		}
	}
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toWalletResponse(summary))
}

// ListTransactions godoc
//
//	@Summary		List wallet transactions
//	@Description	Returns paginated transaction history for the authenticated user's wallet
//	@Tags			wallet
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Param			type		query		string	false	"Transaction type filter (topup, spend, refund, bonus, etc.)"
//	@Param			is_bonus	query		bool	false	"Filter bonus transactions only"
//	@Success		200			{object}	APIResponse{data=[]walletTransactionResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		404			{object}	APIResponse{error=APIError}
//	@Router			/my/wallet/transactions [get]
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
//
//	@Summary		Top up wallet
//	@Description	Initiates a wallet top-up. Amount is in kopecks (min 50000 = 500 RUB, max 3000000 = 30000 RUB)
//	@Tags			wallet
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		topUpRequest	true	"Top-up amount in kopecks"
//	@Success		200		{object}	APIResponse{data=topUpResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/wallet/topup [post]
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

	// Validate amount range against wallet limits
	wallet, err := h.walletService.GetWallet(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	minAmount := domain.TopUpMinForCurrency(wallet.Currency)
	maxAmount := domain.TopUpMaxForCurrency(wallet.Currency)
	if req.Amount < minAmount {
		writeError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("minimum top-up amount is %d", minAmount))
		return
	}
	if req.Amount > maxAmount {
		writeError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("maximum top-up amount is %d", maxAmount))
		return
	}

	if h.paymentProvider == nil {
		writeError(w, http.StatusServiceUnavailable, "service_unavailable", "payment provider is not configured")
		return
	}

	method := req.PaymentMethod
	if method == "" {
		method = "card"
	}

	currency := "RUB"
	if wallet.Currency == domain.WalletCurrencyBYN {
		currency = "BYN"
	}

	result, err := h.paymentProvider.CreatePayment(r.Context(), payment.CreatePaymentRequest{
		Amount:      req.Amount,
		Currency:    currency,
		Description: fmt.Sprintf("Пополнение кошелька на %d коп.", req.Amount),
		ReturnURL:   h.paymentCfg.ReturnURL,
		Metadata: map[string]string{
			"type":    "wallet_topup",
			"user_id": userID.String(),
		},
		Method:  method,
		Capture: true,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, "payment_failed", "failed to create payment")
		return
	}

	writeJSON(w, http.StatusOK, topUpResponse{
		PaymentID:       result.ExternalID,
		Amount:          req.Amount,
		ConfirmationURL: result.ConfirmationURL,
	})
}

// ListHolds godoc
//
//	@Summary		List active wallet holds
//	@Description	Returns active holds (frozen amounts) on the authenticated user's wallet
//	@Tags			wallet
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]walletHoldResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/wallet/holds [get]
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

type adminWalletRequest struct {
	Amount int64  `json:"amount,omitempty"`
	Reason string `json:"reason"`
}

type adminWalletResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Balance  int64  `json:"balance"`
	Held     int64  `json:"held_amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

func toAdminWalletResponse(w *domain.Wallet) adminWalletResponse {
	return adminWalletResponse{
		ID:       w.ID.String(),
		UserID:   w.UserID.String(),
		Balance:  w.Balance,
		Held:     w.HeldAmount,
		Currency: string(w.Currency),
		Status:   string(w.Status),
	}
}

// AdminCreditWallet godoc
//
//	@Summary		Admin credit wallet
//	@Description	Credits (adds funds to) a wallet by admin. Amount is in kopecks. Requires reason for audit log.
//	@Tags			admin-wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Wallet ID (UUID)"
//	@Param			body	body		adminWalletRequest	true	"Credit amount and reason"
//	@Success		200		{object}	APIResponse{data=walletTransactionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/wallets/{id}/credit [post]
func (h *WalletHandler) AdminCreditWallet(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid wallet id")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	var req adminWalletRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "amount must be positive")
		return
	}
	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	tx, err := h.walletService.AdminCredit(r.Context(), walletID, req.Amount, req.Reason, adminID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toWalletTransactionResponse(tx))
}

// AdminDebitWallet godoc
//
//	@Summary		Admin debit wallet
//	@Description	Debits (removes funds from) a wallet by admin. Amount is in kopecks. Requires reason for audit log.
//	@Tags			admin-wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Wallet ID (UUID)"
//	@Param			body	body		adminWalletRequest	true	"Debit amount and reason"
//	@Success		200		{object}	APIResponse{data=walletTransactionResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/wallets/{id}/debit [post]
func (h *WalletHandler) AdminDebitWallet(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid wallet id")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	var req adminWalletRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "amount must be positive")
		return
	}
	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	tx, err := h.walletService.AdminDebit(r.Context(), walletID, req.Amount, req.Reason, adminID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toWalletTransactionResponse(tx))
}

// AdminFreezeWallet godoc
//
//	@Summary		Admin freeze wallet
//	@Description	Freezes a wallet, preventing any transactions. Requires reason for audit log.
//	@Tags			admin-wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Wallet ID (UUID)"
//	@Param			body	body		adminWalletRequest	true	"Reason for freezing"
//	@Success		200		{object}	APIResponse{data=adminWalletResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/wallets/{id}/freeze [post]
func (h *WalletHandler) AdminFreezeWallet(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid wallet id")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	var req adminWalletRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	if err := h.walletService.AdminFreeze(r.Context(), walletID, req.Reason, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	wallet, err := h.walletService.GetWalletByID(r.Context(), walletID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAdminWalletResponse(wallet))
}

// AdminUnfreezeWallet godoc
//
//	@Summary		Admin unfreeze wallet
//	@Description	Unfreezes a frozen wallet, restoring normal operation. Requires reason for audit log.
//	@Tags			admin-wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Wallet ID (UUID)"
//	@Param			body	body		adminWalletRequest	true	"Reason for unfreezing"
//	@Success		200		{object}	APIResponse{data=adminWalletResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/wallets/{id}/unfreeze [post]
func (h *WalletHandler) AdminUnfreezeWallet(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid wallet id")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	var req adminWalletRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	if err := h.walletService.AdminUnfreeze(r.Context(), walletID, req.Reason, adminID); err != nil {
		handleServiceError(w, err)
		return
	}

	wallet, err := h.walletService.GetWalletByID(r.Context(), walletID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAdminWalletResponse(wallet))
}

// AdminGetWallet godoc
//
//	@Summary		Admin get wallet by ID
//	@Description	Returns wallet details by wallet ID. Admin only.
//	@Tags			admin-wallets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Wallet ID (UUID)"
//	@Success		200	{object}	APIResponse{data=adminWalletResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/wallets/{id} [get]
func (h *WalletHandler) AdminGetWallet(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid wallet id")
		return
	}

	wallet, err := h.walletService.GetWalletByID(r.Context(), walletID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAdminWalletResponse(wallet))
}

// AdminBatchCreditWallets godoc
//
//	@Summary		Batch credit wallets
//	@Description	Credits multiple wallets at once (up to 1000). Amount in kopecks. Processes in chunks of 100.
//	@Tags			admin-wallets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		batchCreditRequest	true	"Wallet IDs, amount, and reason"
//	@Success		200		{object}	APIResponse{data=batchDetailedResult}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/wallets/batch-credit [post]
func (h *WalletHandler) AdminBatchCreditWallets(w http.ResponseWriter, r *http.Request) {
	var req batchCreditRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "ids cannot be empty")
		return
	}
	if len(req.IDs) > maxBatchSize {
		writeError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("batch size cannot exceed %d", maxBatchSize))
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "amount must be positive")
		return
	}
	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	result := batchDetailedResult{
		Succeeded: make([]batchItemResult, 0),
		Failed:    make([]batchItemResult, 0),
	}

	for i := 0; i < len(req.IDs); i += batchChunkSize {
		end := i + batchChunkSize
		if end > len(req.IDs) {
			end = len(req.IDs)
		}
		chunk := req.IDs[i:end]

		for _, idStr := range chunk {
			walletID, err := uuid.Parse(idStr)
			if err != nil {
				result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: "invalid uuid"})
				continue
			}

			if _, err := h.walletService.AdminCredit(r.Context(), walletID, req.Amount, req.Reason, adminID); err != nil {
				result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: err.Error()})
			} else {
				result.Succeeded = append(result.Succeeded, batchItemResult{ID: idStr})
			}
		}
	}

	writeJSON(w, http.StatusOK, result)
}
