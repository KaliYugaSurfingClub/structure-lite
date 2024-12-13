package errs

import (
	"errors"
	"fmt"
)

type Op string

// Wrap wraps error
// Op means operation
// It helps to collect stack trace
func Wrap(op Op, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", op, err)
}

func New(op Op, msg string) error {
	return Wrap(op, errors.New(msg))
}

func With(op Op, context string, err error) error {
	if err == nil {
		return nil
	}

	return Wrap(op, fmt.Errorf("%s, %w", context, err))
}
