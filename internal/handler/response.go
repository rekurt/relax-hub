package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nikitaaldaev/bani/internal/domain"
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
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

func writeJSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "already_exists", err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized", err.Error())
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, domain.ErrSlotUnavailable):
		writeError(w, http.StatusConflict, "slot_unavailable", err.Error())
	case errors.Is(err, domain.ErrBookingCancelLate):
		writeError(w, http.StatusBadRequest, "cancel_too_late", err.Error())
	case errors.Is(err, domain.ErrUserBlocked):
		writeError(w, http.StatusForbidden, "user_blocked", err.Error())
	case errors.Is(err, domain.ErrBathhouseNotActive):
		writeError(w, http.StatusBadRequest, "bathhouse_not_active", err.Error())
	case errors.Is(err, domain.ErrBathhouseHasBookings):
		writeError(w, http.StatusConflict, "bathhouse_has_bookings", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

const maxBodySize = 1 << 20 // 1 MB

func readJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return domain.ErrInvalidInput
	}
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return domain.ErrInvalidInput
	}
	return nil
}
