package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/middleware"
)

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
func (h *AdminHandler) BatchListings(w http.ResponseWriter, r *http.Request) {
	var req batchActionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		writeError(w, http.StatusBadRequest, "invalid_input", "action must be 'approve' or 'reject'")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "ids cannot be empty")
		return
	}
	if len(req.IDs) > maxBatchSize {
		writeError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("batch size cannot exceed %d", maxBatchSize))
		return
	}

	result := batchDetailedResult{
		Succeeded: make([]batchItemResult, 0),
		Failed:    make([]batchItemResult, 0),
	}

	for i := 0; i < len(req.IDs); i += batchChunkSize {
		end := i + batchChunkSize
		if end > len(req.IDs) {
			end = len(req.IDs)
		}
		chunk := req.IDs[i:end]

		for _, idStr := range chunk {
			id, err := uuid.Parse(idStr)
			if err != nil {
				result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: "invalid uuid"})
				continue
			}

			switch req.Action {
			case "approve":
				if err := h.bathhouseService.Approve(r.Context(), id); err != nil {
					result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: err.Error()})
				} else {
					result.Succeeded = append(result.Succeeded, batchItemResult{ID: idStr})
				}
			case "reject":
				if err := h.bathhouseService.Reject(r.Context(), id); err != nil {
					result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: err.Error()})
				} else {
					result.Succeeded = append(result.Succeeded, batchItemResult{ID: idStr})
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// BatchUsers godoc
//
//	@Summary		Batch user operations
//	@Description	Performs batch block or unblock on up to 1000 users. Processes in chunks of 100.
//	@Tags			admin-users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		batchActionRequest	true	"Action (block/unblock) and list of IDs"
//	@Success		200		{object}	APIResponse{data=batchDetailedResult}
//	@Failure		400		{object}	APIResponse{error=APIError}
//	@Failure		401		{object}	APIResponse{error=APIError}
//	@Failure		403		{object}	APIResponse{error=APIError}
//	@Router			/admin/users/batch [post]
func (h *AdminHandler) BatchUsers(w http.ResponseWriter, r *http.Request) {
	var req batchActionRequest
	if err := readJSON(w, r, &req); err != nil {
		handleServiceError(w, err)
		return
	}

	if req.Action != "block" && req.Action != "unblock" {
		writeError(w, http.StatusBadRequest, "invalid_input", "action must be 'block' or 'unblock'")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "ids cannot be empty")
		return
	}
	if len(req.IDs) > maxBatchSize {
		writeError(w, http.StatusBadRequest, "invalid_input", fmt.Sprintf("batch size cannot exceed %d", maxBatchSize))
		return
	}

	adminID := middleware.GetUserID(r.Context())

	result := batchDetailedResult{
		Succeeded: make([]batchItemResult, 0),
		Failed:    make([]batchItemResult, 0),
	}

	for i := 0; i < len(req.IDs); i += batchChunkSize {
		end := i + batchChunkSize
		if end > len(req.IDs) {
			end = len(req.IDs)
		}
		chunk := req.IDs[i:end]

		for _, idStr := range chunk {
			id, err := uuid.Parse(idStr)
			if err != nil {
				result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: "invalid uuid"})
				continue
			}

			if id == adminID {
				result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: "cannot block yourself"})
				continue
			}

			switch req.Action {
			case "block":
				if err := h.userService.Block(r.Context(), id); err != nil {
					result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: err.Error()})
				} else {
					result.Succeeded = append(result.Succeeded, batchItemResult{ID: idStr})
				}
			case "unblock":
				if err := h.userService.Unblock(r.Context(), id); err != nil {
					result.Failed = append(result.Failed, batchItemResult{ID: idStr, Error: err.Error()})
				} else {
					result.Succeeded = append(result.Succeeded, batchItemResult{ID: idStr})
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, result)
}
