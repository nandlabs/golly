package uuid

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"time"
)

// UUID represents a universally unique identifier.
type UUID struct {
	bytes []byte
}

// Bytes returns the bytes of the UUID.
func (u *UUID) Bytes() []byte {
	return u.bytes
}

// String returns the string representation of the UUID.
func (u *UUID) String() string {
	// The string representation of a UUID is a sequence of 32 hexadecimal digits,
	// displayed in five groups separated by hyphens.
	// For example, 123e4567-e89b-12d3-a456-426655440000.
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.bytes[0:4], u.bytes[4:6], u.bytes[6:8], u.bytes[8:10], u.bytes[10:])
}

// V1 generates a version 1 UUID.
func V1() (u *UUID, err error) {

	var uuid = make([]byte, 16)

	// Set the version (4 most significant bits of the time_hi_and_version field) to 1.
	uuid[6] = (uuid[6] & 0x0f) | 0x10

	// Set the variant (2 most significant bits of the clock_seq_hi_and_reserved field) to 1.
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	// Generate 6 bytes of random data.
	_, err = rand.Read(uuid[10:])
	if err != nil {
		return
	} else {
		u = &UUID{bytes: uuid}
	}

	// Set the time_low, time_mid, and time_hi fields based on the current time.
	now := time.Now().UTC()
	timestamp := now.UnixNano() / 100
	uuid[0] = byte(timestamp >> 24)
	uuid[1] = byte(timestamp >> 16)
	uuid[2] = byte(timestamp >> 8)
	uuid[3] = byte(timestamp)

	uuid[4] = byte(timestamp >> 40)
	uuid[5] = byte(timestamp >> 32)

	uuid[7] = byte(timestamp >> 56)
	uuid[9] = byte(timestamp >> 48)

	return
}

// V2 generates a version 2 UUID.
func V2() (u *UUID, err error) {
	// Get the MAC address of the machine
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	var mac net.HardwareAddr
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) != 0 {
			mac = iface.HardwareAddr
			break
		}
	}

	if mac == nil {
		err = fmt.Errorf("failed to get MAC address")
		return
	}

	// Get the process ID
	pid := os.Getpid()

	// Get the current timestamp
	now := time.Now().Unix()

	// Generate a hash using MD5
	hash := md5.New()
	_, _ = fmt.Fprintf(hash, "%s%d%d", mac.String(), pid, now)
	hashBytes := hash.Sum(nil)

	// Set the version and variant bits
	hashBytes[6] = (hashBytes[6] & 0x0f) | 0x20 // Set version to 2
	hashBytes[8] = (hashBytes[8] & 0x3f) | 0x80 // Set variant to RFC 4122
	u = &UUID{bytes: hashBytes}
	return
}

// V3 generates a version 3 UUID.
func V3(namespace string, name string) (u *UUID, err error) {
	// Generate a hash using MD5
	hash := md5.New()
	hash.Write([]byte(namespace + name))
	hashBytes := hash.Sum(nil)

	// Set the version and variant bits
	hashBytes[6] = (hashBytes[6] & 0x0f) | 0x30 // Set version to 3
	hashBytes[8] = (hashBytes[8] & 0x3f) | 0x80 // Set variant to RFC 4122
	u = &UUID{bytes: hashBytes}
	return
}

// V4 generates a version 4 UUID.
func V4() (u *UUID, err error) {
	uuid := make([]byte, 16)
	_, err = rand.Read(uuid)
	if err != nil {
		return
	} else {
		u = &UUID{bytes: uuid}
	}
	return
}

// ParseUUID parses a UUID string.
func ParseUUID(s string) (u *UUID, err error) {
	// Remove hyphens from the UUID string
	s = s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:]

	// Parse the UUID string
	uuid, err := hex.DecodeString(s)
	if err != nil {
		return
	} else {
		u = &UUID{bytes: uuid}
	}
	return
}
