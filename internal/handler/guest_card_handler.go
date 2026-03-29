package handler

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type GuestCardHandler struct {
	guestCardService service.GuestCardService
}

func NewGuestCardHandler(guestCardService service.GuestCardService) *GuestCardHandler {
	return &GuestCardHandler{guestCardService: guestCardService}
}

type guestCardResponse struct {
	ID           string   `json:"id"`
	OwnerID      string   `json:"owner_id"`
	ClientID     string   `json:"client_id"`
	BathhouseID  string   `json:"bathhouse_id"`
	FirstVisitAt string   `json:"first_visit_at"`
	LastVisitAt  string   `json:"last_visit_at"`
	VisitCount   int      `json:"visit_count"`
	TotalSpent   int64    `json:"total_spent"`
	AvgCheck     int64    `json:"avg_check"`
	Notes        string   `json:"notes"`
	Tags         []string `json:"tags"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

func toGuestCardResponse(c *domain.GuestCard) guestCardResponse {
	tags := c.Tags
	if tags == nil {
		tags = []string{}
	}
	return guestCardResponse{
		ID:           c.ID.String(),
		OwnerID:      c.OwnerID.String(),
		ClientID:     c.ClientID.String(),
		BathhouseID:  c.BathhouseID.String(),
		FirstVisitAt: c.FirstVisitAt.Format(time.RFC3339),
		LastVisitAt:  c.LastVisitAt.Format(time.RFC3339),
		VisitCount:   c.VisitCount,
		TotalSpent:   c.TotalSpent,
		AvgCheck:     c.AvgCheck,
		Notes:        c.Notes,
		Tags:         tags,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Format(time.RFC3339),
	}
}

type updateGuestNotesRequest struct {
	Notes string   `json:"notes"`
	Tags  []string `json:"tags"`
}

// ListGuests godoc
//
//	@Summary		List CRM guest cards
//	@Description	Get paginated list of guest cards for the owner
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			bathhouse_id	query		string	false	"Filter by bathhouse ID"
//	@Param			search			query		string	false	"Search by client name/email/phone"
//	@Param			tag				query		string	false	"Filter by tag"
//	@Param			date_from		query		string	false	"Filter by last visit date from (RFC3339)"
//	@Param			date_to			query		string	false	"Filter by last visit date to (RFC3339)"
//	@Param			sort_by			query		string	false	"Sort by: last_visit, total_spent, visit_count, avg_check"
//	@Param			page			query		int		false	"Page number"
//	@Param			page_size		query		int		false	"Page size"
//	@Success		200				{object}	APIResponse{data=[]guestCardResponse}
//	@Failure		401				{object}	APIResponse{error=APIError}
//	@Failure		403				{object}	APIResponse{error=APIError}
//	@Router			/my/crm/guests [get]
func (h *GuestCardHandler) ListGuests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	filter := domain.GuestCardFilter{}

	if bhID := r.URL.Query().Get("bathhouse_id"); bhID != "" {
		id, err := uuid.Parse(bhID)
		if err == nil {
			filter.BathhouseID = &id
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}

	if tag := r.URL.Query().Get("tag"); tag != "" {
		filter.Tag = &tag
	}

	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}

	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			filter.DateTo = &t
		}
	}

	filter.SortBy = r.URL.Query().Get("sort_by")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	filter.Page = page
	filter.PageSize = pageSize

	result, err := h.guestCardService.ListGuests(r.Context(), userID, role, filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []guestCardResponse
	for _, c := range result.Items {
		items = append(items, toGuestCardResponse(&c))
	}
	if items == nil {
		items = []guestCardResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// UpdateGuestNotes godoc
//
//	@Summary		Update guest card notes and tags
//	@Description	Update notes and tags for a specific guest card
//	@Tags			crm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Guest card ID (UUID)"
//	@Param			body	body		updateGuestNotesRequest	true	"Notes and tags"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/crm/guests/{id} [put]
func (h *GuestCardHandler) UpdateGuestNotes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	cardID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid guest card ID")
		return
	}

	var req updateGuestNotesRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.guestCardService.UpdateGuestNotes(r.Context(), userID, role, cardID, req.Notes, req.Tags); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "notes updated"})
}

// ExportCSV godoc
//
//	@Summary		Export guest cards as CSV
//	@Description	Export all guest cards as a CSV file
//	@Tags			crm
//	@Produce		text/csv
//	@Security		BearerAuth
//	@Success		200	{file}		file
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/guests/export [get]
func (h *GuestCardHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if role != domain.RoleOwner && role != domain.RoleRepresentative && role != domain.RoleAdmin {
		handleServiceError(w, domain.ErrForbidden)
		return
	}

	filter := domain.GuestCardFilter{}

	var buf bytes.Buffer
	if err := h.guestCardService.ExportCSV(r.Context(), userID, role, filter, &buf); err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=guests.csv")
	_, _ = io.Copy(w, &buf)
}

type segmentResponse struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
}

// ListSegments godoc
//
//	@Summary		List CRM segments
//	@Description	Get all predefined CRM segments with guest counts
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]segmentResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/segments [get]
func (h *GuestCardHandler) ListSegments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	segments, err := h.guestCardService.ListSegments(r.Context(), userID, role)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []segmentResponse
	for _, s := range segments {
		items = append(items, segmentResponse{
			Slug:        string(s.Slug),
			Name:        s.Name,
			Description: s.Description,
			Count:       s.Count,
		})
	}
	if items == nil {
		items = []segmentResponse{}
	}

	writeJSON(w, http.StatusOK, items)
}

// GetGuestsInSegment godoc
//
//	@Summary		Get guests in segment
//	@Description	Get paginated list of guests in a specific CRM segment
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			slug		path		string	true	"Segment slug (new, regular, lost, vip, birthday_soon)"
//	@Param			page		query		int		false	"Page number"
//	@Param			page_size	query		int		false	"Page size"
//	@Success		200			{object}	APIResponse{data=[]guestCardResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/my/crm/segments/{slug}/guests [get]
func (h *GuestCardHandler) GetGuestsInSegment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	slug := chi.URLParam(r, "slug")
	segment := domain.GuestSegmentSlug(slug)

	// Validate segment slug
	valid := false
	for _, s := range domain.AllSegments() {
		if s == segment {
			valid = true
			break
		}
	}
	if !valid {
		writeError(w, http.StatusBadRequest, "invalid_segment", "invalid segment slug")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	result, err := h.guestCardService.GetGuestsInSegment(r.Context(), userID, role, segment, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []guestCardResponse
	for _, c := range result.Items {
		items = append(items, toGuestCardResponse(&c))
	}
	if items == nil {
		items = []guestCardResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// GetStats godoc
//
//	@Summary		Get CRM guest stats
//	@Description	Get summary statistics for all guest cards
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=domain.GuestCardStats}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/stats [get]
func (h *GuestCardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	stats, err := h.guestCardService.GetStats(r.Context(), userID, role)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
