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
	LastMinuteDiscount  int64                  `json:"last_minute_discount,omitempty"`
	AddOnTotal          int64                  `json:"addon_total,omitempty"`
	ServiceFeeAmount    int64                  `json:"service_fee_amount,omitempty"`
	OriginalPrice       int64                  `json:"original_price,omitempty"`
	PromoDiscount       int64                  `json:"promo_discount,omitempty"`
	CertificateDiscount int64                  `json:"certificate_discount,omitempty"`
	CheckedInAt         *time.Time             `json:"checked_in_at,omitempty"`
	CheckedOutAt        *time.Time             `json:"checked_out_at,omitempty"`
	HoldID              string                 `json:"hold_id,omitempty"`
	RejectionReason     string                 `json:"rejection_reason,omitempty"`
	Status              string                 `json:"status"`
	PaymentStatus       string                 `json:"payment_status,omitempty"`
	Comment             string                 `json:"comment"`
	EarnedPoints        int64                  `json:"earned_points,omitempty"`
	LoyaltyDiscount     int64                  `json:"loyalty_discount,omitempty"`
	PointsSpent         int64                  `json:"points_spent,omitempty"`
	ReferralBonusUsed   int64                  `json:"referral_bonus_used,omitempty"`
	IsHolidayPrice      bool                   `json:"is_holiday_price,omitempty"`
	HolidayName         string                 `json:"holiday_name,omitempty"`
	HolidayMultiplier   float64                `json:"holiday_multiplier,omitempty"`
	ModificationCount   int                    `json:"modification_count"`
	DepositAmount       int64                  `json:"deposit_amount,omitempty"`
	DepositStatus       string                 `json:"deposit_status,omitempty"`
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
		LastMinuteDiscount:  b.LastMinuteDiscount,
		AddOnTotal:          b.AddOnTotal,
		ServiceFeeAmount:    b.ServiceFeeAmount,
		CheckedInAt:         b.CheckedInAt,
		CheckedOutAt:        b.CheckedOutAt,
		RejectionReason:     b.RejectionReason,
		Status:              string(b.Status),
		Comment:             b.Comment,
		PointsSpent:         b.PointsSpent,
		ReferralBonusUsed:   b.ReferralBonusUsed,
		ModificationCount:   b.ModificationCount,
		DepositAmount:       b.DepositAmount,
		DepositStatus:       string(b.DepositStatus),
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
	}
}

func toBookingResultResponse(r *service.BookingResult) bookingResponse {
	resp := toBookingResponse(r.Booking)
	if r.Booking.HoldID != nil {
		resp.HoldID = r.Booking.HoldID.String()
	}
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
	resp.LastMinuteDiscount = r.LastMinuteDiscount
	resp.IsHolidayPrice = r.IsHolidayPrice
	resp.HolidayName = r.HolidayName
	resp.HolidayMultiplier = r.HolidayMultiplier
	resp.DepositAmount = r.DepositAmount
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
//
//	@Summary		Create booking
//	@Description	Creates a new booking for a bathhouse. Supports loyalty points, referral bonus, promo codes, and gift certificates as discounts.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		createBookingRequest	true	"Booking data"
//	@Success		201		{object}	APIResponse{data=bookingResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings [post]
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
//
//	@Summary		List my bookings
//	@Description	Returns a paginated list of bookings for the authenticated user
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int	false	"Page number"	default(1)
//	@Param			page_size	query		int	false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]bookingResponse,meta=Meta}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Router			/bookings [get]
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

// GetByID godoc
//
//	@Summary		Get booking by ID
//	@Description	Returns a single booking. Access: admin sees any, client sees their own, owner/representative sees bookings for their bathhouses.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=bookingResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id} [get]
func (h *BookingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	booking, err := h.bookingService.GetByID(r.Context(), userID, role, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := toBookingResponse(booking)
	h.enrichWithPaymentStatus(r.Context(), &resp)
	writeJSON(w, http.StatusOK, resp)
}

type cancelBookingRequest struct {
	RefundTo string `json:"refund_to"` // "wallet" or "card", default "card"
}

// Cancel godoc
//
//	@Summary		Cancel booking
//	@Description	Cancels a booking. Clients can cancel their own bookings, owners/representatives can cancel bookings for their bathhouses. Optional refund_to parameter to choose refund destination.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		cancelBookingRequest	false	"Cancel options"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/cancel [patch]
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req cancelBookingRequest
	// Body is optional for cancel
	_ = readJSON(w, r, &req)

	if req.RefundTo != "" && req.RefundTo != "wallet" && req.RefundTo != "card" {
		writeError(w, http.StatusBadRequest, "invalid_input", "refund_to must be 'wallet' or 'card'")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Cancel(r.Context(), userID, role, bookingID, req.RefundTo); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "cancelled"})
}

// Confirm godoc
//
//	@Summary		Confirm booking
//	@Description	Confirms a pending booking. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/confirm [patch]
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

type rejectBookingRequest struct {
	Reason string `json:"reason"`
}

// Reject godoc
//
//	@Summary		Reject booking
//	@Description	Rejects a pending or pending_owner booking. Only available to bathhouse owners and representatives. Optional rejection reason.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		rejectBookingRequest	false	"Rejection reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/reject [patch]
func (h *BookingHandler) Reject(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req rejectBookingRequest
	// Body is optional for reject
	_ = readJSON(w, r, &req)

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Reject(r.Context(), userID, role, bookingID, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rejected"})
}

// Approve godoc
//
//	@Summary		Approve booking request
//	@Description	Approves a pending_owner booking request. Captures wallet hold. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/approve [patch]
func (h *BookingHandler) Approve(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.Approve(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "approved"})
}

// Complete godoc
//
//	@Summary		Complete booking
//	@Description	Marks a confirmed booking as completed. Awards loyalty points. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=bookingResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/complete [patch]
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
//
//	@Summary		List bathhouse bookings
//	@Description	Returns a paginated list of bookings for a specific bathhouse. Only available to bathhouse owners and representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		string	true	"Bathhouse ID (UUID)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			page_size	query		int		false	"Page size"		default(20)
//	@Success		200			{object}	APIResponse{data=[]bookingResponse,meta=Meta}
//	@Failure		400			{object}	APIResponse{error=APIError}
//	@Failure		401			{object}	APIResponse{error=APIError}
//	@Failure		403			{object}	APIResponse{error=APIError}
//	@Router			/bathhouses/{id}/bookings [get]
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

// CheckIn godoc
//
//	@Summary		Check-in guest
//	@Description	Marks a guest as arrived for a confirmed booking. Available 15 min before to 30 min after booking start time. Only owners/representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/check-in [patch]
func (h *BookingHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.CheckIn(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "checked_in"})
}

// CheckOut godoc
//
//	@Summary		Check-out guest
//	@Description	Marks a guest as departed, completing the booking. Requires prior check-in. Awards loyalty points. Only owners/representatives.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/check-out [patch]
func (h *BookingHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())
	role := middleware.GetUserRole(r.Context())

	if err := h.bookingService.CheckOut(r.Context(), userID, role, bookingID); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "checked_out"})
}

type extendBookingRequest struct {
	ExtraHours    int    `json:"extra_hours"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

type extendBookingResponse struct {
	Booking        bookingResponse `json:"booking"`
	ExtensionPrice int64           `json:"extension_price"`
}

// Extend godoc
//
//	@Summary		Extend booking session
//	@Description	Extends an active booking session by 1-2 hours. Only the booking owner (client) can extend. Booking must be confirmed or checked-in. Extension slots must be available.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		extendBookingRequest	true	"Extension details"
//	@Success		200		{object}	APIResponse{data=extendBookingResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/extend [post]
func (h *BookingHandler) Extend(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req extendBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.ExtraHours < 1 || req.ExtraHours > 2 {
		writeError(w, http.StatusBadRequest, "invalid_input", "extra_hours must be 1 or 2")
		return
	}

	userID := middleware.GetUserID(r.Context())

	result, err := h.bookingService.Extend(r.Context(), userID, bookingID, req.ExtraHours)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := extendBookingResponse{
		Booking:        toBookingResponse(result.Booking),
		ExtensionPrice: result.ExtensionPrice,
	}
	writeJSON(w, http.StatusOK, resp)
}

type rebookAddOnResponse struct {
	AddOnID  string `json:"addon_id"`
	Quantity int    `json:"quantity"`
}

type rebookDataResponse struct {
	BathhouseID   string                `json:"bathhouse_id"`
	DurationHours int                   `json:"duration_hours"`
	TimeFrom      string                `json:"time_from"`
	TimeTo        string                `json:"time_to"`
	GuestCount    int                   `json:"guest_count"`
	AddOns        []rebookAddOnResponse `json:"addons,omitempty"`
}

// GetRebookData godoc
//
//	@Summary		Get re-booking data
//	@Description	Returns pre-filled booking parameters from a past booking (completed or cancelled) for quick re-booking. Prices are not included as they are recalculated at booking time.
//	@Tags			bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Booking ID (UUID)"
//	@Success		200	{object}	APIResponse{data=rebookDataResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/rebook-data [get]
func (h *BookingHandler) GetRebookData(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	userID := middleware.GetUserID(r.Context())

	data, err := h.bookingService.GetRebookData(r.Context(), userID, bookingID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := rebookDataResponse{
		BathhouseID:   data.BathhouseID.String(),
		DurationHours: data.DurationHours,
		TimeFrom:      data.TimeFrom,
		TimeTo:        data.TimeTo,
		GuestCount:    data.GuestCount,
	}
	for _, a := range data.AddOns {
		resp.AddOns = append(resp.AddOns, rebookAddOnResponse{
			AddOnID:  a.AddOnID.String(),
			Quantity: a.Quantity,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

type modifyBookingRequest struct {
	StartTime  string                  `json:"start_time"`
	EndTime    string                  `json:"end_time"`
	GuestCount int                     `json:"guest_count"`
	AddOns     []addOnSelectionRequest `json:"addons,omitempty"`
}

type modifyBookingResponse struct {
	Booking   bookingResponse `json:"booking"`
	OldPrice  int64           `json:"old_price"`
	NewPrice  int64           `json:"new_price"`
	PriceDiff int64           `json:"price_diff"`
}

// Modify godoc
//
//	@Summary		Modify booking
//	@Description	Modifies a pending or confirmed booking. Allows changing date/time, duration, guest count, and add-ons. Maximum 3 modifications per booking. If the new price differs, a refund or additional charge is handled automatically.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		modifyBookingRequest	true	"Modification data"
//	@Success		200		{object}	APIResponse{data=modifyBookingResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Failure		409		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/modify [put]
func (h *BookingHandler) Modify(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req modifyBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
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

	if req.GuestCount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "guest_count must be positive")
		return
	}

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

	userID := middleware.GetUserID(r.Context())

	result, err := h.bookingService.Modify(r.Context(), userID, bookingID, service.ModifyBookingInput{
		StartTime:  startTime,
		EndTime:    endTime,
		GuestCount: req.GuestCount,
		AddOns:     addOnSelections,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := modifyBookingResponse{
		Booking:   toBookingResponse(result.Booking),
		OldPrice:  result.OldPrice,
		NewPrice:  result.NewPrice,
		PriceDiff: result.PriceDiff,
	}
	writeJSON(w, http.StatusOK, resp)
}

type disputeNoShowRequest struct {
	GPSLat  float64 `json:"gps_lat"`
	GPSLon  float64 `json:"gps_lon"`
	Comment string  `json:"comment"`
}

// DisputeNoShow godoc
//
//	@Summary		Dispute no-show
//	@Description	Client disputes a no-show status within 2 hours. Creates a support ticket with GPS coordinates.
//	@Tags			bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Booking ID (UUID)"
//	@Param			body	body		disputeNoShowRequest	true	"Dispute details"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/bookings/{id}/dispute-noshow [post]
func (h *BookingHandler) DisputeNoShow(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req disputeNoShowRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	userID := middleware.GetUserID(r.Context())

	if err := h.bookingService.DisputeNoShow(r.Context(), userID, bookingID, req.GPSLat, req.GPSLon, req.Comment); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "dispute_created"})
}

type adminCancelBookingRequest struct {
	Reason string `json:"reason"`
}

// AdminCancel godoc
//
//	@Summary		Admin cancel booking
//	@Description	Admin cancels any booking regardless of ownership. Issues full wallet refund. Logged to audit.
//	@Tags			admin-bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Booking ID (UUID)"
//	@Param			body	body		adminCancelBookingRequest	true	"Cancel reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/bookings/{id}/cancel [post]
func (h *BookingHandler) AdminCancel(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req adminCancelBookingRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.bookingService.AdminCancel(r.Context(), adminID, bookingID, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "cancelled"})
}

type adminChangeStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// AdminChangeStatus godoc
//
//	@Summary		Admin change booking status
//	@Description	Admin changes the status of any booking. Use for edge cases that normal flows don't cover.
//	@Tags			admin-bookings
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Booking ID (UUID)"
//	@Param			body	body		adminChangeStatusRequest	true	"New status and reason"
//	@Success		200		{object}	APIResponse{data=simpleMessageResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/bookings/{id}/change-status [post]
func (h *BookingHandler) AdminChangeStatus(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid booking id")
		return
	}

	var req adminChangeStatusRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	status := domain.BookingStatus(req.Status)
	if !status.IsValid() {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid status")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "reason is required")
		return
	}

	adminID := middleware.GetUserID(r.Context())

	if err := h.bookingService.AdminChangeStatus(r.Context(), adminID, bookingID, status, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "status_changed"})
}

// AdminListBookings godoc
//
//	@Summary		Admin list bookings
//	@Description	Returns a paginated list of all bookings with optional filters for admin.
//	@Tags			admin-bookings
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id			query		string	false	"Filter by user ID"
//	@Param			bathhouse_id	query		string	false	"Filter by bathhouse ID"
//	@Param			status			query		string	false	"Filter by status"
//	@Param			from_date		query		string	false	"Filter from date (RFC3339)"
//	@Param			to_date			query		string	false	"Filter to date (RFC3339)"
//	@Param			page			query		int		false	"Page number"	default(1)
//	@Param			page_size		query		int		false	"Page size"		default(20)
//	@Success		200				{object}	APIResponse{data=[]bookingResponse,meta=Meta}
//	@Failure		400				{object}	APIResponse{error=APIError}
//	@Failure		401				{object}	APIResponse{error=APIError}
//	@Failure		403				{object}	APIResponse{error=APIError}
//	@Router			/admin/bookings [get]
func (h *BookingHandler) AdminListBookings(w http.ResponseWriter, r *http.Request) {
	filter := domain.AdminBookingFilter{
		Page:     getPage(r.URL.Query().Get("page")),
		PageSize: getPageSize(r.URL.Query().Get("page_size"), 20),
	}

	if uid := r.URL.Query().Get("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid user_id")
			return
		}
		filter.UserID = &id
	}

	if bid := r.URL.Query().Get("bathhouse_id"); bid != "" {
		id, err := uuid.Parse(bid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse_id")
			return
		}
		filter.BathhouseID = &id
	}

	if s := r.URL.Query().Get("status"); s != "" {
		status := domain.BookingStatus(s)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid status")
			return
		}
		filter.Status = &status
	}

	if fd := r.URL.Query().Get("from_date"); fd != "" {
		t, err := time.Parse(time.RFC3339, fd)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid from_date format, use RFC3339")
			return
		}
		filter.FromDate = &t
	}

	if td := r.URL.Query().Get("to_date"); td != "" {
		t, err := time.Parse(time.RFC3339, td)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid to_date format, use RFC3339")
			return
		}
		filter.ToDate = &t
	}

	result, err := h.bookingService.AdminListBookings(r.Context(), filter)
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
