# clients

golly clients is a versatile client library written in Go that provides a
unified interface for interacting with various services with the aim of common
set of features like circuit breaker and retry handler. It allows you to
communicate with different types of services, including REST APIs and messaging
systems along using a consistent and easy-to-use interface.

---

- [Features](#features)
- [Sub-Clients](#sub-clients)
- [Getting Started](#getting-started)
- [Documentation](#documentation)

---

### Features

The Generic Golang Client offers the following features:

1. **Modularity**: The client is designed with a modular architecture, allowing
   you to use specific sub-clients for different types of services.
2. **Sub-clients**: The library provides sub-clients for interacting with
   different services, including REST, messaging, and more. You can selectively
   import and use the required sub-client(s) based on your needs.

### Sub-clients

The Generic Golang Client includes the following sub-clients:

1. **REST Client**<br> The REST client enables communication with RESTful APIs.
   It provides methods for making HTTP requests, handling responses, and
   managing authentication.
2. **Messaging Client**<br> The messaging client allows you to interact with
   messaging systems such as RabbitMQ or Apache Kafka. It provides functionality
   for sending and receiving messages, managing message queues, and handling
   message processing.

## Getting Started

To start using the Generic Golang Client, follow these steps:

1. Install Go and set up your Go development environment.
2. Import the Generic Golang Client into your project:
   ```go
   import "go.nandlabs.io/golly/clients"
   ```
3. Depending on the service you want to interact with, import the relevant
   sub-client:
   ```go
   import "go.nandlabs.io/golly/clients/rest"
   ```
   or
   ```go
   import "go.nandlabs.io/golly/clients/messaging"
   ```
   You can import multiple sub-clients if needed.
4. Initialize the sub-client and start using its functionality. Refer to the
   sub-client's documentation for detailed instructions on how to use it.

### Documentation

For detailed information on how to use the Generic Golang Client and its
sub-clients, refer to the following documentation:

- [REST Client](rest/README.md)
- [Messaging Client](messaging/README.md)
