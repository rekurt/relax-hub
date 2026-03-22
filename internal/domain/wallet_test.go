package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWalletCurrency_IsValid(t *testing.T) {
	tests := []struct {
		currency WalletCurrency
		want     bool
	}{
		{WalletCurrencyRUB, true},
		{WalletCurrencyBYN, true},
		{"USD", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := tt.currency.IsValid(); got != tt.want {
			t.Errorf("WalletCurrency(%q).IsValid() = %v, want %v", tt.currency, got, tt.want)
		}
	}
}

func TestWalletStatus_IsValid(t *testing.T) {
	tests := []struct {
		status WalletStatus
		want   bool
	}{
		{WalletStatusActive, true},
		{WalletStatusFrozen, true},
		{"closed", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := tt.status.IsValid(); got != tt.want {
			t.Errorf("WalletStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestWalletTransactionType_IsValid(t *testing.T) {
	validTypes := []WalletTransactionType{
		WalletTxTopUp, WalletTxSpend, WalletTxRefund, WalletTxBonus,
		WalletTxBonusExpiry, WalletTxHoldCapture, WalletTxHoldRelease,
		WalletTxPayout, WalletTxWelcomeBonus, WalletTxReferralBonus,
	}
	for _, tt := range validTypes {
		if !tt.IsValid() {
			t.Errorf("WalletTransactionType(%q).IsValid() = false, want true", tt)
		}
	}
	if WalletTransactionType("unknown").IsValid() {
		t.Error("WalletTransactionType(unknown).IsValid() = true, want false")
	}
}

func TestWalletTransactionStatus_IsValid(t *testing.T) {
	validStatuses := []WalletTransactionStatus{
		WalletTxStatusPending, WalletTxStatusCompleted, WalletTxStatusFailed, WalletTxStatusCancelled,
	}
	for _, s := range validStatuses {
		if !s.IsValid() {
			t.Errorf("WalletTransactionStatus(%q).IsValid() = false, want true", s)
		}
	}
	if WalletTransactionStatus("unknown").IsValid() {
		t.Error("WalletTransactionStatus(unknown).IsValid() = true, want false")
	}
}

func TestWalletHoldStatus_IsValid(t *testing.T) {
	validStatuses := []WalletHoldStatus{
		WalletHoldStatusActive, WalletHoldStatusCaptured, WalletHoldStatusReleased, WalletHoldStatusExpired,
	}
	for _, s := range validStatuses {
		if !s.IsValid() {
			t.Errorf("WalletHoldStatus(%q).IsValid() = false, want true", s)
		}
	}
	if WalletHoldStatus("unknown").IsValid() {
		t.Error("WalletHoldStatus(unknown).IsValid() = true, want false")
	}
}

func TestWallet_Validate(t *testing.T) {
	validWallet := Wallet{
		UserID:   uuid.New(),
		Currency: WalletCurrencyRUB,
		Balance:  1000,
		Status:   WalletStatusActive,
	}

	tests := []struct {
		name    string
		modify  func(w *Wallet)
		wantErr bool
	}{
		{"valid", func(w *Wallet) {}, false},
		{"empty status ok", func(w *Wallet) { w.Status = "" }, false},
		{"nil user id", func(w *Wallet) { w.UserID = uuid.Nil }, true},
		{"invalid currency", func(w *Wallet) { w.Currency = "USD" }, true},
		{"negative balance", func(w *Wallet) { w.Balance = -1 }, true},
		{"negative held", func(w *Wallet) { w.HeldAmount = -1 }, true},
		{"invalid status", func(w *Wallet) { w.Status = "bad" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := validWallet
			tt.modify(&w)
			err := w.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Wallet.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWallet_AvailableBalance(t *testing.T) {
	w := Wallet{Balance: 10000, HeldAmount: 3000}
	if got := w.AvailableBalance(); got != 7000 {
		t.Errorf("AvailableBalance() = %d, want 7000", got)
	}
}

func TestWallet_IsFrozen(t *testing.T) {
	w := Wallet{Status: WalletStatusActive}
	if w.IsFrozen() {
		t.Error("IsFrozen() = true for active wallet")
	}
	w.Status = WalletStatusFrozen
	if !w.IsFrozen() {
		t.Error("IsFrozen() = false for frozen wallet")
	}
}

func TestWalletTransaction_Validate(t *testing.T) {
	valid := WalletTransaction{
		WalletID: uuid.New(),
		Type:     WalletTxTopUp,
		Amount:   5000,
		Status:   WalletTxStatusCompleted,
	}

	tests := []struct {
		name    string
		modify  func(tx *WalletTransaction)
		wantErr bool
	}{
		{"valid", func(tx *WalletTransaction) {}, false},
		{"nil wallet id", func(tx *WalletTransaction) { tx.WalletID = uuid.Nil }, true},
		{"invalid type", func(tx *WalletTransaction) { tx.Type = "bad" }, true},
		{"zero amount", func(tx *WalletTransaction) { tx.Amount = 0 }, true},
		{"negative amount", func(tx *WalletTransaction) { tx.Amount = -1 }, true},
		{"invalid status", func(tx *WalletTransaction) { tx.Status = "bad" }, true},
		{"empty status ok", func(tx *WalletTransaction) { tx.Status = "" }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := valid
			tt.modify(&tx)
			err := tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WalletTransaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWalletHold_Validate(t *testing.T) {
	valid := WalletHold{
		WalletID:  uuid.New(),
		Amount:    5000,
		ExpiresAt: time.Now().Add(time.Hour),
		Status:    WalletHoldStatusActive,
	}

	tests := []struct {
		name    string
		modify  func(h *WalletHold)
		wantErr bool
	}{
		{"valid", func(h *WalletHold) {}, false},
		{"nil wallet id", func(h *WalletHold) { h.WalletID = uuid.Nil }, true},
		{"zero amount", func(h *WalletHold) { h.Amount = 0 }, true},
		{"zero expiry", func(h *WalletHold) { h.ExpiresAt = time.Time{} }, true},
		{"invalid status", func(h *WalletHold) { h.Status = "bad" }, true},
		{"empty status ok", func(h *WalletHold) { h.Status = "" }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := valid
			tt.modify(&h)
			err := h.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WalletHold.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWalletHold_IsExpired(t *testing.T) {
	h := WalletHold{ExpiresAt: time.Now().Add(-time.Hour)}
	if !h.IsExpired() {
		t.Error("IsExpired() = false for past expiry")
	}
	h.ExpiresAt = time.Now().Add(time.Hour)
	if h.IsExpired() {
		t.Error("IsExpired() = true for future expiry")
	}
}

func TestMaxBalanceForCurrency(t *testing.T) {
	if got := MaxBalanceForCurrency(WalletCurrencyRUB); got != WalletMaxBalanceRUB {
		t.Errorf("MaxBalanceForCurrency(RUB) = %d, want %d", got, WalletMaxBalanceRUB)
	}
	if got := MaxBalanceForCurrency(WalletCurrencyBYN); got != WalletMaxBalanceBYN {
		t.Errorf("MaxBalanceForCurrency(BYN) = %d, want %d", got, WalletMaxBalanceBYN)
	}
}
