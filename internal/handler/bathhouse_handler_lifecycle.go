package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
)

func (h *BathhouseHandler) CheckCompleteness(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	result, err := h.bathhouseService.CheckCompleteness(r.Context(), userID, role, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// @Summary		Submit bathhouse for moderation
// @Description	Submit a bathhouse listing for admin moderation. All required fields must be complete.
// @Tags			bathhouses
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/submit [post]
func (h *BathhouseHandler) SubmitForModeration(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.SubmitForModeration(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "submitted for moderation"})
}

// @Summary		Duplicate bathhouse
// @Description	Create a copy of an existing bathhouse listing with draft status.
// @Tags			bathhouses
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		201	{object}	APIResponse{data=bathhouseResponse}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/duplicate [post]
func (h *BathhouseHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	bh, err := h.bathhouseService.DuplicateBathhouse(r.Context(), userID, role, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBathhouseResponse(bh))
}

// @Summary		Deactivate bathhouse
// @Description	Temporarily deactivate a bathhouse. It will be hidden from search but existing bookings are kept.
// @Tags			bathhouses
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse{data=simpleMessageResponse}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/deactivate [post]
func (h *BathhouseHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.DeactivateBathhouse(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse deactivated"})
}

// @Summary		Activate bathhouse
// @Description	Restore a temporarily deactivated bathhouse back to active status.
// @Tags			bathhouses
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse{data=simpleMessageResponse}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/activate [post]
func (h *BathhouseHandler) Activate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.ActivateBathhouse(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse activated"})
}

// @Summary		Archive bathhouse
// @Description	Permanently archive a bathhouse. Only possible if there are no active bookings. Not reversible via API.
// @Tags			bathhouses
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse{data=simpleMessageResponse}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Failure		409	{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/archive [delete]
func (h *BathhouseHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bathhouseService.ArchiveBathhouse(r.Context(), userID, role, id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse archived"})
}

