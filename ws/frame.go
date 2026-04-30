package ws

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// readFrame reads a single WebSocket frame from the reader.
func readFrame(r io.Reader, maxSize int64) (*frame, error) {
	// Read first 2 bytes: FIN, RSV1-3, opcode, MASK, payload length
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	f := &frame{
		fin:    header[0]&0x80 != 0,
		rsv1:   header[0]&0x40 != 0,
		rsv2:   header[0]&0x20 != 0,
		rsv3:   header[0]&0x10 != 0,
		opcode: Opcode(header[0] & 0x0F),
		masked: header[1]&0x80 != 0,
	}

	// Check reserved bits
	if f.rsv1 || f.rsv2 || f.rsv3 {
		return nil, ErrReservedBits
	}

	// Control frame validation
	if f.opcode.IsControl() {
		if !f.fin {
			return nil, ErrFragmentedControl
		}
	}

	// Read payload length
	payloadLen := uint64(header[1] & 0x7F)
	switch payloadLen {
	case 126:
		extended := make([]byte, 2)
		if _, err := io.ReadFull(r, extended); err != nil {
			return nil, err
		}
		payloadLen = uint64(binary.BigEndian.Uint16(extended))
	case 127:
		extended := make([]byte, 8)
		if _, err := io.ReadFull(r, extended); err != nil {
			return nil, err
		}
		payloadLen = binary.BigEndian.Uint64(extended)
	}

	// Validate control frame size
	if f.opcode.IsControl() && payloadLen > 125 {
		return nil, ErrControlTooLarge
	}

	// Check max message size
	if maxSize > 0 && int64(payloadLen) > maxSize {
		return nil, ErrMessageTooLarge
	}

	// Read mask key if present
	if f.masked {
		if _, err := io.ReadFull(r, f.maskKey[:]); err != nil {
			return nil, err
		}
	}

	// Read payload
	if payloadLen > 0 {
		f.payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(r, f.payload); err != nil {
			return nil, err
		}
		// Unmask payload
		if f.masked {
			maskBytes(f.maskKey, f.payload)
		}
	}

	return f, nil
}

// writeFrame writes a single WebSocket frame to the writer.
func writeFrame(w io.Writer, f *frame) error {
	// Calculate header size
	headerSize := 2
	payloadLen := len(f.payload)
	if payloadLen >= 126 && payloadLen <= 65535 {
		headerSize += 2
	} else if payloadLen > 65535 {
		headerSize += 8
	}
	if f.masked {
		headerSize += 4
	}

	header := make([]byte, headerSize)
	pos := 0

	// First byte: FIN + opcode
	header[pos] = byte(f.opcode)
	if f.fin {
		header[pos] |= 0x80
	}
	pos++

	// Second byte: MASK + payload length
	var lenByte byte
	if f.masked {
		lenByte = 0x80
	}
	switch {
	case payloadLen <= 125:
		header[pos] = lenByte | byte(payloadLen)
		pos++
	case payloadLen <= 65535:
		header[pos] = lenByte | 126
		pos++
		binary.BigEndian.PutUint16(header[pos:], uint16(payloadLen))
		pos += 2
	default:
		header[pos] = lenByte | 127
		pos++
		binary.BigEndian.PutUint64(header[pos:], uint64(payloadLen))
		pos += 8
	}

	// Mask key
	if f.masked {
		if _, err := rand.Read(f.maskKey[:]); err != nil {
			return fmt.Errorf("ws: failed to generate mask key: %w", err)
		}
		copy(header[pos:], f.maskKey[:])
	}

	// Write header
	if _, err := w.Write(header); err != nil {
		return err
	}

	// Write payload (masked if needed)
	if payloadLen > 0 {
		if f.masked {
			masked := make([]byte, payloadLen)
			copy(masked, f.payload)
			maskBytes(f.maskKey, masked)
			_, err := w.Write(masked)
			return err
		}
		_, err := w.Write(f.payload)
		return err
	}

	return nil
}

// maskBytes applies the XOR mask to data in place.
func maskBytes(key [4]byte, data []byte) {
	for i := range data {
		data[i] ^= key[i%4]
	}
}

// makeClosePayload creates a close frame payload with code and optional reason.
func makeClosePayload(code CloseCode, reason string) []byte {
	payload := make([]byte, 2+len(reason))
	binary.BigEndian.PutUint16(payload, uint16(code))
	copy(payload[2:], reason)
	return payload
}
