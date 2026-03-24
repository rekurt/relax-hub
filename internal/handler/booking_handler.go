package handler

import (
	"context"
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
	paymentService service.PaymentService
}

func NewBookingHandler(bookingService service.BookingService, paymentService service.PaymentService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService, paymentService: paymentService}
}

type addOnSelectionRequest struct {
	AddOnID  string `json:"addon_id"`
	Quantity int    `json:"quantity"`
}

type createBookingRequest struct {
	BathhouseID      string                  `json:"bathhouse_id"`
	StartTime        string                  `json:"start_time"`
	EndTime          string                  `json:"end_time"`
	GuestCount       int                     `json:"guest_count"`
	Comment          string                  `json:"comment"`
	UsePoints        int64                   `json:"use_points,omitempty"`
	UseReferralBonus int64                   `json:"use_referral_bonus,omitempty"`
	PromoCode        string                  `json:"promo_code,omitempty"`
	CertificateCode  string                  `json:"certificate_code,omitempty"`
	AddOns           []addOnSelectionRequest `json:"addons,omitempty"`
}

type bookingAddOnResponse struct {
	ID         string `json:"id"`
	AddOnID    string `json:"addon_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	TotalPrice int64  `json:"total_price"`
}

type bookingResponse struct {
	ID                  string                 `json:"id"`
	UserID              string                 `json:"user_id"`
	BathhouseID         string                 `json:"bathhouse_id"`
	StartTime           time.Time              `json:"start_time"`
	EndTime             time.Time              `json:"end_time"`
	GuestCount          int                    `json:"guest_count"`
	TotalPrice          int64                  `json:"total_price"`
	BasePrice           int64                  `json:"base_price,omitempty"`
	LongSessionDiscount int64                  `json:"long_session_discount,omitempty"`
	ExtraGuestSurcharge int64                  `json:"extra_guest_surcharge,omitempty"`
	AddOnTotal          int64                  `json:"addon_total,omitempty"`
	ServiceFeeAmount    int64                  `json:"service_fee_amount,omitempty"`
	OriginalPrice       int64                  `json:"original_price,omitempty"`
	PromoDiscount       int64                  `json:"promo_discount,omitempty"`
	CertificateDiscount int64                  `json:"certificate_discount,omitempty"`
	Status              string                 `json:"status"`
	PaymentStatus       string                 `json:"payment_status,omitempty"`
	Comment             string                 `json:"comment"`
	EarnedPoints        int64                  `json:"earned_points,omitempty"`
	LoyaltyDiscount     int64                  `json:"loyalty_discount,omitempty"`
	PointsSpent         int64                  `json:"points_spent,omitempty"`
	ReferralBonusUsed   int64                  `json:"referral_bonus_used,omitempty"`
	AddOns              []bookingAddOnResponse `json:"addons,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

func toBookingResponse(b *domain.Booking) bookingResponse {
	return bookingResponse{
		ID:                  b.ID.String(),
		UserID:              b.UserID.String(),
		BathhouseID:         b.BathhouseID.String(),
		StartTime:           b.StartTime,
		EndTime:             b.EndTime,
		GuestCount:          b.GuestCount,
		TotalPrice:          b.TotalPrice,
		BasePrice:           b.BasePrice,
		LongSessionDiscount: b.LongSessionDiscount,
		ExtraGuestSurcharge: b.ExtraGuestSurcharge,
		AddOnTotal:          b.AddOnTotal,
		ServiceFeeAmount:    b.ServiceFeeAmount,
		Status:              string(b.Status),
		Comment:             b.Comment,
		PointsSpent:         b.PointsSpent,
		ReferralBonusUsed:   b.ReferralBonusUsed,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
	}
}

func toBookingResultResponse(r *service.BookingResult) bookingResponse {
	resp := toBookingResponse(r.Booking)
	resp.EarnedPoints = r.EarnedPoints
	resp.LoyaltyDiscount = r.LoyaltyDiscount
	resp.PointsSpent = r.PointsSpent
	resp.ReferralBonusUsed = r.ReferralBonusUsed
	resp.OriginalPrice = r.OriginalPrice
	resp.PromoDiscount = r.PromoDiscount
	resp.CertificateDiscount = r.CertificateDiscount
	resp.ServiceFeeAmount = r.ServiceFeeAmount
	resp.BasePrice = r.BasePrice
	resp.LongSessionDiscount = r.LongSessionDiscount
	resp.ExtraGuestSurcharge = r.ExtraGuestSurcharge
	if len(r.AddOns) > 0 {
		resp.AddOns = make([]bookingAddOnResponse, len(r.AddOns))
		for i, a := range r.AddOns {
			resp.AddOns[i] = bookingAddOnResponse{
				ID:         a.ID.String(),
				AddOnID:    a.AddOnID.String(),
				Name:       a.Name,
				Quantity:   a.Quantity,
				UnitPrice:  a.UnitPrice,
				TotalPrice: a.TotalPrice,
			}
		}
	}
	return resp
}

func (h *BookingHandler) enrichWithPaymentStatus(ctx context.Context, resp *bookingResponse) {
	bookingID, err := uuid.Parse(resp.ID)
	if err != nil {
		return
	}
	paymentOwnerID, err := uuid.Parse(resp.UserID)
	if err != nil {
		return
	}
	p, err := h.paymentService.GetPaymentByBooking(ctx, paymentOwnerID, bookingID)
	if err != nil || p == nil {
		return
	}
	resp.PaymentStatus = string(p.Status)
}

// Create godoc
// @Summary      Create booking
// @Description  Creates a new booking for a bathhouse. Supports loyalty points, referral bonus, promo codes, and gift certificates as discounts.
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createBookingRequest  true  "Booking data"
// @Success      201   {object}  APIResponse{data=bookingResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /bookings [post]
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

	var addOnSelections []service.AddOnSelection
	for _, a := range req.AddOns {
		addonID, err := uuid.Parse(a.AddOnID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid addon_id")
			return
		}
		addOnSelections = append(addOnSelections, service.AddOnSelection{
			AddOnID:  addonID,
			Quantity: a.Quantity,
		})
	}

	result, err := h.bookingService.Create(r.Context(), userID, service.CreateBookingInput{
		BathhouseID:      bathhouseID,
		StartTime:        startTime,
		EndTime:          endTime,
		GuestCount:       req.GuestCount,
		Comment:          req.Comment,
		UsePoints:        req.UsePoints,
		UseReferralBonus: req.UseReferralBonus,
		PromoCode:        req.PromoCode,
		CertificateCode:  req.CertificateCode,
		AddOns:           addOnSelections,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBookingResultResponse(result))
}

// ListByUser godoc
// @Summary      List my bookings
// @Description  Returns a paginated list of bookings for the authenticated user
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        page       query     int  false  "Page number"  default(1)
// @Param        page_size  query     int  false  "Page size"    default(20)
// @Success      200        {object}  APIResponse{data=[]bookingResponse,meta=Meta}
// @Failure      401        {object}  APIResponse{error=APIError}
// @Router       /bookings [get]
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
		h.enrichWithPaymentStatus(r.Context(), &items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// Cancel godoc
// @Summary      Cancel booking
// @Description  Cancels a booking. Clients can cancel their own bookings, owners/representatives can cancel bookings for their bathhouses.
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Booking ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /bookings/{id}/cancel [patch]
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

// Confirm godoc
// @Summary      Confirm booking
// @Description  Confirms a pending booking. Only available to bathhouse owners and representatives.
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Booking ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /bookings/{id}/confirm [patch]
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

// Reject godoc
// @Summary      Reject booking
// @Description  Rejects a pending booking. Only available to bathhouse owners and representatives.
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Booking ID (UUID)"
// @Success      200  {object}  APIResponse{data=simpleMessageResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /bookings/{id}/reject [patch]
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

// Complete godoc
// @Summary      Complete booking
// @Description  Marks a confirmed booking as completed. Awards loyalty points. Only available to bathhouse owners and representatives.
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Booking ID (UUID)"
// @Success      200  {object}  APIResponse{data=bookingResponse}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /bookings/{id}/complete [patch]
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

// ListByBathhouse godoc
// @Summary      List bathhouse bookings
// @Description  Returns a paginated list of bookings for a specific bathhouse. Only available to bathhouse owners and representatives.
// @Tags         bookings
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true   "Bathhouse ID (UUID)"
// @Param        page       query     int     false  "Page number"  default(1)
// @Param        page_size  query     int     false  "Page size"    default(20)
// @Success      200        {object}  APIResponse{data=[]bookingResponse,meta=Meta}
// @Failure      400        {object}  APIResponse{error=APIError}
// @Failure      401        {object}  APIResponse{error=APIError}
// @Failure      403        {object}  APIResponse{error=APIError}
// @Router       /bathhouses/{id}/bookings [get]
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
		h.enrichWithPaymentStatus(r.Context(), &items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}
