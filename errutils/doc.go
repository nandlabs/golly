package errutils

// Package errutils provides a set of utilities for working with errors in Go.

// Example:
//
// The following example demonstrates how to use the `Wrap` function to add context to an error:
//
//     package main
//
//     import (
//         "fmt"
//         "oss.nandlabs.io/golly/errutils"
//     )
//
//     func main() {
//         err := someFunction()
//         if err != nil {
//             wrappedErr := errutils.Wrap(err, "failed to perform operation")
//             fmt.Println(wrappedErr)
//         }
//     }
//
// In the above example, the `Wrap` function is used to add context to the original error returned by `someFunction()`.
// The resulting error is then printed using `fmt.Println()`.
