package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/middleware"
	"github.com/rekurt/relax-hub/internal/service"
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
	SavedCardID      string                  `json:"saved_card_id,omitempty"`
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

	var savedCardID uuid.UUID
	if req.SavedCardID != "" {
		savedCardID, err = uuid.Parse(req.SavedCardID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid saved_card_id")
			return
		}
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
		SavedCardID:      savedCardID,
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
