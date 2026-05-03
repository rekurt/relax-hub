package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type BankReconciliationHandler struct {
	bankReconciliationService service.BankReconciliationService
}

func NewBankReconciliationHandler(svc service.BankReconciliationService) *BankReconciliationHandler {
	return &BankReconciliationHandler{bankReconciliationService: svc}
}

type bankStatementEntryResponse struct {
	ID            string  `json:"id"`
	Date          string  `json:"date"`
	Amount        int64   `json:"amount"`
	Description   string  `json:"description"`
	Counterparty  string  `json:"counterparty"`
	ReferenceNum  string  `json:"reference_num,omitempty"`
	MatchedTxID   *string `json:"matched_tx_id,omitempty"`
	MatchedTxType string  `json:"matched_tx_type,omitempty"`
	Status        string  `json:"status"`
	UploadBatchID string  `json:"upload_batch_id"`
	CreatedAt     string  `json:"created_at"`
}

type bankStatementUploadResponse struct {
	ID           string `json:"id"`
	FileName     string `json:"file_name"`
	Format       string `json:"format"`
	TotalRows    int    `json:"total_rows"`
	MatchedCount int    `json:"matched_count"`
	PendingCount int    `json:"pending_count"`
	IgnoredCount int    `json:"ignored_count"`
	CreatedAt    string `json:"created_at"`
}

// UploadBankStatement handles bank statement file upload and auto-matching.
//
//	@Summary	Upload bank statement
//	@Tags		admin,finance
//	@Security	BearerAuth
//	@Accept		multipart/form-data
//	@Param		file	formData	file	true	"Bank statement file (CSV or 1C format)"
//	@Success	200		{object}	APIResponse{data=bankStatementUploadResponse}
//	@Failure	400		{object}	APIResponse
//	@Router		/api/v1/admin/finance/bank-statement [post]
func (h *BankReconciliationHandler) UploadBankStatement(w http.ResponseWriter, r *http.Request) {
	adminID := middleware.GetUserID(r.Context())

	const maxSize = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+1024)

	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_file", "Не удалось прочитать multipart-форму")
		return
	}

	var fileName string
	var format string
	var fileData bytes.Buffer

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_file", "Не удалось прочитать файл")
			return
		}

		switch part.FormName() {
		case "file":
			fileName = filepath.Base(part.FileName())
			if fileName == "." || fileName == string(filepath.Separator) || strings.TrimSpace(fileName) == "" {
				writeError(w, http.StatusBadRequest, "invalid_file", "Не удалось прочитать файл")
				return
			}
			written, err := io.Copy(&fileData, io.LimitReader(part, maxSize+1))
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_file", "Не удалось прочитать файл")
				return
			}
			if written > maxSize {
				writeError(w, http.StatusBadRequest, "file_too_large", "Файл слишком большой (максимум 10 МБ)")
				return
			}
		case "format":
			rawFormat, err := io.ReadAll(io.LimitReader(part, 64))
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_format", "Не удалось прочитать формат файла")
				return
			}
			format = strings.TrimSpace(string(rawFormat))
		}
	}

	if fileName == "" || fileData.Len() == 0 {
		writeError(w, http.StatusBadRequest, "invalid_file", "Не удалось прочитать файл")
		return
	}

	// Detect format from extension or form field
	if format == "" {
		ext := strings.ToLower(filepath.Ext(fileName))
		switch ext {
		case ".csv":
			format = "csv"
		case ".txt", ".1c":
			format = "1c"
		default:
			writeError(w, http.StatusBadRequest, "invalid_format", "Неподдерживаемый формат файла. Допустимые: csv, 1c (txt)")
			return
		}
	}

	upload, err := h.bankReconciliationService.UploadStatement(r.Context(), adminID, fileName, format, bytes.NewReader(fileData.Bytes()))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBankUploadResponse(upload))
}

// ListUnmatched returns unmatched bank statement entries for manual reconciliation.
//
//	@Summary	List unmatched bank entries
//	@Tags		admin,finance
//	@Security	BearerAuth
//	@Param		status		query		string	false	"Filter by status (pending, matched, manual, ignored)"
//	@Param		date_from	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param		date_to		query		string	false	"End date (YYYY-MM-DD)"
//	@Param		page		query		int		false	"Page number"
//	@Param		page_size	query		int		false	"Page size"
//	@Success	200			{object}	APIResponse{data=[]bankStatementEntryResponse}
//	@Router		/api/v1/admin/finance/reconciliation [get]
func (h *BankReconciliationHandler) ListUnmatched(w http.ResponseWriter, r *http.Request) {
	filter := domain.BankStatementFilter{
		Page:     getPage(r.URL.Query().Get("page")),
		PageSize: getPageSize(r.URL.Query().Get("page_size"), 20),
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.BankStatementEntryStatus(status)
		if !s.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_status", "Недопустимый статус")
			return
		}
		filter.Status = &s
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &t
		}
	}

	result, err := h.bankReconciliationService.ListUnmatched(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Ошибка получения данных")
		return
	}

	items := make([]bankStatementEntryResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, toBankEntryResponse(&result.Items[i]))
	}

	pageSize := filter.PageSize
	totalPages := int(result.TotalCount) / pageSize
	if int(result.TotalCount)%pageSize > 0 {
		totalPages++
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalCount: result.TotalCount,
		TotalPages: totalPages,
	})
}

type manualMatchRequest struct {
	TransactionID string `json:"transaction_id"`
	TxType        string `json:"tx_type"` // "payment" or "wallet_transaction"
}

// ManualMatch manually matches a bank entry to an internal transaction.
//
//	@Summary	Manual match bank entry
//	@Tags		admin,finance
//	@Security	BearerAuth
//	@Param		id		path		string				true	"Bank entry ID"
//	@Param		body	body		manualMatchRequest	true	"Match details"
//	@Success	200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure	400		{object}	APIResponse
//	@Failure	404		{object}	APIResponse
//	@Router		/api/v1/admin/finance/reconciliation/{id}/match [put]
func (h *BankReconciliationHandler) ManualMatch(w http.ResponseWriter, r *http.Request) {
	entryIDStr := chi.URLParam(r, "id")
	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Неверный формат ID записи")
		return
	}

	var req manualMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Неверный формат запроса")
		return
	}

	txID, err := uuid.Parse(req.TransactionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_tx_id", "Неверный формат ID транзакции")
		return
	}

	if err := h.bankReconciliationService.ManualMatch(r.Context(), entryID, txID, req.TxType); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "Запись успешно сопоставлена"})
}

func toBankEntryResponse(e *domain.BankStatementEntry) bankStatementEntryResponse {
	resp := bankStatementEntryResponse{
		ID:            e.ID.String(),
		Date:          e.Date.Format("2006-01-02"),
		Amount:        e.Amount,
		Description:   e.Description,
		Counterparty:  e.Counterparty,
		ReferenceNum:  e.ReferenceNum,
		MatchedTxType: e.MatchedTxType,
		Status:        string(e.Status),
		UploadBatchID: e.UploadBatchID.String(),
		CreatedAt:     e.CreatedAt.Format(time.RFC3339),
	}
	if e.MatchedTxID != nil {
		s := e.MatchedTxID.String()
		resp.MatchedTxID = &s
	}
	return resp
}

func toBankUploadResponse(u *domain.BankStatementUpload) bankStatementUploadResponse {
	return bankStatementUploadResponse{
		ID:           u.ID.String(),
		FileName:     u.FileName,
		Format:       u.Format,
		TotalRows:    u.TotalRows,
		MatchedCount: u.MatchedCount,
		PendingCount: u.PendingCount,
		IgnoredCount: u.IgnoredCount,
		CreatedAt:    u.CreatedAt.Format(time.RFC3339),
	}
}
