package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/repository"
)

type ShareHandler struct {
	shareRepo   repository.BookingShareRepository
	frontendURL string
}

func NewShareHandler(shareRepo repository.BookingShareRepository, frontendURL string) *ShareHandler {
	return &ShareHandler{shareRepo: shareRepo, frontendURL: frontendURL}
}

type shareBookingRequest struct {
	BathhouseID string `json:"bathhouse_id"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	GuestCount  int    `json:"guest_count"`
	BookingID   string `json:"booking_id,omitempty"`
}

type shareBookingResponse struct {
	Token    string `json:"token"`
	ShareURL string `json:"share_url"`
}

type resolveShareResponse struct {
	BathhouseID string    `json:"bathhouse_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	GuestCount  int       `json:"guest_count"`
	BookingID   string    `json:"booking_id,omitempty"`
}

// CreateShareLink godoc
// @Summary Create a shareable booking link
// @Tags share
// @Accept json
// @Produce json
// @Param request body shareBookingRequest true "Booking parameters to share"
// @Success 200 {object} APIResponse{data=shareBookingResponse}
// @Failure 400 {object} APIResponse
// @Security BearerAuth
// @Router /api/v1/bookings/share [post]
func (h *ShareHandler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	var req shareBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "Некорректные данные")
		return
	}

	bathhouseID, err := uuid.Parse(req.BathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "Некорректный ID бани")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "Некорректное время начала")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "Некорректное время окончания")
		return
	}

	if req.GuestCount < 1 {
		req.GuestCount = 1
	}

	token, err := generateShareToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "Ошибка генерации ссылки")
		return
	}

	userID := middleware.GetUserID(r.Context())

	share := &domain.BookingShare{
		ID:          uuid.New(),
		Token:       token,
		CreatedBy:   userID,
		BathhouseID: bathhouseID,
		StartTime:   startTime,
		EndTime:     endTime,
		GuestCount:  req.GuestCount,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, 30),
	}

	if req.BookingID != "" {
		bookingID, err := uuid.Parse(req.BookingID)
		if err == nil {
			share.BookingID = &bookingID
		}
	}

	if err := h.shareRepo.Create(r.Context(), share); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "Ошибка сохранения ссылки")
		return
	}

	frontendURL := h.frontendURL
	if frontendURL == "" {
		frontendURL = "https://bani.ru"
	}

	writeJSON(w, http.StatusOK, shareBookingResponse{
		Token:    token,
		ShareURL: frontendURL + "/share/booking/" + token,
	})
}

// ResolveShareLink godoc
// @Summary Resolve a shareable booking link
// @Tags share
// @Produce json
// @Param token path string true "Share token"
// @Success 200 {object} APIResponse{data=resolveShareResponse}
// @Failure 404 {object} APIResponse
// @Router /api/v1/share/booking/{token} [get]
func (h *ShareHandler) ResolveShareLink(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "Токен не указан")
		return
	}

	share, err := h.shareRepo.GetByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Ссылка не найдена")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "Ошибка получения данных")
		return
	}

	if time.Now().After(share.ExpiresAt) {
		writeError(w, http.StatusGone, "expired", "Ссылка истекла")
		return
	}

	resp := resolveShareResponse{
		BathhouseID: share.BathhouseID.String(),
		StartTime:   share.StartTime,
		EndTime:     share.EndTime,
		GuestCount:  share.GuestCount,
	}
	if share.BookingID != nil {
		resp.BookingID = share.BookingID.String()
	}

	writeJSON(w, http.StatusOK, resp)
}

func generateShareToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
