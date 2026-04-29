package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrValidation         = errors.New("validation error")
	ErrInvalidStatus      = errors.New("invalid status transition")
	ErrPendingPayment     = errors.New("household has a pending payment")
	ErrSafetyCheckRequired = errors.New("electronic pickups require safety_check to be true before scheduling")
)
