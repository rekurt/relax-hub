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

func (h *ComplaintHandler) ReportReview(w http.ResponseWriter, r *http.Request) {
	h.report(w, r, domain.ComplaintTargetReview)
}

func (h *ComplaintHandler) ReportBathhouse(w http.ResponseWriter, r *http.Request) {
	h.report(w, r, domain.ComplaintTargetBathhouse)
}

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
		filter.Status = &s
	}
	if v := q.Get("target_type"); v != "" {
		t := domain.ComplaintTargetType(v)
		filter.TargetType = &t
	}
	if v := q.Get("reason"); v != "" {
		reason := domain.ComplaintReason(v)
		filter.Reason = &reason
	}
	if v := q.Get("from_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.FromDate = &t
		}
	}
	if v := q.Get("to_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.ToDate = &t
		}
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
