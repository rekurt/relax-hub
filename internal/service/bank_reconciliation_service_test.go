package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
)

// mockBankReconciliationRepo implements repository.BankReconciliationRepository for testing.
type mockBankReconciliationRepo struct {
	uploads     map[uuid.UUID]*domain.BankStatementUpload
	entries     map[uuid.UUID]*domain.BankStatementEntry
	allEntries  []domain.BankStatementEntry
	payments    []domain.Payment
	matchCalls  int
}

func newMockBankRepo() *mockBankReconciliationRepo {
	return &mockBankReconciliationRepo{
		uploads: make(map[uuid.UUID]*domain.BankStatementUpload),
		entries: make(map[uuid.UUID]*domain.BankStatementEntry),
	}
}

func (m *mockBankReconciliationRepo) CreateUpload(_ context.Context, upload *domain.BankStatementUpload) error {
	m.uploads[upload.ID] = upload
	return nil
}

func (m *mockBankReconciliationRepo) UpdateUploadCounts(_ context.Context, id uuid.UUID, matched, pending, ignored int) error {
	if u, ok := m.uploads[id]; ok {
		u.MatchedCount = matched
		u.PendingCount = pending
		u.IgnoredCount = ignored
	}
	return nil
}

func (m *mockBankReconciliationRepo) CreateEntries(_ context.Context, entries []domain.BankStatementEntry) error {
	for i := range entries {
		m.entries[entries[i].ID] = &entries[i]
		m.allEntries = append(m.allEntries, entries[i])
	}
	return nil
}

func (m *mockBankReconciliationRepo) GetEntryByID(_ context.Context, id uuid.UUID) (*domain.BankStatementEntry, error) {
	if e, ok := m.entries[id]; ok {
		return e, nil
	}
	return nil, domain.ErrBankEntryNotFound
}

func (m *mockBankReconciliationRepo) ListEntries(_ context.Context, filter domain.BankStatementFilter) (*domain.PaginatedResult[domain.BankStatementEntry], error) {
	var filtered []domain.BankStatementEntry
	for _, e := range m.allEntries {
		if filter.Status != nil && e.Status != *filter.Status {
			continue
		}
		filtered = append(filtered, e)
	}
	return &domain.PaginatedResult[domain.BankStatementEntry]{
		Items:      filtered,
		TotalCount: int64(len(filtered)),
		Page:       1,
		PageSize:   20,
	}, nil
}

func (m *mockBankReconciliationRepo) MatchEntry(_ context.Context, entryID uuid.UUID, txID uuid.UUID, txType string) error {
	e, ok := m.entries[entryID]
	if !ok {
		return domain.ErrBankEntryNotFound
	}
	if e.Status != domain.BankEntryPending {
		return domain.ErrBankEntryAlreadyMatched
	}
	e.MatchedTxID = &txID
	e.MatchedTxType = txType
	e.Status = domain.BankEntryManual
	m.matchCalls++
	return nil
}

func (m *mockBankReconciliationRepo) FindPaymentsByAmountAndDate(_ context.Context, amount int64, _, _ time.Time) ([]domain.Payment, error) {
	var result []domain.Payment
	for _, p := range m.payments {
		if p.Amount == amount {
			result = append(result, p)
		}
	}
	return result, nil
}

func TestBankReconciliationService_UploadCSV(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	csv := "date,amount,description,counterparty\n" +
		"15.03.2026,1500.50,Оплата,ООО Тест\n" +
		"16.03.2026,2000,Поступление,ИП Иванов\n"

	adminID := uuid.New()
	upload, err := svc.UploadStatement(context.Background(), adminID, "test.csv", "csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if upload.TotalRows != 2 {
		t.Fatalf("expected 2 rows, got %d", upload.TotalRows)
	}
	if upload.FileName != "test.csv" {
		t.Fatalf("expected test.csv, got %s", upload.FileName)
	}
	if len(repo.allEntries) != 2 {
		t.Fatalf("expected 2 entries in repo, got %d", len(repo.allEntries))
	}
}

func TestBankReconciliationService_UploadCSV_AutoMatch(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	// Add a payment that should be auto-matched
	paymentID := uuid.New()
	repo.payments = []domain.Payment{
		{
			ID:     paymentID,
			Amount: 150050, // 1500.50 in kopecks
			Status: domain.PaymentSucceeded,
		},
	}

	csv := "date,amount\n15.03.2026,1500.50\n"

	upload, err := svc.UploadStatement(context.Background(), uuid.New(), "test.csv", "csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if upload.MatchedCount != 1 {
		t.Fatalf("expected 1 matched, got %d", upload.MatchedCount)
	}
	if upload.PendingCount != 0 {
		t.Fatalf("expected 0 pending, got %d", upload.PendingCount)
	}
}

func TestBankReconciliationService_UploadCSV_NoAutoMatchMultipleCandidates(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	// Add two payments with same amount - should NOT auto-match
	repo.payments = []domain.Payment{
		{ID: uuid.New(), Amount: 150050, Status: domain.PaymentSucceeded},
		{ID: uuid.New(), Amount: 150050, Status: domain.PaymentSucceeded},
	}

	csv := "date,amount\n15.03.2026,1500.50\n"

	upload, err := svc.UploadStatement(context.Background(), uuid.New(), "test.csv", "csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if upload.MatchedCount != 0 {
		t.Fatalf("expected 0 matched (ambiguous), got %d", upload.MatchedCount)
	}
	if upload.PendingCount != 1 {
		t.Fatalf("expected 1 pending, got %d", upload.PendingCount)
	}
}

func TestBankReconciliationService_Upload1C(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	content := "1CClientBankExchange\n" +
		"СекцияДокумент=Платежное поручение\n" +
		"Номер=123\nДата=15.03.2026\nСумма=1500\nПлательщик=ООО Тест\n" +
		"НазначениеПлатежа=Оплата\nКонецДокумента\n"

	upload, err := svc.UploadStatement(context.Background(), uuid.New(), "export.txt", "1c", strings.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if upload.TotalRows != 1 {
		t.Fatalf("expected 1 row, got %d", upload.TotalRows)
	}
}

func TestBankReconciliationService_InvalidFormat(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	_, err := svc.UploadStatement(context.Background(), uuid.New(), "test.xml", "xml", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestBankReconciliationService_ManualMatch(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	entryID := uuid.New()
	repo.entries[entryID] = &domain.BankStatementEntry{
		ID:     entryID,
		Status: domain.BankEntryPending,
	}

	txID := uuid.New()
	err := svc.ManualMatch(context.Background(), entryID, txID, "payment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify entry was matched
	entry := repo.entries[entryID]
	if entry.Status != domain.BankEntryManual {
		t.Fatalf("expected manual status, got %s", entry.Status)
	}
	if entry.MatchedTxID == nil || *entry.MatchedTxID != txID {
		t.Fatal("matched tx ID mismatch")
	}
}

func TestBankReconciliationService_ManualMatch_InvalidTxType(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	err := svc.ManualMatch(context.Background(), uuid.New(), uuid.New(), "invalid")
	if err == nil {
		t.Fatal("expected error for invalid tx type")
	}
}

func TestBankReconciliationService_ManualMatch_AlreadyMatched(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	entryID := uuid.New()
	txID := uuid.New()
	repo.entries[entryID] = &domain.BankStatementEntry{
		ID:          entryID,
		Status:      domain.BankEntryMatched,
		MatchedTxID: &txID,
	}

	err := svc.ManualMatch(context.Background(), entryID, uuid.New(), "payment")
	if err == nil {
		t.Fatal("expected error for already matched entry")
	}
}

func TestBankReconciliationService_ListUnmatched(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	// Add entries with different statuses
	repo.allEntries = []domain.BankStatementEntry{
		{ID: uuid.New(), Status: domain.BankEntryPending},
		{ID: uuid.New(), Status: domain.BankEntryMatched},
		{ID: uuid.New(), Status: domain.BankEntryPending},
	}

	pending := domain.BankEntryPending
	result, err := svc.ListUnmatched(context.Background(), domain.BankStatementFilter{Status: &pending})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("expected 2 pending entries, got %d", len(result.Items))
	}
}

func TestBankReconciliationService_NegativeAmountsSkipAutoMatch(t *testing.T) {
	repo := newMockBankRepo()
	log := logger.New(logger.LevelError)
	svc := NewBankReconciliationService(repo, log)

	// Add a payment
	repo.payments = []domain.Payment{
		{ID: uuid.New(), Amount: 50000, Status: domain.PaymentSucceeded},
	}

	// Negative amount should not be auto-matched
	csv := "date,amount\n15.03.2026,-500.00\n"

	upload, err := svc.UploadStatement(context.Background(), uuid.New(), "test.csv", "csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if upload.MatchedCount != 0 {
		t.Fatalf("expected 0 matched for negative amount, got %d", upload.MatchedCount)
	}
}
