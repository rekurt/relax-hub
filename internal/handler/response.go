package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/nikitaaldaev/bani/internal/domain"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

// swagger response helpers (used only in annotations, not in runtime code)

type simpleMessageResponse struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

type unreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}

type pendingCountResponse struct {
	PendingCount int64 `json:"pending_count"`
}

type toggleFavoriteResponse struct {
	IsFavorite bool `json:"is_favorite"`
}

type myStatsResponse struct {
	TotalVisits int     `json:"total_visits"`
	TotalSpent  int64   `json:"total_spent"`
	AvgCheck    int64   `json:"avg_check"`
	ReviewCount int     `json:"review_count"`
	AvgRating   float64 `json:"avg_rating"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON response: %v\n", err)
	}
}

func writeJSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON response with meta: %v\n", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeErrorWithContext(w, nil, status, code, message)
}

func writeErrorWithContext(w http.ResponseWriter, _ *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode error response: %v\n", err)
	}
}

func handleServiceError(w http.ResponseWriter, err error) {
	handleServiceErrorWithRequest(w, nil, err)
}

func handleServiceErrorWithRequest(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		writeErrorWithContext(w, r, http.StatusConflict, "already_exists", err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		writeErrorWithContext(w, r, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		writeErrorWithContext(w, r, http.StatusUnauthorized, "unauthorized", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		writeErrorWithContext(w, r, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, domain.ErrSlotUnavailable):
		writeErrorWithContext(w, r, http.StatusConflict, "slot_unavailable", err.Error())
	case errors.Is(err, domain.ErrBookingCancelLate):
		writeErrorWithContext(w, r, http.StatusBadRequest, "cancel_too_late", err.Error())
	case errors.Is(err, domain.ErrUserBlocked):
		writeErrorWithContext(w, r, http.StatusForbidden, "user_blocked", err.Error())
	case errors.Is(err, domain.ErrBathhouseNotActive):
		writeErrorWithContext(w, r, http.StatusBadRequest, "bathhouse_not_active", err.Error())
	case errors.Is(err, domain.ErrBathhouseHasBookings):
		writeErrorWithContext(w, r, http.StatusConflict, "bathhouse_has_bookings", err.Error())
	case errors.Is(err, domain.ErrReviewAlreadyResponded):
		writeErrorWithContext(w, r, http.StatusConflict, "review_already_responded", err.Error())
	case errors.Is(err, domain.ErrSocialAccountAlreadyLinked):
		writeErrorWithContext(w, r, http.StatusConflict, "social_account_already_linked", err.Error())
	case errors.Is(err, domain.ErrSocialAccountNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "social_account_not_found", err.Error())
	case errors.Is(err, domain.ErrOAuthExchangeFailed):
		writeErrorWithContext(w, r, http.StatusBadRequest, "oauth_exchange_failed", "oauth code exchange failed")
	case errors.Is(err, domain.ErrSubscriptionNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "subscription_not_found", err.Error())
	case errors.Is(err, domain.ErrSubscriptionAlreadyActive):
		writeErrorWithContext(w, r, http.StatusConflict, "subscription_already_active", err.Error())
	case errors.Is(err, domain.ErrPromotionBudgetExhausted):
		writeErrorWithContext(w, r, http.StatusConflict, "promotion_budget_exhausted", err.Error())
	case errors.Is(err, domain.ErrInsufficientPoints):
		writeErrorWithContext(w, r, http.StatusBadRequest, "insufficient_points", err.Error())
	case errors.Is(err, domain.ErrComplaintNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "complaint_not_found", err.Error())
	case errors.Is(err, domain.ErrAlreadyReported):
		writeErrorWithContext(w, r, http.StatusConflict, "already_reported", err.Error())
	case errors.Is(err, domain.ErrSelfReferral):
		writeErrorWithContext(w, r, http.StatusBadRequest, "self_referral", err.Error())
	case errors.Is(err, domain.ErrAlreadyReferred):
		writeErrorWithContext(w, r, http.StatusConflict, "already_referred", err.Error())
	case errors.Is(err, domain.ErrInsufficientReferralBalance):
		writeErrorWithContext(w, r, http.StatusBadRequest, "insufficient_referral_balance", err.Error())
	case errors.Is(err, domain.ErrCertificateNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "certificate_not_found", err.Error())
	case errors.Is(err, domain.ErrCertificateExpired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "certificate_expired", err.Error())
	case errors.Is(err, domain.ErrCertificateInsufficientBalance):
		writeErrorWithContext(w, r, http.StatusBadRequest, "certificate_insufficient_balance", err.Error())
	case errors.Is(err, domain.ErrPhotoNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "photo_not_found", err.Error())
	case errors.Is(err, domain.ErrPromoNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "promo_not_found", err.Error())
	case errors.Is(err, domain.ErrPromoExpired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "promo_expired", err.Error())
	case errors.Is(err, domain.ErrPromoMaxUses):
		writeErrorWithContext(w, r, http.StatusConflict, "promo_max_uses", err.Error())
	case errors.Is(err, domain.ErrPromoMinAmount):
		writeErrorWithContext(w, r, http.StatusBadRequest, "promo_min_amount", err.Error())
	case errors.Is(err, domain.ErrPromoInvalid):
		writeErrorWithContext(w, r, http.StatusBadRequest, "promo_invalid", err.Error())
	case errors.Is(err, domain.ErrMediaNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "media_not_found", err.Error())
	case errors.Is(err, domain.ErrMediaFileTooLarge):
		writeErrorWithContext(w, r, http.StatusBadRequest, "media_file_too_large", err.Error())
	case errors.Is(err, domain.ErrMediaInvalidType):
		writeErrorWithContext(w, r, http.StatusBadRequest, "media_invalid_type", err.Error())
	case errors.Is(err, domain.ErrMediaLimitReached):
		writeErrorWithContext(w, r, http.StatusConflict, "media_limit_reached", err.Error())
	case errors.Is(err, domain.ErrPaymentNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "payment_not_found", err.Error())
	case errors.Is(err, domain.ErrPaymentAlreadyProcessed):
		writeErrorWithContext(w, r, http.StatusConflict, "payment_already_processed", err.Error())
	case errors.Is(err, domain.ErrRefundExceedsAmount):
		writeErrorWithContext(w, r, http.StatusBadRequest, "refund_exceeds_amount", err.Error())
	case errors.Is(err, domain.ErrPaymentFailed):
		writeErrorWithContext(w, r, http.StatusBadRequest, "payment_failed", err.Error())
	case errors.Is(err, domain.ErrWalletNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "wallet_not_found", err.Error())
	case errors.Is(err, domain.ErrInsufficientWalletBalance):
		writeErrorWithContext(w, r, http.StatusBadRequest, "insufficient_wallet_balance", err.Error())
	case errors.Is(err, domain.ErrWalletLimitExceeded):
		writeErrorWithContext(w, r, http.StatusBadRequest, "wallet_limit_exceeded", err.Error())
	case errors.Is(err, domain.ErrWalletFrozen):
		writeErrorWithContext(w, r, http.StatusForbidden, "wallet_frozen", err.Error())
	case errors.Is(err, domain.ErrWalletConcurrentUpdate):
		writeErrorWithContext(w, r, http.StatusConflict, "wallet_concurrent_update", err.Error())
	case errors.Is(err, domain.ErrHoldNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "hold_not_found", err.Error())
	case errors.Is(err, domain.ErrHoldExpired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "hold_expired", err.Error())
	case errors.Is(err, domain.ErrTopUpBelowMinimum):
		writeErrorWithContext(w, r, http.StatusBadRequest, "topup_below_minimum", err.Error())
	case errors.Is(err, domain.ErrTopUpAboveMaximum):
		writeErrorWithContext(w, r, http.StatusBadRequest, "topup_above_maximum", err.Error())
	case errors.Is(err, domain.ErrPayoutNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "payout_not_found", err.Error())
	case errors.Is(err, domain.ErrPayoutBelowMinimum):
		writeErrorWithContext(w, r, http.StatusBadRequest, "payout_below_minimum", err.Error())
	case errors.Is(err, domain.ErrPayoutDailyLimitExceeded):
		writeErrorWithContext(w, r, http.StatusBadRequest, "payout_daily_limit_exceeded", err.Error())
	case errors.Is(err, domain.ErrPayoutMonthlyLimitExceeded):
		writeErrorWithContext(w, r, http.StatusBadRequest, "payout_monthly_limit_exceeded", err.Error())
	case errors.Is(err, domain.ErrPayoutAlreadyProcessed):
		writeErrorWithContext(w, r, http.StatusConflict, "payout_already_processed", err.Error())
	case errors.Is(err, domain.ErrOTPRateLimited):
		writeErrorWithContext(w, r, http.StatusTooManyRequests, "otp_rate_limited", err.Error())
	case errors.Is(err, domain.ErrOTPInvalid):
		writeErrorWithContext(w, r, http.StatusBadRequest, "otp_invalid", err.Error())
	case errors.Is(err, domain.ErrOTPExpired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "otp_expired", err.Error())
	case errors.Is(err, domain.ErrOTPMaxAttempts):
		writeErrorWithContext(w, r, http.StatusTooManyRequests, "otp_max_attempts", err.Error())
	case errors.Is(err, domain.ErrPhoneRequired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "phone_required", err.Error())
	case errors.Is(err, domain.ErrPhoneInvalid):
		writeErrorWithContext(w, r, http.StatusBadRequest, "phone_invalid", err.Error())
	case errors.Is(err, domain.Err2FARequired):
		writeErrorWithContext(w, r, http.StatusForbidden, "2fa_required", err.Error())
	case errors.Is(err, domain.Err2FAAlreadyEnabled):
		writeErrorWithContext(w, r, http.StatusConflict, "2fa_already_enabled", err.Error())
	case errors.Is(err, domain.Err2FANotEnabled):
		writeErrorWithContext(w, r, http.StatusBadRequest, "2fa_not_enabled", err.Error())
	case errors.Is(err, domain.Err2FAInvalidCode):
		writeErrorWithContext(w, r, http.StatusBadRequest, "2fa_invalid_code", err.Error())
	case errors.Is(err, domain.Err2FAPhoneRequired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "2fa_phone_required", err.Error())
	case errors.Is(err, domain.ErrSessionNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "session_not_found", err.Error())
	case errors.Is(err, domain.ErrSessionExpired):
		writeErrorWithContext(w, r, http.StatusUnauthorized, "session_expired", err.Error())
	case errors.Is(err, domain.ErrResetTokenInvalid):
		writeErrorWithContext(w, r, http.StatusBadRequest, "reset_token_invalid", err.Error())
	case errors.Is(err, domain.ErrResetRateLimited):
		writeErrorWithContext(w, r, http.StatusTooManyRequests, "reset_rate_limited", err.Error())
	case errors.Is(err, domain.ErrAccountDeletionPending):
		writeErrorWithContext(w, r, http.StatusConflict, "account_deletion_pending", err.Error())
	case errors.Is(err, domain.ErrAccountDeletionNotPending):
		writeErrorWithContext(w, r, http.StatusBadRequest, "account_deletion_not_pending", err.Error())
	case errors.Is(err, domain.ErrAccountDeleted):
		writeErrorWithContext(w, r, http.StatusForbidden, "account_deleted", err.Error())
	case errors.Is(err, domain.ErrKYCNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "kyc_not_found", err.Error())
	case errors.Is(err, domain.ErrKYCNotApproved):
		writeErrorWithContext(w, r, http.StatusForbidden, "kyc_not_approved", err.Error())
	case errors.Is(err, domain.ErrKYCPending):
		writeErrorWithContext(w, r, http.StatusConflict, "kyc_pending", err.Error())
	case errors.Is(err, domain.ErrOfferNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "offer_not_found", err.Error())
	case errors.Is(err, domain.ErrOfferNotAccepted):
		writeErrorWithContext(w, r, http.StatusForbidden, "offer_not_accepted", err.Error())
	case errors.Is(err, domain.ErrOfferAlreadyAccepted):
		writeErrorWithContext(w, r, http.StatusConflict, "offer_already_accepted", err.Error())
	case errors.Is(err, domain.ErrPaymentDetailsNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "payment_details_not_found", err.Error())
	case errors.Is(err, domain.ErrPaymentDetailsNotSet):
		writeErrorWithContext(w, r, http.StatusForbidden, "payment_details_not_set", err.Error())
	case errors.Is(err, domain.ErrListingDraftNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "listing_draft_not_found", err.Error())
	case errors.Is(err, domain.ErrListingDraftIncomplete):
		writeErrorWithContext(w, r, http.StatusBadRequest, "listing_draft_incomplete", err.Error())
	case errors.Is(err, domain.ErrListingDraftSubmitted):
		writeErrorWithContext(w, r, http.StatusConflict, "listing_draft_submitted", err.Error())
	case errors.Is(err, domain.ErrListingDraftInvalidStep):
		writeErrorWithContext(w, r, http.StatusBadRequest, "listing_draft_invalid_step", err.Error())
	case errors.Is(err, domain.ErrListingIncomplete):
		writeErrorWithContext(w, r, http.StatusBadRequest, "listing_incomplete", err.Error())
	case errors.Is(err, domain.ErrAddOnNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "addon_not_found", err.Error())
	case errors.Is(err, domain.ErrAddOnLimitReached):
		writeErrorWithContext(w, r, http.StatusConflict, "addon_limit_reached", err.Error())
	case errors.Is(err, domain.ErrSavedSearchNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "saved_search_not_found", err.Error())
	case errors.Is(err, domain.ErrSavedSearchLimitReached):
		writeErrorWithContext(w, r, http.StatusConflict, "saved_search_limit_reached", err.Error())
	case errors.Is(err, domain.ErrBookingModificationLimit):
		writeErrorWithContext(w, r, http.StatusConflict, "booking_modification_limit", err.Error())
	case errors.Is(err, domain.ErrBookingNotModifiable):
		writeErrorWithContext(w, r, http.StatusBadRequest, "booking_not_modifiable", err.Error())
	case errors.Is(err, domain.ErrCheckinTooEarly):
		writeErrorWithContext(w, r, http.StatusBadRequest, "checkin_too_early", err.Error())
	case errors.Is(err, domain.ErrCheckinTooLate):
		writeErrorWithContext(w, r, http.StatusBadRequest, "checkin_too_late", err.Error())
	case errors.Is(err, domain.ErrNotCheckedIn):
		writeErrorWithContext(w, r, http.StatusBadRequest, "not_checked_in", err.Error())
	case errors.Is(err, domain.ErrNoShowDisputeExpired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "noshow_dispute_expired", err.Error())
	case errors.Is(err, domain.ErrEscrowNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "escrow_not_found", err.Error())
	case errors.Is(err, domain.ErrEscrowNotMatured):
		writeErrorWithContext(w, r, http.StatusBadRequest, "escrow_not_matured", err.Error())
	case errors.Is(err, domain.ErrEscrowAlreadyReleased):
		writeErrorWithContext(w, r, http.StatusConflict, "escrow_already_released", err.Error())
	case errors.Is(err, domain.ErrEscrowDisputed):
		writeErrorWithContext(w, r, http.StatusConflict, "escrow_disputed", err.Error())
	case errors.Is(err, domain.ErrBroadcastNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "broadcast_not_found", err.Error())
	case errors.Is(err, domain.ErrBroadcastRateLimit):
		writeErrorWithContext(w, r, http.StatusTooManyRequests, "broadcast_rate_limit", err.Error())
	case errors.Is(err, domain.ErrBroadcastNotDraft):
		writeErrorWithContext(w, r, http.StatusBadRequest, "broadcast_not_draft", err.Error())
	case errors.Is(err, domain.ErrAutoScenarioNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "auto_scenario_not_found", err.Error())
	case errors.Is(err, domain.ErrTemplateNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "template_not_found", err.Error())
	case errors.Is(err, domain.ErrTemplateLimitReached):
		writeErrorWithContext(w, r, http.StatusConflict, "template_limit_reached", err.Error())
	case errors.Is(err, domain.ErrTicketNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "ticket_not_found", err.Error())
	case errors.Is(err, domain.ErrTicketAlreadyClosed):
		writeErrorWithContext(w, r, http.StatusConflict, "ticket_already_closed", err.Error())
	case errors.Is(err, domain.ErrTicketAlreadyResolved):
		writeErrorWithContext(w, r, http.StatusConflict, "ticket_already_resolved", err.Error())
	case errors.Is(err, domain.ErrTicketAlreadyEscalated):
		writeErrorWithContext(w, r, http.StatusConflict, "ticket_already_escalated", err.Error())
	case errors.Is(err, domain.ErrCSATAlreadySubmitted):
		writeErrorWithContext(w, r, http.StatusConflict, "csat_already_submitted", err.Error())
	case errors.Is(err, domain.ErrCSATNotResolved):
		writeErrorWithContext(w, r, http.StatusBadRequest, "csat_not_resolved", err.Error())
	case errors.Is(err, domain.ErrCSATInvalidScore):
		writeErrorWithContext(w, r, http.StatusBadRequest, "csat_invalid_score", err.Error())
	case errors.Is(err, domain.ErrDisputeNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "dispute_not_found", err.Error())
	case errors.Is(err, domain.ErrDisputeAlreadyExists):
		writeErrorWithContext(w, r, http.StatusConflict, "dispute_already_exists", err.Error())
	case errors.Is(err, domain.ErrDisputeAlreadyResolved):
		writeErrorWithContext(w, r, http.StatusConflict, "dispute_already_resolved", err.Error())
	case errors.Is(err, domain.ErrDisputeAlreadyClosed):
		writeErrorWithContext(w, r, http.StatusConflict, "dispute_already_closed", err.Error())
	case errors.Is(err, domain.ErrDisputeEvidenceWindowExpired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "dispute_evidence_window_expired", err.Error())
	case errors.Is(err, domain.ErrDisputeAppealExpired):
		writeErrorWithContext(w, r, http.StatusBadRequest, "dispute_appeal_expired", err.Error())
	case errors.Is(err, domain.ErrDisputeNotResolved):
		writeErrorWithContext(w, r, http.StatusBadRequest, "dispute_not_resolved", err.Error())
	case errors.Is(err, domain.ErrDisputeAlreadyAppealed):
		writeErrorWithContext(w, r, http.StatusConflict, "dispute_already_appealed", err.Error())
	case errors.Is(err, domain.ErrFraudDetected):
		writeErrorWithContext(w, r, http.StatusForbidden, "fraud_detected", err.Error())
	case errors.Is(err, domain.ErrDepositNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "deposit_not_found", err.Error())
	case errors.Is(err, domain.ErrDepositAlreadyReleased):
		writeErrorWithContext(w, r, http.StatusConflict, "deposit_already_released", err.Error())
	case errors.Is(err, domain.ErrDepositAlreadyClaimed):
		writeErrorWithContext(w, r, http.StatusConflict, "deposit_already_claimed", err.Error())
	default:
		writeErrorWithContext(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

const maxBodySize = 1 << 20 // 1 MB

func readJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if r.Body == nil {
		return domain.ErrInvalidInput
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return domain.ErrInvalidInput
	}
	return nil
}
