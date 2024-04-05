package fnutils

import (
	"errors"
	"time"
)

// ExecuteAfterSecs executes the given function after the specified timeout duration,
// expressed in seconds. It returns any error encountered during the execution.
//
// It converts the timeout value from seconds to a duration in seconds and then calls
// the ExecuteAfter function, passing the converted duration and the provided function.
//
// Example:
//
//	err := ExecuteAfterSecs(func() {
//	    fmt.Println("Hello, World!")
//	}, 10)
//
// @param fn The function to be executed.
//
// @param timeout The timeout duration in seconds.
//
// @returns An error encountered during execution, if any.
func ExecuteAfterSecs(fn func(), timeout int) (err error) {
	err = ExecuteAfter(fn, time.Second*time.Duration(timeout))
	return
}

// ExecuteAfterMs executes the given function after the specified timeout duration,
// expressed in milliseconds. It returns any error encountered during the execution.
//
// It converts the timeout value from milliseconds to a duration in seconds and then calls
// the ExecuteAfter function, passing the converted duration and the provided function.
//
// Example:
//
//	err := ExecuteAfterMs(func() {
//	    fmt.Println("Hello, World!")
//	}, 1000)
//
// @param fn The function to be executed.
//
// @param timeout The timeout duration in milliseconds.
//
// @returns An error encountered during execution, if any.
func ExecuteAfterMs(fn func(), timeout int64) (err error) {
	err = ExecuteAfter(fn, time.Millisecond*time.Duration(timeout))
	return
}

// ExecuteAfterMin executes the given function after the specified timeout duration,
// expressed in minutes. It returns any error encountered during the execution.
//
// It converts the timeout value from minutes to a duration in seconds and then calls
// the ExecuteAfter function, passing the converted duration and the provided function.
//
// Example:
//
//	err := ExecuteAfterMin(func() {
//	    fmt.Println("Hello, World!")
//	}, 5)
//
// @param fn The function to be executed.
//
// @param timeout The timeout duration in minutes.
//
// @returns An error encountered during execution, if any.
func ExecuteAfterMin(fn func(), timeout int) (err error) {
	err = ExecuteAfter(fn, time.Minute*time.Duration(timeout))
	return
}

// ExecuteAfter executes the given function after the specified timeout duration.
// It returns any error encountered during the execution.
// It waits for the specified duration and then calls the provided function.
//
// Example:
//
//	err := ExecuteAfter(func() {
//	    fmt.Println("Hello, World!")
//	}, time.Second)
//
// @param fn The function to be executed.
//
// @param timeout The duration to wait before executing the function.
//
// @returns An error encountered during execution, if any.
func ExecuteAfter(fn func(), timeout time.Duration) (err error) {
	if fn == nil {
		err = errors.New("nil function provided")
		return
	}
	if timeout < 0 {
		err = errors.New("timeout cannot be negative")
		return
	}
	select {
	case <-time.After(timeout):
		{
			fn()
		}
	}
	return
}
