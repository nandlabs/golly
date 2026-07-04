package messaging

import (
	"io"
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

// TestKeyed_Roundtrip verifies that a LocalMessage's routing key is
// preserved across SetRoutingKey / RoutingKey calls and reachable via
// the package-level RoutingKeyOf helper.
func TestKeyed_Roundtrip(t *testing.T) {
	msg, err := NewLocalMessage()
	assert.NoError(t, err)

	if RoutingKeyOf(msg) != "" {
		t.Fatalf("expected empty routing key before set, got %q", RoutingKeyOf(msg))
	}

	keyed, ok := msg.(Keyed)
	if !ok {
		t.Fatal("LocalMessage does not implement Keyed")
	}
	keyed.SetRoutingKey("tenant-42")

	if got := RoutingKeyOf(msg); got != "tenant-42" {
		t.Errorf("expected routing key %q, got %q", "tenant-42", got)
	}
	if got := keyed.RoutingKey(); got != "tenant-42" {
		t.Errorf("expected direct routing key %q, got %q", "tenant-42", got)
	}
}

// TestRoutingKeyOf_UnkeyedMessage checks the helper returns "" for
// Message implementations that do NOT implement Keyed.
func TestRoutingKeyOf_UnkeyedMessage(t *testing.T) {
	var msg Message = &plainMessage{}
	// Sanity check: the shim really does NOT implement Keyed.
	if _, ok := msg.(Keyed); ok {
		t.Fatal("plainMessage unexpectedly implements Keyed — test is invalid")
	}
	if got := RoutingKeyOf(msg); got != "" {
		t.Errorf("expected empty routing key for unkeyed message, got %q", got)
	}
}

// plainMessage is a minimal Message that does NOT implement Keyed. It
// exercises the RoutingKeyOf fallback path.
type plainMessage struct{}

func (*plainMessage) Id() string                              { return "plain" }
func (*plainMessage) SetHeader(string, []byte)                {}
func (*plainMessage) SetStrHeader(string, string)             {}
func (*plainMessage) SetBoolHeader(string, bool)              {}
func (*plainMessage) SetIntHeader(string, int)                {}
func (*plainMessage) SetInt8Header(string, int8)              {}
func (*plainMessage) SetInt16Header(string, int16)            {}
func (*plainMessage) SetInt32Header(string, int32)            {}
func (*plainMessage) SetInt64Header(string, int64)            {}
func (*plainMessage) SetFloatHeader(string, float32)          {}
func (*plainMessage) SetFloat64Header(string, float64)        {}
func (*plainMessage) GetHeader(string) ([]byte, bool)         { return nil, false }
func (*plainMessage) GetStrHeader(string) (string, bool)      { return "", false }
func (*plainMessage) GetBoolHeader(string) (bool, bool)       { return false, false }
func (*plainMessage) GetIntHeader(string) (int, bool)         { return 0, false }
func (*plainMessage) GetInt8Header(string) (int8, bool)       { return 0, false }
func (*plainMessage) GetInt16Header(string) (int16, bool)     { return 0, false }
func (*plainMessage) GetInt32Header(string) (int32, bool)     { return 0, false }
func (*plainMessage) GetInt64Header(string) (int64, bool)     { return 0, false }
func (*plainMessage) GetFloatHeader(string) (float32, bool)   { return 0, false }
func (*plainMessage) GetFloat64Header(string) (float64, bool) { return 0, false }
func (*plainMessage) SetBodyStr(string) (int, error)          { return 0, nil }
func (*plainMessage) SetBodyBytes([]byte) (int, error)        { return 0, nil }
func (*plainMessage) SetFrom(io.Reader) (int64, error)        { return 0, nil }
func (*plainMessage) WriteJSON(interface{}) error             { return nil }
func (*plainMessage) WriteXML(interface{}) error              { return nil }
func (*plainMessage) WriteContent(interface{}, string) error  { return nil }
func (*plainMessage) ReadBody() io.Reader                     { return nil }
func (*plainMessage) ReadBytes() []byte                       { return nil }
func (*plainMessage) ReadAsStr() string                       { return "" }
func (*plainMessage) ReadJSON(interface{}) error              { return nil }
func (*plainMessage) ReadXML(interface{}) error               { return nil }
func (*plainMessage) ReadContent(interface{}, string) error   { return nil }
func (*plainMessage) Rsvp(bool, ...Option) error              { return nil }
