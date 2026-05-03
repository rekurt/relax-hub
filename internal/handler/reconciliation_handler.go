package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/service"
)

type ReconciliationHandler struct {
	reconciliationService service.ReconciliationService
}

func NewReconciliationHandler(svc service.ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{reconciliationService: svc}
}

type floatSnapshotResponse struct {
	ID                 string `json:"id"`
	ClientWalletsTotal int64  `json:"client_wallets_total"`
	OwnerWalletsTotal  int64  `json:"owner_wallets_total"`
	EscrowHeldTotal    int64  `json:"escrow_held_total"`
	WalletHoldsTotal   int64  `json:"wallet_holds_total"`
	ExpectedTotal      int64  `json:"expected_total"`
	ActualTotal        int64  `json:"actual_total"`
	Discrepancy        int64  `json:"discrepancy"`
	Status             string `json:"status"`
	ClientWalletsCount int    `json:"client_wallets_count"`
	OwnerWalletsCount  int    `json:"owner_wallets_count"`
	EscrowCount        int    `json:"escrow_count"`
	Notes              string `json:"notes,omitempty"`
	SnapshotDate       string `json:"snapshot_date"`
	CreatedAt          string `json:"created_at"`
}

type reconciliationReportResponse struct {
	ID                    string `json:"id"`
	PeriodStart           string `json:"period_start"`
	PeriodEnd             string `json:"period_end"`
	InternalPaymentsSum   int64  `json:"internal_payments_sum"`
	InternalPaymentsCount int    `json:"internal_payments_count"`
	InternalRefundsSum    int64  `json:"internal_refunds_sum"`
	InternalRefundsCount  int    `json:"internal_refunds_count"`
	ProviderPaymentsSum   int64  `json:"provider_payments_sum"`
	ProviderPaymentsCount int    `json:"provider_payments_count"`
	ProviderRefundsSum    int64  `json:"provider_refunds_sum"`
	ProviderRefundsCount  int    `json:"provider_refunds_count"`
	PaymentDiscrepancy    int64  `json:"payment_discrepancy"`
	RefundDiscrepancy     int64  `json:"refund_discrepancy"`
	Status                string `json:"status"`
	MismatchDetails       string `json:"mismatch_details,omitempty"`
	ErrorMessage          string `json:"error_message,omitempty"`
	CreatedAt             string `json:"created_at"`
}

type floatSummaryResponse struct {
	ClientWalletsTotal int64                         `json:"client_wallets_total"`
	ClientWalletsCount int                           `json:"client_wallets_count"`
	OwnerWalletsTotal  int64                         `json:"owner_wallets_total"`
	OwnerWalletsCount  int                           `json:"owner_wallets_count"`
	EscrowHeldTotal    int64                         `json:"escrow_held_total"`
	EscrowCount        int                           `json:"escrow_count"`
	WalletHoldsTotal   int64                         `json:"wallet_holds_total"`
	PlatformTotal      int64                         `json:"platform_total"`
	LastSnapshot       *floatSnapshotResponse        `json:"last_snapshot,omitempty"`
	LastReport         *reconciliationReportResponse `json:"last_report,omitempty"`
}

// GetFloatSummary returns current float status for the admin dashboard.
//
//	@Summary	Get float summary
//	@Tags		admin,reconciliation
//	@Security	BearerAuth
//	@Success	200	{object}	APIResponse{data=floatSummaryResponse}
//	@Router		/api/v1/admin/reconciliation/summary [get]
func (h *ReconciliationHandler) GetFloatSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.reconciliationService.GetFloatSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Ошибка получения данных")
		return
	}

	resp := floatSummaryResponse{
		ClientWalletsTotal: summary.ClientWalletsTotal,
		ClientWalletsCount: summary.ClientWalletsCount,
		OwnerWalletsTotal:  summary.OwnerWalletsTotal,
		OwnerWalletsCount:  summary.OwnerWalletsCount,
		EscrowHeldTotal:    summary.EscrowHeldTotal,
		EscrowCount:        summary.EscrowCount,
		WalletHoldsTotal:   summary.WalletHoldsTotal,
		PlatformTotal:      summary.PlatformTotal,
	}
	if summary.LastSnapshot != nil {
		resp.LastSnapshot = toFloatSnapshotResponse(summary.LastSnapshot)
	}
	if summary.LastReport != nil {
		resp.LastReport = toReconciliationReportResponse(summary.LastReport)
	}

	writeJSON(w, http.StatusOK, resp)
}

// TakeSnapshot triggers a manual float snapshot.
//
//	@Summary	Take float snapshot
//	@Tags		admin,reconciliation
//	@Security	BearerAuth
//	@Success	200	{object}	APIResponse{data=floatSnapshotResponse}
//	@Router		/api/v1/admin/reconciliation/snapshot [post]
func (h *ReconciliationHandler) TakeSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.reconciliationService.TakeFloatSnapshot(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Ошибка создания снимка")
		return
	}
	writeJSON(w, http.StatusOK, toFloatSnapshotResponse(snapshot))
}

// ListSnapshots returns paginated float snapshots.
//
//	@Summary	List float snapshots
//	@Tags		admin,reconciliation
//	@Security	BearerAuth
//	@Param		date_from	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param		date_to		query		string	false	"End date (YYYY-MM-DD)"
//	@Param		page		query		int		false	"Page number"
//	@Param		page_size	query		int		false	"Page size"
//	@Success	200			{object}	APIResponse{data=[]floatSnapshotResponse}
//	@Router		/api/v1/admin/reconciliation/snapshots [get]
func (h *ReconciliationHandler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	from, to := reconciliationParseDateRange(r)
	page, pageSize := reconciliationParsePagination(r)

	result, err := h.reconciliationService.ListSnapshots(r.Context(), from, to, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Ошибка получения снимков")
		return
	}

	items := make([]floatSnapshotResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, *toFloatSnapshotResponse(&result.Items[i]))
	}

	totalPages := int(result.TotalCount) / pageSize
	if int(result.TotalCount)%pageSize > 0 {
		totalPages++
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: result.TotalCount,
		TotalPages: totalPages,
	})
}

// Reconcile triggers a manual reconciliation for a date period.
//
//	@Summary	Reconcile with provider
//	@Tags		admin,reconciliation
//	@Security	BearerAuth
//	@Param		date_from	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param		date_to		query		string	false	"End date (YYYY-MM-DD)"
//	@Success	200			{object}	APIResponse{data=reconciliationReportResponse}
//	@Router		/api/v1/admin/reconciliation/reconcile [post]
func (h *ReconciliationHandler) Reconcile(w http.ResponseWriter, r *http.Request) {
	from, to := reconciliationParseDateRange(r)

	report, err := h.reconciliationService.ReconcileWithProvider(r.Context(), from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Ошибка сверки")
		return
	}
	writeJSON(w, http.StatusOK, toReconciliationReportResponse(report))
}

// ListReports returns paginated reconciliation reports.
//
//	@Summary	List reconciliation reports
//	@Tags		admin,reconciliation
//	@Security	BearerAuth
//	@Param		page		query		int	false	"Page number"
//	@Param		page_size	query		int	false	"Page size"
//	@Success	200			{object}	APIResponse{data=[]reconciliationReportResponse}
//	@Router		/api/v1/admin/reconciliation/reports [get]
func (h *ReconciliationHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	page, pageSize := reconciliationParsePagination(r)

	result, err := h.reconciliationService.ListReports(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Ошибка получения отчётов")
		return
	}

	items := make([]reconciliationReportResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, *toReconciliationReportResponse(&result.Items[i]))
	}

	totalPages := int(result.TotalCount) / pageSize
	if int(result.TotalCount)%pageSize > 0 {
		totalPages++
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: result.TotalCount,
		TotalPages: totalPages,
	})
}

func reconciliationParseDateRange(r *http.Request) (from, to time.Time) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	if dateFrom != "" {
		from, _ = time.Parse("2006-01-02", dateFrom)
	}
	if from.IsZero() {
		from = time.Now().AddDate(0, 0, -30)
	}

	if dateTo != "" {
		to, _ = time.Parse("2006-01-02", dateTo)
	}
	if to.IsZero() {
		to = time.Now()
	}
	return
}

func reconciliationParsePagination(r *http.Request) (page, pageSize int) {
	page = 1
	pageSize = 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	return
}

func toFloatSnapshotResponse(s *domain.FloatSnapshot) *floatSnapshotResponse {
	return &floatSnapshotResponse{
		ID:                 s.ID.String(),
		ClientWalletsTotal: s.ClientWalletsTotal,
		OwnerWalletsTotal:  s.OwnerWalletsTotal,
		EscrowHeldTotal:    s.EscrowHeldTotal,
		WalletHoldsTotal:   s.WalletHoldsTotal,
		ExpectedTotal:      s.ExpectedTotal,
		ActualTotal:        s.ActualTotal,
		Discrepancy:        s.Discrepancy,
		Status:             string(s.Status),
		ClientWalletsCount: s.ClientWalletsCount,
		OwnerWalletsCount:  s.OwnerWalletsCount,
		EscrowCount:        s.EscrowCount,
		Notes:              s.Notes,
		SnapshotDate:       s.SnapshotDate.Format("2006-01-02"),
		CreatedAt:          s.CreatedAt.Format(time.RFC3339),
	}
}

func toReconciliationReportResponse(r *domain.ReconciliationReport) *reconciliationReportResponse {
	return &reconciliationReportResponse{
		ID:                    r.ID.String(),
		PeriodStart:           r.PeriodStart.Format(time.RFC3339),
		PeriodEnd:             r.PeriodEnd.Format(time.RFC3339),
		InternalPaymentsSum:   r.InternalPaymentsSum,
		InternalPaymentsCount: r.InternalPaymentsCount,
		InternalRefundsSum:    r.InternalRefundsSum,
		InternalRefundsCount:  r.InternalRefundsCount,
		ProviderPaymentsSum:   r.ProviderPaymentsSum,
		ProviderPaymentsCount: r.ProviderPaymentsCount,
		ProviderRefundsSum:    r.ProviderRefundsSum,
		ProviderRefundsCount:  r.ProviderRefundsCount,
		PaymentDiscrepancy:    r.PaymentDiscrepancy,
		RefundDiscrepancy:     r.RefundDiscrepancy,
		Status:                string(r.Status),
		MismatchDetails:       r.MismatchDetails,
		ErrorMessage:          r.ErrorMessage,
		CreatedAt:             r.CreatedAt.Format(time.RFC3339),
	}
}
