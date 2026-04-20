package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
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

type configuredOAuthProvidersResponse struct {
	Providers []string `json:"providers"`
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

// OAuthRedirect godoc
//
//	@Summary		OAuth redirect
//	@Description	Redirects the user to the OAuth provider's authorization page
//	@Tags			oauth
//	@Param			provider		path	string	true	"OAuth provider (vk, yandex, google)"
//	@Param			referral_code	query	string	false	"Referral code"
//	@Success		302
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Router			/auth/oauth/{provider} [get]
//
// OAuthRedirect redirects the user to the OAuth provider's authorization page.
// GET /api/v1/auth/oauth/{provider}
func (h *OAuthHandler) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := domain.OAuthProvider(chi.URLParam(r, "provider"))
	referralCode := r.URL.Query().Get("referral_code")

	url, err := h.oauthService.GetOAuthURL(provider, referralCode)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

// ListConfiguredProviders godoc
//
//	@Summary		List configured OAuth providers
//	@Description	Returns only the providers that are configured on the backend and safe to show in UI
//	@Tags			oauth
//	@Produce		json
//	@Success		200	{object}	APIResponse{data=configuredOAuthProvidersResponse}
//	@Router			/auth/oauth/providers [get]
func (h *OAuthHandler) ListConfiguredProviders(w http.ResponseWriter, _ *http.Request) {
	configuredProviders := h.oauthService.ListConfiguredProviders()
	items := make([]string, 0, len(configuredProviders))
	for _, provider := range configuredProviders {
		items = append(items, string(provider))
	}

	writeJSON(w, http.StatusOK, configuredOAuthProvidersResponse{
		Providers: items,
	})
}

// OAuthCallback godoc
//
//	@Summary		OAuth callback
//	@Description	Handles the OAuth provider callback, authenticates or registers the user
//	@Tags			oauth
//	@Produce		json
//	@Param			provider	path		string	true	"OAuth provider (vk, yandex, google)"
//	@Param			code		query		string	true	"Authorization code"
//	@Param			state		query		string	true	"State parameter"
//	@Success		200			{object}	APIResponse{data=oauthCallbackResponse}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Router			/auth/oauth/{provider}/callback [get]
//
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

// LinkSocialAccount godoc
//
//	@Summary		Link social account
//	@Description	Links a social account to the authenticated user
//	@Tags			oauth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			provider	path		string				true	"OAuth provider (vk, yandex, google)"
//	@Param			body		body		object{code=string}	true	"Authorization code"
//	@Success		200			{object}	APIResponse{data=object{status=string}}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		409			{object}	APIResponse{error=APIError}
//	@Router			/auth/link/{provider} [post]
//
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

// UnlinkSocialAccount godoc
//
//	@Summary		Unlink social account
//	@Description	Unlinks a social account from the authenticated user
//	@Tags			oauth
//	@Produce		json
//	@Security		BearerAuth
//	@Param			provider	path		string	true	"OAuth provider (vk, yandex, google)"
//	@Success		200			{object}	APIResponse{data=object{status=string}}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		404			{object}	APIResponse{error=APIError}
//	@Router			/auth/link/{provider} [delete]
//
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

// ListSocialAccounts godoc
//
//	@Summary		List linked social accounts
//	@Description	Returns the list of social accounts linked to the authenticated user
//	@Tags			oauth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=[]socialAccountResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/auth/me/social-accounts [get]
//
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
