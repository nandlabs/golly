package uuid

import (
	"reflect"
	"testing"
)

func TestUUID_Bytes(t *testing.T) {
	u := &UUID{bytes: []byte{1, 2, 3, 4}}
	want := []byte{1, 2, 3, 4}
	if got := u.Bytes(); !reflect.DeepEqual(got, want) {
		t.Errorf("Bytes() = %v, want %v", got, want)
	}
}

func TestUUID_String(t *testing.T) {
	u := &UUID{bytes: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}}
	want := "01020304-0506-0708-090a-0b0c0d0e0f10"
	if got := u.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestV1(t *testing.T) {
	u, err := V1()
	if err != nil {
		t.Errorf("V1() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V1() generated invalid UUID")
	}
}

func TestV2(t *testing.T) {
	u, err := V2()
	if err != nil {
		t.Errorf("V2() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V2() generated invalid UUID")
	}
}

func TestV3(t *testing.T) {
	u, err := V3("namespace", "name")

	if err != nil {
		t.Errorf("V3() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V3() generated invalid UUID")
	}
}

func TestV4(t *testing.T) {
	u, err := V4()
	if err != nil {
		t.Errorf("V4() error = %v", err)
	}
	if len(u.Bytes()) != 16 {
		t.Errorf("V4() generated invalid UUID")
	}
}

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
