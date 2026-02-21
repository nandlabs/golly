// Package main demonstrates the fsutils package.
package main

import (
	"fmt"

	"oss.nandlabs.io/golly/fsutils"
)

func main() {
	// Check if paths exist
	fmt.Println("Current dir exists:", fsutils.PathExists("."))
	fmt.Println("/tmp exists:", fsutils.DirExists("/tmp"))
	fmt.Println("go.mod exists:", fsutils.FileExists("go.mod"))
	fmt.Println("nonexistent.txt:", fsutils.FileExists("nonexistent.txt"))

	// Detect content type of a file
	ct, err := fsutils.DetectContentType("go.mod")
	if err != nil {
		fmt.Println("DetectContentType error:", err)
	} else {
		fmt.Println("go.mod content type:", ct)
	}

	// Lookup content type by extension
	fmt.Println("Lookup .json:", fsutils.LookupContentType("data.json"))
	fmt.Println("Lookup .html:", fsutils.LookupContentType("page.html"))
	fmt.Println("Lookup .go:", fsutils.LookupContentType("main.go"))
}
