// Package main demonstrates the errutils package.
package main

import (
	"errors"
	"fmt"

	"oss.nandlabs.io/golly/errutils"
)

func main() {
	// CustomError with a template
	notFoundErr := errutils.NewCustomError("resource %s not found")
	err := notFoundErr.Err("user-123")
	fmt.Println("Custom error:", err)

	// MultiError for collecting multiple errors
	multiErr := errutils.NewMultiErr(errors.New("failed to connect to database"))
	multiErr.Add(errors.New("failed to read config"))
	multiErr.Add(errors.New("failed to initialize cache"))
	fmt.Println("Multi error:", multiErr.Error())
	fmt.Println("Has errors:", multiErr.HasErrors())
	fmt.Println("All errors:", multiErr.GetAll())
}
