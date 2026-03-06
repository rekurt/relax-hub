package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository"
	"github.com/nikitaaldaev/bani/internal/service"
)

type WidgetHandler struct {
	bathhouseRepo   repository.BathhouseRepository
	bookingService  service.BookingService
	guestUserID     uuid.UUID
	log             *logger.Logger
}

func NewWidgetHandler(
	bathhouseRepo repository.BathhouseRepository,
	bookingService service.BookingService,
	log *logger.Logger,
) *WidgetHandler {
	// Use a fixed guest user ID for widget bookings
	guestUserID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")

	return &WidgetHandler{
		bathhouseRepo:   bathhouseRepo,
		bookingService:  bookingService,
		guestUserID:     guestUserID,
		log:             log,
	}
}

type widgetBathhouseResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Address      string  `json:"address"`
	PricePerHour int64   `json:"price_per_hour"`
	MinDuration  int     `json:"min_duration"`
	MaxGuests    int     `json:"max_guests"`
	Rating       float64 `json:"rating"`
	ReviewCount  int     `json:"review_count"`
	Images       []string `json:"images"`
}

func toWidgetBathhouseResponse(b *domain.Bathhouse) widgetBathhouseResponse {
	images := b.Images
	if images == nil {
		images = []string{}
	}
	return widgetBathhouseResponse{
		ID:           b.ID.String(),
		Name:         b.Name,
		Description:  b.Description,
		Address:      b.Address,
		PricePerHour: b.PricePerHour,
		MinDuration:  b.MinDuration,
		MaxGuests:    b.MaxGuests,
		Rating:       b.Rating,
		ReviewCount:  b.ReviewCount,
		Images:       images,
	}
}

type widgetSlotResponse struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
	Price     int64     `json:"price"`
}

type widgetBookingRequest struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	GuestCount int    `json:"guest_count"`
	Comment    string `json:"comment"`
}

type widgetBookingResponse struct {
	ID          string    `json:"id"`
	BathhouseID string    `json:"bathhouse_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	GuestCount  int       `json:"guest_count"`
	TotalPrice  int64     `json:"total_price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func toWidgetBookingResponse(b *domain.Booking) widgetBookingResponse {
	return widgetBookingResponse{
		ID:          b.ID.String(),
		BathhouseID: b.BathhouseID.String(),
		StartTime:   b.StartTime,
		EndTime:     b.EndTime,
		GuestCount:  b.GuestCount,
		TotalPrice:  b.TotalPrice,
		Status:      string(b.Status),
		CreatedAt:   b.CreatedAt,
	}
}

func (h *WidgetHandler) GetBathhouse(w http.ResponseWriter, r *http.Request) {
	apiKey := chi.URLParam(r, "api_key")
	if apiKey == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "api_key is required")
		return
	}

	bathhouse, err := h.bathhouseRepo.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		if err == domain.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "bathhouse not found")
		} else {
			h.log.Error("failed to get bathhouse by api_key", "error", err)
			handleServiceError(w, err)
		}
		return
	}

	if bathhouse.Status != domain.BathhouseStatusActive {
		writeError(w, http.StatusNotFound, "bathhouse_not_active", "bathhouse is not active")
		return
	}

	writeJSON(w, http.StatusOK, toWidgetBathhouseResponse(bathhouse))
}

func (h *WidgetHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	apiKey := chi.URLParam(r, "api_key")
	if apiKey == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "api_key is required")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date is required (format: 2006-01-02)")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date format (use 2006-01-02)")
		return
	}

	bathhouse, err := h.bathhouseRepo.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		if err == domain.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "bathhouse not found")
		} else {
			h.log.Error("failed to get bathhouse by api_key", "error", err)
			handleServiceError(w, err)
		}
		return
	}

	if bathhouse.Status != domain.BathhouseStatusActive {
		writeError(w, http.StatusNotFound, "bathhouse_not_active", "bathhouse is not active")
		return
	}

	slots, err := h.bookingService.GetAvailableSlots(r.Context(), bathhouse.ID, date)
	if err != nil {
		h.log.Error("failed to get available slots", "error", err)
		handleServiceError(w, err)
		return
	}

	response := make([]widgetSlotResponse, len(slots))
	for i, slot := range slots {
		response[i] = widgetSlotResponse{
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
			Available: slot.Available,
			Price:     slot.Price,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *WidgetHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	apiKey := chi.URLParam(r, "api_key")
	if apiKey == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "api_key is required")
		return
	}

	var req widgetBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Name == "" || req.Phone == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "name, phone, and email are required")
		return
	}

	if req.GuestCount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "guest_count must be greater than 0")
		return
	}

	if len(req.Comment) > 1000 {
		writeError(w, http.StatusBadRequest, "invalid_input", "comment must not exceed 1000 characters")
		return
	}

	bathhouse, err := h.bathhouseRepo.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		if err == domain.ErrNotFound {
			writeError(w, http.StatusNotFound, "not_found", "bathhouse not found")
		} else {
			h.log.Error("failed to get bathhouse by api_key", "error", err)
			handleServiceError(w, err)
		}
		return
	}

	if bathhouse.Status != domain.BathhouseStatusActive {
		writeError(w, http.StatusNotFound, "bathhouse_not_active", "bathhouse is not active")
		return
	}

	if req.GuestCount > bathhouse.MaxGuests {
		writeError(w, http.StatusBadRequest, "invalid_input", "guest_count exceeds maximum guests allowed")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid start_time format (use RFC3339)")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid end_time format (use RFC3339)")
		return
	}

	booking, err := h.bookingService.Create(r.Context(), h.guestUserID, service.CreateBookingInput{
		BathhouseID: bathhouse.ID,
		StartTime:   startTime,
		EndTime:     endTime,
		GuestCount:  req.GuestCount,
		Comment:     req.Comment,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toWidgetBookingResponse(booking))
}

func (h *WidgetHandler) ServeScript(w http.ResponseWriter, r *http.Request) {
	// Use absolute path based on executable location
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to relative path for development
		exePath, _ = os.Getwd()
	}
	execDir := filepath.Dir(exePath)
	scriptPath := filepath.Join(execDir, "widget/dist/widget.min.js")

	// Read the minified JavaScript file
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		if h.log != nil {
			h.log.Error("failed to read widget script", "error", err)
		}
		writeError(w, http.StatusNotFound, "not_found", "widget script not found")
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(content); err != nil {
		if h.log != nil {
			h.log.Error("failed to write widget script response", "error", err)
		}
	}
}

func (h *WidgetHandler) ServeStyles(w http.ResponseWriter, r *http.Request) {
	// Use absolute path based on executable location
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to relative path for development
		exePath, _ = os.Getwd()
	}
	execDir := filepath.Dir(exePath)
	stylesPath := filepath.Join(execDir, "widget/dist/widget.min.css")

	// Read the minified CSS file
	content, err := os.ReadFile(stylesPath)
	if err != nil {
		if h.log != nil {
			h.log.Error("failed to read widget styles", "error", err)
		}
		writeError(w, http.StatusNotFound, "not_found", "widget styles not found")
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(content); err != nil {
		if h.log != nil {
			h.log.Error("failed to write widget styles response", "error", err)
		}
	}
}
