package errors

import (
	goerrors "errors"
	"fmt"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

const (
	// Used when given a nil error.
	emptyErrorMessage = "missing error"
)

// internalError implements error and is intended to be passed around as such.
// This allows code to seemlessly integrate with libraries and code that do not
// use internal pkgerrors. Helper functions are provided to provide internal data.
type internalError struct {
	// The first element will always be the original internal code.
	// The rest will the be stack trace of internal codes found.
	internalCodeStack []string

	// The original error.
	// Created with pkgerrors.withStack() and will print the stack trace when with %+v.
	err error
}

// New error. This should be called when the application creates a brand new error.
// If an error has been received from an external function, use Wrap().
func New(internalCode string, format string, values ...interface{}) error {
	return newInternalError(internalCode, pkgerrors.New(fmt.Sprintf(format, values...)))
}

// Wrap an existing error.
// "internalCode" should be a unique code to allow developers to easily identify the source of an issue.
func Wrap(internalCode string, err error) error {
	return newInternalError(internalCode, fmt.Errorf("%w", err))
}

// newInternalError creates a new internal error.
// Handles wrapping multiple internal pkgerrors.
func newInternalError(internalCode string, err error) internalError {
	if err == nil {
		// Should never happen, but do not let that break things.
		err = fmt.Errorf(emptyErrorMessage)
	}

	// Error is already an internalError.
	// Simply append the current internal code and error to the stacks and return it.
	var previousInternalError internalError
	if ok := pkgerrors.As(err, &previousInternalError); ok {
		previousInternalError.internalCodeStack = append(previousInternalError.internalCodeStack, internalCode)

		return previousInternalError
	}

	// err is not currently a internal error.
	// Create a new one and return it.
	return internalError{
		internalCodeStack: []string{internalCode},
		err:               pkgerrors.WithStack(err),
	}
}

// Error string of the origin error.
// This will include the internal code.
func (self internalError) Error() string {
	return self.err.Error()
}

// Format the internal error for different situations.
//   * %+v - Print stack trace
//   * %v  - Print error with internal code
func (self internalError) Format(s fmt.State, verb rune) {
	// Determine what to prepend to the error string, if anything.
	switch verb {
	case 'v':
		if s.Flag('+') {
			// Print every internal code given.
			fmt.Fprintf(s, "[%s] ", strings.Join(self.internalCodeStack, ","))
		} else {
			// Print only the origin internal code.
			fmt.Fprintf(s, "[%s] ", self.internalCodeStack[0])
		}
	}

	// Print the stack trace for the error.
	self.err.(fmt.Formatter).Format(s, verb)
}

// Unwrap to get the underlying error.
func (self internalError) Unwrap() error {
	return self.err
}

// InternalCode returns the internal code string of the error origin.
func InternalCode(err error) string {
	var asInternalError internalError
	if ok := pkgerrors.As(err, &asInternalError); ok {
		return asInternalError.internalCodeStack[0]
	}

	// Not a internal error.
	return ""
}

// Is err a type of target error.
// See: errors.Is()
func Is(err, target error) bool {
	return goerrors.Is(err, target)
}

// As converts the target to that type from the error stack.
// See: errors.As()
func As(err error, target interface{}) bool {
	return goerrors.As(err, target)
}
