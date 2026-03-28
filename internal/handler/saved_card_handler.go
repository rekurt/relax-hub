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

// SavedCardHandler handles saved card endpoints.
type SavedCardHandler struct {
	savedCardService service.SavedCardService
}

// NewSavedCardHandler creates a new SavedCardHandler.
func NewSavedCardHandler(savedCardService service.SavedCardService) *SavedCardHandler {
	return &SavedCardHandler{savedCardService: savedCardService}
}

type createSavedCardRequest struct {
	ProviderToken string `json:"provider_token"`
	Last4         string `json:"last4"`
	Brand         string `json:"brand"`
	ExpiryMonth   int    `json:"expiry_month"`
	ExpiryYear    int    `json:"expiry_year"`
	IsDefault     bool   `json:"is_default"`
}

type savedCardResponse struct {
	ID          string    `json:"id"`
	Last4       string    `json:"last4"`
	Brand       string    `json:"brand"`
	ExpiryMonth int       `json:"expiry_month"`
	ExpiryYear  int       `json:"expiry_year"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

func toSavedCardResponse(card *domain.SavedCard) savedCardResponse {
	return savedCardResponse{
		ID:          card.ID.String(),
		Last4:       card.Last4,
		Brand:       card.Brand,
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		IsDefault:   card.IsDefault,
		CreatedAt:   card.CreatedAt,
	}
}

// CreateSavedCard godoc
//
//	@Summary		Save a card for future payments
//	@Description	Creates a new saved card token. User can have max 10 saved cards.
//	@Tags			saved-cards
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createSavedCardRequest	true	"Card data"
//	@Success		201		{object}	APIResponse{data=savedCardResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/saved-cards [post]
func (h *SavedCardHandler) CreateSavedCard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req createSavedCardRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	card := &domain.SavedCard{
		ProviderToken: req.ProviderToken,
		Last4:         req.Last4,
		Brand:         req.Brand,
		ExpiryMonth:   req.ExpiryMonth,
		ExpiryYear:    req.ExpiryYear,
		IsDefault:     req.IsDefault,
	}

	result, err := h.savedCardService.CreateSavedCard(r.Context(), userID, card)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toSavedCardResponse(result))
}

// ListSavedCards godoc
//
//	@Summary		List saved cards
//	@Description	Returns paginated list of user's saved cards
//	@Tags			saved-cards
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]savedCardResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/saved-cards [get]
func (h *SavedCardHandler) ListSavedCards(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.savedCardService.ListSavedCards(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]savedCardResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toSavedCardResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// DeleteSavedCard godoc
//
//	@Summary		Delete saved card
//	@Description	Deletes a saved card by ID
//	@Tags			saved-cards
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Card ID (UUID)"
//	@Success		200	{object}	APIResponse{data=string}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/saved-cards/{id} [delete]
func (h *SavedCardHandler) DeleteSavedCard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	cardID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid card id")
		return
	}

	if err := h.savedCardService.DeleteSavedCard(r.Context(), userID, cardID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "deleted")
}

// SetDefaultCard godoc
//
//	@Summary		Set default saved card
//	@Description	Sets a saved card as the default payment method
//	@Tags			saved-cards
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Card ID (UUID)"
//	@Success		200	{object}	APIResponse{data=savedCardResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/saved-cards/{id}/default [post]
func (h *SavedCardHandler) SetDefaultCard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	cardID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid card id")
		return
	}

	if err := h.savedCardService.SetDefaultCard(r.Context(), userID, cardID); err != nil {
		handleServiceError(w, err)
		return
	}

	card, err := h.savedCardService.GetSavedCard(r.Context(), userID, cardID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toSavedCardResponse(card))
}
