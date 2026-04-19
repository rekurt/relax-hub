package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/service"
)

type MediaHandler struct {
	mediaService service.MediaService
}

func NewMediaHandler(mediaService service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// BathhouseGallery godoc
//
//	@Summary		Get bathhouse gallery
//	@Description	Returns a paginated list of review media (photos/videos) for a bathhouse, aggregated from approved reviews
//	@Tags			review-media
//	@Produce		json
//	@Param			id			path		string	true	"Bathhouse ID (UUID)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]mediaResponse,meta=Meta}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/gallery [get]
func (h *MediaHandler) BathhouseGallery(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.mediaService.ListByBathhouse(r.Context(), bathhouseID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := toMediaResponses(result.Items)

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
