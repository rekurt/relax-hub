package domain

import (
	"testing"
	"time"
)

func TestCancellationPolicy_IsValid(t *testing.T) {
	tests := []struct {
		policy CancellationPolicy
		valid  bool
	}{
		{CancellationPolicyFlexible, true},
		{CancellationPolicyModerate, true},
		{CancellationPolicyStrict, true},
		{"", false},
		{"custom", false},
	}
	for _, tt := range tests {
		if got := tt.policy.IsValid(); got != tt.valid {
			t.Errorf("CancellationPolicy(%q).IsValid() = %v, want %v", tt.policy, got, tt.valid)
		}
	}
}

func TestCancellationPolicy_CalculateRefundPercent(t *testing.T) {
	tests := []struct {
		name           string
		policy         CancellationPolicy
		hoursUntil     float64
		expectedPct    int
	}{
		// Flexible: 100% at 24h+, 50% under 24h
		{"flexible_48h", CancellationPolicyFlexible, 48, 100},
		{"flexible_exactly_24h", CancellationPolicyFlexible, 24, 100},
		{"flexible_23h59m", CancellationPolicyFlexible, 23.98, 50},
		{"flexible_12h", CancellationPolicyFlexible, 12, 50},
		{"flexible_1h", CancellationPolicyFlexible, 1, 50},
		{"flexible_30min", CancellationPolicyFlexible, 0.5, 50},
		{"flexible_0h", CancellationPolicyFlexible, 0, 50},

		// Moderate: 100% at 72h+, 50% at 24-72h, 0% under 24h
		{"moderate_168h", CancellationPolicyModerate, 168, 100},
		{"moderate_exactly_72h", CancellationPolicyModerate, 72, 100},
		{"moderate_71h", CancellationPolicyModerate, 71, 50},
		{"moderate_48h", CancellationPolicyModerate, 48, 50},
		{"moderate_exactly_24h", CancellationPolicyModerate, 24, 50},
		{"moderate_23h", CancellationPolicyModerate, 23, 0},
		{"moderate_1h", CancellationPolicyModerate, 1, 0},
		{"moderate_0h", CancellationPolicyModerate, 0, 0},

		// Strict: 100% at 168h+ (7d), 50% at 72-168h (3-7d), 0% under 72h (3d)
		{"strict_240h", CancellationPolicyStrict, 240, 100},
		{"strict_exactly_168h", CancellationPolicyStrict, 168, 100},
		{"strict_167h", CancellationPolicyStrict, 167, 50},
		{"strict_120h", CancellationPolicyStrict, 120, 50},
		{"strict_exactly_72h", CancellationPolicyStrict, 72, 50},
		{"strict_71h", CancellationPolicyStrict, 71, 0},
		{"strict_24h", CancellationPolicyStrict, 24, 0},
		{"strict_1h", CancellationPolicyStrict, 1, 0},
		{"strict_0h", CancellationPolicyStrict, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur := time.Duration(tt.hoursUntil * float64(time.Hour))
			got := tt.policy.CalculateRefundPercent(dur)
			if got != tt.expectedPct {
				t.Errorf("CalculateRefundPercent(%v) = %d, want %d", dur, got, tt.expectedPct)
			}
		})
	}
}

func TestCancellationPolicy_CalculateRefundAmount(t *testing.T) {
	tests := []struct {
		name       string
		policy     CancellationPolicy
		amount     int64
		hoursUntil float64
		expected   int64
	}{
		{"flexible_full_10000", CancellationPolicyFlexible, 10000, 48, 10000},
		{"flexible_half_10000", CancellationPolicyFlexible, 10000, 12, 5000},
		{"moderate_full_15000", CancellationPolicyModerate, 15000, 72, 15000},
		{"moderate_half_15000", CancellationPolicyModerate, 15000, 48, 7500},
		{"moderate_zero_15000", CancellationPolicyModerate, 15000, 1, 0},
		{"strict_full_20000", CancellationPolicyStrict, 20000, 168, 20000},
		{"strict_half_20000", CancellationPolicyStrict, 20000, 96, 10000},
		{"strict_zero_20000", CancellationPolicyStrict, 20000, 24, 0},
		// Odd amounts - verify integer division
		{"flexible_half_10001", CancellationPolicyFlexible, 10001, 12, 5000},
		{"moderate_half_10001", CancellationPolicyModerate, 10001, 48, 5000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur := time.Duration(tt.hoursUntil * float64(time.Hour))
			got := tt.policy.CalculateRefundAmount(tt.amount, dur)
			if got != tt.expected {
				t.Errorf("CalculateRefundAmount(%d, %v) = %d, want %d", tt.amount, dur, got, tt.expected)
			}
		})
	}
}

func TestCancellationPolicy_GetTiers(t *testing.T) {
	// Verify the number of tiers per policy
	if len(CancellationPolicyFlexible.GetTiers()) != 2 {
		t.Error("flexible should have 2 tiers")
	}
	if len(CancellationPolicyModerate.GetTiers()) != 3 {
		t.Error("moderate should have 3 tiers")
	}
	if len(CancellationPolicyStrict.GetTiers()) != 3 {
		t.Error("strict should have 3 tiers")
	}

	// Unknown policy falls back to flexible
	unknown := CancellationPolicy("unknown")
	if len(unknown.GetTiers()) != 2 {
		t.Error("unknown policy should fall back to flexible (2 tiers)")
	}
}
