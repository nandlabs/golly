# Rest Server

The server package provides a simple and efficient way to expose server side RESTful APIs in Go.
This package is built on top of the `oss.nandlabs.io/golly/turbo` package and provides a simple way to define routes and handle requests and maintain the configuration & lifecycle of the server.

---

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Client](#client)
    - [HTTP Methods : Sending a GET Request](#http-methods--sending-a-get-request)
    - [Retry Configuration](#retry-configuration)
    - [CircuitBreaker Configuration](#circuitbreaker-configuration)
    - [Proxy Configuration](#proxy-configuration)
    - [TLS Configuration](#tls-configuration)
    - [SSL Verification and CA Certs Configuration](#ssl-verification-and-ca-certs-configuration)
  - [Server](#server)
    - [HTTP Methods : Serve a GET Request](#http-methods--serve-a-get-request)
    - [Serve a Post Request](#serve-a-post-request)
    - [Serve a Put Request](#serve-a-put-request)
    - [Serve a Delete Request](#serve-a-delete-request)
    - [Serve a Request with Parameters](#serve-a-request-with-parameters)
- [Documentation](#documentation)

---

## Features

### Client Features

- HTTP methods: GET, POST, PUT, DELETE
- Query parameters
- Request headers
- Retry
- CircuitBreaker Configuration
- Proxy Configuration
- TLS Configuration
- Transport Layer Configuration
  - MaxIdle Connections
  - Connection Timeout
  - TLS Handshake Timeout
- SSL Verification and Configuration
- CA Certs Configuration
- Error handling
  - ErrorOnHttpStatus : sets the list of status codes that can be considered failures

### Server Features

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
go get oss.nandlabs.io/golly/rest
```

## Usage

### Client

#### HTTP Methods : Sending a GET Request

```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/rest/client"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### Retry Configuration

```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  // maxRetries -> 3, wait -> 5 seconds
  client.Retry(3, 5)
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### CircuitBreaker Configuration

```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  client.UseCircuitBreaker(1, 2, 1, 3)
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### Proxy Configuration

```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  err := client.SetProxy("proxy:url", "proxy_user", "proxy_pass")
  if err != nil {
	  fmt.Errorf("unable to set proxy: %v", err)
  }
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### TLS Configuration

```go
package main

import (
  "crypto/tls"
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  client, err := client.SetTLSCerts(tls.Certificate{})
  if err != nil {
    fmt.Errorf("error adding tls certificates: %v", err)
  }
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### SSL Verification and CA Certs Configuration

```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  client, err := client.SSlVerify(true)
  if err != nil {
	  fmt.Errorf("unable to set ssl verification, %v", err)
  }
  client, err = client.SetCACerts("./test-cert.pem", "./test-cert-2.pem")
  if err != nil {
    fmt.Errorf("error adding ca certificates: %v", err)
  }
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

### Server

#### HTTP Methods : Serve a GET Request

```go
package main

import (
	"net/http"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
)

func main() {
	// Create a new server
	// Checkout  New(opts *SrvOptions) for customisng the server properties
	srv, err := rest.DefaultServer()
	if err != nil {
		panic(err)
	}
	// this is the path prefix for each endpoint. Default is empty and no path prefix is added
	srv.Opts().PathPrefix = "/api/v1"
	// Add a GET endpoint
	srv.Get("healthCheck", func(ctx rest.ServerContext) {
		// Set the status code. Remember to set the status code before writing the response
		ctx.SetStatusCode(http.StatusOK)
		ctx.WriteString("Health Check Get")
	})
	// Add a POST endpoint
	srv.Post("healthCheck", func(ctx rest.ServerContext) {
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

#### Serve a Post Request

```go
package main

import (
	"net/http"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
)

func main() {
	// Create a new server
	// Checkout  New(opts *Options) for customisng the server properties
	srv, err := rest.DefaultServer()
	if err != nil {
		panic(err)
	}
	// this is the path prefix for each endpoint. Default is empty and no path prefix is added
	srv.Opts().PathPrefix = "/api/v1"
	// Add a POST endpoint
	srv.Post("healthCheck", func(ctx rest.ServerContext) {
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

#### Serve a Put Request

```go
package main

import (
	"net/http"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
)

func main() {
	// Create a new server
	// Checkout  New(opts *Options) for customisng the server properties
	srv, err := rest.DefaultServer()
	if err != nil {
		panic(err)
	}
	// this is the path prefix for each endpoint. Default is empty and no path prefix is added
	srv.Opts().PathPrefix = "/api/v1"
	// Add a PUT endpoint
	srv.Put("healthCheck", func(ctx rest.ServerContext) {
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

#### Serve a Delete Request

```go
package main

import (
	"net/http"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
)

func main() {
	// Create a new server
	// Checkout  New(opts *Options) for customisng the server properties
	srv, err := rest.DefaultServer()
	if err != nil {
		panic(err)
	}
	// this is the path prefix for each endpoint. Default is empty and no path prefix is added
	srv.Opts().PathPrefix = "/api/v1"
	// Add a DELETE endpoint
	srv.Delete("healthCheck", func(ctx rest.ServerContext) {
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

#### Serve a Request with Parameters

```go
package main

import (
	"net/http"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
)

func main() {
	// Create a new server
	// Checkout  New(opts *Options) for customisng the server properties
	srv, err := rest.DefaultServer()
	if err != nil {
		panic(err)
	}
	// this is the path prefix for each endpoint. Default is empty and no path prefix is added
	srv.Opts().PathPrefix = "/api/v1"
	// Add a GET endpoint
	srv.Get("endpoint/:pathparam", func(ctx rest.ServerContext) {
		// Get the query parameters
		queryParams := ctx.GetParam("paramName",rest.QueryParam)
		// Get the path parameters
		pathParams := ctx.GetParam("pathparam",rest.PathParam)
		// Do something with the query parameters and pathParameters
	})
	// get the component manager
	mgr := lifecycle.NewSimpleComponentManager()
	// Register the server with the component manager
	mgr.Register(srv)
	// Start the server
	mgr.StartAndWait()
}

```

### Documentation

See rest package [documentation](https://pkg.go.dev/oss.nandlabs.io/golly/rest) for more details.
