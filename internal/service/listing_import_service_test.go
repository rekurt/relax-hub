package service

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

type stubBathhouseService struct {
	created []CreateBathhouseInput
	err     error
}

func (s *stubBathhouseService) Create(_ context.Context, _ uuid.UUID, input CreateBathhouseInput) (*domain.Bathhouse, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.created = append(s.created, input)
	return &domain.Bathhouse{ID: uuid.New(), Name: input.Name}, nil
}

func (s *stubBathhouseService) GetByID(context.Context, uuid.UUID) (*domain.Bathhouse, error)   { return nil, nil }
func (s *stubBathhouseService) GetBySlug(context.Context, string) (*domain.Bathhouse, error)     { return nil, nil }
func (s *stubBathhouseService) GetByAPIKey(context.Context, string) (*domain.Bathhouse, error)   { return nil, nil }
func (s *stubBathhouseService) Update(context.Context, uuid.UUID, domain.UserRole, uuid.UUID, UpdateBathhouseInput) (*domain.Bathhouse, error) {
	return nil, nil
}
func (s *stubBathhouseService) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s *stubBathhouseService) Search(context.Context, domain.BathhouseFilter) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}
func (s *stubBathhouseService) ListByOwner(context.Context, uuid.UUID, int, int) (*domain.PaginatedResult[domain.Bathhouse], error) {
	return nil, nil
}
func (s *stubBathhouseService) GetWidgetKey(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) (string, error) {
	return "", nil
}
func (s *stubBathhouseService) RegenerateWidgetKey(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) (string, error) {
	return "", nil
}
func (s *stubBathhouseService) CheckCompleteness(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) (*CompletenessResult, error) {
	return nil, nil
}
func (s *stubBathhouseService) SubmitForModeration(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) error {
	return nil
}
func (s *stubBathhouseService) DuplicateBathhouse(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) (*domain.Bathhouse, error) {
	return nil, nil
}
func (s *stubBathhouseService) DeactivateBathhouse(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) error {
	return nil
}
func (s *stubBathhouseService) ActivateBathhouse(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) error {
	return nil
}
func (s *stubBathhouseService) ArchiveBathhouse(context.Context, uuid.UUID, domain.UserRole, uuid.UUID) error {
	return nil
}
func (s *stubBathhouseService) IncrementViewCount(context.Context, uuid.UUID) error { return nil }
func (s *stubBathhouseService) ComputeBadges(context.Context, *domain.Bathhouse) []string {
	return nil
}
func (s *stubBathhouseService) Approve(context.Context, uuid.UUID) error { return nil }
func (s *stubBathhouseService) Reject(context.Context, uuid.UUID) error  { return nil }

func newTestImportService(bhSvc *stubBathhouseService) ListingImportService {
	return NewListingImportService(bhSvc, logger.New(logger.LevelInfo))
}

func TestImportCSV_Success(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	csv := "name,description,address,city_id,latitude,longitude,price_per_hour_rub,min_duration,max_guests,has_pool,has_sauna,has_steam_room,has_hot_tub,has_bbq,has_karaoke,images\n" +
		"Баня Люкс,Отличная баня,ул. Пушкина 5,1,55.75,37.62,1500,2,10,1,1,0,0,1,0,https://example.com/img1.jpg;https://example.com/img2.jpg\n" +
		"Баня Эконом,,ул. Мира 3,2,55.80,37.70,800,1,6,0,1,1,0,0,0,\n"

	report, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 2, report.TotalRows)
	assert.Equal(t, 2, report.SuccessCount)
	assert.Equal(t, 0, report.ErrorCount)
	assert.Len(t, report.CreatedIDs, 2)

	// Verify first row conversion
	assert.Equal(t, "Баня Люкс", bhSvc.created[0].Name)
	assert.Equal(t, int64(150000), bhSvc.created[0].PricePerHour) // 1500 rub -> 150000 kopecks
	assert.True(t, bhSvc.created[0].HasPool)
	assert.True(t, bhSvc.created[0].HasSauna)
	assert.False(t, bhSvc.created[0].HasSteamRoom)
	assert.Len(t, bhSvc.created[0].Images, 2)

	// Verify second row
	assert.Equal(t, "Баня Эконом", bhSvc.created[1].Name)
	assert.Equal(t, int64(80000), bhSvc.created[1].PricePerHour)
	assert.Empty(t, bhSvc.created[1].Images)
}

func TestImportCSV_MissingRequiredField(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	csv := "name,address,city_id,latitude,longitude,price_per_hour_rub,min_duration,max_guests\n" +
		",ул. Пушкина 5,1,55.75,37.62,1500,2,10\n" // empty name

	report, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, report.TotalRows)
	assert.Equal(t, 0, report.SuccessCount)
	assert.Equal(t, 1, report.ErrorCount)
	assert.Equal(t, "name", report.Errors[0].Field)
}

func TestImportCSV_InvalidNumber(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	csv := "name,address,city_id,latitude,longitude,price_per_hour_rub,min_duration,max_guests\n" +
		"Баня,ул. Ленина 1,abc,55.75,37.62,1500,2,10\n" // city_id not a number

	report, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, report.ErrorCount)
	assert.Equal(t, "city_id", report.Errors[0].Field)
}

func TestImportCSV_MissingHeader(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	csv := "name,address\nБаня,ул. Ленина\n" // missing required columns

	_, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "обязательные столбцы")
}

func TestImportCSV_EmptyFile(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	csv := "name,address,city_id,latitude,longitude,price_per_hour_rub,min_duration,max_guests\n"

	report, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 0, report.TotalRows)
	assert.Equal(t, 0, report.SuccessCount)
}

func TestImportCSV_InvalidLatitude(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	csv := "name,address,city_id,latitude,longitude,price_per_hour_rub,min_duration,max_guests\n" +
		"Баня,ул. Ленина 1,1,200,37.62,1500,2,10\n" // latitude out of range

	report, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, report.ErrorCount)
	assert.Equal(t, "latitude", report.Errors[0].Field)
}

func TestImportCSV_NegativePrice(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	csv := "name,address,city_id,latitude,longitude,price_per_hour_rub,min_duration,max_guests\n" +
		"Баня,ул. Ленина 1,1,55.75,37.62,-100,2,10\n"

	report, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, report.ErrorCount)
	assert.Equal(t, "price_per_hour_rub", report.Errors[0].Field)
}

func TestImportCSV_ServiceError(t *testing.T) {
	bhSvc := &stubBathhouseService{err: domain.ErrForbidden}
	svc := newTestImportService(bhSvc)

	csv := "name,address,city_id,latitude,longitude,price_per_hour_rub,min_duration,max_guests\n" +
		"Баня,ул. Ленина 1,1,55.75,37.62,1500,2,10\n"

	report, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Equal(t, 1, report.TotalRows)
	assert.Equal(t, 0, report.SuccessCount)
	assert.Equal(t, 1, report.ErrorCount)
	assert.Contains(t, report.Errors[0].Message, "ошибка создания")
}

func TestCSVTemplate_HasHeaderAndExample(t *testing.T) {
	tmpl := CSVTemplate()
	lines := strings.Split(strings.TrimSpace(tmpl), "\n")
	assert.GreaterOrEqual(t, len(lines), 2)
	assert.Contains(t, lines[0], "name")
	assert.Contains(t, lines[0], "price_per_hour_rub")
	assert.Contains(t, lines[1], "Баня у Петра")
}

// --- XLSX import tests ---

// createTestXLSX builds an in-memory XLSX with given header and rows.
func createTestXLSX(t *testing.T, header []string, rows [][]interface{}) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	for i, col := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, col)
	}
	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}
	buf, err := f.WriteToBuffer()
	require.NoError(t, err)
	return buf
}

func TestImportXLSX_Success(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	header := []string{"name", "description", "address", "city_id", "latitude", "longitude",
		"price_per_hour_rub", "min_duration", "max_guests", "has_pool", "has_sauna",
		"has_steam_room", "has_hot_tub", "has_bbq", "has_karaoke", "images"}
	rows := [][]interface{}{
		{"Баня Люкс", "Отличная баня", "ул. Пушкина 5", 1, 55.75, 37.62, 1500, 2, 10, 1, 1, 0, 0, 1, 0, "https://example.com/img1.jpg;https://example.com/img2.jpg"},
		{"Баня Эконом", "", "ул. Мира 3", 2, 55.80, 37.70, 800, 1, 6, 0, 1, 1, 0, 0, 0, ""},
	}

	buf := createTestXLSX(t, header, rows)
	report, err := svc.ImportXLSX(context.Background(), uuid.New(), buf)
	require.NoError(t, err)
	assert.Equal(t, 2, report.TotalRows)
	assert.Equal(t, 2, report.SuccessCount)
	assert.Equal(t, 0, report.ErrorCount)
	assert.Len(t, report.CreatedIDs, 2)

	assert.Equal(t, "Баня Люкс", bhSvc.created[0].Name)
	assert.Equal(t, int64(150000), bhSvc.created[0].PricePerHour)
	assert.True(t, bhSvc.created[0].HasPool)
	assert.Len(t, bhSvc.created[0].Images, 2)

	assert.Equal(t, "Баня Эконом", bhSvc.created[1].Name)
	assert.Equal(t, int64(80000), bhSvc.created[1].PricePerHour)
}

func TestImportXLSX_MissingHeader(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	header := []string{"name", "address"} // missing required columns
	rows := [][]interface{}{
		{"Баня", "ул. Ленина"},
	}

	buf := createTestXLSX(t, header, rows)
	_, err := svc.ImportXLSX(context.Background(), uuid.New(), buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "обязательные столбцы")
}

func TestImportXLSX_InvalidData(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	header := []string{"name", "address", "city_id", "latitude", "longitude",
		"price_per_hour_rub", "min_duration", "max_guests"}
	rows := [][]interface{}{
		{"", "ул. Пушкина 5", 1, 55.75, 37.62, 1500, 2, 10}, // empty name
	}

	buf := createTestXLSX(t, header, rows)
	report, err := svc.ImportXLSX(context.Background(), uuid.New(), buf)
	require.NoError(t, err)
	assert.Equal(t, 1, report.ErrorCount)
	assert.Equal(t, "name", report.Errors[0].Field)
}

func TestImportXLSX_EmptyFile(t *testing.T) {
	bhSvc := &stubBathhouseService{}
	svc := newTestImportService(bhSvc)

	header := []string{"name", "address", "city_id", "latitude", "longitude",
		"price_per_hour_rub", "min_duration", "max_guests"}

	buf := createTestXLSX(t, header, nil) // no data rows
	report, err := svc.ImportXLSX(context.Background(), uuid.New(), buf)
	require.NoError(t, err)
	assert.Equal(t, 0, report.TotalRows)
}

func TestImportXLSX_ServiceError(t *testing.T) {
	bhSvc := &stubBathhouseService{err: domain.ErrForbidden}
	svc := newTestImportService(bhSvc)

	header := []string{"name", "address", "city_id", "latitude", "longitude",
		"price_per_hour_rub", "min_duration", "max_guests"}
	rows := [][]interface{}{
		{"Баня", "ул. Ленина 1", 1, 55.75, 37.62, 1500, 2, 10},
	}

	buf := createTestXLSX(t, header, rows)
	report, err := svc.ImportXLSX(context.Background(), uuid.New(), buf)
	require.NoError(t, err)
	assert.Equal(t, 1, report.ErrorCount)
	assert.Contains(t, report.Errors[0].Message, "ошибка создания")
}

func TestXLSXTemplate_Generates(t *testing.T) {
	buf, err := XLSXTemplate()
	require.NoError(t, err)
	require.NotNil(t, buf)
	assert.True(t, buf.Len() > 0)

	// Verify it's a valid XLSX that can be opened
	f, err := excelize.OpenReader(buf)
	require.NoError(t, err)
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rows), 2) // header + sample

	// Check header contains expected columns
	assert.Equal(t, "name", rows[0][0])
	assert.Contains(t, rows[0], "price_per_hour_rub")
}
