package ws

// Opcode represents a WebSocket frame opcode as defined in RFC 6455 Section 5.2.
type Opcode byte

const (
	// OpContinuation denotes a continuation frame.
	OpContinuation Opcode = 0x0
	// OpText denotes a text data frame.
	OpText Opcode = 0x1
	// OpBinary denotes a binary data frame.
	OpBinary Opcode = 0x2
	// OpClose denotes a connection close control frame.
	OpClose Opcode = 0x8
	// OpPing denotes a ping control frame.
	OpPing Opcode = 0x9
	// OpPong denotes a pong control frame.
	OpPong Opcode = 0xA
)

// IsControl returns true if the opcode is a control frame (close, ping, pong).
func (o Opcode) IsControl() bool {
	return o >= OpClose
}

// CloseCode represents a WebSocket close status code as defined in RFC 6455 Section 7.4.
type CloseCode uint16

const (
	CloseNormal           CloseCode = 1000
	CloseGoingAway        CloseCode = 1001
	CloseProtocolError    CloseCode = 1002
	CloseUnsupported      CloseCode = 1003
	CloseNoStatus         CloseCode = 1005
	CloseAbnormal         CloseCode = 1006
	CloseInvalidPayload   CloseCode = 1007
	ClosePolicyViolation  CloseCode = 1008
	CloseMessageTooBig    CloseCode = 1009
	CloseMissingExtension CloseCode = 1010
	CloseInternalError    CloseCode = 1011
	CloseTLSHandshake     CloseCode = 1015
)

// Message represents a WebSocket message consisting of one or more frames.
type Message struct {
	// Type is the opcode of the message (OpText or OpBinary).
	Type Opcode
	// Data is the message payload.
	Data []byte
}

// NewTextMessage creates a new text message with the given payload.
func NewTextMessage(data []byte) Message {
	return Message{Type: OpText, Data: data}
}

// NewBinaryMessage creates a new binary message with the given payload.
func NewBinaryMessage(data []byte) Message {
	return Message{Type: OpBinary, Data: data}
}

// frame represents a single WebSocket frame.
type frame struct {
	fin     bool
	rsv1    bool
	rsv2    bool
	rsv3    bool
	opcode  Opcode
	masked  bool
	maskKey [4]byte
	payload []byte
}
