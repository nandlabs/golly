// Package main demonstrates the semver parsing and comparison utilities.
package main

import (
	"fmt"
	"log"

	"oss.nandlabs.io/golly/semver"
)

func main() {
	// Parse semantic versions
	fmt.Println("=== Parsing Versions ===")
	v1, err := semver.Parse("1.2.3")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parsed:", v1)

	v2, err := semver.Parse("2.0.0-beta.1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parsed:", v2)

	// Compare versions
	result, err := semver.CompareRaw("1.2.3", "2.0.0")
	if err != nil {
		log.Fatal(err)
	}
	switch {
	case result < 0:
		fmt.Println("1.2.3 < 2.0.0")
	case result > 0:
		fmt.Println("1.2.3 > 2.0.0")
	default:
		fmt.Println("1.2.3 == 2.0.0")
	}

	// Parse version with pre-release and build metadata
	v3, err := semver.Parse("1.0.0-alpha+build.123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Major: %d, Minor: %d, Patch: %d\n", v3.CurrentMajor(), v3.CurrentMinor(), v3.CurrentPatch())
	fmt.Println("PreRelease:", v3.CurrentPreRelease())
	fmt.Println("Build:", v3.CurrentBuild())
}
