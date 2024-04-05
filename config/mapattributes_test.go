package config

import (
	"testing"
)

// TestMapAttributes_Set tests the Set method of the MapAttributes struct
func TestMapAttributes_Set(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "value")
	if m.attrs["key"] != "value" {
		t.Errorf("Set failed, expected value 'value', got %v", m.attrs["key"])
	}
}

// TestMapAttributes_Get tests the Get method of the MapAttributes struct
func TestMapAttributes_Get(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "value")
	v := m.Get("key")
	if v != "value" {
		t.Errorf("Get failed, expected value 'value', got %v", v)
	}
}

// TestMapAttributes_GetAsString tests the GetAsString method of the MapAttributes struct
func TestMapAttributes_GetAsString(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "value")
	v := m.GetAsString("key")
	if v != "value" {
		t.Errorf("GetAsString failed, expected value 'value', got %v", v)
	}
}

// TestMapAttributes_GetAsInt tests the GetAsInt method of the MapAttributes struct
func TestMapAttributes_GetAsInt(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", 42)
	v := m.GetAsInt("key")
	if v != 42 {
		t.Errorf("GetAsInt failed, expected value 42, got %v", v)
	}
}

// TestMapAttributes_GetAsFloat tests the GetAsFloat method of the MapAttributes struct
func TestMapAttributes_GetAsFloat(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", 42.42)

	v := m.GetAsFloat("key")
	if v != 42.42 {
		t.Errorf("GetAsFloat failed, expected value 42.42, got %v", v)
	}
}

// TestMapAttributes_GetAsBool tests the GetAsBool method of the MapAttributes struct
func TestMapAttributes_GetAsBool(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", true)

	v := m.GetAsBool("key")
	if v != true {
		t.Errorf("GetAsBool failed, expected value true, got %v", v)
	}
}

// TestMapAttributes_GetAsBytes tests the GetAsBytes method of the MapAttributes struct
func TestMapAttributes_GetAsBytes(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", []byte("value"))

	v := m.GetAsBytes("key")
	if string(v) != "value" {
		t.Errorf("GetAsBytes failed, expected value 'value', got %v", v)
	}
}

// TestMapAttributes_GetAsBytesNil tests the GetAsBytes method of the MapAttributes struct with a nil value
func TestMapAttributes_GetAsBytesNil(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", nil)

	v := m.GetAsBytes("key")
	if v != nil {
		t.Errorf("GetAsBytes failed, expected value nil, got %v", v)
	}
}

// TestMapAttributes_GetAsArray tests the GetAsArray method of the MapAttributes struct
func TestMapAttributes_GetAsArray(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", []any{"value"})

	v := m.GetAsArray("key")
	if v[0] != "value" {
		t.Errorf("GetAsArray failed, expected value 'value', got %v", v)
	}
}

// TestMapAttributes_GetAsArrayNil tests the GetAsArray method of the MapAttributes struct with a nil value
func TestMapAttributes_GetAsArrayNil(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", nil)
	v := m.GetAsArray("key")
	if v != nil {
		t.Errorf("GetAsArray failed, expected value nil, got %v", v)
	}
}

// TestMapAttributes_GetAsMap tests the GetAsMap method of the MapAttributes struct
func TestMapAttributes_GetAsMap(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", map[string]any{"key": "value"})
	v := m.GetAsMap("key")
	if v["key"] != "value" {
		t.Errorf("GetAsMap failed, expected value 'value', got %v", v)
	}
}

// TestMapAttributes_GetAsMapNil tests the GetAsMap method of the MapAttributes struct with a nil value
func TestMapAttributes_GetAsMapNil(t *testing.T) {

	m := NewMapAttributes()
	m.Set("key", nil)
	v := m.GetAsMap("key")
	if v != nil {
		t.Errorf("GetAsMap failed, expected value nil, got %v", v)
	}
}

// TestMapAttributes_AsMap tests the AsMap method of the MapAttributes struct
func TestMapAttributes_AsMap(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "value")
	v := m.AsMap()
	if v["key"] != "value" {
		t.Errorf("AsMap failed, expected value 'value', got %v", v)
	}
}

// TestMapAttributes_Remove tests the Remove method of the MapAttributes struct
func TestMapAttributes_Remove(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "value")
	m.Remove("key")
	if _, ok := m.attrs["key"]; ok {
		t.Errorf("Remove failed, expected key 'key' to be removed")
	}
}

// TestMapAttributes_Merge tests the Merge method of the MapAttributes struct
func TestMapAttributes_Merge(t *testing.T) {
	m1 := NewMapAttributes()
	m1.attrs["key1"] = "value1"
	m2 := NewMapAttributes()
	m2.attrs["key2"] = "value2"
	m1.Merge(m2)
	if m1.attrs["key1"] != "value1" {
		t.Errorf("Merge failed, expected value 'value1', got %v", m1.attrs["key1"])
	}
	if m1.attrs["key2"] != "value2" {
		t.Errorf("Merge failed, expected value 'value2', got %v", m1.attrs["key2"])
	}
}

// Add test cases for all code path in MapAttributes

// TestMapAttributes_GetAsIntNil tests the GetAsInt method of the MapAttributes struct with a nil value
func TestMapAttributes_GetAsIntNil(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", nil)
	v := m.GetAsInt("key")
	if v != 0 {
		t.Errorf("GetAsInt failed, expected value 0, got %v", v)
	}
}

// TestMapAttributes_GetAsFloatNil tests the GetAsFloat method of the MapAttributes struct with a nil value
func TestMapAttributes_GetAsFloatNil(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", nil)
	v := m.GetAsFloat("key")
	if v != 0 {
		t.Errorf("GetAsFloat failed, expected value 0, got %v", v)
	}
}

// TestMapAttributes_GetAsBoolNil tests the GetAsBool method of the MapAttributes struct with a nil value
func TestMapAttributes_GetAsBoolNil(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", nil)
	v := m.GetAsBool("key")
	if v != false {
		t.Errorf("GetAsBool failed, expected value false, got %v", v)
	}
}

// TestMapAttributes_GetAsIntFloat tests the GetAsInt method of the MapAttributes struct with a float value
func TestMapAttributes_GetAsIntFloat(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", 42.42)
	v := m.GetAsInt("key")
	if v != 42 {
		t.Errorf("GetAsInt failed, expected value 42, got %v", v)
	}
}

// TestMapAttributes_GetAsIntString tests the GetAsInt method of the MapAttributes struct with a string value
func TestMapAttributes_GetAsIntString(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "42")
	v := m.GetAsInt("key")
	if v != 42 {
		t.Errorf("GetAsInt failed, expected value 42, got %v", v)
	}
}

// TestMapAttributes_GetAsFloatString tests the GetAsFloat method of the MapAttributes struct with a string value
func TestMapAttributes_GetAsFloatString(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "42.42")
	v := m.GetAsFloat("key")
	if v != 0 {
		t.Errorf("GetAsFloat failed, expected value 0, got %v", v)
	}
}

// TestMapAttributes_GetAsBoolString tests the GetAsBool method of the MapAttributes struct with a string value
func TestMapAttributes_GetAsBoolString(t *testing.T) {
	m := NewMapAttributes()
	m.Set("key", "true")
	v := m.GetAsBool("key")
	if !v {
		t.Errorf("GetAsBool failed, expected value true, got %v", v)
	}
}
