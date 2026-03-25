package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type AuditLogHandler struct {
	auditSvc    service.AuditLogService
	bhSvc       service.BathhouseService
	accessCheck *service.AccessChecker
}

func NewAuditLogHandler(auditSvc service.AuditLogService, bhSvc service.BathhouseService, accessCheck *service.AccessChecker) *AuditLogHandler {
	return &AuditLogHandler{auditSvc: auditSvc, bhSvc: bhSvc, accessCheck: accessCheck}
}

type auditLogResponse struct {
	ID            string          `json:"id"`
	EntityType    string          `json:"entity_type"`
	EntityID      string          `json:"entity_id"`
	UserID        string          `json:"user_id"`
	Action        string          `json:"action"`
	ChangedFields json.RawMessage `json:"changed_fields,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

func toAuditLogResponse(a *domain.AuditLog) auditLogResponse {
	return auditLogResponse{
		ID:            a.ID.String(),
		EntityType:    a.EntityType,
		EntityID:      a.EntityID.String(),
		UserID:        a.UserID.String(),
		Action:        string(a.Action),
		ChangedFields: a.ChangedFields,
		CreatedAt:     a.CreatedAt,
	}
}

func writeAuditLogList(w http.ResponseWriter, result *domain.PaginatedResult[domain.AuditLog]) {
	items := make([]auditLogResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toAuditLogResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// ListAdmin godoc
// @Summary      List audit logs
// @Description  Returns a paginated list of all audit log entries. Admin only. Supports filtering by entity_type, entity_id, user_id, action, and date range.
// @Tags         admin-audit
// @Produce      json
// @Security     BearerAuth
// @Param        page         query     int     false  "Page number"                          default(1)
// @Param        page_size    query     int     false  "Page size"                            default(20)
// @Param        entity_type  query     string  false  "Filter by entity type (e.g. bathhouse)"
// @Param        entity_id    query     string  false  "Filter by entity ID (UUID)"
// @Param        user_id      query     string  false  "Filter by user ID (UUID)"
// @Param        action       query     string  false  "Filter by action (create, update, delete)"
// @Param        from_date    query     string  false  "Filter from date (RFC3339)"
// @Param        to_date      query     string  false  "Filter to date (RFC3339)"
// @Success      200          {object}  APIResponse{data=[]auditLogResponse,meta=Meta}
// @Failure      400          {object}  APIResponse{error=APIError}
// @Failure      401          {object}  APIResponse{error=APIError}
// @Failure      403          {object}  APIResponse{error=APIError}
// @Router       /admin/audit-log [get]
func (h *AuditLogHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := getPage(q.Get("page"))
	pageSize := getPageSize(q.Get("page_size"), 20)

	filter := domain.AuditLogFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if v := q.Get("entity_type"); v != "" {
		filter.EntityType = &v
	}
	if v := q.Get("entity_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid entity_id")
			return
		}
		filter.EntityID = &id
	}
	if v := q.Get("user_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid user_id")
			return
		}
		filter.UserID = &id
	}
	if v := q.Get("action"); v != "" {
		action := domain.AuditAction(v)
		if !action.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid action filter")
			return
		}
		filter.Action = &action
	}
	if v := q.Get("from_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid from_date, expected RFC3339 format")
			return
		}
		filter.FromDate = &t
	}
	if v := q.Get("to_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid to_date, expected RFC3339 format")
			return
		}
		filter.ToDate = &t
	}

	result, err := h.auditSvc.ListAll(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeAuditLogList(w, result)
}

// ListAdminActions godoc
// @Summary      List admin action audit logs
// @Description  Returns a paginated list of admin action audit entries (POST/PUT/PATCH/DELETE on admin routes). Supports filtering by admin_id, action, and date range.
// @Tags         admin-audit
// @Produce      json
// @Security     BearerAuth
// @Param        page         query     int     false  "Page number"                          default(1)
// @Param        page_size    query     int     false  "Page size"                            default(20)
// @Param        admin_id     query     string  false  "Filter by admin user ID (UUID)"
// @Param        action       query     string  false  "Filter by action (create, update, delete)"
// @Param        from_date    query     string  false  "Filter from date (RFC3339)"
// @Param        to_date      query     string  false  "Filter to date (RFC3339)"
// @Success      200          {object}  APIResponse{data=[]auditLogResponse,meta=Meta}
// @Failure      400          {object}  APIResponse{error=APIError}
// @Failure      401          {object}  APIResponse{error=APIError}
// @Failure      403          {object}  APIResponse{error=APIError}
// @Router       /admin/audit-log/actions [get]
func (h *AuditLogHandler) ListAdminActions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := getPage(q.Get("page"))
	pageSize := getPageSize(q.Get("page_size"), 20)

	filter := domain.AuditLogFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if v := q.Get("admin_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid admin_id")
			return
		}
		filter.UserID = &id
	}
	if v := q.Get("action"); v != "" {
		action := domain.AuditAction(v)
		if !action.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid action filter")
			return
		}
		filter.Action = &action
	}
	if v := q.Get("from_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid from_date, expected RFC3339 format")
			return
		}
		filter.FromDate = &t
	}
	if v := q.Get("to_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid to_date, expected RFC3339 format")
			return
		}
		filter.ToDate = &t
	}

	result, err := h.auditSvc.ListAdminActions(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeAuditLogList(w, result)
}

// ListByBathhouse godoc
// @Summary      Get bathhouse edit history
// @Description  Returns a paginated list of audit log entries for a specific bathhouse. Owner or representative only.
// @Tags         bathhouses
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true   "Bathhouse ID (UUID)"
// @Param        page       query     int     false  "Page number"   default(1)
// @Param        page_size  query     int     false  "Page size"     default(20)
// @Success      200        {object}  APIResponse{data=[]auditLogResponse,meta=Meta}
// @Failure      400        {object}  APIResponse{error=APIError}
// @Failure      401        {object}  APIResponse{error=APIError}
// @Failure      403        {object}  APIResponse{error=APIError}
// @Router       /my/bathhouses/{id}/history [get]
func (h *AuditLogHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.accessCheck.CanManageBathhouse(r.Context(), userID, userRole, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	q := r.URL.Query()
	page := getPage(q.Get("page"))
	pageSize := getPageSize(q.Get("page_size"), 20)

	result, err := h.auditSvc.GetHistory(r.Context(), "bathhouse", bathhouseID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeAuditLogList(w, result)
}
