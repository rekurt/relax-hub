package antifraud

import "github.com/nikitaaldaev/bani/internal/domain"

// DefaultRules returns the default set of anti-fraud rules.
func DefaultRules() []Rule {
	return []Rule{
		multiCardTopUpRule(),
		topUpCancelCycleRule(),
		dormantBalanceRule(),
		rapidBookingsRule(),
		selfBookingRule(),
		structuringRule(),
	}
}

// RULE_MULTI_CARD_TOPUP: >3 different cards used for top-ups in 24h
func multiCardTopUpRule() Rule {
	return Rule{
		Name:     domain.FraudRuleMultiCardTopUp,
		Category: RuleCategoryWalletTopUp,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.CardCount > 3 {
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleMultiCardTopUp,
					Severity:  domain.FraudSeverityHigh,
					Action:    domain.FraudActionFreezeWallet,
					Details: map[string]interface{}{
						"card_count_24h": input.CardCount,
						"threshold":      3,
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}

// RULE_TOPUP_CANCEL_CYCLE: >2 top-up then cancel booking cycles in 7 days
func topUpCancelCycleRule() Rule {
	return Rule{
		Name:     domain.FraudRuleTopUpCancelCycle,
		Category: RuleCategoryBookingCancel,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.TopUpCancelCycles > 2 {
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleTopUpCancelCycle,
					Severity:  domain.FraudSeverityHigh,
					Action:    domain.FraudActionBlock,
					Details: map[string]interface{}{
						"cycles_7d": input.TopUpCancelCycles,
						"threshold": 2,
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}

// RULE_DORMANT_BALANCE: >50k balance with no bookings in 30 days
func dormantBalanceRule() Rule {
	return Rule{
		Name:     domain.FraudRuleDormantBalance,
		Category: RuleCategoryDormantBalance,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.Balance > 5000000 && input.LastBookingDaysAgo > 30 { // 50k in kopecks
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleDormantBalance,
					Severity:  domain.FraudSeverityMedium,
					Action:    domain.FraudActionNotifyAdmin,
					Details: map[string]interface{}{
						"balance_kopecks":   input.Balance,
						"last_booking_days": input.LastBookingDaysAgo,
						"balance_threshold": 5000000,
						"inactivity_days":   30,
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}

// RULE_RAPID_BOOKINGS: >5 bookings in 1 hour
func rapidBookingsRule() Rule {
	return Rule{
		Name:     domain.FraudRuleRapidBookings,
		Category: RuleCategoryBookingCreate,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.RecentBookingCount > 5 {
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleRapidBookings,
					Severity:  domain.FraudSeverityMedium,
					Action:    domain.FraudActionBlock,
					Details: map[string]interface{}{
						"bookings_1h": input.RecentBookingCount,
						"threshold":   5,
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}

// RULE_SELF_BOOKING: owner books their own bathhouse
func selfBookingRule() Rule {
	return Rule{
		Name:     domain.FraudRuleSelfBooking,
		Category: RuleCategoryBookingCreate,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.UserID == input.BathhouseOwnerID && input.BathhouseOwnerID != [16]byte{} {
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleSelfBooking,
					Severity:  domain.FraudSeverityHigh,
					Action:    domain.FraudActionBlock,
					Details: map[string]interface{}{
						"user_id":         input.UserID.String(),
						"bathhouse_owner": input.BathhouseOwnerID.String(),
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}

// RULE_STRUCTURING: >3 small withdrawals per day (possible structuring to avoid limits)
func structuringRule() Rule {
	return Rule{
		Name:     domain.FraudRuleStructuring,
		Category: RuleCategoryPayoutRequest,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.DailyWithdrawals > 3 {
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleStructuring,
					Severity:  domain.FraudSeverityHigh,
					Action:    domain.FraudActionFreezeWallet,
					Details: map[string]interface{}{
						"daily_withdrawals": input.DailyWithdrawals,
						"threshold":         3,
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}
