package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/service"
)

type ObjectTypeHandler struct {
	svc service.ObjectTypeService
}

func NewObjectTypeHandler(svc service.ObjectTypeService) *ObjectTypeHandler {
	return &ObjectTypeHandler{svc: svc}
}

type objectTypeResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type createObjectTypeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

type updateObjectTypeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsActive    *bool  `json:"is_active"`
}

func toObjectTypeResponse(o *domain.ObjectType) objectTypeResponse {
	return objectTypeResponse{
		ID:          o.ID.String(),
		Name:        o.Name,
		Description: o.Description,
		SortOrder:   o.SortOrder,
		IsActive:    o.IsActive,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

// ListObjectTypes godoc
//
//	@Summary		List object types
//	@Description	Returns all bathhouse object types
//	@Tags			admin,object-types
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]objectTypeResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/object-types [get]
func (h *ObjectTypeHandler) ListObjectTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.svc.ListAll(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]objectTypeResponse, len(types))
	for i := range types {
		result[i] = toObjectTypeResponse(&types[i])
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateObjectType godoc
//
//	@Summary		Create object type
//	@Description	Creates a new bathhouse object type
//	@Tags			admin,object-types
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createObjectTypeRequest	true	"Object type data"
//	@Success		201		{object}	APIResponse{data=objectTypeResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/object-types [post]
func (h *ObjectTypeHandler) CreateObjectType(w http.ResponseWriter, r *http.Request) {
	var req createObjectTypeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	objType := &domain.ObjectType{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    isActive,
	}

	if err := h.svc.Create(r.Context(), objType); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toObjectTypeResponse(objType))
}

// UpdateObjectType godoc
//
//	@Summary		Update object type
//	@Description	Updates an existing bathhouse object type
//	@Tags			admin,object-types
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Object type ID (UUID)"
//	@Param			body	body		updateObjectTypeRequest	true	"Updated data"
//	@Success		200		{object}	APIResponse{data=objectTypeResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/object-types/{id} [put]
func (h *ObjectTypeHandler) UpdateObjectType(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid object type id")
		return
	}

	var req updateObjectTypeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	objType := &domain.ObjectType{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    isActive,
	}

	if err := h.svc.Update(r.Context(), objType); err != nil {
		handleServiceError(w, err)
		return
	}

	updated, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toObjectTypeResponse(updated))
}

// DeleteObjectType godoc
//
//	@Summary		Delete object type
//	@Description	Deletes a bathhouse object type
//	@Tags			admin,object-types
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Object type ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/object-types/{id} [delete]
func (h *ObjectTypeHandler) DeleteObjectType(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid object type id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "object type deleted"})
}
