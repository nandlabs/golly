// Package ws provides WebSocket client and server implementations following RFC 6455.
//
// It offers a production-ready WebSocket solution with zero external dependencies,
// consistent with golly's minimal-dependency philosophy. The package includes:
//
//   - WebSocket server handler compatible with net/http and the golly turbo router
//   - WebSocket client dialer with automatic reconnection and configurable backoff
//   - Thread-safe connection management with ping/pong heartbeats
//   - Configurable message size limits, buffer sizes, and deadlines
//
// # Server Usage
//
//	handler := ws.NewHandler(
//	    ws.WithReadBufferSize(4096),
//	    ws.WithWriteBufferSize(4096),
//	    ws.WithPingInterval(30 * time.Second),
//	)
//
//	handler.OnMessage(func(conn *ws.Conn, msg ws.Message) {
//	    conn.Send(msg)
//	})
//
//	http.Handle("/ws", handler)
//
// # Client Usage
//
//	client, err := ws.Dial(context.Background(), "wss://example.com/ws",
//	    ws.WithAutoReconnect(true),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	err = client.Send(ws.NewTextMessage([]byte("hello")))
package ws
