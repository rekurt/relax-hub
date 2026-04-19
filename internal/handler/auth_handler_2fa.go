package handler

import (
	"errors"
	"net/http"

	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
)

type totpEnableResponse struct {
	Secret string `json:"secret"`
	QRURL  string `json:"qr_url"`
}

type verify2FARequest struct {
	Code string `json:"code"`
}

type twoFAVerifyLoginRequest struct {
	PartialToken string `json:"partial_token"`
	Code         string `json:"code"`
}

// EnableTOTP godoc
//
//	@Summary		Generate TOTP secret
//	@Description	Generates a TOTP secret and QR code URL for setting up 2FA
//	@Tags			2fa
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=totpEnableResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/auth/2fa/totp/enable [post]
func (h *AuthHandler) EnableTOTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	secret, qrURL, err := h.twoFAService.GenerateTOTPSecret(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, totpEnableResponse{
		Secret: secret,
		QRURL:  qrURL,
	})
}

// VerifyAndActivateTOTP godoc
//
//	@Summary		Verify and activate TOTP
//	@Description	Verifies a TOTP code and activates 2FA for the user
//	@Tags			2fa
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		verify2FARequest	true	"TOTP verification code"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/auth/2fa/totp/verify [post]
func (h *AuthHandler) VerifyAndActivateTOTP(w http.ResponseWriter, r *http.Request) {
	var req verify2FARequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if err := h.twoFAService.EnableTOTP(r.Context(), userID, req.Code); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "TOTP enabled"})
}

// DisableTOTP godoc
//
//	@Summary		Disable TOTP
//	@Description	Disables TOTP 2FA (requires current TOTP code)
//	@Tags			2fa
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		verify2FARequest	true	"Current TOTP code"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/auth/2fa/totp [delete]
func (h *AuthHandler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	var req verify2FARequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if err := h.twoFAService.DisableTOTP(r.Context(), userID, req.Code); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "TOTP disabled"})
}

// EnableSMS2FA godoc
//
//	@Summary		Enable SMS 2FA
//	@Description	Enables SMS-based 2FA (requires verified phone)
//	@Tags			2fa
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/auth/2fa/sms/enable [post]
func (h *AuthHandler) EnableSMS2FA(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.twoFAService.EnableSMS2FA(r.Context(), userID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "SMS 2FA enabled"})
}

// Verify2FALogin godoc
//
//	@Summary		Verify 2FA during login
//	@Description	Verifies the 2FA code using a partial token and returns a full JWT
//	@Tags			2fa
//	@Accept			json
//	@Produce		json
//	@Param			body	body		twoFAVerifyLoginRequest	true	"Partial token and 2FA code"
//	@Success		200		{object}	APIResponse{data=authResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/auth/2fa/verify [post]
func (h *AuthHandler) Verify2FALogin(w http.ResponseWriter, r *http.Request) {
	var req twoFAVerifyLoginRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID, err := h.authService.ParsePartialToken(r.Context(), req.PartialToken)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Determine the user's 2FA method and verify accordingly
	valid, err := h.twoFAService.VerifyTOTP(r.Context(), userID, req.Code)
	if err != nil {
		if errors.Is(err, domain.Err2FANotEnabled) {
			valid, err = h.twoFAService.VerifySMS2FA(r.Context(), userID, req.Code)
			if err != nil {
				handleServiceError(w, err)
				return
			}
		} else {
			handleServiceError(w, err)
			return
		}
	}

	if !valid {
		handleServiceError(w, domain.Err2FAInvalidCode)
		return
	}

	user, token, err := h.authService.Complete2FALogin(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	uReset := toUserResponse(user)
	writeJSON(w, http.StatusOK, authResponse{
		User:  &uReset,
		Token: token,
	})
}

