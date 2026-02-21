[![report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/oss.nandlabs.io/golly)
[![testing](https://img.shields.io/github/actions/workflow/status/nandlabs/golly/go_ci.yml?branch=main&event=push&color=228B22)](https://github.com/nandlabs/golly/actions?query=event%3Apush+branch%3Amain+)
[![release](https://img.shields.io/github/v/release/nandlabs/golly?label=latest&color=228B22)](https://github.com/nandlabs/golly/releases/latest)
[![releaseDate](https://img.shields.io/github/release-date/nandlabs/golly?label=released&color=00ADD8)](https://github.com/nandlabs/golly/releases/latest)
[![godoc](https://godoc.org/oss.nandlabs.io/golly?status.svg)](https://pkg.go.dev/oss.nandlabs.io/golly)

# golly

Golly is a collection of reusable common utilities for go programming language.

## Goals

- Create reusable common collection of utilities targeting enterprise use cases
- Ensure the project is self-contained and minimise external dependencies.

## Installation

```bash
go get oss.nandlabs.io/golly
```

## Core Packages

- [assertion](assertion/README.md)
  - Unified interface for asserting conditions
  - Supports various assertion functions for different types of conditions
- [cli](cli/README.md)
  - Easy to use API for building complex command structures
  - Argument parsing and validation
- [clients](clients/README.md)
  - A common package for all types of client
  - Auth providers, retry with exponential backoff, and circuit breaker
- [codec](codec/README.md)
  - Unified interface for encoding and decoding data
  - Out of the box support for `XML`, `JSON` & `YAML`
  - [codec/validator](codec/validator/README.md) — Input validation utilities
- [collections](collections/README.md)
  - Generic data structures: Stack, Queue, List, LinkedList, Set
  - Synchronized (thread-safe) versions of all collections
- [config](config/README.md)
  - Environment variable helpers (`GetEnvAsString`, `GetEnvAsInt`, `GetEnvAsBool`)
  - Properties file loading and key-value management
- [data](data/README.md)
  - Pipeline key-value data container with typed extraction
  - JSON Schema generation from Go structs via reflection
- [errutils](errutils/README.md)
  - Custom formatted errors and multi-error aggregation
- [fnutils](fnutils/README.md)
  - Deferred and timed function execution utilities
- [fsutils](fsutils/README.md)
  - Filesystem utilities: path/file/dir existence checks, content type detection
- [genai](genai/README.md)
  - Provider-agnostic interface for Generative AI / LLM services
  - Message types: text, binary, file, JSON, YAML, multi-part
  - Prompt templates with variable substitution
  - [genai/impl](genai/impl/README.md) — OpenAI and Ollama provider implementations
- [ioutils](ioutils/README.md)
  - MIME type lookup, channel utilities, and checksum calculation
- [l3](l3/README.md)
  - Lightweight Levelled Logger
  - Multiple logging levels: `OFF`, `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`
  - Console and File based writers
  - Per-package log level configuration
  - Async logging support
- [lifecycle](lifecycle/README.md)
  - Component lifecycle management with dependency ordering
  - Start/stop hooks, state tracking, and change notifications
- [managers](managers/README.md)
  - Generic item manager for registering, retrieving, and listing named items
- [messaging](messaging/README.md)
  - General producer/consumer interfaces for messaging platforms
  - Local (channel-based) provider for in-process messaging
- [pool](pool/README.md)
  - Generic, thread-safe object pool with configurable min/max capacity
  - Automatic lifecycle management via creator/destroyer functions
- [rest](rest/README.md)
  - HTTP server with routing, middleware, TLS, and transport configuration
  - HTTP client with headers, query params, proxy, and error handling
- [secrets](secrets/README.md)
  - AES encryption and decryption for strings and byte slices
- [semver](semver/README.md)
  - Semantic versioning parser and comparator ([SemVer 2.0.0](https://semver.org/spec/v2.0.0.html))
  - Pre-release and build metadata support
- [testing](testing/README.md)
  - [testing/assert](testing/assert/README.md) — Lightweight assertion helpers for unit tests
- [textutils](textutils/README.md)
  - Named ASCII character constants for readable code
- [turbo](turbo/README.md)
  - Enterprise-grade HTTP routing for API development
  - Path/query parameters, filters, and CORS support
  - [turbo/auth](turbo/auth/README.md) — Authentication middleware
  - [turbo/filters](turbo/filters/README.md) — CORS and request/response filters
- [uuid](uuid/README.md)
  - UUID generation (V1, V2, V3, V4) and parsing
- [vfs](vfs/README.md)
  - Virtual File System with unified interface
  - Default local filesystem implementation, extensible for cloud storage

Refer to [pkg.go.dev](https://pkg.go.dev/oss.nandlabs.io/golly) for full API documentation.

## Contributing

We welcome contributions to the project. If you find a bug or would like to
request a new feature, please open an issue on
[GitHub](https://github.com/nandlabs/golly/issues).

## License

This project is licensed under MIT License. See the [License](LICENSE) file for
details.
