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

type RepresentativeHandler struct {
	repService service.RepresentativeService
}

func NewRepresentativeHandler(repService service.RepresentativeService) *RepresentativeHandler {
	return &RepresentativeHandler{repService: repService}
}

type inviteRepresentativeRequest struct {
	UserEmail string `json:"user_email"`
	Role      string `json:"role,omitempty"`
}

type representativeResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BathhouseID string    `json:"bathhouse_id"`
	OwnerID     string    `json:"owner_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

// Invite godoc
//
//	@Summary		Invite representative
//	@Description	Invites a user as a representative for a bathhouse by email. Owner only.
//	@Tags			representatives
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Bathhouse ID (UUID)"
//	@Param			body	body		inviteRepresentativeRequest	true	"User email to invite"
//	@Success		201		{object}	APIResponse{data=representativeResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/representatives [post]
func (h *RepresentativeHandler) Invite(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	var req inviteRepresentativeRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	ownerID := middleware.GetUserID(r.Context())

	rep, err := h.repService.Invite(r.Context(), ownerID, service.InviteRepresentativeInput{
		UserEmail:   req.UserEmail,
		BathhouseID: bathhouseID,
		Role:        domain.RepresentativeRole(req.Role),
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, representativeResponse{
		ID:          rep.ID.String(),
		UserID:      rep.UserID.String(),
		BathhouseID: rep.BathhouseID.String(),
		OwnerID:     rep.OwnerID.String(),
		Role:        string(rep.Role),
		CreatedAt:   rep.CreatedAt,
	})
}

// ListByBathhouse godoc
//
//	@Summary		List representatives
//	@Description	Returns all representatives for a bathhouse. Owner only.
//	@Tags			representatives
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Bathhouse ID (UUID)"
//	@Success		200	{object}	APIResponse{data=[]representativeResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/representatives [get]
func (h *RepresentativeHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	ownerID := middleware.GetUserID(r.Context())

	reps, err := h.repService.ListByBathhouse(r.Context(), ownerID, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]representativeResponse, len(reps))
	for i, rep := range reps {
		items[i] = representativeResponse{
			ID:          rep.ID.String(),
			UserID:      rep.UserID.String(),
			BathhouseID: rep.BathhouseID.String(),
			OwnerID:     rep.OwnerID.String(),
			Role:        string(rep.Role),
			CreatedAt:   rep.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, items)
}

// Revoke godoc
//
//	@Summary		Revoke representative
//	@Description	Revokes a representative's access to a bathhouse. Owner only.
//	@Tags			representatives
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Representative ID (UUID)"
//	@Success		200	{object}	APIResponse
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/representatives/{id} [delete]
func (h *RepresentativeHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	repID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid representative id")
		return
	}

	ownerID := middleware.GetUserID(r.Context())

	if err := h.repService.Revoke(r.Context(), ownerID, repID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "revoked"})
}
