package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type RFMHandler struct {
	rfmService service.RFMService
}

func NewRFMHandler(rfmService service.RFMService) *RFMHandler {
	return &RFMHandler{rfmService: rfmService}
}

type rfmGuestResponse struct {
	ID           string       `json:"id"`
	ClientID     string       `json:"client_id"`
	BathhouseID  string       `json:"bathhouse_id"`
	LastVisitAt  string       `json:"last_visit_at"`
	VisitCount   int          `json:"visit_count"`
	TotalSpent   int64        `json:"total_spent"`
	AvgCheck     int64        `json:"avg_check"`
	Tags         []string     `json:"tags"`
	RFM          rfmScoreResp `json:"rfm"`
}

type rfmScoreResp struct {
	Recency   int `json:"recency"`
	Frequency int `json:"frequency"`
	Monetary  int `json:"monetary"`
}

type rfmMatrixResp struct {
	Recency   int   `json:"recency"`
	Frequency int   `json:"frequency"`
	Count     int64 `json:"count"`
}

type rfmAnalysisResponse struct {
	Guests []rfmGuestResponse `json:"guests"`
	Matrix []rfmMatrixResp    `json:"matrix"`
}

// GetRFMAnalysis godoc
//
//	@Summary		Get RFM analysis
//	@Description	Get RFM (Recency-Frequency-Monetary) analysis for all guests
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=rfmAnalysisResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/rfm [get]
func (h *RFMHandler) GetRFMAnalysis(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	result, err := h.rfmService.GetRFMAnalysis(r.Context(), userID, role)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := rfmAnalysisResponse{
		Guests: make([]rfmGuestResponse, 0, len(result.Guests)),
		Matrix: make([]rfmMatrixResp, 0, len(result.Matrix)),
	}
	for _, g := range result.Guests {
		tags := g.Tags
		if tags == nil {
			tags = []string{}
		}
		resp.Guests = append(resp.Guests, rfmGuestResponse{
			ID:          g.ID.String(),
			ClientID:    g.ClientID.String(),
			BathhouseID: g.BathhouseID.String(),
			LastVisitAt: g.LastVisitAt.Format(time.RFC3339),
			VisitCount:  g.VisitCount,
			TotalSpent:  g.TotalSpent,
			AvgCheck:    g.AvgCheck,
			Tags:        tags,
			RFM: rfmScoreResp{
				Recency:   g.RFM.Recency,
				Frequency: g.RFM.Frequency,
				Monetary:  g.RFM.Monetary,
			},
		})
	}
	for _, m := range result.Matrix {
		resp.Matrix = append(resp.Matrix, rfmMatrixResp{
			Recency:   m.Recency,
			Frequency: m.Frequency,
			Count:     m.Count,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

type customSegmentResponse struct {
	ID          string                       `json:"id"`
	OwnerID     string                       `json:"owner_id"`
	BathhouseID *string                      `json:"bathhouse_id,omitempty"`
	Name        string                       `json:"name"`
	Conditions  domain.CustomSegmentCondition `json:"conditions"`
	GuestCount  int64                        `json:"guest_count"`
	CreatedAt   string                       `json:"created_at"`
	UpdatedAt   string                       `json:"updated_at"`
}

func toCustomSegmentResponse(s *domain.CustomSegment) customSegmentResponse {
	resp := customSegmentResponse{
		ID:         s.ID.String(),
		OwnerID:    s.OwnerID.String(),
		Name:       s.Name,
		Conditions: s.Conditions,
		GuestCount: s.GuestCount,
		CreatedAt:  s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  s.UpdatedAt.Format(time.RFC3339),
	}
	if s.BathhouseID != nil {
		bhID := s.BathhouseID.String()
		resp.BathhouseID = &bhID
	}
	return resp
}

type createCustomSegmentRequest struct {
	Name        string                       `json:"name"`
	BathhouseID *string                      `json:"bathhouse_id,omitempty"`
	Conditions  domain.CustomSegmentCondition `json:"conditions"`
}

// ListCustomSegments godoc
//
//	@Summary		List custom segments
//	@Description	Get all custom CRM segments for the owner
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]customSegmentResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/segments/custom [get]
func (h *RFMHandler) ListCustomSegments(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	segments, err := h.rfmService.ListCustomSegments(r.Context(), userID, role)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []customSegmentResponse
	for _, s := range segments {
		items = append(items, toCustomSegmentResponse(&s))
	}
	if items == nil {
		items = []customSegmentResponse{}
	}

	writeJSON(w, http.StatusOK, items)
}

// CreateCustomSegment godoc
//
//	@Summary		Create custom segment
//	@Description	Create a new custom CRM segment with filter conditions
//	@Tags			crm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createCustomSegmentRequest	true	"Segment data"
//	@Success		201		{object}	APIResponse{data=customSegmentResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/crm/segments/custom [post]
func (h *RFMHandler) CreateCustomSegment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req createCustomSegmentRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	segment := &domain.CustomSegment{
		Name:       req.Name,
		Conditions: req.Conditions,
	}
	if req.BathhouseID != nil {
		bhID, err := uuid.Parse(*req.BathhouseID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_bathhouse_id", "invalid bathhouse ID")
			return
		}
		segment.BathhouseID = &bhID
	}

	if err := h.rfmService.CreateCustomSegment(r.Context(), userID, role, segment); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toCustomSegmentResponse(segment))
}

// UpdateCustomSegment godoc
//
//	@Summary		Update custom segment
//	@Description	Update an existing custom CRM segment
//	@Tags			crm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Segment ID (UUID)"
//	@Param			body	body		createCustomSegmentRequest	true	"Segment data"
//	@Success		200		{object}	APIResponse{data=customSegmentResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/crm/segments/custom/{id} [put]
func (h *RFMHandler) UpdateCustomSegment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	segmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid segment ID")
		return
	}

	var req createCustomSegmentRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	segment := &domain.CustomSegment{
		ID:         segmentID,
		Name:       req.Name,
		Conditions: req.Conditions,
	}
	if req.BathhouseID != nil {
		bhID, err := uuid.Parse(*req.BathhouseID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_bathhouse_id", "invalid bathhouse ID")
			return
		}
		segment.BathhouseID = &bhID
	}

	if err := h.rfmService.UpdateCustomSegment(r.Context(), userID, role, segment); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toCustomSegmentResponse(segment))
}

// DeleteCustomSegment godoc
//
//	@Summary		Delete custom segment
//	@Description	Delete a custom CRM segment
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Segment ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/segments/custom/{id} [delete]
func (h *RFMHandler) DeleteCustomSegment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	segmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid segment ID")
		return
	}

	if err := h.rfmService.DeleteCustomSegment(r.Context(), userID, role, segmentID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "segment deleted"})
}

// GetCustomSegmentGuests godoc
//
//	@Summary		Get guests in custom segment
//	@Description	Get paginated list of guests matching a custom segment's conditions
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Segment ID (UUID)"
//	@Param			page		query		int		false	"Page number"
//	@Param			page_size	query		int		false	"Page size"
//	@Success		200			{object}	APIResponse{data=[]guestCardResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Failure		404			{object}	APIResponse{error=APIError}
//	@Router			/my/crm/segments/custom/{id}/guests [get]
func (h *RFMHandler) GetCustomSegmentGuests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	segmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid segment ID")
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

	result, err := h.rfmService.GetCustomSegmentGuests(r.Context(), userID, role, segmentID, page, pageSize)
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
