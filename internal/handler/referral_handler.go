package handler

import (
	"net/http"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type ReferralHandler struct {
	referralService service.ReferralService
	baseURL         string
}

func NewReferralHandler(referralService service.ReferralService, baseURL string) *ReferralHandler {
	return &ReferralHandler{referralService: referralService, baseURL: baseURL}
}

type referralCodeResponse struct {
	ReferralCode string `json:"referral_code"`
	ReferralLink string `json:"referral_link"`
}

type referralStatsResponse struct {
	TotalInvited   int   `json:"total_invited"`
	TotalCompleted int   `json:"total_completed"`
	TotalEarned    int64 `json:"total_earned"`
}

type referralBalanceResponse struct {
	Balance     int64 `json:"balance"`
	TotalEarned int64 `json:"total_earned"`
}

func toReferralStatsResponse(s *domain.ReferralStats) referralStatsResponse {
	return referralStatsResponse{
		TotalInvited:   s.TotalInvited,
		TotalCompleted: s.TotalCompleted,
		TotalEarned:    s.TotalEarned,
	}
}

func toReferralBalanceResponse(b *domain.ReferralBalance) referralBalanceResponse {
	return referralBalanceResponse{
		Balance:     b.Balance,
		TotalEarned: b.TotalEarned,
	}
}

// GetCode godoc
//
//	@Summary		Get referral code
//	@Description	Returns the referral code and shareable link for the authenticated user
//	@Tags			referral
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=referralCodeResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/referral [get]
func (h *ReferralHandler) GetCode(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	code, err := h.referralService.GenerateCode(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	baseURL := h.baseURL
	if baseURL == "" {
		baseURL = "https://bani.ru"
	}

	writeJSON(w, http.StatusOK, referralCodeResponse{
		ReferralCode: code,
		ReferralLink: baseURL + "/register?ref=" + code,
	})
}

// GetStats godoc
//
//	@Summary		Get referral statistics
//	@Description	Returns referral statistics: total invited, completed, and earned amount
//	@Tags			referral
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=referralStatsResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/referral/stats [get]
func (h *ReferralHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	stats, err := h.referralService.GetStats(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReferralStatsResponse(stats))
}

// GetBalance godoc
//
//	@Summary		Get referral balance
//	@Description	Returns the referral bonus balance and total earned for the authenticated user
//	@Tags			referral
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=referralBalanceResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/referral/balance [get]
func (h *ReferralHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	balance, err := h.referralService.GetBalance(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReferralBalanceResponse(balance))
}
