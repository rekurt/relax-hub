package antifraud

import "github.com/rekurt/relax-hub/internal/domain"

// listingStoplistRule blocks listing creation if the owner is on the antifraud stoplist.
func listingStoplistRule() Rule {
	return Rule{
		Name:     domain.FraudRuleListingStoplist,
		Category: RuleCategoryListingCreate,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.IsOnStoplist {
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleListingStoplist,
					Severity:  domain.FraudSeverityCritical,
					Action:    domain.FraudActionBlock,
					Details: map[string]interface{}{
						"reason": "owner found in antifraud stoplist",
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}

// listingDuplicateRule flags listing creation if the owner shares phone/email/INN
// with another existing owner account (possible duplicate accounts).
func listingDuplicateRule() Rule {
	return Rule{
		Name:     domain.FraudRuleListingDuplicate,
		Category: RuleCategoryListingCreate,
		Evaluate: func(input RuleInput) FraudCheckResult {
			if input.DuplicateOwnerCount > 0 {
				return FraudCheckResult{
					Triggered: true,
					Rule:      domain.FraudRuleListingDuplicate,
					Severity:  domain.FraudSeverityHigh,
					Action:    domain.FraudActionFlag,
					Details: map[string]interface{}{
						"duplicate_owner_count": input.DuplicateOwnerCount,
						"reason":                "owner shares identifiers with other owner accounts",
					},
				}
			}
			return FraudCheckResult{}
		},
	}
}
