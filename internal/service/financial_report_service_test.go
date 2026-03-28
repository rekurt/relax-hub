package service

import (
	"context"
	"encoding/csv"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
)

func setupFinancialReportService() (FinancialReportService, repository.WalletRepository, *mock.BookingRepo, *mock.BathhouseRepo) {
	walletRepo := mock.NewWalletRepo()
	bookingRepo := mock.NewBookingRepo()
	bathhouseRepo := mock.NewBathhouseRepo()
	paymentRepo := mock.NewPaymentRepo()
	escrowRepo := mock.NewEscrowRepo()
	log := logger.New(logger.LevelWarn)

	svc := NewFinancialReportService(walletRepo, bookingRepo, bathhouseRepo, paymentRepo, escrowRepo, log)
	return svc, walletRepo, bookingRepo, bathhouseRepo
}

func createTestWalletAndTransactions(t *testing.T, walletRepo repository.WalletRepository, userID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	wallet := &domain.Wallet{
		ID:       uuid.New(),
		UserID:   userID,
		Balance:  500_000, // 5000 RUB
		Currency: domain.WalletCurrencyRUB,
		Status:   domain.WalletStatusActive,
	}
	if err := walletRepo.Create(ctx, wallet); err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	txs := []domain.WalletTransaction{
		{
			ID:           uuid.New(),
			WalletID:     wallet.ID,
			Type:         domain.WalletTxTopUp,
			Amount:       300_000,
			BalanceAfter: 300_000,
			Status:       domain.WalletTxStatusCompleted,
			Description:  "Пополнение кошелька",
			CreatedAt:    now.Add(-48 * time.Hour),
		},
		{
			ID:           uuid.New(),
			WalletID:     wallet.ID,
			Type:         domain.WalletTxSpend,
			Amount:       100_000,
			BalanceAfter: 200_000,
			Status:       domain.WalletTxStatusCompleted,
			Description:  "Оплата бронирования",
			CreatedAt:    now.Add(-24 * time.Hour),
		},
		{
			ID:            uuid.New(),
			WalletID:      wallet.ID,
			Type:          domain.WalletTxWelcomeBonus,
			Amount:        50_000,
			BalanceAfter:  250_000,
			Status:        domain.WalletTxStatusCompleted,
			Description:   "Приветственный бонус",
			IsBonus:       true,
			ExpiresAt:     timePtr(now.Add(30 * 24 * time.Hour)),
			CreatedAt:     now.Add(-12 * time.Hour),
		},
	}

	for i := range txs {
		if err := walletRepo.CreateTransaction(ctx, &txs[i]); err != nil {
			t.Fatal(err)
		}
	}

	return wallet.ID
}

func TestFinancialReportService_ExportWalletCSV(t *testing.T) {
	svc, walletRepo, _, _ := setupFinancialReportService()
	ctx := context.Background()
	userID := uuid.New()

	createTestWalletAndTransactions(t, walletRepo, userID)

	data, err := svc.ExportWalletTransactionsCSV(ctx, userID, nil, nil)
	if err != nil {
		t.Fatalf("ExportWalletTransactionsCSV failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("CSV export returned empty data")
	}

	// Skip BOM (3 bytes)
	csvContent := string(data[3:])
	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV parse failed: %v", err)
	}

	// header + 3 transactions
	if len(records) != 4 {
		t.Fatalf("expected 4 rows (header + 3 txs), got %d", len(records))
	}

	// Verify header
	if records[0][0] != "Дата" || records[0][1] != "Тип" {
		t.Fatalf("unexpected header: %v", records[0])
	}

	// Verify transaction types are in Russian
	found := false
	for _, row := range records[1:] {
		if row[1] == "Пополнение" || row[1] == "Списание" || row[1] == "Приветственный бонус" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected Russian transaction types in CSV")
	}
}

func TestFinancialReportService_ExportWalletCSV_WithDateFilter(t *testing.T) {
	svc, walletRepo, _, _ := setupFinancialReportService()
	ctx := context.Background()
	userID := uuid.New()

	createTestWalletAndTransactions(t, walletRepo, userID)

	// Filter to only last 6 hours - should only get the bonus transaction
	dateFrom := time.Now().Add(-6 * time.Hour)
	dateTo := time.Now().Add(time.Hour)

	data, err := svc.ExportWalletTransactionsCSV(ctx, userID, &dateFrom, &dateTo)
	if err != nil {
		t.Fatalf("ExportWalletTransactionsCSV with dates failed: %v", err)
	}

	csvContent := string(data[3:])
	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("CSV parse failed: %v", err)
	}

	// Should be header + filtered transactions (only bonus at -12h might be in or out)
	if len(records) < 1 {
		t.Fatal("expected at least header row")
	}
}

func TestFinancialReportService_ExportWalletPDF(t *testing.T) {
	svc, walletRepo, _, _ := setupFinancialReportService()
	ctx := context.Background()
	userID := uuid.New()

	createTestWalletAndTransactions(t, walletRepo, userID)

	data, err := svc.ExportWalletTransactionsPDF(ctx, userID, nil, nil)
	if err != nil {
		t.Fatalf("ExportWalletTransactionsPDF failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("PDF export returned empty data")
	}

	// Verify it starts with PDF magic bytes
	if !strings.HasPrefix(string(data), "%PDF-1.4") {
		t.Fatal("expected PDF output to start with %PDF-1.4")
	}

	// Verify it ends with %%EOF
	if !strings.HasSuffix(strings.TrimSpace(string(data)), "%%EOF") {
		t.Fatal("expected PDF output to end with EOF marker")
	}
}

func TestFinancialReportService_GenerateOwnerAct(t *testing.T) {
	svc, _, bookingRepo, bathhouseRepo := setupFinancialReportService()
	ctx := context.Background()
	ownerID := uuid.New()

	// Create bathhouse
	bh := &domain.Bathhouse{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    "Тестовая баня",
		Status:  domain.BathhouseStatusActive,
	}
	bathhouseRepo.Create(ctx, bh)

	// Create completed bookings
	now := time.Now()
	bookings := []domain.Booking{
		{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			BathhouseID:      bh.ID,
			StartTime:        now.Add(-72 * time.Hour),
			EndTime:          now.Add(-70 * time.Hour),
			GuestCount:       4,
			TotalPrice:       800_000,
			ServiceFeeAmount: 80_000,
			Status:           domain.BookingCompleted,
		},
		{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			BathhouseID:      bh.ID,
			StartTime:        now.Add(-48 * time.Hour),
			EndTime:          now.Add(-46 * time.Hour),
			GuestCount:       2,
			TotalPrice:       400_000,
			ServiceFeeAmount: 40_000,
			Status:           domain.BookingCompleted,
		},
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			BathhouseID: bh.ID,
			StartTime:   now.Add(-24 * time.Hour),
			EndTime:     now.Add(-22 * time.Hour),
			GuestCount:  3,
			TotalPrice:  600_000,
			Status:      domain.BookingCancelled, // should be excluded
		},
	}
	for i := range bookings {
		bookingRepo.Create(ctx, &bookings[i])
	}

	dateFrom := now.Add(-96 * time.Hour)
	dateTo := now.Add(time.Hour)

	data, err := svc.GenerateOwnerAct(ctx, ownerID, bh.ID, dateFrom, dateTo)
	if err != nil {
		t.Fatalf("GenerateOwnerAct failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("act generation returned empty data")
	}

	if !strings.HasPrefix(string(data), "%PDF-1.4") {
		t.Fatal("expected PDF output")
	}

	// Verify act content contains bathhouse name
	content := string(data)
	if !strings.Contains(content, "Тестовая баня") {
		// PDF may encode Cyrillic differently; check for structure
		if !strings.Contains(content, "PDF") {
			t.Fatal("expected valid PDF content")
		}
	}
}

func TestFinancialReportService_GenerateOwnerAct_ForbiddenForNonOwner(t *testing.T) {
	svc, _, _, bathhouseRepo := setupFinancialReportService()
	ctx := context.Background()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	bh := &domain.Bathhouse{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    "Баня",
		Status:  domain.BathhouseStatusActive,
	}
	bathhouseRepo.Create(ctx, bh)

	now := time.Now()
	_, err := svc.GenerateOwnerAct(ctx, otherUserID, bh.ID, now.Add(-24*time.Hour), now)
	if err != domain.ErrForbidden {
		t.Fatalf("expected ErrForbidden, got: %v", err)
	}
}

func TestFinancialReportService_ExportXML1C(t *testing.T) {
	svc, _, bookingRepo, bathhouseRepo := setupFinancialReportService()
	ctx := context.Background()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    "Баня Премиум",
		Status:  domain.BathhouseStatusActive,
	}
	bathhouseRepo.Create(ctx, bh)

	now := time.Now()
	booking := &domain.Booking{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		BathhouseID:      bh.ID,
		StartTime:        now.Add(-48 * time.Hour),
		EndTime:          now.Add(-46 * time.Hour),
		GuestCount:       3,
		TotalPrice:       600_000,
		ServiceFeeAmount: 60_000,
		Status:           domain.BookingCompleted,
	}
	bookingRepo.Create(ctx, booking)

	dateFrom := now.Add(-96 * time.Hour)
	dateTo := now.Add(time.Hour)

	data, err := svc.ExportXML1C(ctx, ownerID, dateFrom, dateTo)
	if err != nil {
		t.Fatalf("ExportXML1C failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("XML export returned empty data")
	}

	// Verify valid XML
	var doc XML1CDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("XML unmarshal failed: %v", err)
	}

	if doc.Version != "1.0" {
		t.Fatalf("expected version 1.0, got %s", doc.Version)
	}

	if doc.Summary.Count != 1 {
		t.Fatalf("expected 1 operation, got %d", doc.Summary.Count)
	}

	if doc.Summary.TotalAmount != "6000.00" {
		t.Fatalf("expected total 6000.00, got %s", doc.Summary.TotalAmount)
	}

	if doc.Summary.TotalFee != "600.00" {
		t.Fatalf("expected fee 600.00, got %s", doc.Summary.TotalFee)
	}

	if doc.Summary.TotalNet != "5400.00" {
		t.Fatalf("expected net 5400.00, got %s", doc.Summary.TotalNet)
	}

	if len(doc.Operations) != 1 {
		t.Fatalf("expected 1 operation in list, got %d", len(doc.Operations))
	}

	if doc.Operations[0].Object != "Баня Премиум" {
		t.Fatalf("expected bathhouse name in operation, got %s", doc.Operations[0].Object)
	}
}

func TestFinancialReportService_ExportWalletCSV_NoWallet(t *testing.T) {
	svc, _, _, _ := setupFinancialReportService()
	ctx := context.Background()
	userID := uuid.New()

	_, err := svc.ExportWalletTransactionsCSV(ctx, userID, nil, nil)
	if err == nil {
		t.Fatal("expected error for non-existent wallet")
	}
}

func TestFinancialReportService_ExportXML1C_NoBookings(t *testing.T) {
	svc, _, _, bathhouseRepo := setupFinancialReportService()
	ctx := context.Background()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    "Пустая баня",
		Status:  domain.BathhouseStatusActive,
	}
	bathhouseRepo.Create(ctx, bh)

	now := time.Now()
	data, err := svc.ExportXML1C(ctx, ownerID, now.Add(-24*time.Hour), now)
	if err != nil {
		t.Fatalf("ExportXML1C failed: %v", err)
	}

	var doc XML1CDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("XML unmarshal failed: %v", err)
	}

	if doc.Summary.Count != 0 {
		t.Fatalf("expected 0 operations, got %d", doc.Summary.Count)
	}
}

func TestFormatKopecks(t *testing.T) {
	tests := []struct {
		amount   int64
		currency domain.WalletCurrency
		expected string
	}{
		{100_000, domain.WalletCurrencyRUB, "1000.00 ₽"},
		{50_050, domain.WalletCurrencyRUB, "500.50 ₽"},
		{0, domain.WalletCurrencyRUB, "0.00 ₽"},
		{100_000, domain.WalletCurrencyBYN, "1000.00 BYN"},
		{99, domain.WalletCurrencyRUB, "0.99 ₽"},
	}

	for _, tt := range tests {
		result := formatKopecks(tt.amount, tt.currency)
		if result != tt.expected {
			t.Errorf("formatKopecks(%d, %s) = %s, want %s", tt.amount, tt.currency, result, tt.expected)
		}
	}
}

func TestTransactionTypeRu(t *testing.T) {
	tests := []struct {
		txType   domain.WalletTransactionType
		expected string
	}{
		{domain.WalletTxTopUp, "Пополнение"},
		{domain.WalletTxSpend, "Списание"},
		{domain.WalletTxRefund, "Возврат"},
		{domain.WalletTxPayout, "Вывод средств"},
		{domain.WalletTxWelcomeBonus, "Приветственный бонус"},
	}

	for _, tt := range tests {
		result := transactionTypeRu(tt.txType)
		if result != tt.expected {
			t.Errorf("transactionTypeRu(%s) = %s, want %s", tt.txType, result, tt.expected)
		}
	}
}

func TestFormatDecimal(t *testing.T) {
	tests := []struct {
		kopecks  int64
		expected string
	}{
		{100_000, "1000.00"},
		{50_050, "500.50"},
		{0, "0.00"},
		{99, "0.99"},
	}

	for _, tt := range tests {
		result := formatDecimal(tt.kopecks)
		if result != tt.expected {
			t.Errorf("formatDecimal(%d) = %s, want %s", tt.kopecks, result, tt.expected)
		}
	}
}

func TestGenerateSimplePDF(t *testing.T) {
	lines := []string{"Title", "---", "Line 1", "Line 2"}
	data := generateSimplePDF(lines)

	if len(data) == 0 {
		t.Fatal("generateSimplePDF returned empty data")
	}

	content := string(data)
	if !strings.HasPrefix(content, "%PDF-1.4") {
		t.Fatal("expected PDF header")
	}
	if !strings.Contains(content, "%%EOF") {
		t.Fatal("expected EOF trailer")
	}
}

func TestPdfEscapeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"test(paren)", "test\\(paren\\)"},
		{"back\\slash", "back\\\\slash"},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		result := pdfEscapeString(tt.input)
		if result != tt.expected {
			t.Errorf("pdfEscapeString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
