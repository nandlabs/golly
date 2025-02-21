package errutils

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"oss.nandlabs.io/golly/textutils"
)

type MultiError struct {
	errs  []error
	mutex sync.Mutex
}

// Add adds an error to the MultiError. If the error is nil, it is not added.
func (m *MultiError) Add(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var sb strings.Builder
	if m.errs != nil {
		for i, e := range m.errs {
			if i != 0 {
				sb.WriteString(textutils.NewLineString)
			}
			sb.WriteString(e.Error())
		}
	}

	return sb.String()
}

// HasErrors will return true if the MultiError has any errors
func (m *MultiError) HasErrors() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.errs) > 0
}

// HasError will return true if the MultiError has any errors of the specified type
func (m *MultiError) HasError(err error) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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
	if err != nil {
		multiErr.Add(err)
	}
	return
}

// CustomError is a struct that holds a template for creating custom errors.
// The template is a string that can contain format verbs.
type CustomError struct {
	template string
}

// Err creates a new error using the template and the parameters.
func (e *CustomError) Err(params ...any) error {
	return fmt.Errorf(e.template, params...)
}

// NewCustomError creates a new CustomError with the given template.
func NewCustomError(template string) *CustomError {
	return &CustomError{template: template}
}
