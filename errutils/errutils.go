package errutils

import (
	"errors"
	"fmt"
	"strings"

	"oss.nandlabs.io/golly/textutils"
)

// FmtError formats an error message using a format string and optional values,
// and returns it as an error object.
//
// It takes a format string f and variadic arguments v, and uses them to construct
// the error message by calling fmt.Sprintf function. The resulting error message
// is then wrapped into an error object using errors.New.
//
// Example:
//
//	err := FmtError("Invalid input: %s", userInput)
//	if err != nil {
//		log.Println(err)
//	}
//
// Deprecated: In favour of fmt.Errorf(...) instead
// @param f The format string specifying the error message.
// @param v Optional values to be inserted into the format string.
// @returns An error object containing the formatted error message.
func FmtError(f string, v ...any) error {
	return errors.New(fmt.Sprintf(f, v...))
}

type MultiError struct {
	errs []error
}

// Add adds an error to the MultiError. If the error is nil, it is not added.
func (m *MultiError) Add(err error) {
	if err != nil {
		m.errs = append(m.errs, err)
	}
}

// GetAll returns all the errors in the MultiError.
func (m *MultiError) GetAll() (errs []error) {
	errs = m.errs
	return
}

// Error function implements the error.Error function of the error interface
func (m *MultiError) Error() string {
	var sb strings.Builder
	if m.errs != nil && len(m.errs) > 0 {
		for i, e := range m.errs {
			if i != 0 {
				sb.WriteString(textutils.NewLineString)
			}
			sb.WriteString(e.Error())
		}
	}

	return sb.String()
}

// HasError will return true if the MultiError has any errors of the specified type
func (m *MultiError) HasError(err error) bool {
	for _, e := range m.errs {
		if errors.Is(e, err) {
			return true
		}
	}
	return false
}

// NewMultiErr creates a new MultiError and adds the given error to it.
func NewMultiErr(err error) (multiErr *MultiError) {
	multiErr = &MultiError{}
	multiErr.Add(err)
	return
}
