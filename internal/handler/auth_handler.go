package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
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
