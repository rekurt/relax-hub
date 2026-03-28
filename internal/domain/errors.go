package domain

import "errors"

var (
	ErrNotFound                    = errors.New("not found")
	ErrAlreadyExists               = errors.New("already exists")
	ErrInvalidInput                = errors.New("invalid input")
	ErrUnauthorized                = errors.New("unauthorized")
	ErrForbidden                   = errors.New("forbidden")
	ErrSlotUnavailable             = errors.New("time slot is unavailable")
	ErrBookingCancelLate           = errors.New("too late to cancel booking")
	ErrUserBlocked                 = errors.New("user is blocked")
	ErrBathhouseNotActive          = errors.New("bathhouse is not active")
	ErrBathhouseHasBookings        = errors.New("cannot delete bathhouse with active bookings")
	ErrReviewAlreadyResponded      = errors.New("review already has owner response")
	ErrSocialAccountAlreadyLinked  = errors.New("social account already linked")
	ErrSocialAccountNotFound       = errors.New("social account not found")
	ErrOAuthExchangeFailed         = errors.New("oauth code exchange failed")
	ErrSubscriptionNotFound        = errors.New("subscription not found")
	ErrSubscriptionAlreadyActive   = errors.New("subscription already active")
	ErrPromotionBudgetExhausted    = errors.New("promotion budget exhausted")
	ErrInsufficientPoints          = errors.New("insufficient loyalty points")
	ErrPromoNotFound               = errors.New("promo code not found")
	ErrPromoExpired                = errors.New("promo code expired")
	ErrPromoMaxUses                = errors.New("promo code usage limit exceeded")
	ErrPromoMinAmount              = errors.New("promo code minimum amount not met")
	ErrPromoInvalid                = errors.New("promo code is invalid")
	ErrComplaintNotFound           = errors.New("complaint not found")
	ErrAlreadyReported             = errors.New("already reported")
	ErrSelfReferral                = errors.New("cannot refer yourself")
	ErrAlreadyReferred             = errors.New("user already referred")
	ErrInsufficientReferralBalance = errors.New("insufficient referral balance")

	ErrCertificateNotFound            = errors.New("gift certificate not found")
	ErrCertificateExpired             = errors.New("gift certificate expired")
	ErrCertificateInsufficientBalance = errors.New("gift certificate insufficient balance")

	ErrPhotoNotFound = errors.New("photo not found")

	ErrMediaNotFound     = errors.New("media not found")
	ErrMediaFileTooLarge = errors.New("media file too large")
	ErrMediaInvalidType  = errors.New("media invalid file type")
	ErrMediaLimitReached = errors.New("media upload limit reached")

	ErrPaymentNotFound         = errors.New("payment not found")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
	ErrRefundExceedsAmount     = errors.New("refund amount exceeds payment amount")
	ErrPaymentFailed           = errors.New("payment failed")

	ErrWalletNotFound            = errors.New("wallet not found")
	ErrInsufficientWalletBalance = errors.New("insufficient wallet balance")
	ErrWalletLimitExceeded       = errors.New("wallet balance limit exceeded")
	ErrWalletFrozen              = errors.New("wallet is frozen")
	ErrWalletConcurrentUpdate    = errors.New("wallet was modified concurrently, please retry")
	ErrHoldNotFound              = errors.New("hold not found")
	ErrHoldExpired               = errors.New("hold has expired")
	ErrTopUpBelowMinimum         = errors.New("top-up amount below minimum")
	ErrTopUpAboveMaximum         = errors.New("top-up amount above maximum")

	ErrPayoutNotFound             = errors.New("payout not found")
	ErrPayoutBelowMinimum         = errors.New("payout amount below minimum")
	ErrPayoutDailyLimitExceeded   = errors.New("daily payout limit exceeded")
	ErrPayoutMonthlyLimitExceeded = errors.New("monthly payout limit exceeded")
	ErrPayoutAlreadyProcessed     = errors.New("payout already processed")

	ErrOTPRateLimited = errors.New("too many OTP requests, try later")
	ErrOTPInvalid     = errors.New("invalid OTP code")
	ErrOTPExpired     = errors.New("OTP code expired")
	ErrOTPMaxAttempts = errors.New("too many verification attempts")
	ErrPhoneRequired  = errors.New("phone number is required")
	ErrPhoneInvalid   = errors.New("invalid phone number format")

	Err2FARequired       = errors.New("two-factor authentication required")
	Err2FAAlreadyEnabled = errors.New("two-factor authentication already enabled")
	Err2FANotEnabled     = errors.New("two-factor authentication not enabled")
	Err2FAInvalidCode    = errors.New("invalid two-factor authentication code")
	Err2FAPhoneRequired  = errors.New("verified phone required for SMS 2FA")

	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")

	ErrResetTokenInvalid = errors.New("invalid or expired password reset token")
	ErrResetRateLimited  = errors.New("too many password reset requests, try later")

	ErrAccountDeletionPending    = errors.New("account deletion already requested")
	ErrAccountDeletionNotPending = errors.New("no pending account deletion")
	ErrAccountDeleted            = errors.New("account has been deleted")

	ErrKYCNotFound    = errors.New("KYC application not found")
	ErrKYCNotApproved = errors.New("KYC verification not approved")
	ErrKYCPending     = errors.New("KYC application already pending review")

	ErrOfferNotFound        = errors.New("offer acceptance not found")
	ErrOfferNotAccepted     = errors.New("offer not accepted")
	ErrOfferAlreadyAccepted = errors.New("offer version already accepted")

	ErrPaymentDetailsNotFound = errors.New("payment details not found")
	ErrPaymentDetailsNotSet   = errors.New("payment details not set")

	ErrListingDraftNotFound    = errors.New("listing draft not found")
	ErrListingDraftIncomplete  = errors.New("listing draft has incomplete steps")
	ErrListingDraftSubmitted   = errors.New("listing draft already submitted")
	ErrListingDraftInvalidStep = errors.New("invalid listing draft step")

	ErrListingIncomplete = errors.New("listing has incomplete required fields")

	ErrAddOnNotFound     = errors.New("add-on not found")
	ErrAddOnLimitReached = errors.New("add-on limit reached")

	ErrSavedSearchNotFound     = errors.New("saved search not found")
	ErrSavedSearchLimitReached = errors.New("saved search limit reached")

	ErrBookingModificationLimit = errors.New("booking modification limit reached")
	ErrBookingNotModifiable     = errors.New("booking cannot be modified in current status")

	ErrCheckinTooEarly = errors.New("check-in is not yet available")
	ErrCheckinTooLate       = errors.New("check-in window has passed")
	ErrNotCheckedIn         = errors.New("guest has not checked in")
	ErrNoShowDisputeExpired = errors.New("no-show dispute window has expired")

	ErrEscrowNotFound        = errors.New("escrow not found")
	ErrEscrowNotMatured      = errors.New("escrow claim period has not ended")
	ErrEscrowAlreadyReleased = errors.New("escrow already released")
	ErrEscrowDisputed        = errors.New("escrow is disputed and cannot be released")

	ErrBroadcastNotFound  = errors.New("broadcast not found")
	ErrBroadcastRateLimit = errors.New("broadcast rate limit exceeded")
	ErrBroadcastNotDraft  = errors.New("broadcast is not in draft status")

	ErrAutoScenarioNotFound = errors.New("auto scenario not found")

	ErrTemplateNotFound     = errors.New("response template not found")
	ErrTemplateLimitReached = errors.New("response template limit reached")

	ErrTicketNotFound         = errors.New("support ticket not found")
	ErrTicketAlreadyClosed    = errors.New("ticket is already closed")
	ErrTicketAlreadyResolved  = errors.New("ticket is already resolved")
	ErrTicketAlreadyEscalated = errors.New("ticket is already at maximum escalation level")
	ErrCSATAlreadySubmitted   = errors.New("CSAT score already submitted")
	ErrCSATNotResolved        = errors.New("ticket must be resolved to submit CSAT")
	ErrCSATInvalidScore       = errors.New("CSAT score must be between 1 and 5")

	ErrDisputeNotFound              = errors.New("dispute not found")
	ErrDisputeAlreadyExists         = errors.New("dispute already exists for this booking")
	ErrDisputeAlreadyResolved       = errors.New("dispute is already resolved")
	ErrDisputeAlreadyClosed         = errors.New("dispute is already closed")
	ErrDisputeEvidenceWindowExpired = errors.New("evidence submission window has expired")
	ErrDisputeAppealExpired         = errors.New("appeal deadline has passed")
	ErrDisputeNotResolved           = errors.New("dispute must be resolved to appeal")
	ErrDisputeAlreadyAppealed       = errors.New("dispute has already been appealed")

	ErrDepositNotFound        = errors.New("security deposit not found")
	ErrDepositAlreadyReleased = errors.New("security deposit already released")
	ErrDepositAlreadyClaimed  = errors.New("security deposit already claimed")

	ErrClientReviewNotFound = errors.New("client review not found")
	ErrReviewBlindPeriod    = errors.New("review is in blind period")

	ErrRegionSwitchBlocked  = errors.New("region switch blocked: resolve wallet balance, active bookings, open disputes, or unactivated certificates first")
	ErrCrossRegionalBooking = errors.New("cross-regional booking not allowed: client region must match bathhouse region")
	ErrRegionSameAsCurrent  = errors.New("already in the requested region")
	ErrRegionInvalid        = errors.New("invalid region")

	ErrSeasonalTariffOverlap  = errors.New("seasonal tariff date range overlaps with existing tariff")
	ErrSeasonalTariffNotFound = errors.New("seasonal tariff not found")

	ErrShareTokenNotFound = errors.New("share token not found")
	ErrShareTokenExpired  = errors.New("share token has expired")

	ErrSavedCardNotFound     = errors.New("saved card not found")
	ErrSavedCardLimitReached = errors.New("saved card limit reached")

	ErrBankEntryNotFound      = errors.New("bank statement entry not found")
	ErrBankEntryAlreadyMatched = errors.New("bank statement entry already matched")

	ErrAdminRoleNotFound       = errors.New("admin role not found")
	ErrAdmin2FARequired        = errors.New("admin accounts require two-factor authentication")
	ErrAdminPermissionDenied   = errors.New("insufficient admin permissions")
	ErrImportValidationFailed  = errors.New("import validation failed")
	ErrImportFileTooLarge      = errors.New("import file too large")

	ErrFAQNotFound = errors.New("FAQ entry not found")

	ErrWebhookNotFound     = errors.New("webhook not found")
	ErrWebhookLimitReached = errors.New("webhook limit reached")
)

// PaymentFailedError wraps ErrPaymentFailed with user-facing details.
type PaymentFailedError struct {
	Code       string // e.g. "insufficient_funds", "timeout"
	MessageRU  string // user-friendly Russian message
	Suggestion string // e.g. "Попробуйте другую карту"
}

func (e *PaymentFailedError) Error() string { return e.MessageRU }
func (e *PaymentFailedError) Unwrap() error { return ErrPaymentFailed }
