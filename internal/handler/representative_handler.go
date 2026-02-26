package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type RepresentativeHandler struct {
	repService service.RepresentativeService
}

func NewRepresentativeHandler(repService service.RepresentativeService) *RepresentativeHandler {
	return &RepresentativeHandler{repService: repService}
}

type inviteRepresentativeRequest struct {
	UserEmail string `json:"user_email"`
}

type representativeResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BathhouseID string    `json:"bathhouse_id"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

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
		CreatedAt:   rep.CreatedAt,
	})
}

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
			CreatedAt:   rep.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, items)
}

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
