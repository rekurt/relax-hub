package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ComplaintHandler struct {
	complaintService service.ComplaintService
}

func NewComplaintHandler(complaintService service.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{complaintService: complaintService}
}

type reportRequest struct {
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

type resolveRequest struct {
	Resolution string `json:"resolution"`
}

type complaintResponse struct {
	ID           string     `json:"id"`
	ReporterID   string     `json:"reporter_id"`
	TargetType   string     `json:"target_type"`
	TargetID     string     `json:"target_id"`
	Reason       string     `json:"reason"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	ResolvedByID *string    `json:"resolved_by_id,omitempty"`
	Resolution   string     `json:"resolution,omitempty"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func toComplaintResponse(c *domain.Complaint) complaintResponse {
	resp := complaintResponse{
		ID:          c.ID.String(),
		ReporterID:  c.ReporterID.String(),
		TargetType:  string(c.TargetType),
		TargetID:    c.TargetID.String(),
		Reason:      string(c.Reason),
		Description: c.Description,
		Status:      string(c.Status),
		Resolution:  c.Resolution,
		ResolvedAt:  c.ResolvedAt,
		CreatedAt:   c.CreatedAt,
	}
	if c.ResolvedByID != nil {
		s := c.ResolvedByID.String()
		resp.ResolvedByID = &s
	}
	return resp
}

// ReportReview godoc
// @Summary      Report a review
// @Description  Submit a complaint about a review (spam, offensive, fake, fraud, other)
// @Tags         complaints
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string         true  "Review ID (UUID)"
// @Param        body  body      reportRequest  true  "Report details"
// @Success      201   {object}  APIResponse{data=complaintResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /reviews/{id}/report [post]
func (h *ComplaintHandler) ReportReview(w http.ResponseWriter, r *http.Request) {
	h.report(w, r, domain.ComplaintTargetReview)
}

// ReportBathhouse godoc
// @Summary      Report a bathhouse
// @Description  Submit a complaint about a bathhouse (spam, offensive, fake, fraud, other)
// @Tags         complaints
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string         true  "Bathhouse ID (UUID)"
// @Param        body  body      reportRequest  true  "Report details"
// @Success      201   {object}  APIResponse{data=complaintResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id}/report [post]
func (h *ComplaintHandler) ReportBathhouse(w http.ResponseWriter, r *http.Request) {
	h.report(w, r, domain.ComplaintTargetBathhouse)
}

// ReportUser godoc
// @Summary      Report a user
// @Description  Submit a complaint about a user (spam, offensive, fake, fraud, other)
// @Tags         complaints
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string         true  "User ID (UUID)"
// @Param        body  body      reportRequest  true  "Report details"
// @Success      201   {object}  APIResponse{data=complaintResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /users/{id}/report [post]
func (h *ComplaintHandler) ReportUser(w http.ResponseWriter, r *http.Request) {
	h.report(w, r, domain.ComplaintTargetUser)
}

func (h *ComplaintHandler) report(w http.ResponseWriter, r *http.Request, targetType domain.ComplaintTargetType) {
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid id")
		return
	}

	var req reportRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	complaint, err := h.complaintService.Report(r.Context(), userID, service.CreateComplaintInput{
		TargetType:  targetType,
		TargetID:    targetID,
		Reason:      domain.ComplaintReason(req.Reason),
		Description: req.Description,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toComplaintResponse(complaint))
}

// List godoc
// @Summary      List complaints
// @Description  Returns a paginated list of complaints. Admin only. Supports filtering by status, target_type, reason, and date range.
// @Tags         admin-complaints
// @Produce      json
// @Security     BearerAuth
// @Param        page         query     int     false  "Page number"                          default(1)
// @Param        page_size    query     int     false  "Page size"                            default(20)
// @Param        status       query     string  false  "Filter by status (pending, resolved, dismissed)"
// @Param        target_type  query     string  false  "Filter by target type (review, bathhouse, user)"
// @Param        reason       query     string  false  "Filter by reason (spam, offensive, fake, fraud, other)"
// @Param        from_date    query     string  false  "Filter from date (RFC3339)"
// @Param        to_date      query     string  false  "Filter to date (RFC3339)"
// @Success      200          {object}  APIResponse{data=[]complaintResponse,meta=Meta}
// @Failure      400          {object}  APIResponse{error=APIError}
// @Failure      401          {object}  APIResponse{error=APIError}
// @Failure      403          {object}  APIResponse{error=APIError}
// @Router       /admin/complaints [get]
func (h *ComplaintHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := getPage(q.Get("page"))
	pageSize := getPageSize(q.Get("page_size"), 20)

	filter := domain.ComplaintFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if v := q.Get("status"); v != "" {
		s := domain.ComplaintStatus(v)
		if !s.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid status filter")
			return
		}
		filter.Status = &s
	}
	if v := q.Get("target_type"); v != "" {
		t := domain.ComplaintTargetType(v)
		if !t.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid target_type filter")
			return
		}
		filter.TargetType = &t
	}
	if v := q.Get("reason"); v != "" {
		reason := domain.ComplaintReason(v)
		if !reason.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid reason filter")
			return
		}
		filter.Reason = &reason
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

	result, err := h.complaintService.List(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]complaintResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toComplaintResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// GetByID godoc
// @Summary      Get complaint by ID
// @Description  Returns a single complaint by its ID. Admin only.
// @Tags         admin-complaints
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Complaint ID (UUID)"
// @Success      200  {object}  APIResponse{data=complaintResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /admin/complaints/{id} [get]
func (h *ComplaintHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid complaint id")
		return
	}

	complaint, err := h.complaintService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toComplaintResponse(complaint))
}

// Resolve godoc
// @Summary      Resolve complaint
// @Description  Resolves a complaint with a resolution note. Admin only.
// @Tags         admin-complaints
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string          true  "Complaint ID (UUID)"
// @Param        body  body      resolveRequest  true  "Resolution details"
// @Success      200   {object}  APIResponse{data=complaintResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /admin/complaints/{id}/resolve [patch]
func (h *ComplaintHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid complaint id")
		return
	}

	var req resolveRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	adminID := middleware.GetUserID(r.Context())

	complaint, err := h.complaintService.Resolve(r.Context(), id, adminID, req.Resolution)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toComplaintResponse(complaint))
}

// Dismiss godoc
// @Summary      Dismiss complaint
// @Description  Dismisses a complaint without action. Admin only.
// @Tags         admin-complaints
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Complaint ID (UUID)"
// @Success      200  {object}  APIResponse{data=complaintResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /admin/complaints/{id}/dismiss [patch]
func (h *ComplaintHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid complaint id")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	complaint, err := h.complaintService.Dismiss(r.Context(), id, adminID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toComplaintResponse(complaint))
}
