package errutils

import (
	"errors"
	"fmt"
	"testing"
)

// TestFmtError tests the FmtError function
func TestFmtError(t *testing.T) {
	t.Run("Testing execution of FmtError", func(t *testing.T) {
		err := FmtError("testing error")
		if err == nil {
			t.Errorf("error not created")
		}
	})
}

// TestMultiError_Add tests the Add function of MultiError
func TestMultiError_Add(t *testing.T) {
	t.Run("Testing execution of MultiError_Add", func(t *testing.T) {
		m := &MultiError{}
		err := errors.New("testing error")
		m.Add(err)
		if len(m.errs) != 1 {
			t.Errorf("error not added to MultiError")
		}
	})
}

// TestMultiError_GetAll tests the GetAll function of MultiError
func TestMultiError_GetAll(t *testing.T) {
	t.Run("Testing execution of MultiError_GetAll", func(t *testing.T) {
		m := &MultiError{}
		err := errors.New("testing error")
		m.Add(err)
		errs := m.GetAll()
		if len(errs) != 1 {
			t.Errorf("error not added to MultiError")
		}
	})
}

// TestMultiError_Error tests the Error function of MultiError
func TestMultiError_Error(t *testing.T) {
	t.Run("Testing execution of MultiError_Error", func(t *testing.T) {
		m := &MultiError{}
		err := errors.New("testing error")
		m.Add(err)
		errs := m.Error()
		if errs != "testing error" {
			t.Errorf("error not added to MultiError")
		}
	})
}

// TestMultiError_HasError tests the HasError function of MultiError
func TestMultiError_HasError(t *testing.T) {
	t.Run("Testing execution of MultiError_HasError", func(t *testing.T) {
		m := &MultiError{}
		err := errors.New("testing error")
		m.Add(err)
		if !m.HasError(err) {
			t.Errorf("error not added to MultiError")
		}
	})
}

// TestNewMultiErr tests the NewMultiErr function of MultiError
func TestNewMultiErr(t *testing.T) {
	t.Run("Testing execution of NewMultiErr", func(t *testing.T) {
		err := errors.New("testing error")
		m := NewMultiErr(err)
		if len(m.errs) != 1 {
			t.Errorf("error not added to MultiError")
		}
	})
}

// TestMultiError_AddNil tests the Add function of MultiError with nil error
func TestMultiError_AddNil(t *testing.T) {
	t.Run("Testing execution of MultiError_AddNil", func(t *testing.T) {
		m := &MultiError{}
		m.Add(nil)
		fmt.Println(m.errs)
		if len(m.errs) != 0 {
			t.Errorf("nil error added to MultiError")
		}
	})
}

// TestMultiError_HasErrorNil tests the HasError function of MultiError with nil error
func TestMultiError_HasErrorNil(t *testing.T) {
	t.Run("Testing execution of MultiError_HasErrorNil", func(t *testing.T) {
		m := &MultiError{}
		if m.HasError(nil) {
			t.Errorf("nil error added to MultiError")
		}
	})
}

// TestMultiError_ErrorNil tests the Error function of MultiError with nil error
func TestMultiError_ErrorNil(t *testing.T) {
	t.Run("Testing execution of MultiError_ErrorNil", func(t *testing.T) {
		m := &MultiError{}
		errs := m.Error()
		if errs != "" {
			t.Errorf("nil error added to MultiError")
		}
	})

}
