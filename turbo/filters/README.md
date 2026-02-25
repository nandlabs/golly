# Turbo Filters Package

The `turbo/filters` package provides HTTP request/response filter middleware for the Turbo router, including CORS support.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- **CORS Filter**: Configurable Cross-Origin Resource Sharing middleware
- Support for allowed origins, methods, headers, and credentials
- Easy integration with the Turbo router

## Usage

```go
import (
    "oss.nandlabs.io/golly/turbo"
    "oss.nandlabs.io/golly/turbo/filters"
)

// Create a CORS filter with allowed origins
cors := filters.NewCorsFilter("https://example.com", "https://app.example.com")

// Configure CORS options
opts := &filters.CorsOptions{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
}

// Attach to a Turbo router
router := turbo.NewRouter()
router.AddCorsFilter(opts)
```
