package genai

import (
	"bytes"
	"testing"

	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/testing/assert"
)

func TestNewExchange(t *testing.T) {
	exchange := NewExchange("test-id")
	assert.NotNil(t, exchange)
	assert.Equal(t, "test-id", exchange.Id())
	assert.NotNil(t, exchange.Attributes())
	assert.Empty(t, exchange.Messages())
}

func TestExchange_MsgsByMime(t *testing.T) {
	exchange := NewExchange("test-id")
	exchange.AddTxtMsg("test message", UserActor)
	messages := exchange.MsgsByMime(ioutils.MimeTextPlain)
	assert.Len(t, messages, 1)
	assert.Equal(t, ioutils.MimeTextPlain, messages[0].Mime())
}

func TestExchange_MsgsByActor(t *testing.T) {
	exchange := NewExchange("test-id")
	exchange.AddTxtMsg("test message", UserActor)
	messages := exchange.MsgsByActors(UserActor)
	assert.Len(t, messages, 1)
	assert.Equal(t, UserActor, messages[0].Actor())
}

func TestExchange_AddMsg(t *testing.T) {
	exchange := NewExchange("test-id")
	exchange.AddTxtMsg("test message", UserActor)

	assert.Len(t, exchange.Messages(), 1)
}

func TestExchange_AddTxtMsg(t *testing.T) {
	exchange := NewExchange("test-id")
	msg, err := exchange.AddTxtMsg("test message", UserActor)
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, "test message", readMessageContent(t, msg))
	assert.Equal(t, UserActor, msg.Actor())
	assert.Equal(t, ioutils.MimeTextPlain, msg.Mime())
}

func TestExchange_AddJsonMsg(t *testing.T) {
	exchange := NewExchange("test-id")
	data := map[string]string{"key": "value"}
	msg, err := exchange.AddJsonMsg(data, UserActor)
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, UserActor, msg.Actor())
	assert.Equal(t, ioutils.MimeApplicationJSON, msg.Mime())
}

func TestExchange_AddFileMsg(t *testing.T) {
	exchange := NewExchange("test-id")
	fileURL := "http://example.com/file.txt"
	msg, err := exchange.AddFileMsg(fileURL, ioutils.MimeTextPlain, UserActor)
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, UserActor, msg.Actor())
	assert.Equal(t, ioutils.MimeTextPlain, msg.Mime())
	assert.Equal(t, fileURL, msg.URL().String())
}

func TestExchange_AddBinMsg(t *testing.T) {
	exchange := NewExchange("test-id")
	data := []byte("binary data")
	msg, err := exchange.AddBinMsg(data, ioutils.MimeApplicationOctetStream, UserActor)
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, UserActor, msg.Actor())
	assert.Equal(t, ioutils.MimeApplicationOctetStream, msg.Mime())
	assert.Equal(t, data, readMessageContentBytes(t, msg))
}

func readMessageContent(t *testing.T, msg *Message) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(msg)
	assert.NoError(t, err)
	return buf.String()
}

func readMessageContentBytes(t *testing.T, msg *Message) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(msg)
	assert.NoError(t, err)
	return buf.Bytes()
}
