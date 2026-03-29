package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ListingImportHandler struct {
	importSvc service.ListingImportService
}

func NewListingImportHandler(importSvc service.ListingImportService) *ListingImportHandler {
	return &ListingImportHandler{importSvc: importSvc}
}

// Import godoc
// @Summary      Import listings from CSV or XLSX
// @Description  Upload a CSV or XLSX file to create multiple bathhouse listings as pending drafts
// @Tags         listings
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "CSV or XLSX file"
// @Success      200 {object} APIResponse{data=service.ImportReport}
// @Failure      400 {object} APIResponse
// @Security     BearerAuth
// @Router       /my/listings/import [post]
func (h *ListingImportHandler) Import(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	// Limit upload to 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file_required", "Загрузите CSV или XLSX файл")
		return
	}
	defer file.Close()

	format := detectImportFormat(header.Filename, header.Header.Get("Content-Type"))

	var report *service.ImportReport

	switch format {
	case "xlsx":
		report, err = h.importSvc.ImportXLSX(r.Context(), userID, file)
	case "csv":
		report, err = h.importSvc.ImportCSV(r.Context(), userID, file)
	default:
		writeError(w, http.StatusBadRequest, "invalid_format", "Поддерживаются форматы CSV и XLSX")
		return
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, "import_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// GetImportTemplate godoc
// @Summary      Get import template (CSV or XLSX)
// @Description  Download a CSV or XLSX template with header and example row for listing import
// @Tags         listings
// @Produce      application/octet-stream
// @Param        format query string false "Template format: csv (default) or xlsx" Enums(csv, xlsx)
// @Success      200 {file} file
// @Security     BearerAuth
// @Router       /my/listings/import/template [get]
func (h *ListingImportHandler) GetImportTemplate(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

	switch format {
	case "xlsx":
		buf, err := service.XLSXTemplate()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "template_error", "Ошибка генерации шаблона")
			return
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=listing_import_template.xlsx")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	default:
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=listing_import_template.csv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM for Excel
		w.Write([]byte(service.CSVTemplate()))
	}
}

// detectImportFormat determines file format from filename extension or content type.
func detectImportFormat(filename, contentType string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx":
		return "xlsx"
	case ".csv":
		return "csv"
	}

	// Fallback to content type
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "spreadsheetml") || strings.Contains(ct, "xlsx") {
		return "xlsx"
	}
	if strings.Contains(ct, "csv") || strings.Contains(ct, "text/plain") {
		return "csv"
	}

	return ""
}
