package ioutils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	SHA256 = "SHA256"
)

// ChkSumCalc interface is used to calculate the checksum of a text or file
type ChkSumCalc interface {
	// Calculate calculates the checksum of the message
	Calculate(content string) (string, error)
	// Verify verifies the checksum of the message
	Verify(content, sum string) (bool, error)
	// CalculateFile calculates the checksum of a file
	CalculateFile(file string) (string, error)
	// VerifyFile verifies the checksum of a file
	VerifyFile(file, sum string) (bool, error)
	// CalculateFor calculates the checksum of the reader
	CalculateFor(reader io.Reader) (string, error)
	// VerifyFor verifies the checksum of the reader
	VerifyFor(reader io.Reader, sum string) (bool, error)
	// Type returns the type of the checksum
	Type() string
}

// Sha256Checksum is a checksum that uses the SHA256 algorithm
type Sha256Checksum struct {
}

// Calculate calculates the checksum of the message
func (s *Sha256Checksum) Calculate(content string) (chksum string, err error) {
	// Calculate the sha256 checksum
	hash := sha256.New()
	_, err = io.Copy(hash, strings.NewReader(content))
	if err == nil {
		chksum = fmt.Sprintf("%x", hash.Sum(nil))
	}
	return
}

// Verify verifies the checksum of the message
func (s *Sha256Checksum) Verify(content, sum string) (b bool, err error) {
	var calcSum string
	// Calculate the checksum of the content
	calcSum, err = s.Calculate(content)
	// Verify the checksum
	b = err == nil && sum == calcSum
	return
}

// CalculateFile calculates the checksum of a file
func (s *Sha256Checksum) CalculateFile(file string) (chksum string, err error) {
	// Calculate the checksum of the file
	hash := sha256.New()
	var f *os.File
	f, err = os.Open(file)
	if err != nil {
		return
	}
	defer CloserFunc(f)
	_, err = io.Copy(hash, f)
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// VerifyFile verifies the checksum of a file
func (s *Sha256Checksum) VerifyFile(file, sum string) (b bool, err error) {
	var calcSum string
	// Calculate the checksum of the file
	calcSum, err = s.CalculateFile(file)
	// Verify the checksum
	b = err == nil && sum == calcSum
	return
}

// CalculateFor calculates the checksum of the reader
func (s *Sha256Checksum) CalculateFor(reader io.Reader) (chksum string, err error) {
	// Calculate the checksum of the reader
	hash := sha256.New()
	_, err = io.Copy(hash, reader)
	if err == nil {
		chksum = fmt.Sprintf("%x", hash.Sum(nil))
	}
	return
}

// VerifyFor verifies the checksum of the reader
func (s *Sha256Checksum) VerifyFor(reader io.Reader, sum string) (b bool, err error) {
	var calcSum string
	// Calculate the checksum of the reader
	calcSum, err = s.CalculateFor(reader)
	// Verify the checksum
	b = err == nil && sum == calcSum
	return
}

// Type returns the type of the checksum
func (s *Sha256Checksum) Type() string {
	return SHA256
}

// NewChkSumCalc creates a new checksum
func NewChkSumCalc(t string) ChkSumCalc {
	switch t {
	case SHA256:
		return &Sha256Checksum{}
	default:
		return nil
	}
}
