package handler

import (
	"net/http"

	"github.com/rekurt/relax-hub/internal/service"
)

type registerPhoneRequest struct {
	Phone        string `json:"phone"`
	Name         string `json:"name"`
	AgeConfirmed bool   `json:"age_confirmed"`
}

type loginPhoneRequest struct {
	Phone string `json:"phone"`
}

type verifyPhoneRequest struct {
	Phone        string `json:"phone"`
	Code         string `json:"code"`
	Name         string `json:"name,omitempty"`
	AgeConfirmed bool   `json:"age_confirmed,omitempty"`
}

type otpSentResponse struct {
	Message string `json:"message"`
}

// RegisterPhone godoc
//
//	@Summary		Register via phone
//	@Description	Registers a new user with phone number and sends OTP
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		registerPhoneRequest	true	"Phone registration data"
//	@Success		200		{object}	APIResponse{data=otpSentResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}
//	@Router			/auth/register-phone [post]
func (h *AuthHandler) RegisterPhone(w http.ResponseWriter, r *http.Request) {
	var req registerPhoneRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	err := h.authService.RegisterPhone(r.Context(), service.RegisterPhoneInput{
		Phone:        req.Phone,
		Name:         req.Name,
		AgeConfirmed: req.AgeConfirmed,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, otpSentResponse{Message: "OTP sent"})
}

// LoginPhone godoc
//
//	@Summary		Login via phone
//	@Description	Sends OTP to an existing user's phone number
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginPhoneRequest	true	"Phone login data"
//	@Success		200		{object}	APIResponse{data=otpSentResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}
//	@Router			/auth/login-phone [post]
func (h *AuthHandler) LoginPhone(w http.ResponseWriter, r *http.Request) {
	var req loginPhoneRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	err := h.authService.LoginPhone(r.Context(), req.Phone)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, otpSentResponse{Message: "OTP sent"})
}

// VerifyPhone godoc
//
//	@Summary		Verify phone OTP
//	@Description	Verifies the OTP code and completes login/registration
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		verifyPhoneRequest	true	"Phone verification data"
//	@Success		200		{object}	APIResponse{data=authResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}
//	@Router			/auth/verify-phone [post]
func (h *AuthHandler) VerifyPhone(w http.ResponseWriter, r *http.Request) {
	var req verifyPhoneRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	result, err := h.authService.VerifyPhone(r.Context(), req.Phone, req.Code, req.Name, req.AgeConfirmed)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := authResponse{
		Token:       result.Token,
		Requires2FA: result.Requires2FA,
	}
	if !result.Requires2FA {
		u := toUserResponse(result.User)
		resp.User = &u
	}
	writeJSON(w, http.StatusOK, resp)
}

