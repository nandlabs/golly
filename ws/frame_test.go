package ws

import (
	"bytes"
	"testing"
)

func TestMaskBytes(t *testing.T) {
	key := [4]byte{0xAA, 0xBB, 0xCC, 0xDD}
	data := []byte("hello")
	original := make([]byte, len(data))
	copy(original, data)

	maskBytes(key, data)
	// Should be different after masking
	if bytes.Equal(data, original) {
		t.Fatal("data should be different after masking")
	}

	// Masking again with same key should restore original
	maskBytes(key, data)
	if !bytes.Equal(data, original) {
		t.Fatalf("expected %q, got %q", original, data)
	}
}

func TestFrameRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		opcode  Opcode
		payload string
		masked  bool
		fin     bool
	}{
		{"text unmasked", OpText, "hello", false, true},
		{"binary unmasked", OpBinary, "binary data", false, true},
		{"text masked", OpText, "masked hello", true, true},
		{"empty payload", OpText, "", false, true},
		{"ping", OpPing, "", false, true},
		{"pong", OpPong, "pong data", false, true},
		{"close", OpClose, "\x03\xe8", false, true}, // code 1000
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Write frame
			f := &frame{
				fin:     tt.fin,
				opcode:  tt.opcode,
				masked:  tt.masked,
				payload: []byte(tt.payload),
			}
			if err := writeFrame(&buf, f); err != nil {
				t.Fatalf("writeFrame error: %v", err)
			}

			// Read frame
			got, err := readFrame(&buf, 0)
			if err != nil {
				t.Fatalf("readFrame error: %v", err)
			}

			if got.opcode != tt.opcode {
				t.Errorf("opcode: got %v, want %v", got.opcode, tt.opcode)
			}
			if got.fin != tt.fin {
				t.Errorf("fin: got %v, want %v", got.fin, tt.fin)
			}
			if !bytes.Equal(got.payload, []byte(tt.payload)) {
				t.Errorf("payload: got %q, want %q", got.payload, tt.payload)
			}
		})
	}
}

func TestFrameLargePayload(t *testing.T) {
	// Test 16-bit extended length (126-65535)
	payload := make([]byte, 300)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	var buf bytes.Buffer
	f := &frame{fin: true, opcode: OpBinary, payload: payload}
	if err := writeFrame(&buf, f); err != nil {
		t.Fatalf("writeFrame error: %v", err)
	}

	got, err := readFrame(&buf, 0)
	if err != nil {
		t.Fatalf("readFrame error: %v", err)
	}
	if !bytes.Equal(got.payload, payload) {
		t.Error("large payload mismatch")
	}
}

func TestFrameMaxSizeEnforced(t *testing.T) {
	var buf bytes.Buffer
	payload := make([]byte, 200)
	f := &frame{fin: true, opcode: OpText, payload: payload}
	if err := writeFrame(&buf, f); err != nil {
		t.Fatalf("writeFrame error: %v", err)
	}

	_, err := readFrame(&buf, 100)
	if err != ErrMessageTooLarge {
		t.Fatalf("expected ErrMessageTooLarge, got %v", err)
	}
}

func TestFrameReservedBits(t *testing.T) {
	// Manually construct a frame with RSV1 set
	data := []byte{0xC1, 0x00} // FIN=1, RSV1=1, opcode=1, no mask, len=0
	buf := bytes.NewBuffer(data)
	_, err := readFrame(buf, 0)
	if err != ErrReservedBits {
		t.Fatalf("expected ErrReservedBits, got %v", err)
	}
}

func TestFrameFragmentedControl(t *testing.T) {
	// Control frame with FIN=0
	data := []byte{0x09, 0x00} // FIN=0, opcode=9 (ping), no mask, len=0
	buf := bytes.NewBuffer(data)
	_, err := readFrame(buf, 0)
	if err != ErrFragmentedControl {
		t.Fatalf("expected ErrFragmentedControl, got %v", err)
	}
}

func TestFrameControlTooLarge(t *testing.T) {
	var buf bytes.Buffer
	payload := make([]byte, 126) // exceeds 125 limit for control frames
	f := &frame{fin: true, opcode: OpPing, payload: payload}
	if err := writeFrame(&buf, f); err != nil {
		t.Fatalf("writeFrame error: %v", err)
	}

	_, err := readFrame(&buf, 0)
	if err != ErrControlTooLarge {
		t.Fatalf("expected ErrControlTooLarge, got %v", err)
	}
}

func TestMakeClosePayload(t *testing.T) {
	payload := makeClosePayload(CloseNormal, "goodbye")
	if len(payload) != 2+len("goodbye") {
		t.Fatalf("expected length %d, got %d", 2+len("goodbye"), len(payload))
	}
	// First two bytes should be 1000 in big-endian
	code := uint16(payload[0])<<8 | uint16(payload[1])
	if code != 1000 {
		t.Fatalf("expected code 1000, got %d", code)
	}
	if string(payload[2:]) != "goodbye" {
		t.Fatalf("expected reason 'goodbye', got %q", string(payload[2:]))
	}
}

func TestOpcodeIsControl(t *testing.T) {
	controls := []Opcode{OpClose, OpPing, OpPong}
	for _, op := range controls {
		if !op.IsControl() {
			t.Errorf("expected %v to be control", op)
		}
	}
	nonControls := []Opcode{OpContinuation, OpText, OpBinary}
	for _, op := range nonControls {
		if op.IsControl() {
			t.Errorf("expected %v to not be control", op)
		}
	}
}

func TestNewTextMessage(t *testing.T) {
	msg := NewTextMessage([]byte("hello"))
	if msg.Type != OpText {
		t.Errorf("expected OpText, got %v", msg.Type)
	}
	if string(msg.Data) != "hello" {
		t.Errorf("expected 'hello', got %q", msg.Data)
	}
}

func TestNewBinaryMessage(t *testing.T) {
	msg := NewBinaryMessage([]byte{0x01, 0x02})
	if msg.Type != OpBinary {
		t.Errorf("expected OpBinary, got %v", msg.Type)
	}
	if len(msg.Data) != 2 {
		t.Errorf("expected 2 bytes, got %d", len(msg.Data))
	}
}

func TestComputeAcceptKey(t *testing.T) {
	// RFC 6455 Section 4.2.2 example
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	expected := "mFpC0hCe2EwfrqT6m61cZJ/7ZqI="
	got := computeAcceptKey(key)
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
