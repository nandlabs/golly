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
- [clients](clients/README.md)
  - A common package for all types of client
  - Checkout clients that leverage this package.
    - [rest](clients/rest/README.md)
    - [messaging](messaging/README.md)
- [cli](cli/README.md)
  - Easy to use API for building complex command structures
  - Argument parsing and validation
- [codec](codec/README.md)
  - Easy to use interface
  - Multiformat support
  - Unifed interface for Endcoding and Decoding data from structured format
  - Out of the box support for `XML` `JSON` & `YAML`
- [collections](collections/README.md)
  - A collection of generic data structures
  - Stack, Queue, List,LinkedList, Set
  - Synchronized versions of the above
- [genai](genai/README.md)
  - Interact with generative AI models
  - Manage sessions, exchanges, and models
  - Handle memory and templates
- [l3](l3/README.md)
  - Lightweight Levelled Logger
  - Multiple logging levels `OFF,ERROR,INFO,DEBUG,TRACE`
  - Console and File based writers
  - Ability to specify log levels for a specific package
  - Async logging support
  - Configuration can be done using either a file,env variables,Struct values at
    runtime.
- [messaging](messaging/README.md)
  - General producer interface for sending messages to different messaging
    platforms.
  - General consumer interface for receiving and processing messages from
    different messaging platforms.
  - A local provider interface for messaging using channels
- rest
  [server](rest/server/README.md)

  - HTTP methods: GET, POST, PUT, DELETE
  - Query parameters
  - Request headers
  - Middleware support
  - TLS Configuration
  - Transport Layer Configuration

  [client](clients/rest/README.md)

  - HTTP methods: GET, POST, PUT, DELETE
  - Query parameters
  - Request headers
  - Proxy Configuration
  - TLS Configuration
  - Transport Layer Configuration
  - SSL Configuration
  - Error handling

- [semver](semver/README.md)
  - Adheres to the [SemVer 2.0.0](https://semver.org/spec/v2.0.0.html)
    specification
  - Easy to use API for parsing, comparing and generating SemVer versions
  - Supports pre-release and build metadata
- [turbo](turbo/README.md)
  - Smart Http Routing Capabilities
  - Aimed for API Development
  - Easy to use
  - Filters
- [vfs](vfs/README.md)
  - Virtual File System
  - Unified interface for multiple file systems
  - Default implementation for local fs available
  - Extensible framework
- [testing/assert](testing/assert/README.md)
  - Flexible and extensible assertion library

And many more...
Refer to [Godocs](https://godoc.org/oss.nandlabs.io/golly?) for more information

## Contributing

We welcome contributions to the project. If you find a bug or would like to
request a new feature, please open an issue on
[GitHub](https://github.com/nandlabs/golly/issues).

## License

This project is licensed under MIT License. See the [License](LICENSE) file for
details.
