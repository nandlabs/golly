// Package main demonstrates the turbo HTTP router package.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"oss.nandlabs.io/golly/turbo"
)

func main() {
	// Create a new router
	router := turbo.NewRouter()

	// Register routes with different HTTP methods
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Welcome to the Turbo router!",
		})
	})

	// Path parameters
	router.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, err := turbo.GetPathParam("id", r)
		if err != nil {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": id,
			"name":    "John Doe",
		})
	})

	// POST route
	router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "user created",
		})
	})

	// PUT route
	router.Put("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := turbo.GetPathParam("id", r)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "user updated",
			"user_id": id,
		})
	})

	// DELETE route
	router.Delete("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		id, _ := turbo.GetPathParam("id", r)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "user deleted",
			"user_id": id,
		})
	})

	// Query parameters
	router.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		query, _ := turbo.GetQueryParam("q", r)
		page, err := turbo.GetQueryParamAsInt("page", r)
		if err != nil {
			page = 1
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"query": query,
			"page":  page,
		})
	})

	// Multiple methods on a single path
	router.Add("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	}, "GET", "HEAD")

	// Start the server
	fmt.Println("Turbo router listening on :8080")
	fmt.Println("Routes:")
	fmt.Println("  GET    /")
	fmt.Println("  GET    /users/:id")
	fmt.Println("  POST   /users")
	fmt.Println("  PUT    /users/:id")
	fmt.Println("  DELETE /users/:id")
	fmt.Println("  GET    /search?q=...&page=...")
	fmt.Println("  GET    /health")
	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Println("Server error:", err)
	}
}
