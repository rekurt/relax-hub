package domain

import "time"

// CancellationPolicy defines the cancellation/refund rules for a bathhouse.
type CancellationPolicy string

const (
	CancellationPolicyFlexible CancellationPolicy = "flexible"
	CancellationPolicyModerate CancellationPolicy = "moderate"
	CancellationPolicyStrict   CancellationPolicy = "strict"
)

func (p CancellationPolicy) IsValid() bool {
	switch p {
	case CancellationPolicyFlexible, CancellationPolicyModerate, CancellationPolicyStrict:
		return true
	}
	return false
}

// CancellationTier represents a single tier in a cancellation policy.
type CancellationTier struct {
	MinHoursBefore int // minimum hours before start for this tier to apply
	RefundPercent  int // refund percentage (0-100)
}

// GetTiers returns the ordered tiers for the policy (checked from most generous to least).
func (p CancellationPolicy) GetTiers() []CancellationTier {
	switch p {
	case CancellationPolicyFlexible:
		return []CancellationTier{
			{MinHoursBefore: 24, RefundPercent: 100},
			{MinHoursBefore: 0, RefundPercent: 50},
		}
	case CancellationPolicyModerate:
		return []CancellationTier{
			{MinHoursBefore: 72, RefundPercent: 100},
			{MinHoursBefore: 24, RefundPercent: 50},
			{MinHoursBefore: 0, RefundPercent: 0},
		}
	case CancellationPolicyStrict:
		return []CancellationTier{
			{MinHoursBefore: 168, RefundPercent: 100}, // 7 days
			{MinHoursBefore: 72, RefundPercent: 50},   // 3 days
			{MinHoursBefore: 0, RefundPercent: 0},
		}
	default:
		// Fallback to flexible
		return CancellationPolicyFlexible.GetTiers()
	}
}

// CalculateRefundPercent returns the refund percentage based on time until booking start.
func (p CancellationPolicy) CalculateRefundPercent(timeUntilStart time.Duration) int {
	hoursUntilStart := timeUntilStart.Hours()
	for _, tier := range p.GetTiers() {
		if hoursUntilStart >= float64(tier.MinHoursBefore) {
			return tier.RefundPercent
		}
	}
	return 0
}

// CalculateRefundAmount returns the refund amount in kopecks based on the policy and time.
func (p CancellationPolicy) CalculateRefundAmount(totalAmount int64, timeUntilStart time.Duration) int64 {
	percent := p.CalculateRefundPercent(timeUntilStart)
	if percent == 0 {
		return 0
	}
	if percent == 100 {
		return totalAmount
	}
	return totalAmount * int64(percent) / 100
}
