package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/service"
)

type WidgetHandler struct {
	bathhouseService service.BathhouseService
	bookingService   service.BookingService
	guestUserID      uuid.UUID
	log              *logger.Logger
}

func NewWidgetHandler(
	bathhouseService service.BathhouseService,
	bookingService service.BookingService,
	log *logger.Logger,
) *WidgetHandler {
	// Use a fixed guest user ID for widget bookings
	guestUserID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")

	return &WidgetHandler{
		bathhouseService: bathhouseService,
		bookingService:   bookingService,
		guestUserID:      guestUserID,
		log:              log,
	}
}

type widgetBathhouseResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Address      string   `json:"address"`
	PricePerHour int64    `json:"price_per_hour"`
	MinDuration  int      `json:"min_duration"`
	MaxGuests    int      `json:"max_guests"`
	Rating       float64  `json:"rating"`
	ReviewCount  int      `json:"review_count"`
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

// GetBathhouse godoc
//
//	@Summary		Get bathhouse (widget)
//	@Description	Returns bathhouse info for the embeddable widget. Public endpoint, authenticated via API key.
//	@Tags			widget
//	@Produce		json
//	@Param			api_key	path		string	true	"Widget API key"
//	@Success		200		{object}	APIResponse{data=widgetBathhouseResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (10/s per API key)"
//	@Router			/widget/{api_key}/bathhouse [get]
func (h *WidgetHandler) GetBathhouse(w http.ResponseWriter, r *http.Request) {
	apiKey := chi.URLParam(r, "api_key")
	if apiKey == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "api_key is required")
		return
	}

	bathhouse, err := h.bathhouseService.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
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

// GetAvailableSlots godoc
//
//	@Summary		Get available slots (widget)
//	@Description	Returns available booking slots for a given date. Public endpoint, authenticated via API key.
//	@Tags			widget
//	@Produce		json
//	@Param			api_key	path		string	true	"Widget API key"
//	@Param			date	query		string	true	"Date (YYYY-MM-DD)"
//	@Success		200		{object}	APIResponse{data=[]widgetSlotResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (10/s per API key)"
//	@Router			/widget/{api_key}/slots [get]
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

	bathhouse, err := h.bathhouseService.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
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

// CreateBooking godoc
//
//	@Summary		Create booking (widget)
//	@Description	Creates a booking through the embeddable widget. Requires guest name, phone, and email. Public endpoint, authenticated via API key.
//	@Tags			widget
//	@Accept			json
//	@Produce		json
//	@Param			api_key	path		string					true	"Widget API key"
//	@Param			body	body		widgetBookingRequest	true	"Booking details"
//	@Success		201		{object}	APIResponse{data=widgetBookingResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Failure		429		{object}	APIResponse{error=APIError}	"Rate limited (10/s per API key)"
//	@Router			/widget/{api_key}/booking [post]
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

	bathhouse, err := h.bathhouseService.GetByAPIKey(r.Context(), apiKey)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
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

	result, err := h.bookingService.Create(r.Context(), h.guestUserID, service.CreateBookingInput{
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

	writeJSON(w, http.StatusCreated, toWidgetBookingResponse(result.Booking))
}

// ServeScript godoc
//
//	@Summary		Serve widget JavaScript
//	@Description	Return the embeddable booking widget JavaScript bundle.
//	@Tags			widget
//	@Produce		plain
//	@Success		200	{string}	string	"Widget JavaScript"
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/widget.js [get]
func (h *WidgetHandler) ServeScript(w http.ResponseWriter, r *http.Request) {
	// Use absolute path based on executable location
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to relative path for development
		exePath, _ = os.Getwd()
	}
	execDir := filepath.Dir(exePath)

	// Read the minified JavaScript file
	root, err := os.OpenRoot(execDir)
	if err != nil {
		if h.log != nil {
			h.log.Error("failed to open widget root", "error", err)
		}
		writeError(w, http.StatusNotFound, "not_found", "widget script not found")
		return
	}
	defer root.Close()

	content, err := root.ReadFile("widget/dist/widget.min.js")
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

// ServeStyles godoc
//
//	@Summary		Serve widget CSS
//	@Description	Return the embeddable booking widget stylesheet.
//	@Tags			widget
//	@Produce		plain
//	@Success		200	{string}	string	"Widget CSS"
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/widget.css [get]
func (h *WidgetHandler) ServeStyles(w http.ResponseWriter, r *http.Request) {
	// Use absolute path based on executable location
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to relative path for development
		exePath, _ = os.Getwd()
	}
	execDir := filepath.Dir(exePath)

	// Read the minified CSS file
	root, err := os.OpenRoot(execDir)
	if err != nil {
		if h.log != nil {
			h.log.Error("failed to open widget root", "error", err)
		}
		writeError(w, http.StatusNotFound, "not_found", "widget styles not found")
		return
	}
	defer root.Close()

	content, err := root.ReadFile("widget/dist/widget.min.css")
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
