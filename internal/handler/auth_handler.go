package handler

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
}

func NewAuthHandler(authService service.AuthService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User  userResponse `json:"user"`
	Token string       `json:"token"`
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	CityID    *int64 `json:"city_id"`
}

type updateProfileRequest struct {
	Name   *string `json:"name,omitempty"`
	Phone  *string `json:"phone,omitempty"`
	Bio    *string `json:"bio,omitempty"`
	CityID **int64 `json:"city_id,omitempty"`
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
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.Name,
		Phone:     u.Phone,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		CityID:    u.CityID,
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	user, token, err := h.authService.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Phone:    req.Phone,
		Role:     domain.UserRole(req.Role),
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		User:  toUserResponse(user),
		Token: token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	user, token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		User:  toUserResponse(user),
		Token: token,
	})
}

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
	"image/webp": ".webp",
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	user, err := h.userService.Update(r.Context(), userID, service.UpdateUserInput{
		Name:   req.Name,
		Phone:  req.Phone,
		Bio:    req.Bio,
		CityID: req.CityID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)

	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "avatar file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// Also try to detect from extension
	ext, ok := allowedAvatarTypes[contentType]
	if !ok {
		// Try by file extension
		fileExt := strings.ToLower(filepath.Ext(header.Filename))
		found := false
		for ct, e := range allowedAvatarTypes {
			if e == fileExt {
				contentType = ct
				ext = e
				found = true
				break
			}
		}
		if !found {
			writeError(w, http.StatusBadRequest, "invalid_input", "unsupported image format, use JPEG, PNG or WebP")
			return
		}
	}

	userID := middleware.GetUserID(r.Context())

	user, err := h.userService.UploadAvatar(r.Context(), userID, service.UploadAvatarInput{
		Data:        file,
		ContentType: contentType,
		Ext:         ext,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *AuthHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.userService.DeleteAvatar(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

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
