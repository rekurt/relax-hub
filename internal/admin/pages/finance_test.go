package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockFinanceProvider is a test implementation of FinanceDataProvider.
type mockFinanceProvider struct {
	data *FinanceData
	err  error
}

func (m *mockFinanceProvider) GetFinanceData(_ context.Context) (*FinanceData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

func sampleFinanceData() *FinanceData {
	return &FinanceData{
		CurrentFloat: FinanceSnapshot{
			ClientWalletsTotal: 5000000,  // 50,000 RUB
			ClientWalletsCount: 120,
			OwnerWalletsTotal:  3000000,  // 30,000 RUB
			OwnerWalletsCount:  25,
			EscrowHeldTotal:    1500000,  // 15,000 RUB
			EscrowCount:        8,
			WalletHoldsTotal:   200000,   // 2,000 RUB
			PlatformTotal:      9500000,  // 95,000 RUB
			Status:             "ok",
		},
		LastSnapshot: &FinanceSnapshot{
			ClientWalletsTotal: 4800000,
			ClientWalletsCount: 118,
			OwnerWalletsTotal:  2900000,
			OwnerWalletsCount:  24,
			EscrowHeldTotal:    1400000,
			EscrowCount:        7,
			WalletHoldsTotal:   180000,
			PlatformTotal:      9100000,
			SnapshotDate:       time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC),
			Status:             "ok",
		},
		LastRecon: &FinanceReconciliation{
			PeriodStart:           time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC),
			PeriodEnd:             time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC),
			InternalPaymentsSum:   2500000,
			InternalPaymentsCount: 15,
			InternalRefundsSum:    300000,
			InternalRefundsCount:  2,
			ProviderPaymentsSum:   2500000,
			ProviderPaymentsCount: 15,
			PaymentDiscrepancy:    0,
			RefundDiscrepancy:     0,
			Status:                "matched",
			CreatedAt:             time.Date(2026, 3, 27, 3, 0, 0, 0, time.UTC),
		},
		RecentSnapshots: []FinanceSnapshot{
			{PlatformTotal: 9500000, SnapshotDate: time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC), Status: "ok"},
			{PlatformTotal: 9100000, SnapshotDate: time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC), Status: "ok"},
		},
		Revenue: RevenueBreakdown{
			ServiceFeesTotal:   850000,  // 8,500 RUB
			ServiceFeesCount:   42,
			SubscriptionsTotal: 450000,  // 4,500 RUB
			SubscriptionsCount: 9,
			PromotionsTotal:    200000,  // 2,000 RUB
			PromotionsCount:    3,
			GrandTotal:         1500000, // 15,000 RUB
			PeriodLabel:        "Текущий месяц",
		},
		Wallet: WalletMetrics{
			TotalBalance:        8000000, // 80,000 RUB
			ActiveWalletsCount:  145,
			WalletPaymentsTotal: 1200000,
			AllPaymentsTotal:    5000000,
			WalletPaymentShare:  24.0,
			ExpiredBonusesTotal: 350000,
			ExpiredBonusesCount: 28,
			PendingBonusesTotal: 120000,
			PendingBonusesCount: 12,
		},
		GeneratedAt: time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC),
	}
}

func TestFinanceHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name       string
		provider   *mockFinanceProvider
		wantStatus int
	}{
		{
			name:       "success with full data",
			provider:   &mockFinanceProvider{data: sampleFinanceData()},
			wantStatus: http.StatusOK,
		},
		{
			name: "success with empty data",
			provider: &mockFinanceProvider{
				data: &FinanceData{
					Revenue: RevenueBreakdown{PeriodLabel: "Текущий месяц"},
					GeneratedAt: time.Now(),
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "provider error",
			provider:   &mockFinanceProvider{err: context.DeadlineExceeded},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewFinanceHandler(tt.provider, testLogger(), "/admin-panel/pages", "/admin-panel")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/finance", nil)

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				ct := rec.Header().Get("Content-Type")
				if ct != "text/html; charset=utf-8" {
					t.Errorf("content-type = %q, want text/html", ct)
				}
				if rec.Body.Len() == 0 {
					t.Error("expected non-empty response body")
				}
			}
		})
	}
}

func TestFinanceHandler_RendersFloatDashboard(t *testing.T) {
	data := sampleFinanceData()
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"Текущий баланс платформы",
		"Кошельки клиентов",
		"50000.00",          // 5000000 kopecks
		"120 кошельков",
		"Кошельки владельцев",
		"30000.00",
		"Эскроу (удержано)",
		"15000.00",
		"Итого на платформе",
		"95000.00",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing float dashboard element %q", c)
		}
	}
}

func TestFinanceHandler_RendersReconciliation(t *testing.T) {
	data := sampleFinanceData()
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"Последняя сверка с провайдером",
		"Совпадает",     // matched status
		"25000.00",      // 2500000 kopecks internal payments
		"badge-ok",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing reconciliation element %q", c)
		}
	}
}

func TestFinanceHandler_RendersRevenueBreakdown(t *testing.T) {
	data := sampleFinanceData()
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"Выручка платформы",
		"Текущий месяц",
		"Сервисные комиссии",
		"8500.00",           // 850000 kopecks
		"42 бронирований",
		"Подписки",
		"4500.00",
		"9 активных",
		"Продвижение",
		"2000.00",
		"3 размещений",
		"Итого выручка",
		"15000.00",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing revenue element %q", c)
		}
	}
}

func TestFinanceHandler_RendersWalletMetrics(t *testing.T) {
	data := sampleFinanceData()
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"Метрики кошельков",
		"Общий баланс кошельков",
		"80000.00",           // 8000000 kopecks
		"145 активных",
		"Доля оплат кошельком",
		"24.0%",
		"Сгоревшие бонусы",
		"3500.00",            // 350000 kopecks
		"28 операций",
		"Сгорают в ближ. 30 дн.",
		"1200.00",            // 120000 kopecks
		"12 бонусов",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing wallet metrics element %q", c)
		}
	}
}

func TestFinanceHandler_RendersBaseLayout(t *testing.T) {
	data := sampleFinanceData()
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	checks := []string{
		"sidebar",
		"<title>Финансовый мониторинг",
		"/admin-panel/pages/finance",
		"Назад в GoAdmin",
		"Последнее обновление:",
	}
	for _, c := range checks {
		if !strings.Contains(body, c) {
			t.Errorf("body missing layout element %q", c)
		}
	}
}

func TestFinanceHandler_NoSnapshotsMessage(t *testing.T) {
	data := &FinanceData{
		Revenue:     RevenueBreakdown{PeriodLabel: "Текущий месяц"},
		GeneratedAt: time.Now(),
	}
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "Снимки ещё не создавались") {
		t.Error("expected empty state message for no snapshots")
	}
	if !strings.Contains(body, "Сверки ещё не проводились") {
		t.Error("expected empty state message for no reconciliation")
	}
}

func TestFinanceHandler_ReconciliationDiscrepancy(t *testing.T) {
	data := sampleFinanceData()
	data.LastRecon.Status = "mismatch"
	data.LastRecon.PaymentDiscrepancy = 50000 // 500 RUB
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "badge-warn") {
		t.Error("expected badge-warn for discrepancy")
	}
	if !strings.Contains(body, "500.00") {
		t.Error("expected discrepancy amount 500.00")
	}
}

func TestFinanceHandler_PendingBonusesWarning(t *testing.T) {
	data := sampleFinanceData()
	data.Wallet.PendingBonusesCount = 15
	data.Wallet.PendingBonusesTotal = 250000
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// The pending bonuses card should have "warn" class when count > 0
	if !strings.Contains(body, `finance-card warn`) {
		t.Error("expected warn class on pending bonuses card when count > 0")
	}
}

func TestFinanceHandler_NoPendingBonusesOk(t *testing.T) {
	data := sampleFinanceData()
	data.Wallet.PendingBonusesCount = 0
	data.Wallet.PendingBonusesTotal = 0
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// When no pending bonuses, the card should show "ok"
	if strings.Count(body, `finance-card ok`) < 2 {
		t.Error("expected ok class on pending bonuses card when count is 0")
	}
}

func TestFinanceHandler_RendersSnapshotHistory(t *testing.T) {
	data := sampleFinanceData()
	provider := &mockFinanceProvider{data: data}
	handler := NewFinanceHandler(provider, testLogger(), "/admin-panel/pages", "/admin-panel")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/finance", nil)
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	if !strings.Contains(body, "История снимков (7 дней)") {
		t.Error("expected snapshot history section")
	}
	if !strings.Contains(body, "27.03.2026") {
		t.Error("expected snapshot date in history")
	}
}

// --- Template function unit tests ---

func TestKopecksToRub(t *testing.T) {
	tests := []struct {
		kopecks int64
		want    string
	}{
		{0, "0.00"},
		{100, "1.00"},
		{5050, "50.50"},
		{1234567, "12345.67"},
		{-500, "-5.00"},
	}
	for _, tt := range tests {
		got := KopecksToRub(tt.kopecks)
		if got != tt.want {
			t.Errorf("KopecksToRub(%d) = %q, want %q", tt.kopecks, got, tt.want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		t    time.Time
		want string
	}{
		{time.Time{}, "—"},
		{time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC), "28.03.2026"},
		{time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), "05.01.2026"},
	}
	for _, tt := range tests {
		got := FormatDate(tt.t)
		if got != tt.want {
			t.Errorf("FormatDate(%v) = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func TestReconStatus(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"ok", "OK"},
		{"matched", "Совпадает"},
		{"mismatch", "Расхождение"},
		{"discrepancy", "Расхождение"},
		{"error", "Ошибка"},
		{"pending", "Ожидает"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		got := ReconStatus(tt.status)
		if got != tt.want {
			t.Errorf("ReconStatus(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestPctOf(t *testing.T) {
	tests := []struct {
		a, b int64
		want string
	}{
		{0, 100, "0.0"},
		{50, 100, "50.0"},
		{1, 3, "33.3"},
		{100, 0, "0.0"},
	}
	for _, tt := range tests {
		got := PctOf(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("PctOf(%d, %d) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}
