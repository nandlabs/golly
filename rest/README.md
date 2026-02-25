# REST Package

The `rest` package provides a comprehensive HTTP client and server framework for building RESTful APIs in Go. It includes a feature-rich HTTP client with retry, circuit breaker, and context support, as well as a server built on top of the `oss.nandlabs.io/golly/turbo` router with lifecycle management.

---

- [Features](#features)
- [Installation](#installation)
- [Client](#client)
  - [Creating a Client](#creating-a-client)
  - [Making Requests](#making-requests)
  - [Request Context](#request-context)
  - [Request Body and Content Types](#request-body-and-content-types)
  - [Query Parameters](#query-parameters)
  - [Path Parameters](#path-parameters)
  - [Headers](#headers)
  - [Multipart File Upload](#multipart-file-upload)
  - [Handling Responses](#handling-responses)
  - [Retry Configuration](#retry-configuration)
  - [Circuit Breaker](#circuit-breaker)
  - [Error on HTTP Status](#error-on-http-status)
  - [Authentication](#authentication)
  - [OAuth 2.0](#oauth-20)
  - [Proxy Configuration](#proxy-configuration)
  - [TLS and SSL Configuration](#tls-and-ssl-configuration)
  - [Timeout Configuration](#timeout-configuration)
  - [Client Options Builder](#client-options-builder)
- [Server](#server)
  - [Creating a Server](#creating-a-server)
  - [Server Options](#server-options)
  - [Registering Routes](#registering-routes)
  - [ServerContext](#servercontext)
  - [Reading Request Data](#reading-request-data)
  - [Writing Responses](#writing-responses)
  - [Path and Query Parameters](#path-and-query-parameters)
  - [Request Context (Server)](#request-context-server)
  - [Unhandled and Unsupported Routes](#unhandled-and-unsupported-routes)
  - [Global Filters](#global-filters)
  - [CORS Configuration](#cors-configuration)
  - [TLS / HTTPS Server](#tls--https-server)
  - [Loading Server Config from File](#loading-server-config-from-file)
  - [Lifecycle Management](#lifecycle-management)
  - [Complete Server Example](#complete-server-example)
- [Documentation](#documentation)

---

## Features

### Client Features

- **HTTP Methods** — GET, POST, PUT, DELETE, PATCH and any custom method
- **Request Building** — fluent API for headers, query params, path params, body, and content type
- **Context Support** — `context.Context` propagation for cancellation and deadlines
- **Automatic Codec** — request/response serialization via JSON, XML, or YAML based on content type
- **Retry** — configurable retry with exponential backoff, respects context cancellation
- **Circuit Breaker** — prevents cascading failures by short-circuiting on repeated errors
- **Error on Status** — treat specific HTTP status codes as errors to trigger retries
- **Authentication** — built-in Basic and Bearer auth handlers, extensible via custom `AuthHandlerFunc`
- **OAuth 2.0** — client-credentials flow with automatic token refresh
- **Proxy** — HTTP proxy with optional basic authentication
- **TLS/SSL** — custom CA certificates, client certificates, and SSL verification toggle
- **Transport Tuning** — idle connections, timeouts, TLS handshake timeout, expect-continue timeout
- **Cookie Jar** — optional `http.CookieJar` for session management
- **Multipart Upload** — file uploads via multipart/form-data
- **Base URL** — set a base URL so requests only need relative paths

### Server Features

- **HTTP Methods** — `Get`, `Post`, `Put`, `Delete`, and generic `AddRoute` for any method
- **Turbo Router** — high-performance routing via `oss.nandlabs.io/golly/turbo`
- **ServerContext** — unified request/response context with codec-aware read/write helpers
- **Context Propagation** — access the request's `context.Context` for cancellation and deadlines
- **Codec-Aware I/O** — `Read`, `WriteJSON`, `WriteXML`, `WriteYAML`, and generic `Write` by content type
- **Path & Query Parameters** — extract path (`:param`) and query parameters
- **Path Prefix** — configure a global path prefix for all routes
- **Unhandled / Unsupported** — custom handlers for 404 and 405 responses
- **Global Filters** — middleware applied to all routes (logging, auth, etc.)
- **CORS** — built-in CORS filter with configurable origins, methods, and max age
- **TLS / HTTPS** — serve over HTTPS with certificate and private key
- **Config from File** — load `SrvOptions` from JSON, YAML, or XML configuration files
- **Lifecycle Management** — integrates with `oss.nandlabs.io/golly/lifecycle` for graceful start/stop

## Installation

```bash
go get oss.nandlabs.io/golly/rest
```

---

## Client

### Creating a Client

```go
// Default client with sensible defaults
client := rest.NewClient()

// Client with custom options via builder
opts := rest.CliOptsBuilder().
    RequestTimeoutMs(30000).
    MaxIdlePerHost(50).
    Build()
client := rest.NewClientWithOptions(opts)
```

**Default values:**

| Setting                   | Default |
| ------------------------- | ------- |
| Max idle connections/host | 20      |
| Request timeout           | 60s     |
| Idle connection timeout   | 90s     |
| TLS handshake timeout     | 10s     |
| SSL verification          | enabled |

### Making Requests

```go
// Create a request
req, err := client.NewRequest("https://api.example.com/users", http.MethodGet)
if err != nil {
    log.Fatal(err)
}

// Execute the request
resp, err := client.Execute(req)
if err != nil {
    log.Fatal(err)
}
```

### Request Context

Requests support `context.Context` for cancellation, deadlines, and timeouts. By default, every request is initialized with `context.Background()`.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, err := client.NewRequest("https://api.example.com/users", http.MethodGet)
if err != nil {
    log.Fatal(err)
}

// Attach context — returns error if ctx is nil
req, err = req.WithContext(ctx)
if err != nil {
    log.Fatal(err)
}

resp, err := client.Execute(req)
if err != nil {
    // May return context.DeadlineExceeded or context.Canceled
    log.Fatal(err)
}
```

Context is respected during retries — if the context is cancelled between retry attempts, the retry loop exits immediately.

### Request Body and Content Types

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

req, _ := client.NewRequest("https://api.example.com/users", http.MethodPost)

// Set a struct body — automatically serialized using the codec matching the content type
req.SetBody(User{Name: "Alice", Email: "alice@example.com"})
req.SetContentType("application/json")

// Or use a raw io.Reader
req.SeBodyReader(strings.NewReader(`{"name":"Alice"}`))
req.SetContentType("application/json")
```

Supported content types include `application/json`, `application/xml`, `text/xml`, and `text/yaml`. The codec is resolved automatically from the content type.

### Query Parameters

```go
req, _ := client.NewRequest("https://api.example.com/search", http.MethodGet)

// Fluent chaining
req.AddQueryParam("q", "golang").
    AddQueryParam("page", "1").
    AddQueryParam("limit", "20")
// Resulting URL: https://api.example.com/search?q=golang&page=1&limit=20
```

### Path Parameters

Path parameters use `${paramName}` placeholders in the URL:

```go
req, _ := client.NewRequest("https://api.example.com/users/${userId}/posts/${postId}", http.MethodGet)
req.AddPathParam("userId", "42").
    AddPathParam("postId", "100")
// Resulting URL: https://api.example.com/users/42/posts/100
```

### Headers

```go
req, _ := client.NewRequest("https://api.example.com/data", http.MethodGet)
req.AddHeader("Authorization", "Bearer my-token").
    AddHeader("Accept", "application/json").
    AddHeader("X-Custom-Header", "value1", "value2") // multiple values
```

### Multipart File Upload

```go
req, _ := client.NewRequest("https://api.example.com/upload", http.MethodPost)
req.SetMultipartFiles(
    &rest.MultipartFile{ParamName: "file", FilePath: "/path/to/document.pdf"},
    &rest.MultipartFile{ParamName: "avatar", FilePath: "/path/to/photo.jpg"},
)
```

Multipart uploads are only allowed with POST, PUT, and PATCH methods.

### Handling Responses

```go
resp, err := client.Execute(req)
if err != nil {
    log.Fatal(err)
}

// Check success (status 200–204)
if resp.IsSuccess() {
    var user User
    err := resp.Decode(&user) // auto-detects codec from Content-Type header
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(user.Name)
} else {
    fmt.Printf("Error: %s (status %d)\n", resp.Status(), resp.StatusCode())
}

// Access the raw *http.Response if needed
raw := resp.Raw()
```

### Retry Configuration

Configure retries via the `ClientOptsBuilder`. The `RetryPolicy` method accepts parameters for max retries, base backoff interval, whether to use exponential backoff, and a max backoff cap:

```go
// Fixed backoff: retry up to 3 times, wait 1s between each attempt
opts := rest.CliOptsBuilder()
opts.RetryPolicy(3, 1000, false, 0)
client := rest.NewClientWithOptions(opts.Build())
```

**Exponential backoff:** The wait time doubles on each retry (base × 2^attempt), capped at a maximum:

```go
// Exponential backoff: retry up to 5 times
//   attempt 0: 500ms, attempt 1: 1s, attempt 2: 2s, attempt 3: 4s, attempt 4: 5s (capped)
opts := rest.CliOptsBuilder()
opts.RetryPolicy(
    5,     // maxRetries
    500,   // backoffIntervalMs (base wait)
    true,  // exponential
    5000,  // maxBackoffInMs (cap)
)
client := rest.NewClientWithOptions(opts.Build())
```

**Combining with error-on-status:** Treat specific status codes as errors to trigger retries:

```go
opts := rest.CliOptsBuilder().
    ErrOnStatus(429, 500, 502, 503, 504)
opts.RetryPolicy(3, 1000, true, 10000)
client := rest.NewClientWithOptions(opts.Build())
```

Retries are context-aware — if a request's context is cancelled or times out during the backoff wait, the retry loop exits immediately with the context error.

### Circuit Breaker

The circuit breaker prevents sending requests to an endpoint that is consistently failing. It transitions between **closed** (normal), **open** (blocking), and **half-open** (probing) states:

```go
opts := rest.CliOptsBuilder()
opts.CircuitBreaker(
    3,  // failureThreshold — consecutive failures to trip the circuit
    2,  // successThreshold — consecutive successes in half-open to close it
    1,  // maxHalfOpen — max concurrent requests allowed in half-open state
    30, // timeout — seconds before transitioning from open to half-open
)
client := rest.NewClientWithOptions(opts.Build())
```

When the circuit is open, `Execute` returns an error immediately without making the HTTP call.

### Error on HTTP Status

Treat specific HTTP status codes as errors (e.g., to trigger retries on 429 or 503):

```go
opts := rest.CliOptsBuilder().
    ErrOnStatus(429, 500, 502, 503, 504).
    Build()

client := rest.NewClientWithOptions(opts)
```

### Authentication

**Basic Auth:**

```go
opts := rest.CliOptsBuilder()
opts.AuthType(clients.AuthTypeBasic).
    AuthUser("username").
    AuthPass("password")

client := rest.NewClientWithOptions(opts.Build())
```

**Bearer Token:**

```go
opts := rest.CliOptsBuilder()
opts.AuthType(clients.AuthTypeBearer).
    AuthToken("my-api-token")

client := rest.NewClientWithOptions(opts.Build())
```

**Custom Auth Handler:**

```go
opts := rest.CliOptsBuilder().Build()
opts.AuthHandlers[myCustomAuthType] = func(client *rest.Client, req *http.Request) error {
    req.Header.Set("X-API-Key", "my-key")
    return nil
}

client := rest.NewClientWithOptions(opts)
```

### OAuth 2.0

The package includes a built-in OAuth 2.0 client-credentials provider with automatic token refresh:

```go
provider := rest.NewOAuth2Provider(
    "client-id",
    "client-secret",
    "client_credentials",
    "https://auth.example.com/oauth2/token",
)

// Optionally add extra form parameters
provider.AddParam("scope", "read write")

// Use with the client options builder
opts := rest.CliOptsBuilder()
opts.AuthType(clients.AuthTypeBearer).
    AuthProvider(provider)

client := rest.NewClientWithOptions(opts.Build())
```

Tokens are cached and refreshed automatically before expiry.

### Proxy Configuration

```go
opts := rest.CliOptsBuilder().
    ProxyAuth("proxy_user", "proxy_password", "").
    Build()

client := rest.NewClientWithOptions(opts)
```

### TLS and SSL Configuration

```go
opts := rest.CliOptsBuilder().
    // Disable SSL verification (not recommended for production)
    SSLVerify(true).
    // Add custom CA certificates
    CaCerts("/path/to/ca-cert.pem", "/path/to/ca-cert-2.pem").
    // Add client TLS certificates
    TlsCerts(clientCert).
    Build()

client := rest.NewClientWithOptions(opts)
```

### Timeout Configuration

```go
opts := rest.CliOptsBuilder().
    RequestTimeoutMs(30000).           // Overall request timeout: 30s
    IdleTimeoutMs(60000).              // Idle connection timeout: 60s
    TlsHandShakeTimeoutMs(15000).      // TLS handshake timeout: 15s
    ExpectContinueTimeoutMs(2000).     // Expect-Continue timeout: 2s
    Build()

client := rest.NewClientWithOptions(opts)
```

### Client Options Builder

Full example showing all builder options:

```go
opts := rest.CliOptsBuilder().
    MaxIdlePerHost(50).
    RequestTimeoutMs(30000).
    IdleTimeoutMs(60000).
    TlsHandShakeTimeoutMs(15000).
    ExpectContinueTimeoutMs(2000).
    SSLVerify(false).
    CaCerts("/path/to/ca.pem").
    TlsCerts(cert).
    ErrOnStatus(429, 500, 503).
    ProxyAuth("user", "pass", "").
    CookieJar(jar).
    CodecOpts(map[string]any{"indent": true}).
    Build()

client := rest.NewClientWithOptions(opts)
defer client.Close() // closes idle connections
```

---

## Server

### Creating a Server

```go
// Default server on localhost:8080
srv, err := rest.DefaultServer()
if err != nil {
    log.Fatal(err)
}

// Server with custom options
opts := rest.DefaultSrvOptions()
opts.ListenHost = "0.0.0.0"
opts.ListenPort = 9090
opts.PathPrefix = "/api/v1"

srv, err := rest.NewServer(opts)
if err != nil {
    log.Fatal(err)
}
```

### Server Options

The `SrvOptions` struct controls server behavior and can be configured programmatically or loaded from a file.

| Field            | Type           | Default                 | Description                       |
| ---------------- | -------------- | ----------------------- | --------------------------------- |
| `Id`             | `string`       | `"default-http-server"` | Unique server identifier          |
| `PathPrefix`     | `string`       | `"/"`                   | Global path prefix for all routes |
| `ListenHost`     | `string`       | `"localhost"`           | Bind address                      |
| `ListenPort`     | `int16`        | `8080`                  | Listen port                       |
| `ReadTimeout`    | `int64`        | `20000`                 | Read timeout in milliseconds      |
| `WriteTimeout`   | `int64`        | `20000`                 | Write timeout in milliseconds     |
| `EnableTLS`      | `bool`         | `false`                 | Enable HTTPS                      |
| `PrivateKeyPath` | `string`       | `""`                    | Path to TLS private key file      |
| `CertPath`       | `string`       | `""`                    | Path to TLS certificate file      |
| `Cors`           | `*CorsOptions` | _(see CORS section)_    | CORS configuration                |

Options support fluent setters:

```go
opts := rest.EmptySrvOptions()
opts.SetListenHost("0.0.0.0").
    SetListenPort(3000).
    SetEnableTLS(true).
    SetPrivateKeyPath("/path/to/key.pem").
    SetCertPath("/path/to/cert.pem")
opts.Id = "my-api-server"
```

### Registering Routes

Routes are registered using the HTTP method helpers or the generic `AddRoute`:

```go
// Method-specific helpers
srv.Get("users", handler)
srv.Post("users", handler)
srv.Put("users/:id", handler)
srv.Delete("users/:id", handler)

// Generic route with multiple methods
srv.AddRoute("users/:id", handler, http.MethodGet, http.MethodHead)
```

The `PathPrefix` is automatically prepended. For example, with `PathPrefix = "/api/v1"`, registering `"users"` creates the route `/api/v1/users`.

### ServerContext

Every handler receives a `ServerContext` that wraps the underlying `http.Request` and `http.ResponseWriter`. It provides convenience methods for reading input and writing output.

```go
type HandlerFunc func(context ServerContext)
```

**Key methods:**

| Method                                 | Description                                         |
| -------------------------------------- | --------------------------------------------------- |
| `Context() context.Context`            | Returns the request's context                       |
| `GetMethod() string`                   | HTTP method (GET, POST, etc.)                       |
| `GetURL() string`                      | Full request URL                                    |
| `GetHeader(name) string`               | Get a request header value                          |
| `InHeaders() http.Header`              | Clone of all request headers                        |
| `GetParam(name, type) (string, error)` | Get path or query parameter                         |
| `GetBody() (io.Reader, error)`         | Raw request body reader                             |
| `Read(obj) error`                      | Decode body into struct (auto-detects content type) |
| `GetRequest() *http.Request`           | Access underlying `*http.Request`                   |
| `SetStatusCode(code)`                  | Set response status code (call before writing)      |
| `SetHeader(name, value)`               | Set a response header                               |
| `SetContentType(contentType)`          | Set response Content-Type header                    |
| `SetCookie(cookie)`                    | Set a response cookie                               |
| `WriteJSON(data) error`                | Write JSON response                                 |
| `WriteXML(data) error`                 | Write XML response                                  |
| `WriteYAML(data) error`                | Write YAML response                                 |
| `Write(data, contentType) error`       | Write with explicit content type                    |
| `WriteData([]byte) (int, error)`       | Write raw bytes                                     |
| `WriteString(string)`                  | Write a string response                             |
| `WriteFrom(io.Reader)`                 | Stream from a reader to response                    |
| `HttpResWriter() http.ResponseWriter`  | Access underlying `http.ResponseWriter`             |

### Reading Request Data

**Decode a JSON/XML/YAML body into a struct:**

```go
type CreateUserInput struct {
    Name  string `json:"name" xml:"name"`
    Email string `json:"email" xml:"email"`
}

srv.Post("users", func(ctx rest.ServerContext) {
    var input CreateUserInput
    if err := ctx.Read(&input); err != nil {
        ctx.SetStatusCode(http.StatusBadRequest)
        ctx.WriteJSON(map[string]string{"error": err.Error()})
        return
    }
    // input is populated based on the request's Content-Type header
    fmt.Println(input.Name, input.Email)
})
```

**Read raw body:**

```go
srv.Post("raw", func(ctx rest.ServerContext) {
    body, _ := ctx.GetBody()
    data, _ := io.ReadAll(body)
    fmt.Println(string(data))
})
```

### Writing Responses

> **Important:** Always call `SetStatusCode` before writing the response body.

**JSON response:**

```go
srv.Get("users/:id", func(ctx rest.ServerContext) {
    user := User{Name: "Alice", Email: "alice@example.com"}
    ctx.SetStatusCode(http.StatusOK)
    ctx.WriteJSON(user)
})
```

**XML response:**

```go
srv.Get("users/:id", func(ctx rest.ServerContext) {
    user := User{Name: "Alice", Email: "alice@example.com"}
    ctx.SetStatusCode(http.StatusOK)
    ctx.WriteXML(user)
})
```

**Dynamic content type (matches Accept or explicit type):**

```go
srv.Get("users/:id", func(ctx rest.ServerContext) {
    user := User{Name: "Alice", Email: "alice@example.com"}
    ctx.SetStatusCode(http.StatusOK)
    ctx.Write(user, "application/yaml")
})
```

**String / raw bytes:**

```go
srv.Get("health", func(ctx rest.ServerContext) {
    ctx.SetStatusCode(http.StatusOK)
    ctx.WriteString("OK")
})
```

### Path and Query Parameters

Path parameters use the `:param` syntax in the route pattern. Query parameters are extracted from the URL query string.

```go
srv.Get("users/:userId/posts/:postId", func(ctx rest.ServerContext) {
    userId, err := ctx.GetParam("userId", rest.PathParam)
    if err != nil {
        ctx.SetStatusCode(http.StatusBadRequest)
        ctx.WriteJSON(map[string]string{"error": "missing userId"})
        return
    }

    postId, _ := ctx.GetParam("postId", rest.PathParam)

    // Query param: /users/42/posts/1?format=brief
    format, _ := ctx.GetParam("format", rest.QueryParam)

    ctx.SetStatusCode(http.StatusOK)
    ctx.WriteJSON(map[string]string{
        "userId": userId,
        "postId": postId,
        "format": format,
    })
})
```

### Request Context (Server)

Access the incoming request's `context.Context` for passing to downstream services, databases, or other context-aware operations:

```go
srv.Get("users/:id", func(ctx rest.ServerContext) {
    // Get the request context (carries deadlines, cancellation, and values)
    reqCtx := ctx.Context()

    user, err := userService.GetByID(reqCtx, userId)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            return // client disconnected
        }
        ctx.SetStatusCode(http.StatusInternalServerError)
        ctx.WriteJSON(map[string]string{"error": err.Error()})
        return
    }
    ctx.SetStatusCode(http.StatusOK)
    ctx.WriteJSON(user)
})
```

### Unhandled and Unsupported Routes

Customize the response for routes that don't match (404) or methods that aren't allowed (405):

```go
// Custom 404 handler
srv.Unhandled(func(ctx rest.ServerContext) {
    ctx.SetStatusCode(http.StatusNotFound)
    ctx.WriteJSON(map[string]string{
        "error":   "not_found",
        "message": "The requested resource does not exist",
    })
})

// Custom 405 handler
srv.Unsupported(func(ctx rest.ServerContext) {
    ctx.SetStatusCode(http.StatusMethodNotAllowed)
    ctx.WriteJSON(map[string]string{
        "error":   "method_not_allowed",
        "message": "This HTTP method is not supported for this endpoint",
    })
})
```

### Global Filters

Add middleware functions that run on every request (e.g., logging, authentication, request ID injection):

```go
srv.AddGlobalFilter(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
    start := time.Now()
    log.Printf("→ %s %s", r.Method, r.URL.Path)
    next.ServeHTTP(w, r)
    log.Printf("← %s %s [%v]", r.Method, r.URL.Path, time.Since(start))
})
```

### CORS Configuration

CORS is configured via the `Cors` field in `SrvOptions`. The default configuration allows all origins with GET, POST, PUT, and DELETE methods.

```go
opts := rest.DefaultSrvOptions()
opts.Cors = &filters.CorsOptions{
    MaxAge:         3600,
    AllowedOrigins: []string{"https://example.com", "https://app.example.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
    ResponseStatus: http.StatusNoContent,
}

srv, _ := rest.NewServer(opts)
```

### TLS / HTTPS Server

```go
opts := rest.DefaultSrvOptions()
opts.EnableTLS = true
opts.CertPath = "/path/to/cert.pem"
opts.PrivateKeyPath = "/path/to/key.pem"

srv, err := rest.NewServer(opts)
if err != nil {
    log.Fatal(err)
}
```

### Loading Server Config from File

Load server options from a JSON, YAML, or XML file:

```json
{
  "id": "my-api-server",
  "path_prefix": "/api/v1",
  "listen_host": "0.0.0.0",
  "listen_port": 8080,
  "read_timeout": 30000,
  "write_timeout": 30000,
  "enable_tls": false,
  "cors": {
    "max_age": 3600,
    "allowed_origins": ["*"],
    "allowed_methods": ["GET", "POST", "PUT", "DELETE"]
  }
}
```

```go
srv, err := rest.NewServerFrom("config/server.json")
if err != nil {
    log.Fatal(err)
}
```

### Lifecycle Management

The server implements `lifecycle.Component` and integrates with the component manager for graceful start and shutdown:

```go
srv, _ := rest.DefaultServer()

// Register routes...
srv.Get("health", healthHandler)

// Create a component manager
mgr := lifecycle.NewSimpleComponentManager()

// Register the server — the manager handles Start/Stop lifecycle
mgr.Register(srv)

// Block until the process receives a termination signal, then gracefully shuts down
mgr.StartAndWait()
```

### Complete Server Example

A full example with multiple routes, JSON I/O, parameters, error handling, filters, and lifecycle management:

```go
package main

import (
    "log"
    "net/http"
    "time"

    "oss.nandlabs.io/golly/lifecycle"
    "oss.nandlabs.io/golly/rest"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // Create server with custom options
    opts := rest.DefaultSrvOptions()
    opts.PathPrefix = "/api/v1"
    opts.ListenHost = "0.0.0.0"
    opts.ListenPort = 8080

    srv, err := rest.NewServer(opts)
    if err != nil {
        log.Fatal(err)
    }

    // Global logging filter
    srv.AddGlobalFilter(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
        start := time.Now()
        log.Printf("→ %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
        log.Printf("← %s %s [%v]", r.Method, r.URL.Path, time.Since(start))
    })

    // Health check
    srv.Get("health", func(ctx rest.ServerContext) {
        ctx.SetStatusCode(http.StatusOK)
        ctx.WriteJSON(map[string]string{"status": "healthy"})
    })

    // Get user by ID
    srv.Get("users/:id", func(ctx rest.ServerContext) {
        id, err := ctx.GetParam("id", rest.PathParam)
        if err != nil {
            ctx.SetStatusCode(http.StatusBadRequest)
            ctx.WriteJSON(map[string]string{"error": "missing id"})
            return
        }
        user := User{ID: id, Name: "Alice", Email: "alice@example.com"}
        ctx.SetStatusCode(http.StatusOK)
        ctx.WriteJSON(user)
    })

    // Create user
    srv.Post("users", func(ctx rest.ServerContext) {
        var user User
        if err := ctx.Read(&user); err != nil {
            ctx.SetStatusCode(http.StatusBadRequest)
            ctx.WriteJSON(map[string]string{"error": err.Error()})
            return
        }
        user.ID = "new-id"
        ctx.SetStatusCode(http.StatusCreated)
        ctx.WriteJSON(user)
    })

    // Update user
    srv.Put("users/:id", func(ctx rest.ServerContext) {
        id, _ := ctx.GetParam("id", rest.PathParam)
        var user User
        if err := ctx.Read(&user); err != nil {
            ctx.SetStatusCode(http.StatusBadRequest)
            ctx.WriteJSON(map[string]string{"error": err.Error()})
            return
        }
        user.ID = id
        ctx.SetStatusCode(http.StatusOK)
        ctx.WriteJSON(user)
    })

    // Delete user
    srv.Delete("users/:id", func(ctx rest.ServerContext) {
        id, _ := ctx.GetParam("id", rest.PathParam)
        ctx.SetStatusCode(http.StatusOK)
        ctx.WriteJSON(map[string]string{"deleted": id})
    })

    // Custom 404
    srv.Unhandled(func(ctx rest.ServerContext) {
        ctx.SetStatusCode(http.StatusNotFound)
        ctx.WriteJSON(map[string]string{"error": "not found"})
    })

    // Start with lifecycle management
    mgr := lifecycle.NewSimpleComponentManager()
    mgr.Register(srv)
    mgr.StartAndWait()
}
```

---

## Documentation

See the full API reference on [pkg.go.dev](https://pkg.go.dev/oss.nandlabs.io/golly/rest).
