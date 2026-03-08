package handler

import (
	"net/http"

	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ReferralHandler struct {
	referralService service.ReferralService
}

func NewReferralHandler(referralService service.ReferralService) *ReferralHandler {
	return &ReferralHandler{referralService: referralService}
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

// GetCode returns the referral code and link for the authenticated user.
func (h *ReferralHandler) GetCode(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	code, err := h.referralService.GenerateCode(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	baseURL := scheme + "://" + r.Host

	writeJSON(w, http.StatusOK, referralCodeResponse{
		ReferralCode: code,
		ReferralLink: baseURL + "/register?ref=" + code,
	})
}

// GetStats returns referral statistics for the authenticated user.
func (h *ReferralHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	stats, err := h.referralService.GetStats(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReferralStatsResponse(stats))
}

// GetBalance returns the referral bonus balance for the authenticated user.
func (h *ReferralHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	balance, err := h.referralService.GetBalance(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toReferralBalanceResponse(balance))
}
