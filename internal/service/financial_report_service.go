package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type FinancialReportService interface {
	ExportWalletTransactionsCSV(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error)
	ExportWalletTransactionsPDF(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error)
	GenerateOwnerAct(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID, dateFrom, dateTo time.Time) ([]byte, error)
	ExportXML1C(ctx context.Context, ownerID uuid.UUID, dateFrom, dateTo time.Time) ([]byte, error)
}

type financialReportService struct {
	walletRepo    repository.WalletRepository
	bookingRepo   repository.BookingRepository
	bathhouseRepo repository.BathhouseRepository
	paymentRepo   repository.PaymentRepository
	escrowRepo    repository.EscrowRepository
	logger        *logger.Logger
}

func NewFinancialReportService(
	walletRepo repository.WalletRepository,
	bookingRepo repository.BookingRepository,
	bathhouseRepo repository.BathhouseRepository,
	paymentRepo repository.PaymentRepository,
	escrowRepo repository.EscrowRepository,
	log *logger.Logger,
) FinancialReportService {
	return &financialReportService{
		walletRepo:    walletRepo,
		bookingRepo:   bookingRepo,
		bathhouseRepo: bathhouseRepo,
		paymentRepo:   paymentRepo,
		escrowRepo:    escrowRepo,
		logger:        log,
	}
}

func (s *financialReportService) ExportWalletTransactionsCSV(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error) {
	txs, err := s.fetchAllTransactions(ctx, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	// BOM for Excel UTF-8 support
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	w.Comma = ';'

	header := []string{"Дата", "Тип", "Сумма", "Баланс после", "Статус", "Описание", "Бонус", "Срок бонуса"}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, tx := range txs {
		row := []string{
			tx.CreatedAt.Format("02.01.2006 15:04"),
			transactionTypeRu(tx.Type),
			formatKopecks(tx.Amount, wallet.Currency),
			formatKopecks(tx.BalanceAfter, wallet.Currency),
			transactionStatusRu(tx.Status),
			tx.Description,
			boolRu(tx.IsBonus),
			formatOptionalTime(tx.ExpiresAt),
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *financialReportService) ExportWalletTransactionsPDF(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error) {
	txs, err := s.fetchAllTransactions(ctx, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	title := "Выписка по кошельку"
	if dateFrom != nil && dateTo != nil {
		title += fmt.Sprintf(" за период %s — %s", dateFrom.Format("02.01.2006"), dateTo.Format("02.01.2006"))
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, fmt.Sprintf("Валюта: %s", string(wallet.Currency)))
	lines = append(lines, fmt.Sprintf("Текущий баланс: %s", formatKopecks(wallet.Balance, wallet.Currency)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%-20s %-18s %12s %12s %-12s %s",
		"Дата", "Тип", "Сумма", "Баланс", "Статус", "Описание"))
	lines = append(lines, "---")

	for _, tx := range txs {
		line := fmt.Sprintf("%-20s %-18s %12s %12s %-12s %s",
			tx.CreatedAt.Format("02.01.2006 15:04"),
			transactionTypeRu(tx.Type),
			formatKopecks(tx.Amount, wallet.Currency),
			formatKopecks(tx.BalanceAfter, wallet.Currency),
			transactionStatusRu(tx.Status),
			tx.Description,
		)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Итого операций: %d", len(txs)))

	return generateSimplePDF(lines), nil
}

func (s *financialReportService) GenerateOwnerAct(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID, dateFrom, dateTo time.Time) ([]byte, error) {
	bh, err := s.bathhouseRepo.GetByID(ctx, bathhouseID)
	if err != nil {
		return nil, err
	}
	if bh.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}

	bookings, err := s.fetchCompletedBookings(ctx, bathhouseID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Акт оказанных услуг"))
	lines = append(lines, fmt.Sprintf("Объект: %s", bh.Name))
	lines = append(lines, fmt.Sprintf("Период: %s — %s", dateFrom.Format("02.01.2006"), dateTo.Format("02.01.2006")))
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("%-5s %-20s %-20s %8s %12s %12s",
		"#", "Начало", "Конец", "Гости", "Сумма", "Комиссия"))
	lines = append(lines, "---")

	var totalAmount, totalFee int64
	for i, b := range bookings {
		fee := b.ServiceFeeAmount
		totalAmount += b.TotalPrice
		totalFee += fee

		line := fmt.Sprintf("%-5d %-20s %-20s %8d %12s %12s",
			i+1,
			b.StartTime.Format("02.01.2006 15:04"),
			b.EndTime.Format("02.01.2006 15:04"),
			b.GuestCount,
			formatKopecksRUB(b.TotalPrice),
			formatKopecksRUB(fee),
		)
		lines = append(lines, line)
	}

	lines = append(lines, "---")
	lines = append(lines, fmt.Sprintf("Итого бронирований: %d", len(bookings)))
	lines = append(lines, fmt.Sprintf("Общая сумма: %s", formatKopecksRUB(totalAmount)))
	lines = append(lines, fmt.Sprintf("Комиссия платформы: %s", formatKopecksRUB(totalFee)))
	lines = append(lines, fmt.Sprintf("К выплате: %s", formatKopecksRUB(totalAmount-totalFee)))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Дата формирования: %s", time.Now().Format("02.01.2006")))

	return generateSimplePDF(lines), nil
}

// XML1CDocument represents a 1C-compatible XML document for financial data exchange.
type XML1CDocument struct {
	XMLName    xml.Name       `xml:"Документ"`
	Version    string         `xml:"Версия,attr"`
	DateFrom   string         `xml:"ПериодС"`
	DateTo     string         `xml:"ПериодПо"`
	Owner      string         `xml:"Владелец"`
	Operations []XML1COperation `xml:"Операции>Операция"`
	Summary    XML1CSummary   `xml:"Итого"`
}

// XML1COperation represents a single financial operation in 1C XML format.
type XML1COperation struct {
	Number    int    `xml:"Номер"`
	Date      string `xml:"Дата"`
	Type      string `xml:"Тип"`
	Amount    string `xml:"Сумма"`
	Fee       string `xml:"Комиссия"`
	NetAmount string `xml:"КВыплате"`
	Object    string `xml:"Объект"`
}

// XML1CSummary represents totals in 1C XML format.
type XML1CSummary struct {
	TotalAmount string `xml:"Сумма"`
	TotalFee    string `xml:"Комиссия"`
	TotalNet    string `xml:"КВыплате"`
	Count       int    `xml:"КоличествоОпераций"`
}

func (s *financialReportService) ExportXML1C(ctx context.Context, ownerID uuid.UUID, dateFrom, dateTo time.Time) ([]byte, error) {
	bathhouses, err := s.bathhouseRepo.ListByOwner(ctx, ownerID, 1, 100)
	if err != nil {
		return nil, err
	}

	doc := XML1CDocument{
		Version:  "1.0",
		DateFrom: dateFrom.Format("2006-01-02"),
		DateTo:   dateTo.Format("2006-01-02"),
		Owner:    ownerID.String(),
	}

	var totalAmount, totalFee int64
	num := 0

	for _, bh := range bathhouses.Items {
		bookings, err := s.fetchCompletedBookings(ctx, bh.ID, dateFrom, dateTo)
		if err != nil {
			return nil, err
		}

		for _, b := range bookings {
			num++
			fee := b.ServiceFeeAmount
			net := b.TotalPrice - fee
			totalAmount += b.TotalPrice
			totalFee += fee

			doc.Operations = append(doc.Operations, XML1COperation{
				Number:    num,
				Date:      b.StartTime.Format("2006-01-02"),
				Type:      "Бронирование",
				Amount:    formatDecimal(b.TotalPrice),
				Fee:       formatDecimal(fee),
				NetAmount: formatDecimal(net),
				Object:    bh.Name,
			})
		}
	}

	doc.Summary = XML1CSummary{
		TotalAmount: formatDecimal(totalAmount),
		TotalFee:    formatDecimal(totalFee),
		TotalNet:    formatDecimal(totalAmount - totalFee),
		Count:       num,
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("xml encode: %w", err)
	}

	return buf.Bytes(), nil
}

// fetchAllTransactions retrieves all wallet transactions for a user within the date range.
func (s *financialReportService) fetchAllTransactions(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]domain.WalletTransaction, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var all []domain.WalletTransaction
	page := 1
	pageSize := 500

	for {
		filter := domain.WalletTransactionFilter{
			WalletID: &wallet.ID,
			DateFrom: dateFrom,
			DateTo:   dateTo,
			Page:     page,
			PageSize: pageSize,
		}

		result, err := s.walletRepo.ListTransactions(ctx, filter)
		if err != nil {
			return nil, err
		}

		all = append(all, result.Items...)

		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return all, nil
}

// fetchCompletedBookings retrieves completed bookings for a bathhouse within date range.
func (s *financialReportService) fetchCompletedBookings(ctx context.Context, bathhouseID uuid.UUID, dateFrom, dateTo time.Time) ([]domain.Booking, error) {
	var completed []domain.Booking
	page := 1
	pageSize := 100

	for {
		result, err := s.bookingRepo.ListByBathhouse(ctx, bathhouseID, page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, b := range result.Items {
			if b.Status != domain.BookingCompleted {
				continue
			}
			if b.StartTime.Before(dateFrom) || b.StartTime.After(dateTo) {
				continue
			}
			completed = append(completed, b)
		}

		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return completed, nil
}

func transactionTypeRu(t domain.WalletTransactionType) string {
	switch t {
	case domain.WalletTxTopUp:
		return "Пополнение"
	case domain.WalletTxSpend:
		return "Списание"
	case domain.WalletTxRefund:
		return "Возврат"
	case domain.WalletTxBonus:
		return "Бонус"
	case domain.WalletTxBonusExpiry:
		return "Сгорание бонуса"
	case domain.WalletTxHoldCapture:
		return "Списание холда"
	case domain.WalletTxHoldRelease:
		return "Возврат холда"
	case domain.WalletTxPayout:
		return "Вывод средств"
	case domain.WalletTxWelcomeBonus:
		return "Приветственный бонус"
	case domain.WalletTxReferralBonus:
		return "Реферальный бонус"
	default:
		return string(t)
	}
}

func transactionStatusRu(s domain.WalletTransactionStatus) string {
	switch s {
	case domain.WalletTxStatusPending:
		return "В обработке"
	case domain.WalletTxStatusCompleted:
		return "Завершена"
	case domain.WalletTxStatusFailed:
		return "Ошибка"
	case domain.WalletTxStatusCancelled:
		return "Отменена"
	default:
		return string(s)
	}
}

func formatKopecks(amount int64, currency domain.WalletCurrency) string {
	rubles := amount / 100
	kopecks := amount % 100
	sym := "₽"
	if currency == domain.WalletCurrencyBYN {
		sym = "BYN"
	}
	return fmt.Sprintf("%d.%02d %s", rubles, kopecks, sym)
}

func formatKopecksRUB(amount int64) string {
	return formatKopecks(amount, domain.WalletCurrencyRUB)
}

func formatDecimal(kopecks int64) string {
	return fmt.Sprintf("%.2f", float64(kopecks)/100)
}

func boolRu(v bool) string {
	if v {
		return "Да"
	}
	return "Нет"
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("02.01.2006")
}

// generateSimplePDF creates a minimal valid PDF with text content.
// Uses PDF 1.4 spec directly - no external dependencies needed.
func generateSimplePDF(lines []string) []byte {
	var buf bytes.Buffer

	// PDF header
	buf.WriteString("%PDF-1.4\n")

	// Object 1: Catalog
	obj1Offset := buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages
	obj2Offset := buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Build content stream
	var content bytes.Buffer
	content.WriteString("BT\n")
	content.WriteString("/F1 10 Tf\n")
	y := 800
	for _, line := range lines {
		if line == "---" {
			content.WriteString(fmt.Sprintf("1 0 0 1 50 %d Tm\n", y))
			content.WriteString("(────────────────────────────────────────────────────) Tj\n")
		} else {
			content.WriteString(fmt.Sprintf("1 0 0 1 50 %d Tm\n", y))
			escaped := pdfEscapeString(line)
			content.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
		}
		y -= 14
		if y < 50 {
			break
		}
	}
	content.WriteString("ET\n")
	contentBytes := content.Bytes()

	// Object 3: Page
	obj3Offset := buf.Len()
	buf.WriteString(fmt.Sprintf("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n"))

	// Object 4: Content stream
	obj4Offset := buf.Len()
	buf.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n", len(contentBytes)))
	buf.Write(contentBytes)
	buf.WriteString("\nendstream\nendobj\n")

	// Object 5: Font
	obj5Offset := buf.Len()
	buf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>\nendobj\n")

	// Cross-reference table
	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString("0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj1Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj2Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj3Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj4Offset))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", obj5Offset))

	// Trailer
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset))

	return buf.Bytes()
}

func pdfEscapeString(s string) string {
	var buf bytes.Buffer
	for _, c := range s {
		switch c {
		case '(', ')', '\\':
			buf.WriteByte('\\')
			buf.WriteRune(c)
		default:
			buf.WriteRune(c)
		}
	}
	return buf.String()
}
