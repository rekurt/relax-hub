package handler

import (
	"net/http"

	"github.com/rekurt/relax-hub/internal/middleware"
)

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ForgotPassword godoc
//
//	@Summary		Request password reset
//	@Description	Sends a password reset link to the user's email
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		forgotPasswordRequest	true	"Email address"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (3/15min)"
//	@Router			/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.passwordResetService.ForgotPassword(r.Context(), req.Email); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "If the email exists, a reset link has been sent"})
}

// ResetPassword godoc
//
//	@Summary		Reset password with token
//	@Description	Resets the user's password using a reset token from email
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		resetPasswordRequest	true	"Reset token and new password"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Router			/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.passwordResetService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "Password has been reset successfully"})
}

// DeleteAccount godoc
//
//	@Summary		Request account deletion
//	@Description	Requests account deletion with a 30-day grace period
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		409	{object}	APIResponse{error=APIError}
//	@Router			/auth/delete-account [post]
func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.accountDeletionService.RequestDeletion(r.Context(), userID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "Account deletion scheduled. You have 30 days to restore your account."})
}

// RestoreAccount godoc
//
//	@Summary		Restore account during grace period
//	@Description	Cancels a pending account deletion during the 30-day grace period
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/auth/restore-account [post]
func (h *AuthHandler) RestoreAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.accountDeletionService.RestoreAccount(r.Context(), userID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "Account deletion cancelled. Your account has been restored."})
}

// GetProfileCompleteness godoc
//
//	@Summary		Get profile completeness
//	@Description	Returns profile completeness percentage and individual field statuses
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=profileCompletenessResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/profile-completeness [get]
func (h *AuthHandler) GetProfileCompleteness(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	result, err := h.userService.GetProfileCompleteness(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CompleteOnboarding godoc
//
//	@Summary		Complete onboarding tour
//	@Description	Marks the onboarding tour as completed for the authenticated user
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/onboarding/complete [post]
func (h *AuthHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.userService.CompleteOnboarding(r.Context(), userID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, simpleMessageResponse{Message: "Onboarding completed"})
}
