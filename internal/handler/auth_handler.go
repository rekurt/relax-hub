package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type AuthHandler struct {
	authService            service.AuthService
	userService            service.UserService
	twoFAService           service.TwoFAService
	passwordResetService   service.PasswordResetService
	accountDeletionService service.AccountDeletionService
}

func NewAuthHandler(authService service.AuthService, userService service.UserService, twoFAService service.TwoFAService, passwordResetService service.PasswordResetService, accountDeletionService service.AccountDeletionService) *AuthHandler {
	return &AuthHandler{
		authService:            authService,
		userService:            userService,
		twoFAService:           twoFAService,
		passwordResetService:   passwordResetService,
		accountDeletionService: accountDeletionService,
	}
}

type registerRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Role         string `json:"role"`
	ReferralCode string `json:"referral_code,omitempty"`
	AgeConfirmed bool   `json:"age_confirmed"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User        *userResponse `json:"user,omitempty"`
	Token       string        `json:"token"`
	Requires2FA bool          `json:"requires_2fa,omitempty"`
}

type userResponse struct {
	ID                  string `json:"id"`
	Email               string `json:"email"`
	Name                string `json:"name"`
	Phone               string `json:"phone"`
	Role                string `json:"role"`
	IsActive            bool   `json:"is_active"`
	AvatarURL           string `json:"avatar_url"`
	Bio                 string `json:"bio"`
	CityID              *int64 `json:"city_id"`
	Region              string `json:"region"`
	OnboardingCompleted bool   `json:"onboarding_completed"`
}

type updateProfileRequest struct {
	Name   *string            `json:"name,omitempty"`
	Phone  *string            `json:"phone,omitempty"`
	Bio    *string            `json:"bio,omitempty"`
	CityID nullableInt64Field `json:"city_id" swaggertype:"integer"`
}

// nullableInt64Field distinguishes three JSON states: absent, null, and value.
// Absent: Set=false. Null: Set=true, Value=nil. Value: Set=true, Value=&v.
type nullableInt64Field struct {
	Value *int64
	Set   bool
}

func (n *nullableInt64Field) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		return nil
	}
	n.Value = new(int64)
	return json.Unmarshal(data, n.Value)
}

type publicProfileResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	CityName    string    `json:"city_name"`
	MemberSince time.Time `json:"member_since"`
	ReviewCount int       `json:"review_count"`
	VisitCount  int       `json:"visit_count"`
	AvgRating   float64   `json:"avg_rating"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:                  u.ID.String(),
		Email:               u.Email,
		Name:                u.Name,
		Phone:               u.Phone,
		Role:                string(u.Role),
		IsActive:            u.IsActive,
		AvatarURL:           u.AvatarURL,
		Bio:                 u.Bio,
		CityID:              u.CityID,
		Region:              string(u.Region),
		OnboardingCompleted: u.OnboardingCompleted,
	}
}

func toPublicProfileResponse(p *domain.UserProfile) publicProfileResponse {
	return publicProfileResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		AvatarURL:   p.AvatarURL,
		Bio:         p.Bio,
		CityName:    p.CityName,
		MemberSince: p.MemberSince,
		ReviewCount: p.ReviewCount,
		VisitCount:  p.VisitCount,
		AvgRating:   p.AvgRating,
	}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new user account with the given credentials and role
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		registerRequest	true	"Registration data"
//	@Success		201		{object}	APIResponse{data=authResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (5/min)"
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	user, token, err := h.authService.Register(r.Context(), service.RegisterInput{
		Email:        req.Email,
		Password:     req.Password,
		Name:         req.Name,
		Phone:        req.Phone,
		Role:         domain.UserRole(req.Role),
		ReferralCode: req.ReferralCode,
		AgeConfirmed: req.AgeConfirmed,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	u := toUserResponse(user)
	writeJSON(w, http.StatusCreated, authResponse{
		User:  &u,
		Token: token,
	})
}

// Login godoc
//
//	@Summary		Login
//	@Description	Authenticates a user and returns a JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"Login credentials"
//	@Success		200		{object}	APIResponse{data=authResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (10/min)"
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	result, err := h.authService.Login(r.Context(), req.Email, req.Password)
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

// Me godoc
//
//	@Summary		Get current user
//	@Description	Returns the authenticated user's profile
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=userResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

const maxAvatarSize = 5 << 20 // 5 MB

var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Updates the authenticated user's profile fields
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		updateProfileRequest	true	"Profile fields to update"
//	@Success		200		{object}	APIResponse{data=userResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/auth/me [put]
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	input := service.UpdateUserInput{
		Name:  req.Name,
		Phone: req.Phone,
		Bio:   req.Bio,
	}
	if req.CityID.Set {
		input.CityID = &req.CityID.Value
	}

	user, err := h.userService.Update(r.Context(), userID, input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// UploadAvatar godoc
//
//	@Summary		Upload avatar
//	@Description	Uploads a new avatar image for the authenticated user (JPEG or PNG, max 5MB)
//	@Tags			auth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			avatar	formData	file	true	"Avatar image file"
//	@Success		200		{object}	APIResponse{data=userResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Router			/auth/me/avatar [post]
func (h *AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)

	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "avatar file is required")
		return
	}
	defer file.Close()

	// Detect actual content type from file bytes (not trusting client headers)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_input", "failed to read avatar file")
		return
	}
	if n == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "empty avatar file")
		return
	}
	detectedType := http.DetectContentType(buf[:n])
	ext, ok := allowedAvatarTypes[detectedType]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_input", "unsupported image format, use JPEG or PNG")
		return
	}

	// Reconstruct the reader with the already-read bytes prepended
	data := io.MultiReader(bytes.NewReader(buf[:n]), file)

	userID := middleware.GetUserID(r.Context())

	user, err := h.userService.UploadAvatar(r.Context(), userID, service.UploadAvatarInput{
		Data:        data,
		ContentType: detectedType,
		Ext:         ext,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// DeleteAvatar godoc
//
//	@Summary		Delete avatar
//	@Description	Removes the authenticated user's avatar
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=userResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/auth/me/avatar [delete]
func (h *AuthHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.userService.DeleteAvatar(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// GetPublicProfile godoc
//
//	@Summary		Get public user profile
//	@Description	Returns a user's public profile by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID (UUID)"
//	@Success		200	{object}	APIResponse{data=publicProfileResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/users/{id}/profile [get]
func (h *AuthHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}

	profile, err := h.userService.GetPublicProfile(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toPublicProfileResponse(profile))
}

// GetMyStats godoc
//
//	@Summary		Get my statistics
//	@Description	Returns booking and review statistics for the authenticated user
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=myStatsResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Router			/my/stats [get]
func (h *AuthHandler) GetMyStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	stats, err := h.userService.GetMyStats(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

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
//	@Success		200	{object}	APIResponse{data=service.ProfileCompletenessOutput}
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
