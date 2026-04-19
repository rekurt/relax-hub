package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type TemplateHandler struct {
	templateService service.TemplateService
}

func NewTemplateHandler(templateService service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

type createTemplateRequest struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	SortOrder int    `json:"sort_order"`
}

type updateTemplateRequest struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	SortOrder int    `json:"sort_order"`
}

type templateResponse struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	IsDefault bool   `json:"is_default"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
}

func toTemplateResponse(t *domain.ResponseTemplate) templateResponse {
	return templateResponse{
		ID:        t.ID.String(),
		OwnerID:   t.OwnerID.String(),
		Title:     t.Title,
		Body:      t.Body,
		IsDefault: t.IsDefault,
		SortOrder: t.SortOrder,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
}

// CreateTemplate godoc
//
//	@Summary		Create a response template
//	@Description	Create a new response template for review responses
//	@Tags			crm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createTemplateRequest	true	"Template data"
//	@Success		201		{object}	APIResponse{data=templateResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/crm/templates [post]
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	var req createTemplateRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	template := &domain.ResponseTemplate{
		ID:        uuid.New(),
		Title:     req.Title,
		Body:      req.Body,
		SortOrder: req.SortOrder,
	}

	if err := h.templateService.Create(r.Context(), userID, role, template); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTemplateResponse(template))
}

// ListTemplates godoc
//
//	@Summary		List response templates
//	@Description	Get all response templates for the owner (seeds defaults on first use)
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]templateResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/templates [get]
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	templates, err := h.templateService.List(r.Context(), userID, role)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var items []templateResponse
	for _, t := range templates {
		items = append(items, toTemplateResponse(&t))
	}
	if items == nil {
		items = []templateResponse{}
	}

	writeJSON(w, http.StatusOK, items)
}

// UpdateTemplate godoc
//
//	@Summary		Update a response template
//	@Description	Update an existing response template
//	@Tags			crm
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Template ID (UUID)"
//	@Param			body	body		updateTemplateRequest	true	"Template data"
//	@Success		200		{object}	APIResponse{data=templateResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/my/crm/templates/{id} [put]
func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	templateID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid template ID")
		return
	}

	var req updateTemplateRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	template := &domain.ResponseTemplate{
		ID:        templateID,
		Title:     req.Title,
		Body:      req.Body,
		SortOrder: req.SortOrder,
	}

	if err := h.templateService.Update(r.Context(), userID, role, template); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTemplateResponse(template))
}

// DeleteTemplate godoc
//
//	@Summary		Delete a response template
//	@Description	Delete a response template
//	@Tags			crm
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Template ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/crm/templates/{id} [delete]
func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	templateID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "invalid template ID")
		return
	}

	if err := h.templateService.Delete(r.Context(), userID, role, templateID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "template deleted"})
}
