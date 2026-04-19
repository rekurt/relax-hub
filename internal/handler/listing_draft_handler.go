package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type ListingDraftHandler struct {
	draftService service.ListingDraftService
}

func NewListingDraftHandler(draftService service.ListingDraftService) *ListingDraftHandler {
	return &ListingDraftHandler{draftService: draftService}
}

type listingDraftResponse struct {
	ID          string                     `json:"id"`
	UserID      string                     `json:"user_id"`
	Status      string                     `json:"status"`
	CurrentStep int                        `json:"current_step"`
	StepData    map[string]json.RawMessage `json:"step_data"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

type listingDraftListItem struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	CurrentStep int       `json:"current_step"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type saveStepRequest struct {
	Data json.RawMessage `json:"data"`
}

func makeDraftResponse(id, userID uuid.UUID, status string, currentStep int, stepData map[int]json.RawMessage, createdAt, updatedAt time.Time) listingDraftResponse {
	sd := make(map[string]json.RawMessage, len(stepData))
	for k, v := range stepData {
		sd[strconv.Itoa(k)] = v
	}
	return listingDraftResponse{
		ID:          id.String(),
		UserID:      userID.String(),
		Status:      status,
		CurrentStep: currentStep,
		StepData:    sd,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// CreateDraft godoc
//
//	@Summary		Create listing draft
//	@Description	Create a new listing draft (requires approved KYC and accepted offer)
//	@Tags			listing-drafts
//	@Produce		json
//	@Security		BearerAuth
//	@Success		201	{object}	APIResponse{data=listingDraftResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/listing-drafts [post]
func (h *ListingDraftHandler) CreateDraft(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	draft, err := h.draftService.Create(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, makeDraftResponse(
		draft.ID, draft.UserID, string(draft.Status), draft.CurrentStep,
		draft.StepData, draft.CreatedAt, draft.UpdatedAt,
	))
}

// SaveStep godoc
//
//	@Summary		Save draft step data
//	@Description	Save data for a specific step of the listing draft wizard
//	@Tags			listing-drafts
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Draft ID (UUID)"
//	@Param			step	path		int				true	"Step number (1-7)"
//	@Param			body	body		saveStepRequest	true	"Step data"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/listing-drafts/{id}/step/{step} [put]
func (h *ListingDraftHandler) SaveStep(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	draftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid draft ID")
		return
	}

	step, err := strconv.Atoi(chi.URLParam(r, "step"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_step", "invalid step number")
		return
	}

	var req saveStepRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}
	if len(req.Data) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "step data is required")
		return
	}

	if err := h.draftService.SaveStep(r.Context(), draftID, userID, step, req.Data); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "step saved"})
}

// GetDraft godoc
//
//	@Summary		Get listing draft
//	@Description	Get a listing draft with all step data
//	@Tags			listing-drafts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Draft ID (UUID)"
//	@Success		200	{object}	APIResponse{data=listingDraftResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/listing-drafts/{id} [get]
func (h *ListingDraftHandler) GetDraft(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	draftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid draft ID")
		return
	}

	draft, err := h.draftService.GetDraft(r.Context(), draftID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, makeDraftResponse(
		draft.ID, draft.UserID, string(draft.Status), draft.CurrentStep,
		draft.StepData, draft.CreatedAt, draft.UpdatedAt,
	))
}

// ListDrafts godoc
//
//	@Summary		List user's listing drafts
//	@Description	Get all listing drafts for the authenticated user
//	@Tags			listing-drafts
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]listingDraftListItem}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/listing-drafts [get]
func (h *ListingDraftHandler) ListDrafts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	drafts, err := h.draftService.ListDrafts(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]listingDraftListItem, 0, len(drafts))
	for _, d := range drafts {
		items = append(items, listingDraftListItem{
			ID:          d.ID.String(),
			Status:      string(d.Status),
			CurrentStep: d.CurrentStep,
			CreatedAt:   d.CreatedAt,
			UpdatedAt:   d.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

// SubmitDraft godoc
//
//	@Summary		Submit listing draft
//	@Description	Submit a completed draft to create a bathhouse listing (sends to moderation)
//	@Tags			listing-drafts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Draft ID (UUID)"
//	@Success		200	{object}	APIResponse{data=listingDraftResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/my/listing-drafts/{id}/submit [post]
func (h *ListingDraftHandler) SubmitDraft(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	draftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid draft ID")
		return
	}

	draft, err := h.draftService.Submit(r.Context(), draftID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, makeDraftResponse(
		draft.ID, draft.UserID, string(draft.Status), draft.CurrentStep,
		draft.StepData, draft.CreatedAt, draft.UpdatedAt,
	))
}

// DeleteDraft godoc
//
//	@Summary		Delete listing draft
//	@Description	Delete (discard) a listing draft
//	@Tags			listing-drafts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Draft ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/listing-drafts/{id} [delete]
func (h *ListingDraftHandler) DeleteDraft(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	draftID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid draft ID")
		return
	}

	if err := h.draftService.Delete(r.Context(), draftID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "draft deleted"})
}
