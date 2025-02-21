package genai

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"oss.nandlabs.io/golly/ioutils"
)

// Message represents a structure that holds information about a message.
type Message struct {
	u        *url.URL
	rwer     io.ReadWriter
	mimeType string
	msgActor Actor
	done     bool
}

func NewTextMessage(text string, mimeType string) *Message {
	return &Message{
		rwer:     bytes.NewBufferString(text),
		mimeType: mimeType,
	}
}

func NewBinMessage(data []byte, mimeType string) *Message {
	return &Message{
		rwer:     bytes.NewBuffer(data),
		mimeType: mimeType,
	}
}
func NewFileMessage(u string, mimeType string) *Message {
	return &Message{
		u:        &url.URL{Path: u},
		mimeType: mimeType,
	}
}

// Mime returns the MIME type of the message

func (m *Message) Mime() string {
	return m.mimeType
}

// Actor returns the actor that sent the message
func (m *Message) Actor() Actor {
	return m.msgActor
}

// Read implements the io.Reader interface
func (m *Message) Read(p []byte) (n int, err error) {
	return m.rwer.Read(p)
}

// ReadFrom implements the io.ReaderFrom interface
func (m *Message) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(m.rwer, r)
}

// SetActor sets the actor that sent the message
func (m *Message) SetActor(actor Actor) {
	m.msgActor = actor
}

// SetMime sets the MIME type of the message
func (m *Message) SetMime(mime string) {
	m.mimeType = mime
}

// Write implements the io.Writer interface
func (m *Message) Write(p []byte) (n int, err error) {
	return m.rwer.Write(p)
}

func (m *Message) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, m.rwer)
}

// URL returns the URL of the message message
func (b *Message) URL() *url.URL {
	return b.u
}

// IsDone returns true if the message is done
func (m *Message) IsDone() bool {
	return m.done
}

// Done marks the message as done
func (m *Message) Done() {
	m.done = true
}

// String returns the string representation of the message
func (m *Message) String() string {

	switch m.mimeType {
	case ioutils.MimeTextPlain,
		ioutils.MimeTextHTML,
		ioutils.MimeMarkDown,
		ioutils.MimeTextYAML,
		ioutils.MimeApplicationJSON,
		ioutils.MimeApplicationXML,
		ioutils.MimeTextXML,
		ioutils.MimeTextCSS,
		ioutils.MimeTextCSV:
		return m.rwer.(*bytes.Buffer).String()
	default:
		return fmt.Sprintf("{mimeType: %s, actor: %s}", m.mimeType, m.msgActor)

	}
}
