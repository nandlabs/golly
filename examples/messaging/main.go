// Package main demonstrates the messaging package with the local (channel-based) provider.
package main

import (
	"fmt"
	"net/url"
	"time"

	"oss.nandlabs.io/golly/messaging"
)

func main() {
	// Get the messaging manager (local provider is registered by default)
	manager := messaging.GetManager()

	// Parse a destination URL — the "chan" scheme uses in-memory channels
	destination, _ := url.Parse("chan://example/notifications")

	// Send messages
	fmt.Println("--- Sending messages ---")
	for i := 1; i <= 3; i++ {
		msg, _ := messaging.NewLocalMessage()
		msg.SetBodyBytes([]byte(fmt.Sprintf("Hello #%d", i)))
		err := manager.Send(destination, msg)
		if err != nil {
			fmt.Println("Send error:", err)
		} else {
			fmt.Printf("Sent message #%d\n", i)
		}
	}

	// Receive messages one at a time
	fmt.Println("\n--- Receiving messages ---")
	for i := 0; i < 3; i++ {
		msg, err := manager.Receive(destination)
		if err != nil {
			fmt.Println("Receive error:", err)
			break
		}
		body := msg.ReadAsStr()
		fmt.Printf("Received: %s\n", body)
	}

	// Demonstrate listener-based (push) model
	fmt.Println("\n--- Using a listener ---")
	listenerDest, _ := url.Parse("chan://example/events")
	done := make(chan struct{})

	err := manager.AddListener(listenerDest, func(msg messaging.Message) {
		body := msg.ReadAsStr()
		fmt.Printf("Listener got: %s\n", body)
		close(done)
	})
	if err != nil {
		fmt.Println("AddListener error:", err)
		return
	}

	// Send a message that the listener will receive
	msg, _ := messaging.NewLocalMessage()
	msg.SetBodyBytes([]byte("Event: user-signup"))
	_ = manager.Send(listenerDest, msg)

	// Wait for listener to process
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		fmt.Println("Timeout waiting for listener")
	}

	fmt.Println("\nDone.")
}
