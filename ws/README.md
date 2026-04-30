# ws — WebSocket Client & Server

A zero-dependency WebSocket implementation for Go following [RFC 6455](https://datatracker.ietf.org/doc/html/rfc6455).

## Features

- **Server handler** — `http.Handler` compatible, works with `net/http` and `turbo` router
- **Client dialer** — with automatic reconnection and exponential backoff
- **Thread-safe** — concurrent `Send()` from multiple goroutines
- **Ping/pong heartbeats** — configurable interval and timeout
- **Message size limits** — protect against oversized payloads
- **Fragmented messages** — automatic reassembly of multi-frame messages
- **Connection management** — broadcast, iterate, and track active connections
- **Functional options** — clean configuration via `With*` option functions

## Installation

```bash
go get oss.nandlabs.io/golly
```

## Quick Start

### Server

```go
package main

import (
    "fmt"
    "net/http"

    "oss.nandlabs.io/golly/ws"
)

func main() {
    handler := ws.NewHandler(
        ws.WithPingInterval(30 * time.Second),
        ws.WithMaxMessageSize(64 * 1024),
    )

    handler.OnConnect(func(conn *ws.Conn) {
        fmt.Printf("connected: %s\n", conn.ID())
    })

    handler.OnMessage(func(conn *ws.Conn, msg ws.Message) {
        // Echo back
        conn.Send(msg)
    })

    handler.OnDisconnect(func(conn *ws.Conn, err error) {
        fmt.Printf("disconnected: %s\n", conn.ID())
    })

    http.Handle("/ws", handler)
    http.ListenAndServe(":8080", nil)
}
```

### Client

```go
package main

import (
    "context"
    "fmt"

    "oss.nandlabs.io/golly/ws"
)

func main() {
    client, err := ws.Dial(context.Background(), "ws://localhost:8080/ws",
        ws.WithAutoReconnect(true),
        ws.WithMaxReconnectWait(30 * time.Second),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Send a message
    client.Send(ws.NewTextMessage([]byte("hello")))

    // Read messages
    for msg := range client.Messages() {
        fmt.Printf("received: %s\n", msg.Data)
    }
}
```

### Standalone Upgrade

For one-off upgrades without the full `Handler`:

```go
http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    conn, err := ws.Upgrade(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer conn.Close()

    // Use conn.Send() and set conn.OnMessage directly
})
```

### Broadcasting

```go
handler.OnMessage(func(conn *ws.Conn, msg ws.Message) {
    // Send to all connected clients
    handler.Broadcast(msg)
})
```

## Configuration Options

| Option                  | Default | Description                                  |
| ----------------------- | ------- | -------------------------------------------- |
| `WithReadBufferSize`    | 4096    | Read buffer size in bytes                    |
| `WithWriteBufferSize`   | 4096    | Write buffer size in bytes                   |
| `WithMaxMessageSize`    | 64KB    | Maximum allowed message size                 |
| `WithPingInterval`      | 30s     | Interval between ping frames (0 to disable)  |
| `WithPongTimeout`       | 10s     | Time to wait for pong response               |
| `WithWriteTimeout`      | 10s     | Timeout for write operations                 |
| `WithHandshakeTimeout`  | 10s     | Timeout for WebSocket handshake              |
| `WithTLSConfig`         | nil     | TLS configuration for `wss://` connections   |
| `WithAutoReconnect`     | false   | Enable automatic client reconnection         |
| `WithMaxReconnectWait`  | 30s     | Maximum wait between reconnection attempts   |
| `WithCheckOrigin`       | nil     | Origin validation function for server        |

## Architecture

```
ws/
├── doc.go       # Package documentation
├── pkg.go       # Package-level logger
├── errors.go    # Sentinel errors
├── message.go   # Message types, opcodes, close codes
├── frame.go     # RFC 6455 frame encoding/decoding
├── options.go   # Functional options
├── conn.go      # WebSocket connection (read/write pumps, ping/pong)
├── server.go    # HTTP handler, upgrade, connection management
└── client.go    # Dialer, auto-reconnect, message channel
```
