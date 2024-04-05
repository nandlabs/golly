package ioutils

import (
	"strings"
	"testing"
)

func TestSha256Checksum_Calculate(t *testing.T) {
	checksum := NewChkSumCalc(SHA256)
	content := "Hello, World!"
	expectedChecksum := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	calculatedChecksum, err := checksum.Calculate(content)
	if err != nil {
		t.Errorf("Error calculating checksum: %v", err)
	}

	if calculatedChecksum != expectedChecksum {
		t.Errorf("Expected checksum %s, but got %s", expectedChecksum, calculatedChecksum)
	}
}

func TestSha256Checksum_Verify(t *testing.T) {
	checksum := NewChkSumCalc(SHA256)
	content := "Hello, World!"
	expectedChecksum := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	valid, err := checksum.Verify(content, expectedChecksum)
	if err != nil {
		t.Errorf("Error verifying checksum: %v", err)
	}

	if !valid {
		t.Errorf("Expected checksum %s to be valid, but it was not", expectedChecksum)
	}
}

func TestSha256Checksum_CalculateFile(t *testing.T) {
	checksum := NewChkSumCalc(SHA256)
	file := "./testdata/hello-world.txt"
	expectedChecksum := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	calculatedChecksum, err := checksum.CalculateFile(file)
	if err != nil {
		t.Errorf("Error calculating file checksum: %v", err)
	}

	if calculatedChecksum != expectedChecksum {
		t.Errorf("Expected file checksum %s, but got %s", expectedChecksum, calculatedChecksum)
	}
}

func TestSha256Checksum_VerifyFile(t *testing.T) {
	checksum := NewChkSumCalc(SHA256)
	file := "./testdata/hello-world.txt"
	expectedChecksum := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	valid, err := checksum.VerifyFile(file, expectedChecksum)
	if err != nil {
		t.Errorf("Error verifying file checksum: %v", err)
	}

	if !valid {
		t.Errorf("Expected file checksum %s to be valid, but it was not", expectedChecksum)
	}
}

func TestSha256Checksum_CalculateFor(t *testing.T) {
	checksum := NewChkSumCalc(SHA256)
	content := "Hello, World!"
	expectedChecksum := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	calculatedChecksum, err := checksum.CalculateFor(strings.NewReader(content))
	if err != nil {
		t.Errorf("Error calculating reader checksum: %v", err)
	}

	if calculatedChecksum != expectedChecksum {
		t.Errorf("Expected reader checksum %s, but got %s", expectedChecksum, calculatedChecksum)
	}
}

func TestSha256Checksum_VerifyFor(t *testing.T) {
	checksum := NewChkSumCalc(SHA256)
	content := "Hello, World!"
	expectedChecksum := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	valid, err := checksum.VerifyFor(strings.NewReader(content), expectedChecksum)
	if err != nil {
		t.Errorf("Error verifying reader checksum: %v", err)
	}

	if !valid {
		t.Errorf("Expected reader checksum %s to be valid, but it was not", expectedChecksum)
	}
}
