package handler

import (
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
)

type widgetKeyResponse struct {
	ApiKey string `json:"api_key"`
}

type widgetCodeRequest struct {
	Color      string `json:"color"`
	FontFamily string `json:"font_family"`
	ShowPrice  bool   `json:"show_price"`
	ShowRating bool   `json:"show_rating"`
	Language   string `json:"language"`
}

type widgetCodeResponse struct {
	Code      string `json:"code"`
	ApiKey    string `json:"api_key"`
	ScriptURL string `json:"script_url"`
	StyleURL  string `json:"style_url"`
}

// @Summary		Get widget API key
// @Description	Get the widget API key for a bathhouse. Owner or representative only.
// @Tags			widgets
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse{data=widgetKeyResponse}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/widget-key [get]
func (h *BathhouseHandler) GetWidgetKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	bathhouseID := chi.URLParam(r, "id")

	id, err := uuid.Parse(bathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	apiKey, err := h.bathhouseService.GetWidgetKey(r.Context(), userID, userRole, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widgetKeyResponse{ApiKey: apiKey})
}

// @Summary		Regenerate widget API key
// @Description	Generate a new widget API key, invalidating the old one.
// @Tags			widgets
// @Produce		json
// @Security		BearerAuth
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse{data=widgetKeyResponse}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		401	{object}	APIResponse{error=APIError}
// @Failure		403	{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/widget-key/regenerate [post]
func (h *BathhouseHandler) RegenerateWidgetKey(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	bathhouseID := chi.URLParam(r, "id")

	id, err := uuid.Parse(bathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	newKey, err := h.bathhouseService.RegenerateWidgetKey(r.Context(), userID, userRole, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, widgetKeyResponse{ApiKey: newKey})
}

// @Summary		Get widget embed code
// @Description	Generate HTML embed code for the booking widget with customization options.
// @Tags			widgets
// @Produce		json
// @Security		BearerAuth
// @Param			id			path		string	true	"Bathhouse ID (UUID)"
// @Param			color		query		string	false	"Widget accent color (#RRGGBB)"	default(#4CAF50)
// @Param			font_family	query		string	false	"Font family for the widget"
// @Param			show_price	query		bool	false	"Show price in widget"
// @Param			show_rating	query		bool	false	"Show rating in widget"
// @Param			language	query		string	false	"Widget language"	default(en)
// @Success		200			{object}	APIResponse{data=widgetCodeResponse}
// @Failure		400			{object}	APIResponse{error=APIError}
// @Failure		401			{object}	APIResponse{error=APIError}
// @Failure		403			{object}	APIResponse{error=APIError}
// @Router			/my/bathhouses/{id}/widget-code [get]
func (h *BathhouseHandler) GetWidgetCode(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	bathhouseID := chi.URLParam(r, "id")

	id, err := uuid.Parse(bathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	// Get the API key (this verifies authorization via the service)
	apiKey, err := h.bathhouseService.GetWidgetKey(r.Context(), userID, userRole, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Parse optional request parameters
	var req widgetCodeRequest
	req.Color = r.URL.Query().Get("color")
	req.FontFamily = r.URL.Query().Get("font_family")
	req.ShowPrice = r.URL.Query().Get("show_price") == "true"
	req.ShowRating = r.URL.Query().Get("show_rating") == "true"
	req.Language = r.URL.Query().Get("language")
	if req.Language == "" {
		req.Language = "en"
	}

	// Validate parameters
	if req.Color != "" && !isValidColor(req.Color) {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid color format")
		return
	}
	if req.FontFamily != "" && !isValidFontFamily(req.FontFamily) {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid font family")
		return
	}

	// Set defaults
	if req.Color == "" {
		req.Color = "#4CAF50"
	}
	if req.FontFamily == "" {
		req.FontFamily = "Helvetica, Arial, sans-serif"
	}

	// Generate HTML embed code with absolute URLs
	code := generateWidgetCode(apiKey, req, r)

	// Get API base URL for response
	scheme := "https"
	if r.Header.Get("X-Forwarded-Proto") != "" {
		scheme = r.Header.Get("X-Forwarded-Proto")
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	apiBaseURL := scheme + "://" + host

	writeJSON(w, http.StatusOK, widgetCodeResponse{
		Code:      code,
		ApiKey:    apiKey,
		ScriptURL: apiBaseURL + "/widget.js",
		StyleURL:  apiBaseURL + "/widget.css",
	})
}

func generateWidgetCode(apiKey string, req widgetCodeRequest, r *http.Request) string {
	// Calculate API base URL
	scheme := "https"
	if r.Header.Get("X-Forwarded-Proto") != "" {
		scheme = r.Header.Get("X-Forwarded-Proto")
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	apiBaseURL := scheme + "://" + host

	// Properly escape all user-supplied parameters to prevent injection
	escapedApiKey := html.EscapeString(apiKey)
	escapedColor := html.EscapeString(req.Color)
	escapedFontFamily := html.EscapeString(req.FontFamily)
	escapedLanguage := html.EscapeString(req.Language)
	escapedApiBaseURL := html.EscapeString(apiBaseURL)

	return `<div id="bani-widget" data-api-key="` + escapedApiKey + `" data-color="` + escapedColor + `" data-font-family="` + escapedFontFamily + `" data-language="` + escapedLanguage + `" data-show-price="` + boolToString(req.ShowPrice) + `" data-show-rating="` + boolToString(req.ShowRating) + `"></div>
<link rel="stylesheet" href="` + escapedApiBaseURL + `/widget.css">
<script src="` + escapedApiBaseURL + `/widget.js"></script>
<script>
  window.BANI_WIDGET_API_URL = '` + strings.ReplaceAll(strings.ReplaceAll(escapedApiBaseURL, `\`, `\\`), `'`, `\'`) + `';
</script>`
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// isValidColor validates that a color is in valid hex format (#RRGGBB)
func isValidColor(color string) bool {
	colorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	return colorRegex.MatchString(color)
}

// isValidFontFamily validates font family to prevent CSS injection
func isValidFontFamily(fontFamily string) bool {
	// Allow comma-separated font families with basic validation
	// Reject if it contains special CSS characters that could be injection vectors
	forbidden := []string{";", "}", "{", "(", ")", "!", "@"}
	for _, char := range forbidden {
		if strings.Contains(fontFamily, char) {
			return false
		}
	}
	// Ensure it's not too long (reasonable max is 256 chars)
	if len(fontFamily) > 256 {
		return false
	}
	return true
}

// @Summary		Get bathhouse SEO meta tags
// @Description	Get SEO meta tags (title, description, og:image, canonical) for a bathhouse.
// @Tags			bathhouses
// @Produce		json
// @Param			id	path		string	true	"Bathhouse ID (UUID)"
// @Success		200	{object}	APIResponse{data=seo.MetaTags}
// @Failure		400	{object}	APIResponse{error=APIError}
// @Failure		404	{object}	APIResponse{error=APIError}
// @Router			/bathhouses/{id}/meta [get]
