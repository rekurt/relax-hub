package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
)

// SavedSearchHandler handles recently viewed and saved search endpoints.
type SavedSearchHandler struct {
	savedSearchService service.SavedSearchService
	bathhouseService   service.BathhouseService
}

// NewSavedSearchHandler creates a new SavedSearchHandler.
func NewSavedSearchHandler(
	savedSearchService service.SavedSearchService,
	bathhouseService service.BathhouseService,
) *SavedSearchHandler {
	return &SavedSearchHandler{
		savedSearchService: savedSearchService,
		bathhouseService:   bathhouseService,
	}
}

type recentlyViewedResponse struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Slug     string    `json:"slug"`
	ViewedAt time.Time `json:"viewed_at,omitempty"`
}

// ListRecentlyViewed godoc
//
//	@Summary		List recently viewed bathhouses
//	@Description	Returns the user's recently viewed bathhouses (up to 20)
//	@Tags			saved-searches
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Max results (default 20)"
//	@Success		200		{object}	APIResponse{data=[]recentlyViewedResponse}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/my/recently-viewed [get]
func (h *SavedSearchHandler) ListRecentlyViewed(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed := getPage(v); parsed > 0 && parsed <= 20 {
			limit = parsed
		}
	}

	ids, err := h.savedSearchService.ListRecentlyViewed(r.Context(), userID, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]recentlyViewedResponse, 0, len(ids))
	for _, id := range ids {
		bh, err := h.bathhouseService.GetByID(r.Context(), id)
		if err != nil {
			continue
		}
		items = append(items, recentlyViewedResponse{
			ID:   bh.ID.String(),
			Name: bh.Name,
			Slug: bh.Slug,
		})
	}

	writeJSON(w, http.StatusOK, items)
}

type recordRecentlyViewedRequest struct {
	BathhouseID string `json:"bathhouse_id"`
}

// RecordRecentlyViewed godoc
//
//	@Summary		Record a bathhouse view
//	@Description	Records that the current user viewed a specific bathhouse (for recently-viewed list).
//	@Tags			saved-searches
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		recordRecentlyViewedRequest	true	"Bathhouse ID"
//	@Success		204
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/my/recently-viewed [post]
func (h *SavedSearchHandler) RecordRecentlyViewed(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req recordRecentlyViewedRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	bathhouseID, err := uuid.Parse(req.BathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse_id")
		return
	}

	if err := h.savedSearchService.RecordView(r.Context(), userID, bathhouseID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type savedSearchResponse struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	Name        string          `json:"name"`
	Filters     json.RawMessage `json:"filters"`
	NotifyOnNew bool            `json:"notify_on_new"`
	CreatedAt   time.Time       `json:"created_at"`
}

type createSavedSearchRequest struct {
	Name        string          `json:"name"`
	Filters     json.RawMessage `json:"filters"`
	NotifyOnNew bool            `json:"notify_on_new"`
}

// CreateSavedSearch godoc
//
//	@Summary		Save a search
//	@Description	Saves a search with filters for later use, optionally with notifications for new matches
//	@Tags			saved-searches
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createSavedSearchRequest	true	"Saved search data"
//	@Success		201		{object}	APIResponse{data=savedSearchResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/my/saved-searches [post]
func (h *SavedSearchHandler) CreateSavedSearch(w http.ResponseWriter, r *http.Request) {
	var req createSavedSearchRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid request body")
		return
	}

	if len(req.Filters) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "filters are required")
		return
	}

	userID := middleware.GetUserID(r.Context())

	search, err := h.savedSearchService.SaveSearch(r.Context(), userID, req.Name, req.Filters, req.NotifyOnNew)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, savedSearchResponse{
		ID:          search.ID.String(),
		UserID:      search.UserID.String(),
		Name:        search.Name,
		Filters:     search.Filters,
		NotifyOnNew: search.NotifyOnNew,
		CreatedAt:   search.CreatedAt,
	})
}

// ListSavedSearches godoc
//
//	@Summary		List saved searches
//	@Description	Returns paginated list of the user's saved searches
//	@Tags			saved-searches
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]savedSearchResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/my/saved-searches [get]
func (h *SavedSearchHandler) ListSavedSearches(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.savedSearchService.ListSavedSearches(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]savedSearchResponse, len(result.Items))
	for i, s := range result.Items {
		items[i] = savedSearchResponse{
			ID:          s.ID.String(),
			UserID:      s.UserID.String(),
			Name:        s.Name,
			Filters:     s.Filters,
			NotifyOnNew: s.NotifyOnNew,
			CreatedAt:   s.CreatedAt,
		}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// DeleteSavedSearch godoc
//
//	@Summary		Delete a saved search
//	@Description	Deletes a saved search by ID
//	@Tags			saved-searches
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Saved search ID (UUID)"
//	@Success		200	{object}	APIResponse{data=string}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/my/saved-searches/{id} [delete]
func (h *SavedSearchHandler) DeleteSavedSearch(w http.ResponseWriter, r *http.Request) {
	searchID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid search id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	if err := h.savedSearchService.DeleteSavedSearch(r.Context(), userID, searchID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "deleted")
}
