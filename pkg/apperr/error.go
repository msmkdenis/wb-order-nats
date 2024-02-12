// Package url_err implements the application errors.
package apperr

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
)

// ValueError is an error that represents a value error.
type ValueError struct {
	caller  string
	message string
	err     error
}

// NewValueError creates a new ValueError with the given message, caller, and error.
//
// Parameters:
//
//	message string - the error message
//	caller string - the caller of the function tracing place in the code
//	err error - the original error
//
// Return type:
//
//	error - the newly created ValueError
func NewValueError(message string, caller string, err error) error {
	return &ValueError{
		caller:  caller,
		message: message,
		err:     err,
	}
}

// Error returns a string representing the error.
//
// No parameters.
// Returns a string.
func (v *ValueError) Error() string {
	return fmt.Sprintf("%s %s %s", v.caller, v.message, v.err)
}

// Unwrap returns the error that has been wrapped by ValueError.
// No parameters. Returns an error.
func (v *ValueError) Unwrap() error {
	return v.err
}

// Caller returns file name and line number of function call
func Caller() string {
	_, file, lineNo, ok := runtime.Caller(1)
	if !ok {
		return "runtime.Caller() failed"
	}

	fileName := path.Base(file)
	dir := filepath.Base(filepath.Dir(file))
	return fmt.Sprintf("%s/%s:%d", dir, fileName, lineNo)
}
