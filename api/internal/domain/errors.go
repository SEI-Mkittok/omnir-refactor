package domain

import "errors"

// Sentinel errors for domain-level conditions.
var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation error")
)
