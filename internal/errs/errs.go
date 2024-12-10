package errs

import "fmt"

type Op string

// W wraps error
// Op means operation
// It helps to collect stack trace
func W(op Op, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", op, err)
}
