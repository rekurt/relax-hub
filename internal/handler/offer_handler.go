package handler

import (
	"net/http"
	"time"

	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

type OfferHandler struct {
	offerService service.OfferService
}

func NewOfferHandler(offerService service.OfferService) *OfferHandler {
	return &OfferHandler{offerService: offerService}
}

type offerAcceptanceResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	OfferVersion string    `json:"offer_version"`
	AcceptedAt   time.Time `json:"accepted_at"`
}

type offerStatusResponse struct {
	Accepted       bool                     `json:"accepted"`
	CurrentVersion string                   `json:"current_version"`
	Acceptance     *offerAcceptanceResponse `json:"acceptance,omitempty"`
}

// AcceptOffer godoc
//
//	@Summary		Accept platform offer
//	@Description	Accept the current version of the platform offer/contract (owner only)
//	@Tags			offer
//	@Produce		json
//	@Security		BearerAuth
//	@Success		201	{object}	APIResponse{data=offerAcceptanceResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/my/offer/accept [post]
func (h *OfferHandler) AcceptOffer(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	acceptance, err := h.offerService.Accept(r.Context(), userID, service.AcceptOfferInput{
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, offerAcceptanceResponse{
		ID:           acceptance.ID.String(),
		UserID:       acceptance.UserID.String(),
		OfferVersion: acceptance.OfferVersion,
		AcceptedAt:   acceptance.AcceptedAt,
	})
}

// GetOfferStatus godoc
//
//	@Summary		Get offer acceptance status
//	@Description	Check if the current user has accepted the latest offer version (owner only)
//	@Tags			offer
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=offerStatusResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/offer/status [get]
func (h *OfferHandler) GetOfferStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	status, err := h.offerService.GetStatus(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := offerStatusResponse{
		Accepted:       status.Accepted,
		CurrentVersion: status.CurrentVersion,
	}
	if status.Acceptance != nil {
		resp.Acceptance = &offerAcceptanceResponse{
			ID:           status.Acceptance.ID.String(),
			UserID:       status.Acceptance.UserID.String(),
			OfferVersion: status.Acceptance.OfferVersion,
			AcceptedAt:   status.Acceptance.AcceptedAt,
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
