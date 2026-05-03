package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rekurt/relax-hub/internal/domain"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "test_error", "Test error message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false")
	}

	if resp.Error == nil {
		t.Fatal("Expected error to be set")
	}

	if resp.Error.Code != "test_error" {
		t.Errorf("Expected code 'test_error', got '%s'", resp.Error.Code)
	}

	if resp.Error.Message != "Test error message" {
		t.Errorf("Expected message 'Test error message', got '%s'", resp.Error.Message)
	}
}

func TestWriteErrorWithContext_5xxLogging(t *testing.T) {
	// Test that 5xx errors are logged
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	writeErrorWithContext(w, r, http.StatusInternalServerError, "internal_error", "Internal error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("Expected success=false")
	}
}

func TestHandleServiceError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrNotFound)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "not_found" {
		t.Error("Expected error code 'not_found'")
	}
}

func TestHandleServiceError_AlreadyExists(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrAlreadyExists)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "already_exists" {
		t.Error("Expected error code 'already_exists'")
	}
}

func TestHandleServiceError_InvalidInput(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrInvalidInput)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "invalid_input" {
		t.Error("Expected error code 'invalid_input'")
	}
}

func TestHandleServiceError_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrUnauthorized)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "unauthorized" {
		t.Error("Expected error code 'unauthorized'")
	}
}

func TestHandleServiceError_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	handleServiceError(w, domain.ErrForbidden)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "forbidden" {
		t.Error("Expected error code 'forbidden'")
	}
}

func TestHandleServiceError_InternalError(t *testing.T) {
	w := httptest.NewRecorder()
	// Test with a generic error that doesn't match any domain error
	unknownErr := errors.New("unknown error")
	handleServiceError(w, unknownErr)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "internal_error" {
		t.Error("Expected error code 'internal_error'")
	}
}

func TestReadJSON_Valid(t *testing.T) {
	type testData struct {
		Name string `json:"name"`
	}

	body := bytes.NewReader([]byte(`{"name":"test"}`))
	r := httptest.NewRequest("POST", "/test", body)
	w := httptest.NewRecorder()

	var data testData
	err := readJSON(w, r, &data)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if data.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", data.Name)
	}
}

func TestReadJSON_InvalidJSON(t *testing.T) {
	body := bytes.NewReader([]byte(`invalid json`))
	r := httptest.NewRequest("POST", "/test", body)
	w := httptest.NewRecorder()

	var data map[string]interface{}
	err := readJSON(w, r, &data)

	if err != domain.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestReadJSON_NilBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	var data map[string]interface{}
	err := readJSON(w, r, &data)

	if err != domain.ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestHandleServiceError_OAuthExchangeFailed_NoInternalDetails(t *testing.T) {
	w := httptest.NewRecorder()
	// Wrap with internal details that should NOT be exposed
	err := errors.Join(domain.ErrOAuthExchangeFailed, errors.New("Post \"https://accounts.google.com/o/oauth2/token\": 400 Bad Request"))
	handleServiceError(w, err)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Error == nil || resp.Error.Code != "oauth_exchange_failed" {
		t.Error("Expected error code 'oauth_exchange_failed'")
	}

	if resp.Error.Message != "oauth code exchange failed" {
		t.Errorf("Expected static message 'oauth code exchange failed', got '%s'", resp.Error.Message)
	}
}

func TestWriteJSONWithMeta(t *testing.T) {
	w := httptest.NewRecorder()
	data := []int{1, 2, 3}
	meta := &Meta{
		Page:       1,
		PageSize:   10,
		TotalCount: 100,
		TotalPages: 10,
	}

	writeJSONWithMeta(w, http.StatusOK, data, meta)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}

	if resp.Meta == nil {
		t.Fatal("Expected meta to be set")
	}

	if resp.Meta.Page != 1 {
		t.Errorf("Expected page 1, got %d", resp.Meta.Page)
	}

	if resp.Meta.TotalPages != 10 {
		t.Errorf("Expected total_pages 10, got %d", resp.Meta.TotalPages)
	}
}

// TestHandleServiceError_ErrorMappings exercises every domain-error -> HTTP-status
// branch in handleServiceErrorWithRequest. Failing cases here mean either a
// mapping was added without a matching case or the HTTP status / code drifted.
func TestHandleServiceError_ErrorMappings(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrNotFound, http.StatusNotFound, "not_found"},
		{domain.ErrAlreadyExists, http.StatusConflict, "already_exists"},
		{domain.ErrInvalidInput, http.StatusBadRequest, "invalid_input"},
		{domain.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
		{domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{domain.ErrSlotUnavailable, http.StatusConflict, "slot_unavailable"},
		{domain.ErrBookingCancelLate, http.StatusBadRequest, "cancel_too_late"},
		{domain.ErrUserBlocked, http.StatusForbidden, "user_blocked"},
		{domain.ErrBathhouseNotActive, http.StatusBadRequest, "bathhouse_not_active"},
		{domain.ErrBathhouseHasBookings, http.StatusConflict, "bathhouse_has_bookings"},
		{domain.ErrReviewAlreadyResponded, http.StatusConflict, "review_already_responded"},
		{domain.ErrSocialAccountAlreadyLinked, http.StatusConflict, "social_account_already_linked"},
		{domain.ErrSocialAccountNotFound, http.StatusNotFound, "social_account_not_found"},
		{domain.ErrOAuthExchangeFailed, http.StatusBadRequest, "oauth_exchange_failed"},
		{domain.ErrSubscriptionNotFound, http.StatusNotFound, "subscription_not_found"},
		{domain.ErrSubscriptionAlreadyActive, http.StatusConflict, "subscription_already_active"},
		{domain.ErrPromotionBudgetExhausted, http.StatusConflict, "promotion_budget_exhausted"},
		{domain.ErrInsufficientPoints, http.StatusBadRequest, "insufficient_points"},
		{domain.ErrComplaintNotFound, http.StatusNotFound, "complaint_not_found"},
		{domain.ErrAlreadyReported, http.StatusConflict, "already_reported"},
		{domain.ErrSelfReferral, http.StatusBadRequest, "self_referral"},
		{domain.ErrAlreadyReferred, http.StatusConflict, "already_referred"},
		{domain.ErrInsufficientReferralBalance, http.StatusBadRequest, "insufficient_referral_balance"},
		{domain.ErrCertificateNotFound, http.StatusNotFound, "certificate_not_found"},
		{domain.ErrCertificateOrderNotFound, http.StatusNotFound, "certificate_order_not_found"},
		{domain.ErrCertificateExpired, http.StatusBadRequest, "certificate_expired"},
		{domain.ErrCertificateInsufficientBalance, http.StatusBadRequest, "certificate_insufficient_balance"},
		{domain.ErrPhotoNotFound, http.StatusNotFound, "photo_not_found"},
		{domain.ErrPromoNotFound, http.StatusNotFound, "promo_not_found"},
		{domain.ErrPromoExpired, http.StatusBadRequest, "promo_expired"},
		{domain.ErrPromoMaxUses, http.StatusConflict, "promo_max_uses"},
		{domain.ErrPromoMinAmount, http.StatusBadRequest, "promo_min_amount"},
		{domain.ErrPromoInvalid, http.StatusBadRequest, "promo_invalid"},
		{domain.ErrMediaNotFound, http.StatusNotFound, "media_not_found"},
		{domain.ErrMediaFileTooLarge, http.StatusBadRequest, "media_file_too_large"},
		{domain.ErrMediaInvalidType, http.StatusBadRequest, "media_invalid_type"},
		{domain.ErrMediaLimitReached, http.StatusConflict, "media_limit_reached"},
		{domain.ErrPaymentNotFound, http.StatusNotFound, "payment_not_found"},
		{domain.ErrPaymentAlreadyProcessed, http.StatusConflict, "payment_already_processed"},
		{domain.ErrRefundExceedsAmount, http.StatusBadRequest, "refund_exceeds_amount"},
		{domain.ErrPaymentFailed, http.StatusBadRequest, "payment_failed"},
		{domain.ErrWalletNotFound, http.StatusNotFound, "wallet_not_found"},
		{domain.ErrInsufficientWalletBalance, http.StatusBadRequest, "insufficient_wallet_balance"},
		{domain.ErrWalletLimitExceeded, http.StatusBadRequest, "wallet_limit_exceeded"},
		{domain.ErrWalletFrozen, http.StatusForbidden, "wallet_frozen"},
		{domain.ErrWalletConcurrentUpdate, http.StatusConflict, "wallet_concurrent_update"},
		{domain.ErrHoldNotFound, http.StatusNotFound, "hold_not_found"},
		{domain.ErrHoldExpired, http.StatusBadRequest, "hold_expired"},
		{domain.ErrTopUpBelowMinimum, http.StatusBadRequest, "topup_below_minimum"},
		{domain.ErrTopUpAboveMaximum, http.StatusBadRequest, "topup_above_maximum"},
		{domain.ErrPayoutNotFound, http.StatusNotFound, "payout_not_found"},
		{domain.ErrPayoutBelowMinimum, http.StatusBadRequest, "payout_below_minimum"},
		{domain.ErrPayoutDailyLimitExceeded, http.StatusBadRequest, "payout_daily_limit_exceeded"},
		{domain.ErrPayoutMonthlyLimitExceeded, http.StatusBadRequest, "payout_monthly_limit_exceeded"},
		{domain.ErrPayoutAlreadyProcessed, http.StatusConflict, "payout_already_processed"},
		{domain.ErrOTPRateLimited, http.StatusTooManyRequests, "otp_rate_limited"},
		{domain.ErrOTPInvalid, http.StatusBadRequest, "otp_invalid"},
		{domain.ErrOTPExpired, http.StatusBadRequest, "otp_expired"},
		{domain.ErrOTPMaxAttempts, http.StatusTooManyRequests, "otp_max_attempts"},
		{domain.ErrPhoneRequired, http.StatusBadRequest, "phone_required"},
		{domain.ErrPhoneInvalid, http.StatusBadRequest, "phone_invalid"},
		{domain.Err2FARequired, http.StatusForbidden, "2fa_required"},
		{domain.Err2FAAlreadyEnabled, http.StatusConflict, "2fa_already_enabled"},
		{domain.Err2FANotEnabled, http.StatusBadRequest, "2fa_not_enabled"},
		{domain.Err2FAInvalidCode, http.StatusBadRequest, "2fa_invalid_code"},
		{domain.Err2FAPhoneRequired, http.StatusBadRequest, "2fa_phone_required"},
		{domain.ErrSessionNotFound, http.StatusNotFound, "session_not_found"},
		{domain.ErrSessionExpired, http.StatusUnauthorized, "session_expired"},
		{domain.ErrResetTokenInvalid, http.StatusBadRequest, "reset_token_invalid"},
		{domain.ErrResetRateLimited, http.StatusTooManyRequests, "reset_rate_limited"},
		{domain.ErrAccountDeletionPending, http.StatusConflict, "account_deletion_pending"},
		{domain.ErrAccountDeletionNotPending, http.StatusBadRequest, "account_deletion_not_pending"},
		{domain.ErrAccountDeleted, http.StatusForbidden, "account_deleted"},
		{domain.ErrKYCNotFound, http.StatusNotFound, "kyc_not_found"},
		{domain.ErrKYCNotApproved, http.StatusForbidden, "kyc_not_approved"},
		{domain.ErrKYCPending, http.StatusConflict, "kyc_pending"},
		{domain.ErrOfferNotFound, http.StatusNotFound, "offer_not_found"},
		{domain.ErrOfferNotAccepted, http.StatusForbidden, "offer_not_accepted"},
		{domain.ErrOfferAlreadyAccepted, http.StatusConflict, "offer_already_accepted"},
		{domain.ErrPaymentDetailsNotFound, http.StatusNotFound, "payment_details_not_found"},
		{domain.ErrPaymentDetailsNotSet, http.StatusForbidden, "payment_details_not_set"},
		{domain.ErrListingDraftNotFound, http.StatusNotFound, "listing_draft_not_found"},
		{domain.ErrListingDraftIncomplete, http.StatusBadRequest, "listing_draft_incomplete"},
		{domain.ErrListingDraftSubmitted, http.StatusConflict, "listing_draft_submitted"},
		{domain.ErrListingDraftInvalidStep, http.StatusBadRequest, "listing_draft_invalid_step"},
		{domain.ErrListingIncomplete, http.StatusBadRequest, "listing_incomplete"},
		{domain.ErrAddOnNotFound, http.StatusNotFound, "addon_not_found"},
		{domain.ErrAddOnLimitReached, http.StatusConflict, "addon_limit_reached"},
		{domain.ErrSavedSearchNotFound, http.StatusNotFound, "saved_search_not_found"},
		{domain.ErrSavedSearchLimitReached, http.StatusConflict, "saved_search_limit_reached"},
		{domain.ErrSavedCardNotFound, http.StatusNotFound, "saved_card_not_found"},
		{domain.ErrSavedCardLimitReached, http.StatusConflict, "saved_card_limit_reached"},
		{domain.ErrBookingModificationLimit, http.StatusBadRequest, "booking_modification_limit"},
		{domain.ErrBookingNotModifiable, http.StatusBadRequest, "booking_not_modifiable"},
		{domain.ErrModificationRequestNotFound, http.StatusNotFound, "modification_request_not_found"},
		{domain.ErrModificationRequestPending, http.StatusConflict, "modification_request_pending"},
		{domain.ErrModificationRequestExpired, http.StatusBadRequest, "modification_request_expired"},
		{domain.ErrExtensionRequestNotFound, http.StatusNotFound, "extension_request_not_found"},
		{domain.ErrExtensionRequestPending, http.StatusConflict, "extension_request_pending"},
		{domain.ErrExtensionRequestExpired, http.StatusBadRequest, "extension_request_expired"},
		{domain.ErrCheckinTooEarly, http.StatusBadRequest, "checkin_too_early"},
		{domain.ErrCheckinTooLate, http.StatusBadRequest, "checkin_too_late"},
		{domain.ErrNotCheckedIn, http.StatusBadRequest, "not_checked_in"},
		{domain.ErrNoShowDisputeExpired, http.StatusBadRequest, "noshow_dispute_expired"},
		{domain.ErrEscrowNotFound, http.StatusNotFound, "escrow_not_found"},
		{domain.ErrEscrowNotMatured, http.StatusBadRequest, "escrow_not_matured"},
		{domain.ErrEscrowAlreadyReleased, http.StatusConflict, "escrow_already_released"},
		{domain.ErrEscrowDisputed, http.StatusConflict, "escrow_disputed"},
		{domain.ErrBroadcastNotFound, http.StatusNotFound, "broadcast_not_found"},
		{domain.ErrBroadcastRateLimit, http.StatusTooManyRequests, "broadcast_rate_limit"},
		{domain.ErrBroadcastNotDraft, http.StatusBadRequest, "broadcast_not_draft"},
		{domain.ErrAutoScenarioNotFound, http.StatusNotFound, "auto_scenario_not_found"},
		{domain.ErrTemplateNotFound, http.StatusNotFound, "template_not_found"},
		{domain.ErrTemplateLimitReached, http.StatusConflict, "template_limit_reached"},
		{domain.ErrTicketNotFound, http.StatusNotFound, "ticket_not_found"},
		{domain.ErrTicketAlreadyClosed, http.StatusConflict, "ticket_already_closed"},
		{domain.ErrTicketAlreadyResolved, http.StatusConflict, "ticket_already_resolved"},
		{domain.ErrTicketAlreadyEscalated, http.StatusConflict, "ticket_already_escalated"},
		{domain.ErrCSATAlreadySubmitted, http.StatusConflict, "csat_already_submitted"},
		{domain.ErrCSATNotResolved, http.StatusBadRequest, "csat_not_resolved"},
		{domain.ErrCSATInvalidScore, http.StatusBadRequest, "csat_invalid_score"},
		{domain.ErrDisputeNotFound, http.StatusNotFound, "dispute_not_found"},
		{domain.ErrDisputeAlreadyExists, http.StatusConflict, "dispute_already_exists"},
		{domain.ErrDisputeAlreadyResolved, http.StatusConflict, "dispute_already_resolved"},
		{domain.ErrDisputeAlreadyClosed, http.StatusConflict, "dispute_already_closed"},
		{domain.ErrDisputeEvidenceWindowExpired, http.StatusBadRequest, "dispute_evidence_window_expired"},
		{domain.ErrDisputeAppealExpired, http.StatusBadRequest, "dispute_appeal_expired"},
		{domain.ErrDisputeNotResolved, http.StatusBadRequest, "dispute_not_resolved"},
		{domain.ErrDisputeAlreadyAppealed, http.StatusConflict, "dispute_already_appealed"},
		{domain.ErrBankEntryNotFound, http.StatusNotFound, "bank_entry_not_found"},
		{domain.ErrBankEntryAlreadyMatched, http.StatusConflict, "bank_entry_already_matched"},
		{domain.ErrFraudDetected, http.StatusForbidden, "fraud_detected"},
		{domain.ErrDepositNotFound, http.StatusNotFound, "deposit_not_found"},
		{domain.ErrDepositAlreadyReleased, http.StatusConflict, "deposit_already_released"},
		{domain.ErrDepositAlreadyClaimed, http.StatusConflict, "deposit_already_claimed"},
		{domain.ErrClientReviewNotFound, http.StatusNotFound, "client_review_not_found"},
		{domain.ErrReviewBlindPeriod, http.StatusForbidden, "review_blind_period"},
		{domain.ErrRegionSwitchBlocked, http.StatusConflict, "region_switch_blocked"},
		{domain.ErrRegionSameAsCurrent, http.StatusBadRequest, "region_same_as_current"},
		{domain.ErrRegionInvalid, http.StatusBadRequest, "region_invalid"},
		{domain.ErrAdminRoleNotFound, http.StatusNotFound, "admin_role_not_found"},
		{domain.ErrAdmin2FARequired, http.StatusForbidden, "admin_2fa_required"},
		{domain.ErrAdminPermissionDenied, http.StatusForbidden, "admin_permission_denied"},
		{domain.ErrFAQNotFound, http.StatusNotFound, "faq_not_found"},
		{domain.ErrWebhookNotFound, http.StatusNotFound, "webhook_not_found"},
		{domain.ErrWebhookLimitReached, http.StatusConflict, "webhook_limit_reached"},
		{domain.ErrPMSConnectionNotFound, http.StatusNotFound, "pms_connection_not_found"},
		{domain.ErrPMSConnectionLimitReached, http.StatusConflict, "pms_connection_limit_reached"},
		{domain.ErrPMSConnectionAlreadyExists, http.StatusConflict, "pms_connection_already_exists"},
		{domain.ErrPMSSyncFailed, http.StatusBadRequest, "pms_sync_failed"},
		{domain.ErrPhotoOrderNotFound, http.StatusNotFound, "photo_order_not_found"},
		{domain.ErrPhotoOrderInvalidStatus, http.StatusBadRequest, "photo_order_invalid_status"},
	}

	r := httptest.NewRequest("GET", "/test", nil)
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			w := httptest.NewRecorder()
			handleServiceErrorWithRequest(w, r, tc.err)

			if w.Code != tc.status {
				t.Errorf("status: got %d, want %d", w.Code, tc.status)
			}

			var resp APIResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Success {
				t.Error("Success should be false")
			}
			if resp.Error == nil || resp.Error.Code != tc.code {
				got := ""
				if resp.Error != nil {
					got = resp.Error.Code
				}
				t.Errorf("code: got %q, want %q", got, tc.code)
			}
		})
	}
}

// TestHandleServiceError_PaymentFailedError covers the inner branch where
// ErrPaymentFailed is wrapped by a *domain.PaymentFailedError carrying a
// localised message and suggestion.
func TestHandleServiceError_PaymentFailedError(t *testing.T) {
	pfe := &domain.PaymentFailedError{
		Code:       "card_declined",
		MessageRU:  "Карта отклонена банком",
		Suggestion: "Попробуйте другую карту",
	}

	w := httptest.NewRecorder()
	handleServiceError(w, pfe)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("Error must be set")
	}
	if resp.Error.Code != "card_declined" {
		t.Errorf("code: got %q, want %q", resp.Error.Code, "card_declined")
	}
	if resp.Error.Message != "Карта отклонена банком" {
		t.Errorf("message: got %q, want localised RU message", resp.Error.Message)
	}
	if resp.Error.Suggestion != "Попробуйте другую карту" {
		t.Errorf("suggestion: got %q", resp.Error.Suggestion)
	}
}
