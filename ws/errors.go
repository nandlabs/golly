package ws

import "errors"

var (
	// ErrConnClosed is returned when an operation is attempted on a closed connection.
	ErrConnClosed = errors.New("ws: connection closed")
	// ErrMessageTooLarge is returned when a message exceeds the maximum allowed size.
	ErrMessageTooLarge = errors.New("ws: message too large")
	// ErrInvalidOpcode is returned when an unknown frame opcode is received.
	ErrInvalidOpcode = errors.New("ws: invalid opcode")
	// ErrInvalidCloseCode is returned when a close frame has an invalid status code.
	ErrInvalidCloseCode = errors.New("ws: invalid close code")
	// ErrHandshakeFailed is returned when the WebSocket handshake fails.
	ErrHandshakeFailed = errors.New("ws: handshake failed")
	// ErrInvalidURL is returned when the provided WebSocket URL is invalid.
	ErrInvalidURL = errors.New("ws: invalid url")
	// ErrReservedBits is returned when reserved bits are set in a frame header.
	ErrReservedBits = errors.New("ws: reserved bits set")
	// ErrFragmentedControl is returned when a control frame is fragmented.
	ErrFragmentedControl = errors.New("ws: fragmented control frame")
	// ErrControlTooLarge is returned when a control frame payload exceeds 125 bytes.
	ErrControlTooLarge = errors.New("ws: control frame too large")
	// ErrUnauthorized is returned when an UpgradeAuthFunc rejects the handshake.
	ErrUnauthorized = errors.New("ws: unauthorized")
)
