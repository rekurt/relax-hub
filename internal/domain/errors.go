package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrSlotUnavailable    = errors.New("time slot is unavailable")
	ErrBookingCancelLate  = errors.New("too late to cancel booking")
	ErrUserBlocked        = errors.New("user is blocked")
	ErrBathhouseNotActive   = errors.New("bathhouse is not active")
	ErrBathhouseHasBookings = errors.New("cannot delete bathhouse with active bookings")
)
