// Package genai contains tests for the Message struct and its methods.
//
// TestMessage_Mime tests the Mime method of the Message struct to ensure it returns the correct MIME type.
//
// TestMessage_Actor tests the Actor method of the Message struct to ensure it returns the correct actor.
//
// TestMessage_Read tests the Read method of the Message struct to ensure it reads data correctly.
//
// TestMessage_SetActor tests the SetActor method of the Message struct to ensure it sets the actor correctly.
//
// TestMessage_SetMime tests the SetMime method of the Message struct to ensure it sets the MIME type correctly.
//
// TestMessage_Write tests the Write method of the Message struct to ensure it writes data correctly.
//
// TestMessage_URL tests the URL method of the Message struct to ensure it returns the correct URL.
package genai

import (
	"bytes"
	"net/url"
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestMessage_Mime tests the Mime method of the Message struct.
// It verifies that the Mime method returns the correct MIME type
// that was set in the Message instance.
func TestMessage_Mime(t *testing.T) {
	msg := &Message{mimeType: "text/plain"}
	assert.Equal(t, "text/plain", msg.Mime())
}

// TestMessage_Actor tests the Actor method of the Message struct.
// It verifies that the Actor method returns the correct actor that was set in the Message.
func TestMessage_Actor(t *testing.T) {
	actor := UserActor
	msg := &Message{msgActor: actor}
	assert.Equal(t, actor, msg.Actor())
}

// TestMessage_Read tests the Read method of the Message struct.
// It verifies that the method reads the correct number of bytes
// and that the data read matches the expected data.
func TestMessage_Read(t *testing.T) {
	data := []byte("test data")
	rwer := bytes.NewBuffer(data)
	msg := &Message{rwer: rwer}

	buf := make([]byte, len(data))
	n, err := msg.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, buf)
}

// TestMessage_SetActor tests the SetActor method of the Message struct.
// It verifies that the actor is correctly set and retrieved using the Actor method.
func TestMessage_SetActor(t *testing.T) {
	actor := UserActor
	msg := &Message{}
	msg.SetActor(actor)
	assert.Equal(t, actor, msg.Actor())
}

// TestMessage_SetMime tests the SetMime method of the Message struct.
// It verifies that the MIME type is correctly set and retrieved.
func TestMessage_SetMime(t *testing.T) {
	mime := "application/json"
	msg := &Message{}
	msg.SetMime(mime)
	assert.Equal(t, mime, msg.Mime())
}

// TestMessage_Write tests the Write method of the Message struct.
// It verifies that the data is correctly written to the underlying writer
// and that the number of bytes written and the written data match the expected values.
func TestMessage_Write(t *testing.T) {
	data := []byte("test data")
	rwer := bytes.NewBuffer(nil)
	msg := &Message{rwer: rwer}

	n, err := msg.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, rwer.Bytes())
}

// TestMessage_URL tests the URL method of the Message struct.
// It ensures that the URL method returns the correct URL that was set during the creation of the Message instance.
func TestMessage_URL(t *testing.T) {
	u, err := url.Parse("http://example.com")
	assert.NoError(t, err)
	msg := &Message{u: u}
	assert.Equal(t, u, msg.URL())
}
