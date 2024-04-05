package assert

import (
	"errors"
	"testing"
)

func TestEqual(t *testing.T) {
	expected := 10
	actual := 10
	Equal(t, expected, actual)
}

func TestNotEqual(t *testing.T) {
	expected := 10
	actual := 20
	NotEqual(t, expected, actual)
}

func TestTrue(t *testing.T) {
	condition := true
	True(t, condition)
}

func TestFalse(t *testing.T) {
	condition := false
	False(t, condition)
}

func TestNil(t *testing.T) {
	var value interface{}
	Nil(t, value)
}

func TestNotNil(t *testing.T) {
	value := "Hello"
	NotNil(t, value)
}

func TestError(t *testing.T) {
	err := errors.New("Something went wrong")
	Error(t, err)
}

func TestNoError(t *testing.T) {
	err := error(nil)
	NoError(t, err)
}

func TestMapContains(t *testing.T) {
	m := map[string]interface{}{
		"key": "value",
	}
	key := "key"
	value := "value"
	MapContains(t, m, key, value)
}

func TestMapMissing(t *testing.T) {
	m := map[string]interface{}{
		"key": "value",
	}
	key := "key"
	value := "value1"
	MapMissing(t, m, key, value)
}

func TestHasKey(t *testing.T) {
	m := map[string]interface{}{
		"key": "value",
	}
	key := "key"
	HasKey(t, m, key)
}

func TestHasValue(t *testing.T) {
	m := map[string]interface{}{
		"key": "value",
	}
	value := "value"
	HasValue(t, m, value)
}

func TestListHas(t *testing.T) {
	value := 10
	list := []interface{}{1, 2, 3, 10, 5}
	ListHas(t, value, list...)
}

func TestListMissing(t *testing.T) {
	value := 10
	list := []interface{}{1, 2, 3, 4, 5}
	ListMissing(t, value, list...)
}
