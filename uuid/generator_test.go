package uuid

import (
	"reflect"
	"testing"
)

// TestUUID_Bytes tests the Bytes method of the UUID struct.
// It verifies that the Bytes method returns the correct byte slice.
func TestUUID_Bytes(t *testing.T) {
	u := &UUID{bytes: []byte{1, 2, 3, 4}}
	want := []byte{1, 2, 3, 4}
	if got := u.Bytes(); !reflect.DeepEqual(got, want) {
		t.Errorf("Bytes() = %v, want %v", got, want)
	}
}

// TestUUID_String tests the String method of the UUID struct.
// It verifies that the String method returns the correct string representation of the UUID.
func TestUUID_String(t *testing.T) {
	u := &UUID{bytes: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}
	want := "01020304-0506-0708-090a-0b0c0d0e0f10"
	if got := u.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

// TestV1 tests the V1 function.
// It verifies that the V1 function generates a valid UUID.
func TestV1(t *testing.T) {
	u, err := V1()
	if err != nil {
		t.Errorf("V1() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V1() generated invalid UUID")
	}
}

// TestV2 tests the V2 function.
// It verifies that the V2 function generates a valid UUID.
func TestV2(t *testing.T) {
	u, err := V2()
	if err != nil {
		t.Errorf("V2() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V2() generated invalid UUID")
	}
}

// TestV3 tests the V3 function.
// It verifies that the V3 function generates a valid UUID.
func TestV3(t *testing.T) {
	u, err := V3("namespace", "name")

	if err != nil {
		t.Errorf("V3() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V3() generated invalid UUID")
	}
}

// TestV4 tests the V4 function.
// It verifies that the V4 function generates a valid UUID.
func TestV4(t *testing.T) {
	u, err := V4()
	if err != nil {
		t.Errorf("V4() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V4() generated invalid UUID")
	}
}

// TestParseUUID tests the ParseUUID function.
// It verifies that the function correctly parses a UUID string and returns the expected byte slice.
func TestParseUUID(t *testing.T) {
	s := "01020304-0506-0708-090a-0b0c0d0e0f10"
	want := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	u, err := ParseUUID(s)
	if err != nil {
		t.Errorf("ParseUUID() error = %v", err)
	}
	if !reflect.DeepEqual(u.Bytes(), want) {
		t.Errorf("ParseUUID() = %v, want %v", u.Bytes(), want)
	}
}
