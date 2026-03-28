package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type FinancialReportHandler struct {
	reportService service.FinancialReportService
}

func NewFinancialReportHandler(reportService service.FinancialReportService) *FinancialReportHandler {
	return &FinancialReportHandler{reportService: reportService}
}

// ExportWalletTransactions godoc
//
//	@Summary		Export wallet transactions
//	@Description	Export wallet transaction history in CSV or PDF format
//	@Tags			wallet
//	@Produce		application/csv,application/pdf
//	@Security		BearerAuth
//	@Param			format		query		string	true	"Export format: csv or pdf"
//	@Param			date_from	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param			date_to		query		string	false	"End date (YYYY-MM-DD)"
//	@Success		200			{file}		file
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/wallet/export [get]
func (h *FinancialReportHandler) ExportWalletTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	format := r.URL.Query().Get("format")

	if format != "csv" && format != "pdf" {
		writeError(w, http.StatusBadRequest, "invalid_input", "format must be csv or pdf")
		return
	}

	dateFrom, dateTo, err := parseDateRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}

	switch format {
	case "csv":
		data, err := h.reportService.ExportWalletTransactionsCSV(r.Context(), userID, dateFrom, dateTo)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=wallet_transactions.csv")
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	case "pdf":
		data, err := h.reportService.ExportWalletTransactionsPDF(r.Context(), userID, dateFrom, dateTo)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=wallet_transactions.pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// GenerateAct godoc
//
//	@Summary		Generate act for owner
//	@Description	Generate act of services rendered (PDF) for a specific bathhouse and date range
//	@Tags			finance
//	@Produce		application/pdf
//	@Security		BearerAuth
//	@Param			bathhouse_id	path		string	true	"Bathhouse ID"
//	@Param			date_from		query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			date_to			query		string	true	"End date (YYYY-MM-DD)"
//	@Success		200				{file}		file
//	@Failure		400				{object}	APIResponse{error=APIError}
//	@Failure		401				{object}	APIResponse{error=APIError}
//	@Failure		403				{object}	APIResponse{error=APIError}
//	@Router			/my/finance/acts/{bathhouse_id} [get]
func (h *FinancialReportHandler) GenerateAct(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	bathhouseIDStr := chi.URLParam(r, "bathhouse_id")
	bathhouseID, err := uuid.Parse(bathhouseIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse_id")
		return
	}

	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")

	if dateFromStr == "" || dateToStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date_from and date_to are required")
		return
	}

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from format, use YYYY-MM-DD")
		return
	}
	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to format, use YYYY-MM-DD")
		return
	}
	dateTo = dateTo.Add(24*time.Hour - time.Second)

	data, err := h.reportService.GenerateOwnerAct(r.Context(), userID, bathhouseID, dateFrom, dateTo)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	filename := fmt.Sprintf("act_%s_%s_%s.pdf", bathhouseIDStr[:8], dateFromStr, dateToStr)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ExportXML1C godoc
//
//	@Summary		Export financial data in 1C XML format
//	@Description	Export financial data for legal entities in 1C-compatible XML format
//	@Tags			finance
//	@Produce		application/xml
//	@Security		BearerAuth
//	@Param			date_from	query		string	true	"Start date (YYYY-MM-DD)"
//	@Param			date_to		query		string	true	"End date (YYYY-MM-DD)"
//	@Success		200			{file}		file
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/finance/export-xml [get]
func (h *FinancialReportHandler) ExportXML1C(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")

	if dateFromStr == "" || dateToStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date_from and date_to are required")
		return
	}

	dateFrom, err := time.Parse("2006-01-02", dateFromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from format, use YYYY-MM-DD")
		return
	}
	dateTo, err := time.Parse("2006-01-02", dateToStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to format, use YYYY-MM-DD")
		return
	}
	dateTo = dateTo.Add(24*time.Hour - time.Second)

	data, err := h.reportService.ExportXML1C(r.Context(), userID, dateFrom, dateTo)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	filename := fmt.Sprintf("finance_%s_%s.xml", dateFromStr, dateToStr)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func parseDateRange(r *http.Request) (*time.Time, *time.Time, error) {
	var dateFrom, dateTo *time.Time

	if s := r.URL.Query().Get("date_from"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid date_from format, use YYYY-MM-DD")
		}
		dateFrom = &t
	}

	if s := r.URL.Query().Get("date_to"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid date_to format, use YYYY-MM-DD")
		}
		end := t.Add(24*time.Hour - time.Second)
		dateTo = &end
	}

	return dateFrom, dateTo, nil
}
