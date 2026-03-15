package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/middleware"
	"github.com/nikitaaldaev/bani/internal/service"
)

type AdminHandler struct {
	userService      service.UserService
	bathhouseService service.BathhouseService
	cityService      service.CityService
	reviewService    service.ReviewService
	notifService     service.NotificationService
}

func NewAdminHandler(
	userService service.UserService,
	bathhouseService service.BathhouseService,
	cityService service.CityService,
	reviewService service.ReviewService,
	notifService service.NotificationService,
) *AdminHandler {
	return &AdminHandler{
		userService:      userService,
		bathhouseService: bathhouseService,
		cityService:      cityService,
		reviewService:    reviewService,
		notifService:     notifService,
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	result, err := h.userService.List(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]userResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toUserResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *AdminHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}

	adminID := middleware.GetUserID(r.Context())
	if id == adminID {
		writeError(w, http.StatusBadRequest, "invalid_input", "cannot block yourself")
		return
	}

	if err := h.userService.Block(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user blocked"})
}

func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid user id")
		return
	}

	if err := h.userService.Unblock(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user unblocked"})
}

func (h *AdminHandler) ApproveBathhouse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	if err := h.bathhouseService.Approve(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse approved"})
}

func (h *AdminHandler) RejectBathhouse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse id")
		return
	}

	if err := h.bathhouseService.Reject(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "bathhouse rejected"})
}

func (h *AdminHandler) ListBathhouses(w http.ResponseWriter, r *http.Request) {
	filter := domain.BathhouseFilter{
		Page:     getPage(r.URL.Query().Get("page")),
		PageSize: getPageSize(r.URL.Query().Get("page_size"), 20),
	}

	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.BathhouseStatus(v)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid status value")
			return
		}
		filter.Status = &status
	} else {
		filter.ShowAllStatuses = true
	}

	result, err := h.bathhouseService.Search(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]bathhouseResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toBathhouseResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

type createCityRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type updateCityRequest struct {
	Name      *string  `json:"name"`
	Slug      *string  `json:"slug"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// @Summary      Create city
// @Description  Create a new city. Admin only.
// @Tags         admin-cities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createCityRequest  true  "City data"
// @Success      201   {object}  APIResponse{data=cityResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      409   {object}  APIResponse{error=APIError}
// @Router       /admin/cities [post]
func (h *AdminHandler) CreateCity(w http.ResponseWriter, r *http.Request) {
	var req createCityRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	city, err := h.cityService.Create(r.Context(), service.CreateCityInput{
		Name:      req.Name,
		Slug:      req.Slug,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, cityResponse{
		ID:        city.ID,
		Name:      city.Name,
		Slug:      city.Slug,
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
	})
}

// @Summary      Update city
// @Description  Update an existing city. Admin only.
// @Tags         admin-cities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                true  "City ID"
// @Param        body  body      updateCityRequest  true  "Fields to update"
// @Success      200   {object}  APIResponse{data=cityResponse}
// @Failure      400   {object}  APIResponse{error=APIError}
// @Failure      401   {object}  APIResponse{error=APIError}
// @Failure      403   {object}  APIResponse{error=APIError}
// @Failure      404   {object}  APIResponse{error=APIError}
// @Router       /admin/cities/{id} [put]
func (h *AdminHandler) UpdateCity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid city id")
		return
	}

	var req updateCityRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	city, err := h.cityService.Update(r.Context(), id, service.UpdateCityInput{
		Name:      req.Name,
		Slug:      req.Slug,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cityResponse{
		ID:        city.ID,
		Name:      city.Name,
		Slug:      city.Slug,
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
	})
}

// @Summary      Delete city
// @Description  Delete a city. Admin only.
// @Tags         admin-cities
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "City ID"
// @Success      200  {object}  APIResponse
// @Failure      400  {object}  APIResponse{error=APIError}
// @Failure      401  {object}  APIResponse{error=APIError}
// @Failure      403  {object}  APIResponse{error=APIError}
// @Failure      404  {object}  APIResponse{error=APIError}
// @Router       /admin/cities/{id} [delete]
func (h *AdminHandler) DeleteCity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid city id")
		return
	}

	if err := h.cityService.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "city deleted"})
}

// Review Moderation Handlers

type adminReviewResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	BathhouseID      string     `json:"bathhouse_id"`
	BookingID        string     `json:"booking_id"`
	Rating           int        `json:"rating"`
	Text             string     `json:"text"`
	Status           string     `json:"status"`
	RejectionReasons []string   `json:"rejection_reasons,omitempty"`
	OwnerResponse    string     `json:"owner_response,omitempty"`
	OwnerResponseAt  *time.Time `json:"owner_response_at,omitempty"`
	Images           []string   `json:"images,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func toAdminReviewResponse(rev *domain.Review) adminReviewResponse {
	return adminReviewResponse{
		ID:               rev.ID.String(),
		UserID:           rev.UserID.String(),
		BathhouseID:      rev.BathhouseID.String(),
		BookingID:        rev.BookingID.String(),
		Rating:           rev.Rating,
		Text:             rev.Text,
		Status:           string(rev.Status),
		RejectionReasons: rev.RejectionReasons,
		OwnerResponse:    rev.OwnerResponse,
		OwnerResponseAt:  rev.OwnerResponseAt,
		Images:           rev.Images,
		CreatedAt:        rev.CreatedAt,
		UpdatedAt:        rev.UpdatedAt,
	}
}

func (h *AdminHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	page := getPage(r.URL.Query().Get("page"))
	pageSize := getPageSize(r.URL.Query().Get("page_size"), 20)

	filter := domain.AdminReviewFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if v := r.URL.Query().Get("status"); v != "" {
		status := domain.ReviewStatus(v)
		if !status.IsValid() {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid status value")
			return
		}
		filter.Status = &status
	}

	if v := r.URL.Query().Get("bathhouse_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid bathhouse_id")
			return
		}
		filter.BathhouseID = &id
	}

	if v := r.URL.Query().Get("min_rating"); v != "" {
		if minRating, err := strconv.Atoi(v); err == nil && minRating > 0 {
			filter.MinRating = &minRating
		}
	}

	if v := r.URL.Query().Get("max_rating"); v != "" {
		if maxRating, err := strconv.Atoi(v); err == nil && maxRating > 0 {
			filter.MaxRating = &maxRating
		}
	}

	if v := r.URL.Query().Get("from_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			filter.FromDate = &t
		}
	}

	if v := r.URL.Query().Get("to_date"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			filter.ToDate = &t
		}
	}

	result, err := h.reviewService.ListAllReviews(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items := make([]adminReviewResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toAdminReviewResponse(&result.Items[i])
	}

	writeJSONWithMeta(w, http.StatusOK, items, &Meta{
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *AdminHandler) GetPendingCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.reviewService.CountPendingReviews(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"pending_count": count})
}

type approveRejectRequest struct {
	Reason string `json:"reason,omitempty"`
}

func (h *AdminHandler) ApproveReview(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review id")
		return
	}

	review, err := h.reviewService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	if err := h.reviewService.UpdateStatus(r.Context(), id, domain.ReviewStatusApproved); err != nil {
		handleServiceError(w, err)
		return
	}

	review.Status = domain.ReviewStatusApproved

	// Notify review author about approval
	data := map[string]string{
		"review_id":    review.ID.String(),
		"bathhouse_id": review.BathhouseID.String(),
	}
	_ = h.notifService.Send(r.Context(), review.UserID, domain.NotifReviewApproved,
		"Ваш отзыв одобрен",
		"Ваш отзыв успешно прошел модерацию и опубликован",
		data,
	)

	writeJSON(w, http.StatusOK, toAdminReviewResponse(review))
}

func (h *AdminHandler) RejectReview(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid review id")
		return
	}

	var req approveRejectRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	review, err := h.reviewService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	var reasons []string
	if req.Reason != "" {
		reasons = []string{req.Reason}
	}

	if err := h.reviewService.UpdateStatusWithReasons(r.Context(), id, domain.ReviewStatusRejected, reasons); err != nil {
		handleServiceError(w, err)
		return
	}

	review.Status = domain.ReviewStatusRejected
	review.RejectionReasons = reasons

	// Notify review author about rejection with reason
	data := map[string]string{
		"review_id":    review.ID.String(),
		"bathhouse_id": review.BathhouseID.String(),
	}
	reasonText := "Ваш отзыв не соответствует политике платформы"
	if req.Reason != "" {
		reasonText = req.Reason
	}
	_ = h.notifService.Send(r.Context(), review.UserID, domain.NotifReviewRejected,
		"Ваш отзыв отклонен",
		reasonText,
		data,
	)

	writeJSON(w, http.StatusOK, toAdminReviewResponse(review))
}

type batchIDsRequest struct {
	IDs []string `json:"ids"`
}

type batchResult struct {
	Successful int `json:"successful"`
	Failed     int `json:"failed"`
}

func (h *AdminHandler) BatchApproveReviews(w http.ResponseWriter, r *http.Request) {
	var req batchIDsRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "ids cannot be empty")
		return
	}

	if len(req.IDs) > 100 {
		writeError(w, http.StatusBadRequest, "invalid_input", "batch size cannot exceed 100")
		return
	}

	result := batchResult{}
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			result.Failed++
			continue
		}

		review, err := h.reviewService.GetByID(r.Context(), id)
		if err != nil {
			result.Failed++
			continue
		}

		if err := h.reviewService.UpdateStatus(r.Context(), id, domain.ReviewStatusApproved); err != nil {
			result.Failed++
			continue
		}

		// Notify review author about approval
		data := map[string]string{
			"review_id":    review.ID.String(),
			"bathhouse_id": review.BathhouseID.String(),
		}
		_ = h.notifService.Send(r.Context(), review.UserID, domain.NotifReviewApproved,
			"Ваш отзыв одобрен",
			"Ваш отзыв успешно прошел модерацию и опубликован",
			data,
		)

		result.Successful++
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *AdminHandler) BatchRejectReviews(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs    []string `json:"ids"`
		Reason string   `json:"reason,omitempty"`
	}
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "ids cannot be empty")
		return
	}

	if len(req.IDs) > 100 {
		writeError(w, http.StatusBadRequest, "invalid_input", "batch size cannot exceed 100")
		return
	}

	result := batchResult{}
	var reasons []string
	if req.Reason != "" {
		reasons = []string{req.Reason}
	}

	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			result.Failed++
			continue
		}

		review, err := h.reviewService.GetByID(r.Context(), id)
		if err != nil {
			result.Failed++
			continue
		}

		if err := h.reviewService.UpdateStatusWithReasons(r.Context(), id, domain.ReviewStatusRejected, reasons); err != nil {
			result.Failed++
			continue
		}

		// Notify review author about rejection
		data := map[string]string{
			"review_id":    review.ID.String(),
			"bathhouse_id": review.BathhouseID.String(),
		}
		reasonText := "Ваш отзыв не соответствует политике платформы"
		if req.Reason != "" {
			reasonText = req.Reason
		}
		_ = h.notifService.Send(r.Context(), review.UserID, domain.NotifReviewRejected,
			"Ваш отзыв отклонен",
			reasonText,
			data,
		)

		result.Successful++
	}

	writeJSON(w, http.StatusOK, result)
}
