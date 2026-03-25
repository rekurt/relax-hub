package antifraud

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

// FraudEngine checks operations against anti-fraud rules.
type FraudEngine interface {
	CheckWalletTopUp(ctx context.Context, userID uuid.UUID, amount int64, cardCount int) error
	CheckBookingCreate(ctx context.Context, userID uuid.UUID, bathhouseOwnerID uuid.UUID, recentBookingCount int) error
	CheckBookingCancel(ctx context.Context, userID uuid.UUID, topUpCancelCycles int) error
	CheckPayoutRequest(ctx context.Context, userID uuid.UUID, amount int64, dailyWithdrawals int) error
	CheckDormantBalance(ctx context.Context, userID uuid.UUID, balance int64, lastBookingDaysAgo int) error
}

// FraudCheckResult captures the outcome of a single rule evaluation.
type FraudCheckResult struct {
	Triggered bool
	Rule      domain.FraudRuleName
	Severity  domain.FraudSeverity
	Action    domain.FraudAction
	Details   map[string]interface{}
}

type fraudEngine struct {
	flagRepo repository.FraudFlagRepository
	rules    []Rule
	logger   *logger.Logger
}

func NewFraudEngine(flagRepo repository.FraudFlagRepository, log *logger.Logger) FraudEngine {
	return &fraudEngine{
		flagRepo: flagRepo,
		rules:    DefaultRules(),
		logger:   log,
	}
}

func (e *fraudEngine) CheckWalletTopUp(ctx context.Context, userID uuid.UUID, amount int64, cardCount int) error {
	input := RuleInput{
		UserID:    userID,
		Amount:    amount,
		CardCount: cardCount,
	}
	return e.evaluate(ctx, input, RuleCategoryWalletTopUp)
}

func (e *fraudEngine) CheckBookingCreate(ctx context.Context, userID uuid.UUID, bathhouseOwnerID uuid.UUID, recentBookingCount int) error {
	input := RuleInput{
		UserID:             userID,
		BathhouseOwnerID:   bathhouseOwnerID,
		RecentBookingCount: recentBookingCount,
	}
	return e.evaluate(ctx, input, RuleCategoryBookingCreate)
}

func (e *fraudEngine) CheckBookingCancel(ctx context.Context, userID uuid.UUID, topUpCancelCycles int) error {
	input := RuleInput{
		UserID:           userID,
		TopUpCancelCycles: topUpCancelCycles,
	}
	return e.evaluate(ctx, input, RuleCategoryBookingCancel)
}

func (e *fraudEngine) CheckPayoutRequest(ctx context.Context, userID uuid.UUID, amount int64, dailyWithdrawals int) error {
	input := RuleInput{
		UserID:           userID,
		Amount:           amount,
		DailyWithdrawals: dailyWithdrawals,
	}
	return e.evaluate(ctx, input, RuleCategoryPayoutRequest)
}

func (e *fraudEngine) CheckDormantBalance(ctx context.Context, userID uuid.UUID, balance int64, lastBookingDaysAgo int) error {
	input := RuleInput{
		UserID:             userID,
		Balance:            balance,
		LastBookingDaysAgo: lastBookingDaysAgo,
	}
	return e.evaluate(ctx, input, RuleCategoryDormantBalance)
}

func (e *fraudEngine) evaluate(ctx context.Context, input RuleInput, category RuleCategory) error {
	for _, rule := range e.rules {
		if rule.Category != category {
			continue
		}

		result := rule.Evaluate(input)
		if !result.Triggered {
			continue
		}

		detailsJSON, _ := json.Marshal(result.Details)

		flag := &domain.FraudFlag{
			UserID:   input.UserID,
			Rule:     result.Rule,
			Severity: result.Severity,
			Action:   result.Action,
			Details:  detailsJSON,
		}

		if err := e.flagRepo.Create(ctx, flag); err != nil {
			e.logger.Error("failed to create fraud flag", "error", err, "rule", result.Rule, "user_id", input.UserID)
		}

		e.logger.Warn("fraud rule triggered",
			"rule", result.Rule,
			"severity", result.Severity,
			"action", result.Action,
			"user_id", input.UserID,
		)

		// "block" action prevents the operation
		if result.Action == domain.FraudActionBlock {
			return domain.ErrFraudDetected
		}
	}

	return nil
}

// RuleCategory groups rules by the operation they check.
type RuleCategory string

const (
	RuleCategoryWalletTopUp    RuleCategory = "wallet_topup"
	RuleCategoryBookingCreate  RuleCategory = "booking_create"
	RuleCategoryBookingCancel  RuleCategory = "booking_cancel"
	RuleCategoryPayoutRequest  RuleCategory = "payout_request"
	RuleCategoryDormantBalance RuleCategory = "dormant_balance"
)

// RuleInput contains all possible inputs for fraud rule evaluation.
type RuleInput struct {
	UserID             uuid.UUID
	BathhouseOwnerID   uuid.UUID
	Amount             int64
	Balance            int64
	CardCount          int
	RecentBookingCount int
	TopUpCancelCycles  int
	DailyWithdrawals   int
	LastBookingDaysAgo int
	Timestamp          time.Time
}

// Rule defines a single anti-fraud rule.
type Rule struct {
	Name     domain.FraudRuleName
	Category RuleCategory
	Evaluate func(input RuleInput) FraudCheckResult
}
