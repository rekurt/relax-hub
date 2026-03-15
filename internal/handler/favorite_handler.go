package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type FavoriteHandler struct {
	favoriteService service.FavoriteService
}

func NewFavoriteHandler(favoriteService service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

type favoriteResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	BathhouseID string    `json:"bathhouse_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Toggle godoc
// @Summary      Toggle favorite
// @Description  Adds or removes a bathhouse from the user's favorites
// @Tags         favorites
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Bathhouse ID (UUID)"
// @Success      200  {object}  APIResponse{data=toggleFavoriteResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id}/favorite [post]
func (h *FavoriteHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	isFavorite, err := h.favoriteService.Toggle(r.Context(), userID, bathhouseID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"is_favorite": isFavorite})
}

// List godoc
// @Summary      List favorites
// @Description  Returns paginated list of the user's favorite bathhouses
// @Tags         favorites
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int  false  "Page number"  default(1)
// @Param        page_size  query     int  false  "Page size"    default(20)
// @Success      200  {object}  APIResponse{data=[]favoriteResponse,meta=Meta}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Router       /my/favorites [get]
func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.favoriteService.List(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]favoriteResponse, len(result.Items))
	for i, fav := range result.Items {
		items[i] = favoriteResponse{
			ID:          fav.ID.String(),
			UserID:      fav.UserID.String(),
			BathhouseID: fav.BathhouseID.String(),
			CreatedAt:   fav.CreatedAt,
		}
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
