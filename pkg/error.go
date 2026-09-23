package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrInternal = errors.New("internal")

	// client errors
	ErrEmpty              = errors.New("empty")
	ErrExceedsMax         = errors.New("exceeds_max")
	ErrBelowMin           = errors.New("below_min")
	ErrInvalidInput       = errors.New("invalid_input")
	ErrPreconditionNotMet = errors.New("precondition_not_met")
)

type FieldError struct {
	Key   string
	Msg   string
	cause error
}

func (e *FieldError) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("%s: %s", e.Key, e.Msg)
	}
	return fmt.Sprintf("%s: %s: %s", e.Key, e.Msg, e.cause.Error())
}

func (e *FieldError) Unwrap() error {
	return e.cause
}

type LimitError struct {
	Key   string
	Limit int
	cause error
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("%s: %d: %s", e.Key, e.Limit, e.cause.Error())
}

func (e *LimitError) Unwrap() error {
	return e.cause
}

func NewFieldError(key, msg string, err error) *FieldError {
	return &FieldError{Msg: msg, Key: key, cause: err}
}

func NewLimitError(key string, limit int, err error) *LimitError {
	return &LimitError{Key: key, Limit: limit, cause: err}
}
