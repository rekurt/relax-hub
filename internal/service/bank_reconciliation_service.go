package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

// BankReconciliationService handles bank statement import and matching.
type BankReconciliationService interface {
	// UploadStatement parses and imports a bank statement file.
	UploadStatement(ctx context.Context, adminID uuid.UUID, fileName string, format string, r io.Reader) (*domain.BankStatementUpload, error)

	// ListUnmatched returns pending (unmatched) bank statement entries.
	ListUnmatched(ctx context.Context, filter domain.BankStatementFilter) (*domain.PaginatedResult[domain.BankStatementEntry], error)

	// ManualMatch manually matches a bank entry to an internal transaction.
	ManualMatch(ctx context.Context, entryID uuid.UUID, txID uuid.UUID, txType string) error
}

type bankReconciliationService struct {
	repo   repository.BankReconciliationRepository
	logger *logger.Logger
}

func NewBankReconciliationService(
	repo repository.BankReconciliationRepository,
	log *logger.Logger,
) BankReconciliationService {
	return &bankReconciliationService{
		repo:   repo,
		logger: log,
	}
}

func (s *bankReconciliationService) UploadStatement(ctx context.Context, adminID uuid.UUID, fileName string, format string, r io.Reader) (*domain.BankStatementUpload, error) {
	// Parse the file based on format
	var parsed []ParsedBankEntry
	var err error

	switch strings.ToLower(format) {
	case "csv":
		parsed, err = ParseCSVBankStatement(r)
	case "1c":
		parsed, err = Parse1CBankStatement(r)
	default:
		return nil, fmt.Errorf("%w: неподдерживаемый формат файла: %s (допустимые: csv, 1c)", domain.ErrInvalidInput, format)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err.Error())
	}

	// Create upload record
	batchID := uuid.New()
	upload := &domain.BankStatementUpload{
		ID:         batchID,
		FileName:   fileName,
		Format:     format,
		TotalRows:  len(parsed),
		UploadedBy: adminID,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateUpload(ctx, upload); err != nil {
		return nil, fmt.Errorf("create upload record: %w", err)
	}

	// Convert to domain entries
	entries := ToBankStatementEntries(parsed, batchID)

	// Save entries
	if err := s.repo.CreateEntries(ctx, entries); err != nil {
		return nil, fmt.Errorf("create entries: %w", err)
	}

	// Auto-match entries against internal payments
	matchedCount := 0
	for i := range entries {
		matched, err := s.autoMatchEntry(ctx, &entries[i])
		if err != nil {
			s.logger.Error("auto-match failed", "entry_id", entries[i].ID, "error", err)
			continue
		}
		if matched {
			matchedCount++
		}
	}

	pendingCount := len(entries) - matchedCount
	upload.MatchedCount = matchedCount
	upload.PendingCount = pendingCount

	if err := s.repo.UpdateUploadCounts(ctx, batchID, matchedCount, pendingCount, 0); err != nil {
		s.logger.Error("update upload counts failed", "batch_id", batchID, "error", err)
	}

	s.logger.Info("bank statement uploaded",
		"batch_id", batchID,
		"file", fileName,
		"format", format,
		"total", len(entries),
		"matched", matchedCount,
		"pending", pendingCount,
	)

	return upload, nil
}

// autoMatchEntry tries to match a bank statement entry to an internal payment.
// Matching criteria: exact amount + date within +/-1 day + optional reference number.
func (s *bankReconciliationService) autoMatchEntry(ctx context.Context, entry *domain.BankStatementEntry) (bool, error) {
	if entry.Amount <= 0 {
		// Only match incoming payments (positive amounts)
		return false, nil
	}

	// Search for payments with same amount within +/-1 day window
	dateFrom := entry.Date.AddDate(0, 0, -1)
	dateTo := entry.Date.AddDate(0, 0, 2) // +1 day exclusive

	payments, err := s.repo.FindPaymentsByAmountAndDate(ctx, entry.Amount, dateFrom, dateTo)
	if err != nil {
		return false, err
	}

	if len(payments) == 0 {
		return false, nil
	}

	// If reference number is available, try to match by it first
	if entry.ReferenceNum != "" {
		for _, p := range payments {
			if p.ExternalID == entry.ReferenceNum {
				if err := s.repo.MatchEntry(ctx, entry.ID, p.ID, "payment"); err != nil {
					return false, err
				}
				entry.Status = domain.BankEntryMatched
				return true, nil
			}
		}
	}

	// If only one payment matches amount+date, auto-match it
	if len(payments) == 1 {
		if err := s.repo.MatchEntry(ctx, entry.ID, payments[0].ID, "payment"); err != nil {
			return false, err
		}
		entry.Status = domain.BankEntryMatched
		return true, nil
	}

	// Multiple candidates - leave as pending for manual matching
	return false, nil
}

func (s *bankReconciliationService) ListUnmatched(ctx context.Context, filter domain.BankStatementFilter) (*domain.PaginatedResult[domain.BankStatementEntry], error) {
	return s.repo.ListEntries(ctx, filter)
}

func (s *bankReconciliationService) ManualMatch(ctx context.Context, entryID uuid.UUID, txID uuid.UUID, txType string) error {
	if txType != "payment" && txType != "wallet_transaction" {
		return fmt.Errorf("%w: тип транзакции должен быть 'payment' или 'wallet_transaction'", domain.ErrInvalidInput)
	}

	// Verify entry exists and is pending
	entry, err := s.repo.GetEntryByID(ctx, entryID)
	if err != nil {
		return domain.ErrBankEntryNotFound
	}
	if entry.Status != domain.BankEntryPending {
		return domain.ErrBankEntryAlreadyMatched
	}

	return s.repo.MatchEntry(ctx, entryID, txID, txType)
}
