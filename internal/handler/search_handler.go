package handler

import (
	"net/http"
	"strconv"

	"github.com/nikitaaldaev/bani/internal/service"
)

// SearchHandler handles search-related endpoints.
type SearchHandler struct {
	suggestionService service.SearchSuggestionService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(suggestionService service.SearchSuggestionService) *SearchHandler {
	return &SearchHandler{suggestionService: suggestionService}
}

type suggestionResponse struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// GetSuggestions godoc
// @Summary      Get search suggestions
// @Description  Returns autocomplete suggestions for the given query, including bathhouse names, city names, and popular searches
// @Tags         search
// @Produce      json
// @Param        q      query    string  true   "Search query (min 2 characters)"
// @Param        limit  query    int     false  "Max results (default 10, max 20)"
// @Success      200    {object} APIResponse{data=[]suggestionResponse}
// @Router       /search/suggestions [get]
func (h *SearchHandler) GetSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusOK, []suggestionResponse{})
		return
	}

	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 20 {
		limit = 20
	}

	suggestions, err := h.suggestionService.GetSuggestions(r.Context(), query, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	result := make([]suggestionResponse, len(suggestions))
	for i, s := range suggestions {
		result[i] = suggestionResponse{
			Text: s.Text,
			Type: string(s.Type),
		}
	}

	writeJSON(w, http.StatusOK, result)
}
