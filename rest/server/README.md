# Rest Server

The server package provides a simple and efficient way to expose server side RESTful APIs in Go.
This package is built on top of the `oss.nandlabs.io/golly/turbo` package and provides a simple way to define routes and handle requests and maintain the configuration & lifecycle of the server.

---

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)

---

## Features

- HTTP methods: GET, POST, PUT, DELETE
- Query parameters
- Request headers
- TLS Configuration
- Transport Layer Configuration
  - Connection Timeout
  - Read Timeout

## Installation

To install the REST server, use the following command:

```bash
go get oss.nandlabs.io/golly/rest/server
```

## Usage

To use the REST client in your Go project, you first need to import the package:

```go
import "oss.nandlabs.io/golly/rest/server"
```

#### HTTP Methods : Sevve a GET Request

```go
package main

import (
	"net/http"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest/server"
)

func main() {
	// Create a new server
	// Checkout  New(opts *Options) for customisng the server properties
	srv, err := server.Default()
	if err != nil {
		panic(err)
	}
	// this is the path prefix for each endpoint. Default is empty and no path prefix is added
	srv.Opts().PathPrefix = "/api/v1"
	// Add a GET endpoint
	srv.Get("healthCheck", func(ctx server.Context) {
		// Set the status code. Remember to set the status code before writing the response
		ctx.SetStatusCode(http.StatusOK)
		ctx.WriteString("Health Check Get")
	})
	// Add a POST endpoint
	srv.Post("healthCheck", func(ctx server.Context) {
		input, _ := ctx.GetBody()
		// Set the status code. Remember to set the status code before writing the response
		ctx.SetStatusCode(http.StatusOK)
		// Write the response
		ctx.WriteFrom(input)
	})
	// get the component manager
	mgr := lifecycle.NewSimpleComponentManager()
	// Register the server with the component manager
	mgr.Register(srv)
	// Start the server
	mgr.StartAndWait()
}

```
