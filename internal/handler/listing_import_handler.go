package handler

import (
	"net/http"

	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ListingImportHandler struct {
	importSvc service.ListingImportService
}

func NewListingImportHandler(importSvc service.ListingImportService) *ListingImportHandler {
	return &ListingImportHandler{importSvc: importSvc}
}

// ImportCSV godoc
// @Summary      Import listings from CSV
// @Description  Upload a CSV file to create multiple bathhouse listings as pending drafts
// @Tags         listings
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "CSV file"
// @Success      200 {object} APIResponse{data=service.ImportReport}
// @Failure      400 {object} APIResponse
// @Security     BearerAuth
// @Router       /my/listings/import [post]
func (h *ListingImportHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	// Limit upload to 10 MB to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file_required", "Загрузите CSV файл")
		return
	}
	defer file.Close()

	report, err := h.importSvc.ImportCSV(r.Context(), userID, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "import_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

// GetImportTemplate godoc
// @Summary      Get CSV import template
// @Description  Download a CSV template with header and example row for listing import
// @Tags         listings
// @Produce      text/csv
// @Success      200 {string} string
// @Security     BearerAuth
// @Router       /my/listings/import/template [get]
func (h *ListingImportHandler) GetImportTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=listing_import_template.csv")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM for Excel
	w.Write([]byte(service.CSVTemplate()))
}
