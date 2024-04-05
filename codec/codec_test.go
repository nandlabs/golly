package codec

import (
	"bytes"
	"strings"
	"testing"
)

type Message struct {
	Name string `json:"name"`
	Body string `json:"body"`
	Time int64  `json:"time"`
}

type XMLMessage struct {
	Name string `xml:"name"`
	Body string `xml:"body"`
	Time int64  `xml:"time"`
}

type Message2 struct {
	Name string `json:"name" constraints:"min-length=5"`
	Body string `json:"body" constraints:"max-length=50"`
	Time int64  `json:"time" constraints:"min=10"`
}

func TestNewJson(t *testing.T) {
	m := Message2{"TestUser", "Hello", 123124124}
	c, _ := Get("application/json", nil)
	buf := new(bytes.Buffer)
	if err := c.Write(m, buf); err != nil {
		t.Errorf("error in write: %d", err)
	}

	const want = "{\"name\":\"TestUser\",\"body\":\"Hello\",\"time\":123124124}\n"
	if got := buf; got.String() != want {
		t.Errorf("got %q, want %q", got.String(), want)
	}
}

func TestNewDefaultJson(t *testing.T) {
	m := Message2{"TestUser", "Hello", 123124124}
	c, _ := GetDefault("application/json")
	buf := new(bytes.Buffer)
	if err := c.Write(m, buf); err != nil {
		t.Errorf("error in write: %d", err)
	}

	const want = "{\"name\":\"TestUser\",\"body\":\"Hello\",\"time\":123124124}\n"
	if got := buf; got.String() != want {
		t.Errorf("got %q, want %q", got.String(), want)
	}
}

func TestNewJsonCodec2(t *testing.T) {
	var m Message
	c, _ := Get("application/json", nil)
	const input = `{"name":"Test","body":"Hello","time":123124124}`
	b := strings.NewReader(input)
	if err := c.Read(b, &m); err != nil {
		t.Errorf("error in read: %d", err)
	}
	want := Message{
		Name: "Test",
		Body: "Hello",
		Time: 123124124,
	}
	if m != want {
		t.Errorf("got %q, want %q", m, want)
	}
}

func TestNewXmlCodec(t *testing.T) {
	m := XMLMessage{"Test", "Hello", 123124124}
	c, _ := Get("text/xml", nil)
	buf := new(bytes.Buffer)
	if err := c.Write(m, buf); err != nil {
		t.Errorf("error in write: %d", err)
	}
	const want = `<XMLMessage><name>Test</name><body>Hello</body><time>123124124</time></XMLMessage>`
	if got := buf; got.String() != want {
		t.Errorf("got %q, want %q", got.String(), want)
	}
}

func TestNewXmlCodec2(t *testing.T) {
	var m XMLMessage
	c, _ := Get("text/xml", nil)
	const input = `<XMLMessage><name>Test</name><body>Hello</body><time>123124124</time></XMLMessage>`
	b := strings.NewReader(input)
	if err := c.Read(b, &m); err != nil {
		t.Errorf("error in read: %d", err)
	}
	want := XMLMessage{
		Name: "Test",
		Body: "Hello",
		Time: 123124124,
	}
	if m != want {
		t.Errorf("got %q, want %q", m, want)
	}
}

func TestNewInvalid(t *testing.T) {
	_, err := Get("text/plain", nil)

	if err == nil {
		t.Error("got nil wanted err")

	}
}
