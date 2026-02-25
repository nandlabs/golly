// Package main demonstrates the REST server for building HTTP APIs.
package main

import (
	"fmt"
	"log"
	"net/http"

	"oss.nandlabs.io/golly/rest"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Create server with options
	opts := rest.DefaultSrvOptions()
	opts.Id = "example-api"
	opts.ListenHost = "localhost"
	opts.ListenPort = 9090
	opts.PathPrefix = "/api"

	server, err := rest.NewServer(opts)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Health check endpoint
	_, err = server.Get("/health", func(ctx rest.ServerContext) {
		resp := HealthResponse{Status: "ok", Version: "1.0.0"}
		ctx.SetStatusCode(http.StatusOK)
		if err := ctx.WriteJSON(resp); err != nil {
			log.Println("Write error:", err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	// GET /items - list items
	_, err = server.Get("/items", func(ctx rest.ServerContext) {
		items := []Item{
			{ID: "1", Name: "Widget"},
			{ID: "2", Name: "Gadget"},
		}
		ctx.SetStatusCode(http.StatusOK)
		if err := ctx.WriteJSON(items); err != nil {
			log.Println("Write error:", err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	// POST /items - create item
	_, err = server.Post("/items", func(ctx rest.ServerContext) {
		var item Item
		if err := ctx.Read(&item); err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			_ = ctx.WriteJSON(map[string]string{"error": err.Error()})
			return
		}
		ctx.SetStatusCode(http.StatusCreated)
		_ = ctx.WriteJSON(item)
	})
	if err != nil {
		log.Fatal(err)
	}

	// GET /items/:id - get item by ID
	_, err = server.Get("/items/:id", func(ctx rest.ServerContext) {
		id, err := ctx.GetParam("id", rest.PathParam)
		if err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			_ = ctx.WriteJSON(map[string]string{"error": "missing id"})
			return
		}
		item := Item{ID: id, Name: "Item-" + id}
		ctx.SetStatusCode(http.StatusOK)
		_ = ctx.WriteJSON(item)
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting server at http://%s:%d%s\n", opts.ListenHost, opts.ListenPort, opts.PathPrefix)

	// Start the server (blocks)
	if err := server.Start(); err != nil {
		log.Fatal("Server error:", err)
	}
}
