package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

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
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON response: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "failed to encode JSON response with meta: %v\n", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeErrorWithContext(w, nil, status, code, message)
}

func writeErrorWithContext(w http.ResponseWriter, _ *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode error response: %v\n", err)
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
	case errors.Is(err, domain.ErrSocialAccountAlreadyLinked):
		writeErrorWithContext(w, r, http.StatusConflict, "social_account_already_linked", err.Error())
	case errors.Is(err, domain.ErrSocialAccountNotFound):
		writeErrorWithContext(w, r, http.StatusNotFound, "social_account_not_found", err.Error())
	case errors.Is(err, domain.ErrOAuthExchangeFailed):
		writeErrorWithContext(w, r, http.StatusBadRequest, "oauth_exchange_failed", err.Error())
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
