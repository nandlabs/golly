// Package rest provides a set of utilities for making RESTful API calls.
//
// This package includes functions for sending HTTP requests, handling responses,
// and managing authentication and authorization headers.
//
// Example usage:
//
//	// Create a new REST client
//	client := rest.NewClient()
//
//	// Set the base URL for the API
//	client.SetBaseURL("https://api.example.com")
//
//	// Set the authentication token
//	client.SetAuthToken("YOUR_AUTH_TOKEN")
//
//	// Send a GET request
//	response, err := client.Get("/users")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Print the response body
//	fmt.Println(response.Body)
//
//	// Close the response body
//	response.Close()
//
// For more information and examples, please refer to the package documentation.
package rest
