# Turbo Auth Package

The `turbo/auth` package provides authentication middleware (filters) for the Turbo HTTP router.

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

- **Basic Auth Filter**: Middleware for HTTP Basic Authentication
- Implements the `Authenticator` interface for extensibility

## Usage

```go
import "oss.nandlabs.io/golly/turbo/auth"

// Create a Basic Auth filter
basicAuth := auth.CreateBasicAuthAuthenticator()
```

The `Authenticator` interface can be implemented for custom authentication strategies.
