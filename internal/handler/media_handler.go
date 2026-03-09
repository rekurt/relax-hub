package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/service"
)

type MediaHandler struct {
	mediaService service.MediaService
}

func NewMediaHandler(mediaService service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

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
