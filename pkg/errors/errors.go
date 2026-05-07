// Package errors provides helpers for error wrapping following D-05
// (standard %w wrapping to support errors.Is / errors.As).
package errors

import (
	"errors"
	"fmt"
)

// Wrap wraps err with the supplied message. If err is nil, Wrap returns nil.
// The format follows "context: original" so callers can chain errors.Is checks.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Wrapf wraps err with a formatted message. Returns nil when err is nil.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// Is is a convenience re-export of errors.Is so callers can import only
// this package for error handling.
var Is = errors.Is

// As is a convenience re-export of errors.As.
var As = errors.As

// New creates a new error with the given message.
var New = errors.New
