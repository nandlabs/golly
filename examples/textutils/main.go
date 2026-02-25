// Package main demonstrates the textutils package with ASCII character constants.
package main

import (
	"fmt"

	"oss.nandlabs.io/golly/textutils"
)

func main() {
	// textutils provides named ASCII character constants for readable code.
	// Instead of using magic rune/byte literals, use descriptive names.

	fmt.Println("=== ASCII Character Constants ===")

	// Letter constants
	fmt.Printf("Uppercase A: %c (%d)\n", textutils.AUpperChar, textutils.AUpperChar)
	fmt.Printf("Lowercase a: %c (%d)\n", textutils.ALowerChar, textutils.ALowerChar)
	fmt.Printf("Uppercase Z: %c (%d)\n", textutils.ZUpperChar, textutils.ZUpperChar)
	fmt.Printf("Lowercase z: %c (%d)\n", textutils.ZLowerChar, textutils.ZLowerChar)

	// Use constants in string building and comparison
	fmt.Println("\n=== Using Constants ===")

	// Check if a character is uppercase
	ch := 'G'
	if ch >= textutils.AUpperChar && ch <= textutils.ZUpperChar {
		fmt.Printf("'%c' is an uppercase letter\n", ch)
	}

	// Convert uppercase to lowercase using constant offsets
	lower := ch + (textutils.ALowerChar - textutils.AUpperChar)
	fmt.Printf("'%c' -> '%c'\n", ch, lower)

	// Build a simple alphabet range
	fmt.Print("\nAlphabet: ")
	for c := textutils.AUpperChar; c <= textutils.ZUpperChar; c++ {
		fmt.Printf("%c", c)
	}
	fmt.Println()
}
