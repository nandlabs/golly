// Package main demonstrates the managers package for generic item management.
package main

import (
	"fmt"

	"oss.nandlabs.io/golly/managers"
)

// Service represents a simple service with a name and port.
type Service struct {
	Name string
	Port int
}

func main() {
	// Create a typed item manager
	mgr := managers.NewItemManager[*Service]()

	// Register services
	mgr.Register("auth", &Service{Name: "Auth Service", Port: 8081})
	mgr.Register("users", &Service{Name: "User Service", Port: 8082})
	mgr.Register("orders", &Service{Name: "Order Service", Port: 8083})

	// Retrieve a service by name
	authSvc := mgr.Get("auth")
	fmt.Printf("Auth service: %s (port %d)\n", authSvc.Name, authSvc.Port)

	// List all registered services
	fmt.Println("\nAll services:")
	for _, svc := range mgr.Items() {
		fmt.Printf("  - %s (port %d)\n", svc.Name, svc.Port)
	}

	// Unregister a service
	mgr.Unregister("orders")
	fmt.Println("\nAfter unregistering 'orders':")
	for _, svc := range mgr.Items() {
		fmt.Printf("  - %s (port %d)\n", svc.Name, svc.Port)
	}
}
