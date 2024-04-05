// Package messaging provides a set of utilities for working with messaging systems.
// It includes functionality for sending and receiving messages, as well as managing message queues.

// Example usage:
//
//	// Create a new message sender
//	sender := messaging.NewSender("localhost:5672", "guest", "guest")
//
//	// Send a message to a queue
//	err := sender.SendMessage("my-queue", "Hello, World!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a new message receiver
//	receiver := messaging.NewReceiver("localhost:5672", "guest", "guest")
//
//	// Receive messages from a queue
//	messages, err := receiver.ReceiveMessages("my-queue", 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, message := range messages {
//	    fmt.Println(message)
//	}
//
//	// Close the sender and receiver
//	sender.Close()
//	receiver.Close()
//
// This package supports various messaging protocols, including AMQP and MQTT.
// It provides a simple and consistent API for interacting with different messaging systems.
// The `Sender` type is used for sending messages, while the `Receiver` type is used for receiving messages.
// Both types provide methods for connecting to a messaging server, sending/receiving messages, and closing the connection.
//
// Note: This package requires a messaging server to be running in order to send/receive messages.
// Please refer to the documentation of the specific messaging protocol for more information on how to set up a server.
package messaging
