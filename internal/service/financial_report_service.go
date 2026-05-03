package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository"
)

type FinancialReportService interface {
	ExportWalletTransactionsCSV(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error)
	ExportWalletTransactionsPDF(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error)
	ExportPayoutsCSV(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error)
	ExportPayoutsPDF(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error)
	GenerateOwnerAct(ctx context.Context, ownerID uuid.UUID, bathhouseID uuid.UUID, dateFrom, dateTo time.Time) ([]byte, error)
	ExportXML1C(ctx context.Context, ownerID uuid.UUID, dateFrom, dateTo time.Time) ([]byte, error)
}

type financialReportService struct {
	walletRepo    repository.WalletRepository
	bookingRepo   repository.BookingRepository
	bathhouseRepo repository.BathhouseRepository
	paymentRepo   repository.PaymentRepository
	escrowRepo    repository.EscrowRepository
	payoutRepo    repository.PayoutRepository
	logger        *logger.Logger
}

func NewFinancialReportService(
	walletRepo repository.WalletRepository,
	bookingRepo repository.BookingRepository,
	bathhouseRepo repository.BathhouseRepository,
	paymentRepo repository.PaymentRepository,
	escrowRepo repository.EscrowRepository,
	payoutRepo repository.PayoutRepository,
	log *logger.Logger,
) FinancialReportService {
	return &financialReportService{
		walletRepo:    walletRepo,
		bookingRepo:   bookingRepo,
		bathhouseRepo: bathhouseRepo,
		paymentRepo:   paymentRepo,
		escrowRepo:    escrowRepo,
		payoutRepo:    payoutRepo,
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
	lines = append(lines, "Акт оказанных услуг")
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
	XMLName    xml.Name         `xml:"Документ"`
	Version    string           `xml:"Версия,attr"`
	DateFrom   string           `xml:"ПериодС"`
	DateTo     string           `xml:"ПериодПо"`
	Owner      string           `xml:"Владелец"`
	Operations []XML1COperation `xml:"Операции>Операция"`
	Summary    XML1CSummary     `xml:"Итого"`
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
	var allBathhouses []domain.Bathhouse
	page := 1
	for {
		result, err := s.bathhouseRepo.ListByOwner(ctx, ownerID, page, 100)
		if err != nil {
			return nil, err
		}
		allBathhouses = append(allBathhouses, result.Items...)
		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	doc := XML1CDocument{
		Version:  "1.0",
		DateFrom: dateFrom.Format("2006-01-02"),
		DateTo:   dateTo.Format("2006-01-02"),
		Owner:    ownerID.String(),
	}

	var totalAmount, totalFee int64
	num := 0

	for _, bh := range allBathhouses {
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

func (s *financialReportService) ExportPayoutsCSV(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error) {
	payouts, err := s.fetchAllPayouts(ctx, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(&buf)
	w.Comma = ';'

	header := []string{"Дата запроса", "Сумма", "Статус", "Дата обработки", "Причина отказа"}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, p := range payouts {
		row := []string{
			p.RequestedAt.Format("02.01.2006 15:04"),
			formatKopecksRUB(p.Amount),
			payoutStatusRu(p.Status),
			formatOptionalTimePtr(p.ProcessedAt),
			p.FailureReason,
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

func (s *financialReportService) ExportPayoutsPDF(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]byte, error) {
	payouts, err := s.fetchAllPayouts(ctx, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}

	title := "Выписка по выплатам"
	if dateFrom != nil && dateTo != nil {
		title += fmt.Sprintf(" за период %s — %s", dateFrom.Format("02.01.2006"), dateTo.Format("02.01.2006"))
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%-20s %12s %-16s %-20s %s",
		"Дата запроса", "Сумма", "Статус", "Дата обработки", "Причина отказа"))
	lines = append(lines, "---")

	var totalAmount int64
	for _, p := range payouts {
		if p.Status == domain.PayoutStatusCompleted {
			totalAmount += p.Amount
		}
		line := fmt.Sprintf("%-20s %12s %-16s %-20s %s",
			p.RequestedAt.Format("02.01.2006 15:04"),
			formatKopecksRUB(p.Amount),
			payoutStatusRu(p.Status),
			formatOptionalTimePtr(p.ProcessedAt),
			p.FailureReason,
		)
		lines = append(lines, line)
	}

	lines = append(lines, "---")
	lines = append(lines, fmt.Sprintf("Итого выплат: %d", len(payouts)))
	lines = append(lines, fmt.Sprintf("Общая сумма: %s", formatKopecksRUB(totalAmount)))

	return generateSimplePDF(lines), nil
}

// fetchAllPayouts retrieves all payouts for a user within the date range.
func (s *financialReportService) fetchAllPayouts(ctx context.Context, userID uuid.UUID, dateFrom, dateTo *time.Time) ([]domain.Payout, error) {
	var all []domain.Payout
	page := 1
	pageSize := 500

	for {
		result, err := s.payoutRepo.ListByUser(ctx, userID, page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, p := range result.Items {
			if dateFrom != nil && p.RequestedAt.Before(*dateFrom) {
				continue
			}
			if dateTo != nil && p.RequestedAt.After(*dateTo) {
				continue
			}
			all = append(all, p)
		}

		if page >= result.TotalPages || len(result.Items) == 0 {
			break
		}
		page++
	}

	return all, nil
}

func payoutStatusRu(s domain.PayoutStatus) string {
	switch s {
	case domain.PayoutStatusPending:
		return "Ожидает"
	case domain.PayoutStatusProcessing:
		return "Обрабатывается"
	case domain.PayoutStatusCompleted:
		return "Выполнена"
	case domain.PayoutStatusFailed:
		return "Ошибка"
	default:
		return string(s)
	}
}

func formatOptionalTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("02.01.2006 15:04")
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
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	rubles := amount / 100
	kopecks := amount % 100
	sym := "₽"
	if currency == domain.WalletCurrencyBYN {
		sym = "BYN"
	}
	return fmt.Sprintf("%s%d.%02d %s", sign, rubles, kopecks, sym)
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

// generateSimplePDF generates an HTML document styled for printing as PDF.
// Standard PDF Type1 fonts (Courier, Helvetica, etc.) do not support Cyrillic glyphs,
// so we use HTML with UTF-8 encoding which handles all Unicode characters natively.
// Users can print to PDF from their browser if a true PDF is needed.
func generateSimplePDF(lines []string) []byte {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<style>
  @page { size: A4; margin: 20mm; }
  body { font-family: monospace; font-size: 10pt; line-height: 1.4; }
  hr { border: none; border-top: 1px solid #000; margin: 4px 0; }
  .line { white-space: pre-wrap; }
</style>
</head>
<body>
`)

	for _, line := range lines {
		if line == "---" {
			buf.WriteString("<hr>\n")
		} else {
			buf.WriteString(`<div class="line">`)
			buf.WriteString(htmlEscapeString(line))
			buf.WriteString("</div>\n")
		}
	}

	buf.WriteString("</body>\n</html>\n")
	return buf.Bytes()
}

func htmlEscapeString(s string) string {
	var buf bytes.Buffer
	for _, c := range s {
		switch c {
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '&':
			buf.WriteString("&amp;")
		case '"':
			buf.WriteString("&quot;")
		default:
			buf.WriteRune(c)
		}
	}
	return buf.String()
}
