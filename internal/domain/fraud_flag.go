package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type FraudRuleName string

const (
	FraudRuleMultiCardTopUp   FraudRuleName = "RULE_MULTI_CARD_TOPUP"
	FraudRuleTopUpCancelCycle FraudRuleName = "RULE_TOPUP_CANCEL_CYCLE"
	FraudRuleDormantBalance   FraudRuleName = "RULE_DORMANT_BALANCE"
	FraudRuleRapidBookings    FraudRuleName = "RULE_RAPID_BOOKINGS"
	FraudRuleSelfBooking      FraudRuleName = "RULE_SELF_BOOKING"
	FraudRuleStructuring      FraudRuleName = "RULE_STRUCTURING"
	FraudRuleFakeReviews      FraudRuleName = "RULE_FAKE_REVIEWS"
)

func (r FraudRuleName) IsValid() bool {
	switch r {
	case FraudRuleMultiCardTopUp, FraudRuleTopUpCancelCycle, FraudRuleDormantBalance,
		FraudRuleRapidBookings, FraudRuleSelfBooking, FraudRuleStructuring, FraudRuleFakeReviews:
		return true
	}
	return false
}

type FraudSeverity string

const (
	FraudSeverityLow      FraudSeverity = "low"
	FraudSeverityMedium   FraudSeverity = "medium"
	FraudSeverityHigh     FraudSeverity = "high"
	FraudSeverityCritical FraudSeverity = "critical"
)

func (s FraudSeverity) IsValid() bool {
	switch s {
	case FraudSeverityLow, FraudSeverityMedium, FraudSeverityHigh, FraudSeverityCritical:
		return true
	}
	return false
}

type FraudFlagStatus string

const (
	FraudFlagStatusPending      FraudFlagStatus = "pending"
	FraudFlagStatusReviewed     FraudFlagStatus = "reviewed"
	FraudFlagStatusDismissed    FraudFlagStatus = "dismissed"
	FraudFlagStatusActionTaken  FraudFlagStatus = "action_taken"
)

func (s FraudFlagStatus) IsValid() bool {
	switch s {
	case FraudFlagStatusPending, FraudFlagStatusReviewed, FraudFlagStatusDismissed, FraudFlagStatusActionTaken:
		return true
	}
	return false
}

type FraudAction string

const (
	FraudActionFlag         FraudAction = "flag"
	FraudActionFreezeWallet FraudAction = "freeze_wallet"
	FraudActionBlockUser    FraudAction = "block_user"
	FraudActionNotifyAdmin  FraudAction = "notify_admin"
	FraudActionBlock        FraudAction = "block"
)

func (a FraudAction) IsValid() bool {
	switch a {
	case FraudActionFlag, FraudActionFreezeWallet, FraudActionBlockUser, FraudActionNotifyAdmin, FraudActionBlock:
		return true
	}
	return false
}

type FraudFlag struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Rule       FraudRuleName
	Severity   FraudSeverity
	Status     FraudFlagStatus
	Action     FraudAction
	Details    json.RawMessage
	CreatedAt  time.Time
	ReviewedAt *time.Time
	ReviewedBy *uuid.UUID
}

type FraudFlagFilter struct {
	Status *FraudFlagStatus
	UserID *uuid.UUID
	Rule   *FraudRuleName
	Page   int
	PageSize int
}

var (
	ErrFraudDetected = errors.New("operation blocked by anti-fraud check")
)
