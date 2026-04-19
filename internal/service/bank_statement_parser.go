package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

// ParsedBankEntry represents a parsed bank statement row before persisting.
type ParsedBankEntry struct {
	Date         time.Time
	Amount       int64  // kopecks, positive = incoming, negative = outgoing
	Description  string
	Counterparty string
	ReferenceNum string
}

// ParseCSVBankStatement parses a CSV bank statement.
// Expected columns: date, amount, description, counterparty, reference_num
// Date format: DD.MM.YYYY or YYYY-MM-DD
// Amount: rubles with optional kopecks (e.g., "1500.50" or "1500,50"), negative for outgoing
func ParseCSVBankStatement(r io.Reader) ([]ParsedBankEntry, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	// Try to auto-detect separator
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("чтение заголовка CSV: %w", err)
	}

	// If only one column detected, try semicolon separator
	if len(header) == 1 && strings.Contains(header[0], ";") {
		return nil, fmt.Errorf("обнаружен разделитель ';', используйте формат 1C для загрузки")
	}

	colIndex := buildBankColumnIndex(header)
	if err := validateBankHeader(colIndex); err != nil {
		return nil, err
	}

	var entries []ParsedBankEntry
	rowNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rowNum++
			continue
		}
		rowNum++

		entry, err := parseBankRow(record, colIndex, rowNum)
		if err != nil {
			continue // skip unparseable rows
		}
		entries = append(entries, *entry)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("не найдено ни одной строки для импорта")
	}

	return entries, nil
}

// Parse1CBankStatement parses a 1C bank exchange format.
// 1C format uses sections delimited by "СекцияДокумент=" and "КонецДокумента"
// with key=value pairs for each transaction.
func Parse1CBankStatement(r io.Reader) ([]ParsedBankEntry, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("чтение файла 1C: %w", err)
	}

	content := string(data)
	var entries []ParsedBankEntry

	// Split by document sections
	sections := strings.Split(content, "СекцияДокумент=")
	for _, section := range sections[1:] { // skip everything before first section
		entry, err := parse1CSection(section)
		if err != nil {
			continue // skip unparseable sections
		}
		entries = append(entries, *entry)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("не найдено ни одной транзакции в файле 1C")
	}

	return entries, nil
}

func parse1CSection(section string) (*ParsedBankEntry, error) {
	lines := strings.Split(section, "\n")
	kv := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "КонецДокумента" {
			break
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			kv[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Parse date
	dateStr := kv["Дата"]
	if dateStr == "" {
		return nil, fmt.Errorf("отсутствует поле Дата")
	}
	date, err := parseBankDate(dateStr)
	if err != nil {
		return nil, err
	}

	// Parse amount - 1C uses "Сумма" field
	amountStr := kv["Сумма"]
	if amountStr == "" {
		return nil, fmt.Errorf("отсутствует поле Сумма")
	}
	amount, err := parseBankAmount(amountStr)
	if err != nil {
		return nil, err
	}

	// Determine direction: ДебетКредит or check ПлательщикСчет vs ПолучательСчет
	// In 1C format, "Кредит" means incoming, "Дебет" means outgoing for our account
	if dc, ok := kv["ВидДокумента"]; ok {
		if strings.Contains(strings.ToLower(dc), "списан") {
			amount = -amount
		}
	}

	description := kv["НазначениеПлатежа"]
	counterparty := kv["Плательщик"]
	if counterparty == "" {
		counterparty = kv["Получатель"]
	}
	referenceNum := kv["Номер"]

	return &ParsedBankEntry{
		Date:         date,
		Amount:       amount,
		Description:  description,
		Counterparty: counterparty,
		ReferenceNum: referenceNum,
	}, nil
}

func buildBankColumnIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, col := range header {
		normalized := strings.TrimSpace(strings.ToLower(col))
		idx[normalized] = i
		// Also map Russian column names to English equivalents
		switch normalized {
		case "дата":
			idx["date"] = i
		case "сумма":
			idx["amount"] = i
		case "назначение платежа", "назначение", "описание":
			idx["description"] = i
		case "контрагент", "плательщик/получатель":
			idx["counterparty"] = i
		case "номер документа", "номер", "референс":
			idx["reference_num"] = i
		}
	}
	return idx
}

func validateBankHeader(colIndex map[string]int) error {
	required := []string{"date", "amount"}
	var missing []string
	for _, col := range required {
		if _, ok := colIndex[col]; !ok {
			missing = append(missing, col)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("отсутствуют обязательные столбцы: %s (ожидаются: date/дата, amount/сумма)", strings.Join(missing, ", "))
	}
	return nil
}

func parseBankRow(record []string, colIndex map[string]int, rowNum int) (*ParsedBankEntry, error) {
	getCol := func(name string) string {
		i, ok := colIndex[name]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	dateStr := getCol("date")
	if dateStr == "" {
		return nil, fmt.Errorf("строка %d: пустая дата", rowNum)
	}
	date, err := parseBankDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("строка %d: %w", rowNum, err)
	}

	amountStr := getCol("amount")
	if amountStr == "" {
		return nil, fmt.Errorf("строка %d: пустая сумма", rowNum)
	}
	amount, err := parseBankAmount(amountStr)
	if err != nil {
		return nil, fmt.Errorf("строка %d: %w", rowNum, err)
	}

	return &ParsedBankEntry{
		Date:         date,
		Amount:       amount,
		Description:  getCol("description"),
		Counterparty: getCol("counterparty"),
		ReferenceNum: getCol("reference_num"),
	}, nil
}

// parseBankDate parses dates in DD.MM.YYYY or YYYY-MM-DD format.
func parseBankDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"02.01.2006", "2006-01-02", "02/01/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("неверный формат даты: %s (ожидается ДД.ММ.ГГГГ или ГГГГ-ММ-ДД)", s)
}

// parseBankAmount parses a monetary amount string to kopecks.
// Handles: "1500.50", "1500,50", "-1500.50", "1 500,50"
func parseBankAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")  // remove thousand separators
	s = strings.ReplaceAll(s, "\u00a0", "") // remove non-breaking spaces
	s = strings.ReplaceAll(s, ",", ".") // normalize decimal separator

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("неверный формат суммы: %s", s)
	}
	if f == 0 {
		return 0, fmt.Errorf("сумма не может быть нулевой")
	}

	// Convert to kopecks
	return int64(math.Round(f * 100)), nil
}

// ToBankStatementEntries converts parsed entries into domain entries with a batch ID.
func ToBankStatementEntries(parsed []ParsedBankEntry, batchID uuid.UUID) []domain.BankStatementEntry {
	now := time.Now()
	entries := make([]domain.BankStatementEntry, len(parsed))
	for i, p := range parsed {
		entries[i] = domain.BankStatementEntry{
			ID:            uuid.New(),
			Date:          p.Date,
			Amount:        p.Amount,
			Description:   p.Description,
			Counterparty:  p.Counterparty,
			ReferenceNum:  p.ReferenceNum,
			Status:        domain.BankEntryPending,
			UploadBatchID: batchID,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
	}
	return entries
}
