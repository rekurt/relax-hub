package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/nikitaaldaev/bani/internal/logger"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	}); err != nil {
		log := logger.New(logger.LevelError)
		log.Error("Failed to encode JSON response", "error", err)
	}
}

func writeJSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}); err != nil {
		log := logger.New(logger.LevelError)
		log.Error("Failed to encode JSON response with meta", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeErrorWithContext(w, nil, status, code, message)
}

func writeErrorWithContext(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	log := logger.New(logger.LevelError)

	// Log 5xx errors
	if status >= 500 {
		var reqID string
		if r != nil {
			reqID = chiMiddleware.GetReqID(r.Context())
		}
		log.Error("HTTP error", "status", status, "code", code, "message", message, "request_id", reqID)
	}

	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}); err != nil {
		log.Error("Failed to encode error response", "error", err)
	}
}

func handleServiceError(w http.ResponseWriter, err error) {
	handleServiceErrorWithRequest(w, nil, err)
}

func handleServiceErrorWithRequest(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		writeErrorWithContext(w, r, http.StatusConflict, "already_exists", err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		writeErrorWithContext(w, r, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		writeErrorWithContext(w, r, http.StatusUnauthorized, "unauthorized", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		writeErrorWithContext(w, r, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, domain.ErrSlotUnavailable):
		writeErrorWithContext(w, r, http.StatusConflict, "slot_unavailable", err.Error())
	case errors.Is(err, domain.ErrBookingCancelLate):
		writeErrorWithContext(w, r, http.StatusBadRequest, "cancel_too_late", err.Error())
	case errors.Is(err, domain.ErrUserBlocked):
		writeErrorWithContext(w, r, http.StatusForbidden, "user_blocked", err.Error())
	case errors.Is(err, domain.ErrBathhouseNotActive):
		writeErrorWithContext(w, r, http.StatusBadRequest, "bathhouse_not_active", err.Error())
	case errors.Is(err, domain.ErrBathhouseHasBookings):
		writeErrorWithContext(w, r, http.StatusConflict, "bathhouse_has_bookings", err.Error())
	case errors.Is(err, domain.ErrReviewAlreadyResponded):
		writeErrorWithContext(w, r, http.StatusConflict, "review_already_responded", err.Error())
	default:
		writeErrorWithContext(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

const maxBodySize = 1 << 20 // 1 MB

func readJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if r.Body == nil {
		return domain.ErrInvalidInput
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return domain.ErrInvalidInput
	}
	return nil
}
