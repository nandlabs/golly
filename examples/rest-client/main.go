// Package main demonstrates the REST client for making HTTP requests.
package main

import (
	"fmt"
	"log"
	"net/http"

	"oss.nandlabs.io/golly/rest"
)

func main() {
	// Create a REST client
	client := rest.NewClient()
	defer client.Close()

	// Simple GET request
	req, err := client.NewRequest("https://httpbin.org/get", http.MethodGet)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Add headers and query parameters
	req.AddHeader("Accept", "application/json").
		AddQueryParam("name", "golly").
		AddQueryParam("version", "1.0")

	// Execute the request
	resp, err := client.Execute(req)
	if err != nil {
		log.Fatal("Error executing request:", err)
	}

	fmt.Println("Status:", resp.Status())
	fmt.Println("Status Code:", resp.StatusCode())
	fmt.Println("Success:", resp.IsSuccess())

	// Decode response into a map
	var result map[string]interface{}
	if err := resp.Decode(&result); err != nil {
		log.Fatal("Error decoding:", err)
	}
	fmt.Println("Args:", result["args"])

	// POST request with JSON body
	body := map[string]string{
		"message": "Hello from Golly REST client",
	}
	postReq, err := client.NewRequest("https://httpbin.org/post", http.MethodPost)
	if err != nil {
		log.Fatal(err)
	}
	postReq.SetContentType("application/json").SetBody(body)

	postResp, err := client.Execute(postReq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nPOST Status:", postResp.StatusCode())
}
