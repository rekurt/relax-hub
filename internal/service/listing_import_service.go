package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/logger"
)

// ListingImportService handles bulk CSV import of bathhouse listings.
type ListingImportService interface {
	ImportCSV(ctx context.Context, ownerID uuid.UUID, r io.Reader) (*ImportReport, error)
}

// ImportReport summarises the result of a CSV import.
type ImportReport struct {
	TotalRows    int           `json:"total_rows"`
	SuccessCount int           `json:"success_count"`
	ErrorCount   int           `json:"error_count"`
	Errors       []ImportError `json:"errors,omitempty"`
	CreatedIDs   []uuid.UUID   `json:"created_ids,omitempty"`
}

// ImportError describes a validation or creation error for a specific CSV row.
type ImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// csvColumns defines the expected column order in the import CSV.
var csvColumns = []string{
	"name", "description", "address", "city_id",
	"latitude", "longitude", "price_per_hour_rub",
	"min_duration", "max_guests",
	"has_pool", "has_sauna", "has_steam_room",
	"has_hot_tub", "has_bbq", "has_karaoke",
	"images",
}

type listingImportService struct {
	bathhouseSvc BathhouseService
	logger       *logger.Logger
}

func NewListingImportService(
	bathhouseSvc BathhouseService,
	log *logger.Logger,
) ListingImportService {
	return &listingImportService{
		bathhouseSvc: bathhouseSvc,
		logger:       log,
	}
}

func (s *listingImportService) ImportCSV(ctx context.Context, ownerID uuid.UUID, r io.Reader) (*ImportReport, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	colIndex := buildColumnIndex(header)
	if err := validateHeader(colIndex); err != nil {
		return nil, err
	}

	const maxRows = 500

	report := &ImportReport{}
	rowNum := 1 // 1-based, header is row 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			report.Errors = append(report.Errors, ImportError{
				Row: rowNum + 1, Message: fmt.Sprintf("ошибка чтения строки: %v", err),
			})
			report.ErrorCount++
			rowNum++
			continue
		}
		rowNum++
		report.TotalRows++

		if report.TotalRows > maxRows {
			return nil, fmt.Errorf("превышен лимит строк: максимум %d", maxRows)
		}

		input, rowErrors := parseRow(record, colIndex, rowNum)
		if len(rowErrors) > 0 {
			report.Errors = append(report.Errors, rowErrors...)
			report.ErrorCount++
			continue
		}

		bh, err := s.bathhouseSvc.Create(ctx, ownerID, *input)
		if err != nil {
			report.Errors = append(report.Errors, ImportError{
				Row: rowNum, Message: fmt.Sprintf("ошибка создания: %v", err),
			})
			report.ErrorCount++
			continue
		}

		report.SuccessCount++
		report.CreatedIDs = append(report.CreatedIDs, bh.ID)
	}

	s.logger.Info("CSV import complete",
		"owner_id", ownerID,
		"total", report.TotalRows,
		"success", report.SuccessCount,
		"errors", report.ErrorCount,
	)

	return report, nil
}

func buildColumnIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, col := range header {
		idx[strings.TrimSpace(strings.ToLower(col))] = i
	}
	return idx
}

func validateHeader(colIndex map[string]int) error {
	required := []string{"name", "address", "city_id", "latitude", "longitude", "price_per_hour_rub", "min_duration", "max_guests"}
	var missing []string
	for _, col := range required {
		if _, ok := colIndex[col]; !ok {
			missing = append(missing, col)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("отсутствуют обязательные столбцы: %s", strings.Join(missing, ", "))
	}
	return nil
}

func parseRow(record []string, colIndex map[string]int, rowNum int) (*CreateBathhouseInput, []ImportError) {
	var errs []ImportError

	getCol := func(name string) string {
		i, ok := colIndex[name]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	requireStr := func(name string) string {
		v := getCol(name)
		if v == "" {
			errs = append(errs, ImportError{Row: rowNum, Field: name, Message: "обязательное поле пустое"})
		}
		return v
	}

	parseInt := func(name string, required bool) int64 {
		v := getCol(name)
		if v == "" {
			if required {
				errs = append(errs, ImportError{Row: rowNum, Field: name, Message: "обязательное поле пустое"})
			}
			return 0
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			errs = append(errs, ImportError{Row: rowNum, Field: name, Message: "ожидается целое число"})
			return 0
		}
		return n
	}

	parseFloat := func(name string) float64 {
		v := getCol(name)
		if v == "" {
			errs = append(errs, ImportError{Row: rowNum, Field: name, Message: "обязательное поле пустое"})
			return 0
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			errs = append(errs, ImportError{Row: rowNum, Field: name, Message: "ожидается число"})
			return 0
		}
		return f
	}

	parseBool := func(name string) bool {
		v := strings.ToLower(getCol(name))
		return v == "1" || v == "true" || v == "да" || v == "yes"
	}

	name := requireStr("name")
	address := requireStr("address")
	cityID := parseInt("city_id", true)
	lat := parseFloat("latitude")
	lon := parseFloat("longitude")
	priceRub := parseInt("price_per_hour_rub", true)
	minDuration := int(parseInt("min_duration", true))
	maxGuests := int(parseInt("max_guests", true))
	description := getCol("description")

	// Validate ranges
	if priceRub <= 0 && len(errs) == 0 {
		errs = append(errs, ImportError{Row: rowNum, Field: "price_per_hour_rub", Message: "цена должна быть больше 0"})
	}
	if minDuration <= 0 && len(errs) == 0 {
		errs = append(errs, ImportError{Row: rowNum, Field: "min_duration", Message: "минимальная длительность должна быть больше 0"})
	}
	if maxGuests <= 0 && len(errs) == 0 {
		errs = append(errs, ImportError{Row: rowNum, Field: "max_guests", Message: "макс. гостей должно быть больше 0"})
	}
	if lat < -90 || lat > 90 {
		errs = append(errs, ImportError{Row: rowNum, Field: "latitude", Message: "широта должна быть от -90 до 90"})
	}
	if lon < -180 || lon > 180 {
		errs = append(errs, ImportError{Row: rowNum, Field: "longitude", Message: "долгота должна быть от -180 до 180"})
	}

	// Parse images (semicolon-separated URLs)
	var images []string
	if raw := getCol("images"); raw != "" {
		for _, imgURL := range strings.Split(raw, ";") {
			imgURL = strings.TrimSpace(imgURL)
			if imgURL == "" {
				continue
			}
			if !strings.HasPrefix(imgURL, "https://") && !strings.HasPrefix(imgURL, "http://") {
				errs = append(errs, ImportError{Row: rowNum, Field: "images", Message: "URL изображения должен начинаться с http:// или https://"})
				continue
			}
			images = append(images, imgURL)
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}

	return &CreateBathhouseInput{
		Name:         name,
		Description:  description,
		Address:      address,
		CityID:       cityID,
		Latitude:     lat,
		Longitude:    lon,
		PricePerHour: priceRub * 100, // rubles → kopecks
		MinDuration:  minDuration,
		MaxGuests:    maxGuests,
		HasPool:      parseBool("has_pool"),
		HasSauna:     parseBool("has_sauna"),
		HasSteamRoom: parseBool("has_steam_room"),
		HasHotTub:    parseBool("has_hot_tub"),
		HasBBQ:       parseBool("has_bbq"),
		HasKaraoke:   parseBool("has_karaoke"),
		Images:       images,
	}, nil
}

// CSVTemplate returns the CSV header line for the import template.
func CSVTemplate() string {
	return strings.Join(csvColumns, ",") + "\n" +
		"Баня у Петра,Лучшая баня в городе,ул. Ленина 10,1,55.7558,37.6173,1500,2,10,1,1,0,0,1,0,https://example.com/photo1.jpg;https://example.com/photo2.jpg\n"
}

// ValidateImportRow checks a single parsed row without creating a listing (for dry-run).
func ValidateImportRow(record []string, colIndex map[string]int, rowNum int) []ImportError {
	_, errs := parseRow(record, colIndex, rowNum)
	return errs
}
