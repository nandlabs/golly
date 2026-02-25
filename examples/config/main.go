// Package main demonstrates the config package for application configuration.
package main

import (
	"fmt"
	"os"
	"strings"

	"oss.nandlabs.io/golly/config"
)

func main() {
	// --- Environment variable helpers ---
	// Set some env vars for demo
	os.Setenv("APP_NAME", "golly-demo")
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_DEBUG", "true")
	defer os.Unsetenv("APP_NAME")
	defer os.Unsetenv("APP_PORT")
	defer os.Unsetenv("APP_DEBUG")

	name := config.GetEnvAsString("APP_NAME", "default-app")
	fmt.Println("App name:", name)

	port, err := config.GetEnvAsInt("APP_PORT", 3000)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Port:", port)

	debug, err := config.GetEnvAsBool("APP_DEBUG", false)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Debug:", debug)

	// Default value when env var is not set
	missing := config.GetEnvAsString("APP_MISSING", "fallback-value")
	fmt.Println("Missing env:", missing)

	// --- Properties configuration ---
	props := config.NewProperties()

	// Load from a reader (Java-style properties format)
	propsData := `
db.host=localhost
db.port=5432
db.name=golly_db
app.version=1.0.0
`
	err = props.Load(strings.NewReader(propsData))
	if err != nil {
		fmt.Println("Error loading properties:", err)
	}

	fmt.Println("\nProperties:")
	fmt.Println("db.host:", props.Get("db.host", ""))
	dbPort, _ := props.GetAsInt("db.port", 3306)
	fmt.Println("db.port:", dbPort)
	fmt.Println("db.name:", props.Get("db.name", ""))
	fmt.Println("app.version:", props.Get("app.version", "0.0.0"))

	// Put new values
	props.Put("app.env", "production")
	fmt.Println("app.env:", props.Get("app.env", ""))
}
