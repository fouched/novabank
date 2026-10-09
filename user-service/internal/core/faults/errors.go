package faults

import (
	"errors"
	"fmt"
)

// Define core sentinel variables for common business failures
var (
	ErrNotFound   = errors.New("resource not found")
	ErrBadRequest = errors.New("bad request")
)

// ApplicationError allows attaching custom, granular validation context if needed
type ApplicationError struct {
	Err     error
	Message string
}

func (e *ApplicationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ApplicationError) Unwrap() error {
	return e.Err
}

func NewNotFound(msg string) error {
	return &ApplicationError{Err: ErrNotFound, Message: msg}
}

func NewBadRequest(msg string) error {
	return &ApplicationError{Err: ErrBadRequest, Message: msg}
}
