package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

type createBookingRequest struct {
	BathhouseID string `json:"bathhouse_id"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	GuestCount  int    `json:"guest_count"`
	Comment     string `json:"comment"`
	UsePoints   int64  `json:"use_points,omitempty"`
}

type bookingResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	BathhouseID     string    `json:"bathhouse_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	GuestCount      int       `json:"guest_count"`
	TotalPrice      int64     `json:"total_price"`
	Status          string    `json:"status"`
	Comment         string    `json:"comment"`
	EarnedPoints    int64     `json:"earned_points,omitempty"`
	LoyaltyDiscount int64     `json:"loyalty_discount,omitempty"`
	PointsSpent     int64     `json:"points_spent,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toBookingResponse(b *domain.Booking) bookingResponse {
	return bookingResponse{
		ID:          b.ID.String(),
		UserID:      b.UserID.String(),
		BathhouseID: b.BathhouseID.String(),
		StartTime:   b.StartTime,
		EndTime:     b.EndTime,
		GuestCount:  b.GuestCount,
		TotalPrice:  b.TotalPrice,
		Status:      string(b.Status),
		Comment:     b.Comment,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func toBookingResultResponse(r *service.BookingResult) bookingResponse {
	resp := toBookingResponse(r.Booking)
	resp.EarnedPoints = r.EarnedPoints
	resp.LoyaltyDiscount = r.LoyaltyDiscount
	resp.PointsSpent = r.PointsSpent
	return resp
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	bathhouseID, err := uuid.Parse(req.BathhouseID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse_id")
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid start_time format, use RFC3339")
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid end_time format, use RFC3339")
		return
	}

	userID := middleware.GetUserID(r.Context())

	result, err := h.bookingService.Create(r.Context(), userID, service.CreateBookingInput{
		BathhouseID: bathhouseID,
		StartTime:   startTime,
		EndTime:     endTime,
		GuestCount:  req.GuestCount,
		Comment:     req.Comment,
		UsePoints:   req.UsePoints,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBookingResultResponse(result))
}

func (h *BookingHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.bookingService.ListByUser(r.Context(), userID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bookingResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBookingResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Cancel(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "cancelled"})
}

func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Confirm(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "confirmed"})
}

func (h *BookingHandler) Reject(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Reject(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rejected"})
}

func (h *BookingHandler) Complete(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	result, err := h.bookingService.Complete(r.Context(), userID, role, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBookingResultResponse(result))
}

func (h *BookingHandler) ListByBathhouse(w http.ResponseWriter, r *http.Request) {
	bathhouseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.bookingService.ListByBathhouse(r.Context(), userID, role, bathhouseID, page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bookingResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBookingResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
