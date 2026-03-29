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

type DisputeHandler struct {
	disputeService service.DisputeService
}

func NewDisputeHandler(disputeService service.DisputeService) *DisputeHandler {
	return &DisputeHandler{disputeService: disputeService}
}

type openDisputeRequest struct {
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

type submitEvidenceRequest struct {
	Type        string `json:"type"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type resolveDisputeRequest struct {
	Resolution         string `json:"resolution"`
	RefundAmount       int64  `json:"refund_amount"`
	CompensationAmount int64  `json:"compensation_amount"`
	MediatorNotes      string `json:"mediator_notes"`
}

type assignDisputeRequest struct {
	MediatorID string `json:"mediator_id"`
}

type disputeResponse struct {
	ID                 string  `json:"id"`
	BookingID          string  `json:"booking_id"`
	InitiatorID        string  `json:"initiator_id"`
	RespondentID       string  `json:"respondent_id"`
	Reason             string  `json:"reason"`
	Description        string  `json:"description"`
	Status             string  `json:"status"`
	Resolution         *string `json:"resolution,omitempty"`
	RefundAmount       int64   `json:"refund_amount"`
	CompensationAmount int64   `json:"compensation_amount"`
	MediatorID         *string `json:"mediator_id,omitempty"`
	MediatorNotes      string  `json:"mediator_notes,omitempty"`
	AppealDeadline     *string `json:"appeal_deadline,omitempty"`
	EvidenceDeadline   *string `json:"evidence_deadline,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	ResolvedAt         *string `json:"resolved_at,omitempty"`
}

type disputeEvidenceResponse struct {
	ID          string `json:"id"`
	DisputeID   string `json:"dispute_id"`
	UserID      string `json:"user_id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func toDisputeResponse(d *domain.Dispute) disputeResponse {
	resp := disputeResponse{
		ID:                 d.ID.String(),
		BookingID:          d.BookingID.String(),
		InitiatorID:        d.InitiatorID.String(),
		RespondentID:       d.RespondentID.String(),
		Reason:             string(d.Reason),
		Description:        d.Description,
		Status:             string(d.Status),
		RefundAmount:       d.RefundAmount,
		CompensationAmount: d.CompensationAmount,
		MediatorNotes:      d.MediatorNotes,
		CreatedAt:          d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          d.UpdatedAt.Format(time.RFC3339),
	}
	if d.Resolution != nil {
		s := string(*d.Resolution)
		resp.Resolution = &s
	}
	if d.MediatorID != nil {
		s := d.MediatorID.String()
		resp.MediatorID = &s
	}
	if d.AppealDeadline != nil {
		s := d.AppealDeadline.Format(time.RFC3339)
		resp.AppealDeadline = &s
	}
	if d.EvidenceDeadline != nil {
		s := d.EvidenceDeadline.Format(time.RFC3339)
		resp.EvidenceDeadline = &s
	}
	if d.ResolvedAt != nil {
		s := d.ResolvedAt.Format(time.RFC3339)
		resp.ResolvedAt = &s
	}
	return resp
}

func toDisputeEvidenceResponse(e *domain.DisputeEvidence) disputeEvidenceResponse {
	return disputeEvidenceResponse{
		ID:          e.ID.String(),
		DisputeID:   e.DisputeID.String(),
		UserID:      e.UserID.String(),
		Type:        string(e.Type),
		URL:         e.URL,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
	}
}

// OpenDispute godoc
//
//	@Summary		Open a dispute on a booking
//	@Description	Open a dispute for a specific booking
//	@Tags			disputes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Booking ID (UUID)"
//	@Param			body	body		openDisputeRequest	true	"Dispute data"
//	@Success		201		{object}	APIResponse{data=disputeResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/dispute [post]
func (h *DisputeHandler) OpenDispute(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid booking ID")
		return
	}

	var req openDisputeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	dispute, err := h.disputeService.OpenDispute(r.Context(), userID, bookingID, domain.DisputeReason(req.Reason), req.Description)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toDisputeResponse(dispute))
}

// SubmitEvidence godoc
//
//	@Summary		Submit evidence for a dispute
//	@Description	Submit evidence for an open dispute within the evidence collection window
//	@Tags			disputes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Dispute ID (UUID)"
//	@Param			body	body		submitEvidenceRequest	true	"Evidence data"
//	@Success		201		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/my/disputes/{id}/evidence [post]
func (h *DisputeHandler) SubmitEvidence(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	var req submitEvidenceRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	evidence := &domain.DisputeEvidence{
		ID:          uuid.New(),
		Type:        domain.DisputeEvidenceType(req.Type),
		URL:         req.URL,
		Description: req.Description,
	}

	if err := h.disputeService.SubmitEvidence(r.Context(), userID, disputeID, evidence); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, simpleMessageResponse{Message: "evidence submitted"})
}

// GetDispute godoc
//
//	@Summary		Get dispute details
//	@Description	Get details of a specific dispute
//	@Tags			disputes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Dispute ID (UUID)"
//	@Success		200	{object}	APIResponse{data=disputeResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/disputes/{id} [get]
func (h *DisputeHandler) GetDispute(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	dispute, err := h.disputeService.GetDispute(r.Context(), userID, role, disputeID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toDisputeResponse(dispute))
}

// ListUserDisputes godoc
//
//	@Summary		List my disputes
//	@Description	Get all disputes for the authenticated user
//	@Tags			disputes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]disputeResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/disputes [get]
func (h *DisputeHandler) ListUserDisputes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	result, err := h.disputeService.ListUserDisputes(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []disputeResponse
	for _, d := range result.Items {
		items = append(items, toDisputeResponse(&d))
	}
	if items == nil {
		items = []disputeResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// ListEvidence godoc
//
//	@Summary		List dispute evidence
//	@Description	Get all evidence for a dispute
//	@Tags			disputes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Dispute ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]disputeEvidenceResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/disputes/{id}/evidence [get]
func (h *DisputeHandler) ListEvidence(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	evidence, err := h.disputeService.ListEvidence(r.Context(), userID, role, disputeID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []disputeEvidenceResponse
	for _, e := range evidence {
		items = append(items, toDisputeEvidenceResponse(&e))
	}
	if items == nil {
		items = []disputeEvidenceResponse{}
	}

	writeJSON(w, http.StatusOK, items)
}

// AppealDispute godoc
//
//	@Summary		Appeal a dispute resolution
//	@Description	Appeal a resolved dispute within the appeal window (7 days)
//	@Tags			disputes
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Dispute ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/my/disputes/{id}/appeal [post]
func (h *DisputeHandler) AppealDispute(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	if err := h.disputeService.AppealDispute(r.Context(), userID, disputeID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "dispute appealed"})
}

// --- Admin endpoints ---

// AdminGetDispute godoc
//
//	@Summary		Get dispute details (admin)
//	@Description	Get details of a specific dispute as admin
//	@Tags			disputes-admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Dispute ID (UUID)"
//	@Success		200	{object}	APIResponse{data=disputeResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/disputes/{id} [get]
func (h *DisputeHandler) AdminGetDispute(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	dispute, err := h.disputeService.GetDispute(r.Context(), userID, role, disputeID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toDisputeResponse(dispute))
}

// AdminListDisputes godoc
//
//	@Summary		List all disputes (admin)
//	@Description	List all disputes with optional status filter
//	@Tags			disputes-admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status		query		string	false	"Filter by status"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]disputeResponse}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/admin/disputes [get]
func (h *DisputeHandler) AdminListDisputes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	filter := domain.DisputeFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.DisputeStatus(s)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_status", "invalid dispute status")
			return
		}
		filter.Status = &status
	}

	result, err := h.disputeService.ListAllDisputes(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []disputeResponse
	for _, d := range result.Items {
		items = append(items, toDisputeResponse(&d))
	}
	if items == nil {
		items = []disputeResponse{}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// AdminAssignDispute godoc
//
//	@Summary		Assign dispute to mediator
//	@Description	Assign a dispute to an admin mediator
//	@Tags			disputes-admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Dispute ID (UUID)"
//	@Param			body	body		assignDisputeRequest	true	"Assignment data"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/disputes/{id}/assign [patch]
func (h *DisputeHandler) AdminAssignDispute(w http.ResponseWriter, r *http.Request) {
	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	var req assignDisputeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	mediatorID, err := uuid.Parse(req.MediatorID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_mediator_id", "invalid mediator ID")
		return
	}

	if err := h.disputeService.AssignDispute(r.Context(), disputeID, mediatorID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "dispute assigned"})
}

// AdminResolveDispute godoc
//
//	@Summary		Resolve a dispute
//	@Description	Resolve a dispute with a decision (full/partial/no refund)
//	@Tags			disputes-admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Dispute ID (UUID)"
//	@Param			body	body		resolveDisputeRequest	true	"Resolution data"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/admin/disputes/{id}/resolve [patch]
func (h *DisputeHandler) AdminResolveDispute(w http.ResponseWriter, r *http.Request) {
	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	var req resolveDisputeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.disputeService.ResolveDispute(r.Context(), disputeID, domain.DisputeResolution(req.Resolution), req.RefundAmount, req.CompensationAmount, req.MediatorNotes); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "dispute resolved"})
}

// AdminCloseDispute godoc
//
//	@Summary		Close a dispute
//	@Description	Close a dispute (after resolution or appeal)
//	@Tags			disputes-admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Dispute ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/admin/disputes/{id}/close [patch]
func (h *DisputeHandler) AdminCloseDispute(w http.ResponseWriter, r *http.Request) {
	disputeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid dispute ID")
		return
	}

	if err := h.disputeService.CloseDispute(r.Context(), disputeID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "dispute closed"})
}
