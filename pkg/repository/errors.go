package repository

import (
	"fmt"
)

// ErrNotFound is returned when a resource is not found.
type ErrNotFound struct {
	Entity string
	ID     string
}

// Error implements the error interface.
func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Entity, e.ID)
}

// ErrUniqueConstraintFailed is returned when a unique constraint is violated.
type ErrUniqueConstraintFailed struct {
	Cause error
}

// Error implements the error interface.
func (e *ErrUniqueConstraintFailed) Error() string {
	return fmt.Sprintf("unique constraint failed: %v", e.Cause)
}

// Unwrap returns the underlying cause of the error.
func (e *ErrUniqueConstraintFailed) Unwrap() error {
	return e.Cause
}
