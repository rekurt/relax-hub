package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
)

type adminReviewResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	BathhouseID      string     `json:"bathhouse_id"`
	BookingID        string     `json:"booking_id"`
	Rating           int        `json:"rating"`
	Cleanliness      *float64   `json:"cleanliness,omitempty"`
	Accuracy         *float64   `json:"accuracy,omitempty"`
	Communication    *float64   `json:"communication,omitempty"`
	ValueForMoney    *float64   `json:"value_for_money,omitempty"`
	Text             string     `json:"text"`
	Status           string     `json:"status"`
	ModerationScore  *float64   `json:"moderation_score,omitempty"`
	ModerationFlags  []string   `json:"moderation_flags,omitempty"`
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
		Cleanliness:      rev.Cleanliness,
		Accuracy:         rev.Accuracy,
		Communication:    rev.Communication,
		ValueForMoney:    rev.ValueForMoney,
		Text:             rev.Text,
		Status:           string(rev.Status),
		ModerationScore:  rev.ModerationScore,
		ModerationFlags:  rev.ModerationFlags,
		RejectionReasons: rev.RejectionReasons,
		OwnerResponse:    rev.OwnerResponse,
		OwnerResponseAt:  rev.OwnerResponseAt,
		Images:           rev.Images,
		CreatedAt:        rev.CreatedAt,
		UpdatedAt:        rev.UpdatedAt,
	}
}

// ListReviews godoc
//
//	@Summary		List reviews (admin)
//	@Description	Returns a paginated list of all reviews for moderation. Supports filtering by status, bathhouse, rating range, and date range. Admin only.
//	@Tags			admin-reviews
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page			query		int		false	"Page number"	default(1)
//	@Param			page_size		query		int		false	"Page size"		default(20)
//	@Param			status			query		string	false	"Filter by status (pending, approved, rejected, hidden)"
//	@Param			bathhouse_id	query		string	false	"Filter by bathhouse ID (UUID)"
//	@Param			min_rating		query		int		false	"Minimum rating"
//	@Param			max_rating		query		int		false	"Maximum rating"
//	@Param			from_date		query		string	false	"Filter from date (RFC3339)"
//	@Param			to_date			query		string	false	"Filter to date (RFC3339)"
//	@Success		200				{object}	APIResponse{data=[]adminReviewResponse,meta=Meta}
//	@Failure		400				{object}	APIResponse{error=APIError}
//	@Failure		401				{object}	APIResponse{error=APIError}
//	@Failure		403				{object}	APIResponse{error=APIError}
//	@Router			/admin/reviews [get]
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

// GetPendingCount godoc
//
//	@Summary		Get pending reviews count
//	@Description	Returns the number of reviews awaiting moderation. Admin only.
//	@Tags			admin-reviews
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	APIResponse{data=pendingCountResponse}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Router			/admin/reviews/pending-count [get]
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

// ApproveReview godoc
//
//	@Summary		Approve review
//	@Description	Approves a pending review and notifies the author. Admin only.
//	@Tags			admin-reviews
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Review ID (UUID)"
//	@Success		200	{object}	APIResponse{data=adminReviewResponse}
//	@Failure		400	{object}	APIResponse{error=APIError}
//	@Failure		401	{object}	APIResponse{error=APIError}
//	@Failure		403	{object}	APIResponse{error=APIError}
//	@Failure		404	{object}	APIResponse{error=APIError}
//	@Router			/admin/reviews/{id}/approve [patch]
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

// RejectReview godoc
//
//	@Summary		Reject review
//	@Description	Rejects a review with an optional reason and notifies the author. Admin only.
//	@Tags			admin-reviews
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Review ID (UUID)"
//	@Param			body	body		approveRejectRequest	true	"Rejection reason (optional)"
//	@Success		200		{object}	APIResponse{data=adminReviewResponse}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Failure		404		{object}	APIResponse{error=APIError}
//	@Router			/admin/reviews/{id}/reject [patch]
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

// BatchApproveReviews godoc
//
//	@Summary		Batch approve reviews
//	@Description	Approves multiple reviews at once (max 100). Notifies each author. Admin only.
//	@Tags			admin-reviews
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		batchIDsRequest	true	"List of review IDs to approve"
//	@Success		200		{object}	APIResponse{data=batchResult}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/reviews/batch-approve [post]
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

type batchRejectRequest struct {
	IDs    []string `json:"ids"`
	Reason string   `json:"reason,omitempty"`
}

// --- Enhanced batch types for mass operations (up to 1000) ---

const maxBatchSize = 1000
const batchChunkSize = 100

type batchActionRequest struct {
	Action string   `json:"action"`
	IDs    []string `json:"ids"`
	Reason string   `json:"reason,omitempty"`
}

type batchCreditRequest struct {
	IDs    []string `json:"ids"`
	Amount int64    `json:"amount"`
	Reason string   `json:"reason"`
}

type batchItemResult struct {
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

type batchDetailedResult struct {
	Succeeded []batchItemResult `json:"succeeded"`
	Failed    []batchItemResult `json:"failed"`
}

// BatchRejectReviews godoc
//
//	@Summary		Batch reject reviews
//	@Description	Rejects multiple reviews at once (max 100) with an optional reason. Notifies each author. Admin only.
//	@Tags			admin-reviews
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		batchRejectRequest	true	"List of review IDs and optional rejection reason"
//	@Success		200		{object}	APIResponse{data=batchResult}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/reviews/batch-reject [post]
func (h *AdminHandler) BatchRejectReviews(w http.ResponseWriter, r *http.Request) {
	var req batchRejectRequest
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

// BatchListings godoc
//
//	@Summary		Batch listing operations
//	@Description	Performs batch approve or reject on up to 1000 listings. Processes in chunks of 100.
//	@Tags			admin-bathhouses
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		batchActionRequest	true	"Action (approve/reject) and list of IDs"
//	@Success		200		{object}	APIResponse{data=batchDetailedResult}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/listings/batch [post]
