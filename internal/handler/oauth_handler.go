package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type OAuthHandler struct {
	oauthService service.OAuthService
}

func NewOAuthHandler(oauthService service.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthService: oauthService}
}

type oauthCallbackResponse struct {
	User  userResponse `json:"user"`
	Token string       `json:"token"`
}

type socialAccountResponse struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	LinkedAt  time.Time `json:"linked_at"`
}

func toSocialAccountResponse(sa *domain.SocialAccount) socialAccountResponse {
	return socialAccountResponse{
		ID:        sa.ID.String(),
		Provider:  string(sa.Provider),
		Email:     sa.Email,
		Name:      sa.Name,
		AvatarURL: sa.AvatarURL,
		LinkedAt:  sa.LinkedAt,
	}
}

func toSocialAccountListResponse(accounts []domain.SocialAccount) []socialAccountResponse {
	result := make([]socialAccountResponse, len(accounts))
	for i := range accounts {
		result[i] = toSocialAccountResponse(&accounts[i])
	}
	return result
}

// OAuthRedirect redirects the user to the OAuth provider's authorization page.
// GET /api/v1/auth/oauth/{provider}
func (h *OAuthHandler) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := domain.OAuthProvider(chi.URLParam(r, "provider"))

	url, err := h.oauthService.GetOAuthURL(provider)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

// OAuthCallback handles the OAuth provider callback with an authorization code.
// GET /api/v1/auth/oauth/{provider}/callback?code=...&state=...
func (h *OAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := domain.OAuthProvider(chi.URLParam(r, "provider"))

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		desc := r.URL.Query().Get("error_description")
		if desc == "" {
			desc = errParam
		}
		writeError(w, http.StatusBadRequest, "oauth_error", desc)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "authorization code is required")
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "state parameter is required")
		return
	}

	user, token, err := h.oauthService.OAuthCallback(r.Context(), provider, code, state)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, oauthCallbackResponse{
		User:  toUserResponse(user),
		Token: token,
	})
}

// LinkSocialAccount links a social account to the authenticated user.
// POST /api/v1/auth/link/{provider}
func (h *OAuthHandler) LinkSocialAccount(w http.ResponseWriter, r *http.Request) {
	provider := domain.OAuthProvider(chi.URLParam(r, "provider"))
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Code string `json:"code"`
	}
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "authorization code is required")
		return
	}

	if err := h.oauthService.LinkSocialAccount(r.Context(), userID, provider, req.Code); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "linked"})
}

// UnlinkSocialAccount unlinks a social account from the authenticated user.
// DELETE /api/v1/auth/link/{provider}
func (h *OAuthHandler) UnlinkSocialAccount(w http.ResponseWriter, r *http.Request) {
	provider := domain.OAuthProvider(chi.URLParam(r, "provider"))
	userID := middleware.GetUserID(r.Context())

	if err := h.oauthService.UnlinkSocialAccount(r.Context(), userID, provider); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "unlinked"})
}

// ListSocialAccounts returns the list of linked social accounts for the authenticated user.
// GET /api/v1/auth/me/social-accounts
func (h *OAuthHandler) ListSocialAccounts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	accounts, err := h.oauthService.ListSocialAccounts(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toSocialAccountListResponse(accounts))
}
