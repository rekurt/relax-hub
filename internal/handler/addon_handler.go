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

type AddOnHandler struct {
	addOnService service.AddOnService
}

func NewAddOnHandler(addOnService service.AddOnService) *AddOnHandler {
	return &AddOnHandler{addOnService: addOnService}
}

type createAddOnRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Unit        string `json:"unit"`
	SortOrder   int    `json:"sort_order"`
}

type updateAddOnRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Unit        string `json:"unit"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

type addOnResponse struct {
	ID          string    `json:"id"`
	BathhouseID string    `json:"bathhouse_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Unit        string    `json:"unit"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toAddOnResponse(a *domain.AddOn) addOnResponse {
	return addOnResponse{
		ID:          a.ID.String(),
		BathhouseID: a.BathhouseID.String(),
		Name:        a.Name,
		Description: a.Description,
		Price:       a.Price,
		Unit:        string(a.Unit),
		IsActive:    a.IsActive,
		SortOrder:   a.SortOrder,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func toAddOnListResponse(addons []domain.AddOn) []addOnResponse {
	result := make([]addOnResponse, len(addons))
	for i := range addons {
		result[i] = toAddOnResponse(&addons[i])
	}
	return result
}

// CreateAddOn godoc
//
//	@Summary		Create add-on for bathhouse
//	@Description	Creates a new add-on service for a bathhouse. Available to owners and representatives.
//	@Tags			add-ons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Bathhouse ID (UUID)"
//	@Param			body	body		createAddOnRequest	true	"Add-on data"
//	@Success		201		{object}	APIResponse{data=addOnResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/addons [post]
func (h *AddOnHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req createAddOnRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	addon := &domain.AddOn{
		BathhouseID: bathhouseID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Unit:        domain.AddOnUnit(req.Unit),
		SortOrder:   req.SortOrder,
	}

	result, err := h.addOnService.CreateAddOn(r.Context(), userID, userRole, addon)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toAddOnResponse(result))
}

// UpdateAddOn godoc
//
//	@Summary		Update add-on
//	@Description	Updates an existing add-on. Available to owners and representatives.
//	@Tags			add-ons
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Add-on ID (UUID)"
//	@Param			body	body		updateAddOnRequest	true	"Updated add-on data"
//	@Success		200		{object}	APIResponse{data=addOnResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/addons/{id} [put]
func (h *AddOnHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	addonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid add-on id")
		return
	}

	var req updateAddOnRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	addon := &domain.AddOn{
		ID:          addonID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Unit:        domain.AddOnUnit(req.Unit),
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}

	result, err := h.addOnService.UpdateAddOn(r.Context(), userID, userRole, addon)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAddOnResponse(result))
}

// DeleteAddOn godoc
//
//	@Summary		Delete add-on
//	@Description	Deletes an add-on. Available to owners and representatives.
//	@Tags			add-ons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Add-on ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/addons/{id} [delete]
func (h *AddOnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	addonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid add-on id")
		return
	}

	if err := h.addOnService.DeleteAddOn(r.Context(), userID, userRole, addonID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "add-on deleted"})
}

// ListByBathhouse godoc
//
//	@Summary		List add-ons for bathhouse (owner)
//	@Description	Lists all add-ons for a bathhouse. Available to owners and representatives.
//	@Tags			add-ons
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]addOnResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/bathhouses/{id}/addons [get]
func (h *AddOnHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	addons, err := h.addOnService.ListAddOnsManaged(r.Context(), userID, userRole, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAddOnListResponse(addons))
}

// ListPublic godoc
//
//	@Summary		List active add-ons for bathhouse (public)
//	@Description	Lists active add-ons for a bathhouse. Available publicly.
//	@Tags			add-ons
//	@Produce		json
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]addOnResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/addons [get]
func (h *AddOnHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	addons, err := h.addOnService.ListActiveAddOns(r.Context(), bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toAddOnListResponse(addons))
}
