package messaging

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestLocalMessage_SetBodyBytes(t *testing.T) {
	message := NewLocalMessage()
	input := []byte("this is a test string")
	_, err := message.SetBodyBytes(input)
	if err != nil {
		t.Errorf("Error SetBodyBytes: %v", err)
	}
	res := message.ReadBytes()
	if string(res) != string(input) {
		t.Errorf("Error ReadAsStr, want= %v, got= %v", input, res)
	}
}

func TestLocalMessage_SetBodyStr(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	input := "this is a test string"
	_, err := message.SetBodyStr(input)
	if err != nil {
		t.Errorf("Error SetBodyBytes: %v", err)
	}
	res := message.ReadAsStr()
	if res != input {
		t.Errorf("Error ReadAsStr: %v", err)
	}
}

func TestLocalMessage_SetFrom(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	input := strings.NewReader("some io.Reader stream to be read")
	_, err := message.SetFrom(input)
	if err != nil {
		t.Errorf("Error SetFrom: %v", err)
	}
	res := message.ReadBody()
	b1 := make([]byte, 2)
	r1, _ := res.Read(b1)
	if r1 != len(b1) {
		t.Error("Error SetFrom")
	}
}

func TestLocalMessage_WriteJSON(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	input := interface{}(`{"name":"Test","body":"Hello","time":123124124}`)
	err := message.WriteJSON(input)
	if err != nil {
		t.Errorf("Error WriteJSON: %v", err)
	}
	var got interface{}
	err = message.ReadJSON(&got)
	if !reflect.DeepEqual(input, got) {
		t.Errorf("Error ReadJSON: %v", err)
	}
}

//func TestLocalMessage_WriteXML(t *testing.T) {
//	message := &BaseMessage{
//		headers:     make(map[string]interface{}),
//		headerTypes: make(map[string]reflect.Kind),
//		body:        &bytes.Buffer{},
//	}
//	input := interface{}(`<XMLMessage><name>Test</name><body>Hello</body><time>123124124</time></XMLMessage>`)
//	err := message.WriteXML(input)
//	if err != nil {
//		t.Errorf("Error WriteXML: %v", err)
//	}
//	var got interface{}
//	err = message.ReadXML(&got)
//	fmt.Println(got)
//	if !reflect.DeepEqual(input, got) {
//		t.Errorf("Error ReadXML: %v", err)
//	}
//}

func TestLocalMessage_Header(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetHeader("key", []byte("test"))
	got, exists := message.GetHeader("key")
	if exists {
		if string(got) != "test" {
			t.Errorf("Error Header Setters/Getters, got : %v, want : test", got)
		}
	}
}

func TestLocalMessage_BoolHeader(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetBoolHeader("isPresent", true)
	got, exists := message.GetBoolHeader("isPresent")
	if exists {
		if got != true {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v", got)
		}
	}
}

func TestLocalMessage_StrHeader(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetStrHeader("header-obj", "header-value")
	got, exists := message.GetStrHeader("header-obj")
	if exists {
		if got != "header-value" {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : header-value", got)
		}
	}
}

func TestLocalMessage_IntHeader(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetIntHeader("header-obj", 10)
	got, exists := message.GetIntHeader("header-obj")
	if exists {
		if got != 10 {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : 10", got)
		}
	}
}

func TestLocalMessage_Int8Header(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetInt8Header("header-obj", 120)
	got, exists := message.GetInt8Header("header-obj")
	if exists {
		if got != 120 {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : 120", got)
		}
	}
}

func TestLocalMessage_Int16Header(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetInt16Header("header-obj", 32767)
	got, exists := message.GetInt16Header("header-obj")
	if exists {
		if got != 32767 {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : 120", got)
		}
	}
}

func TestLocalMessage_Int32Header(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetInt32Header("header-obj", 2147483647)
	got, exists := message.GetInt32Header("header-obj")
	if exists {
		if got != 2147483647 {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : 2147483647", got)
		}
	}
}

func TestLocalMessage_Int64Header(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetInt64Header("header-obj", 9223372036854775807)
	got, exists := message.GetInt64Header("header-obj")
	if exists {
		if got != 9223372036854775807 {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : 9223372036854775807", got)
		}
	}
}

func TestLocalMessage_Float32Header(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetFloatHeader("header-obj", 200.0)
	got, exists := message.GetFloatHeader("header-obj")
	if exists {
		if got != 200.0 {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : 200.0", got)
		}
	}
}

func TestLocalMessage_Float64Header(t *testing.T) {
	message := &BaseMessage{
		headers:     make(map[string]interface{}),
		headerTypes: make(map[string]reflect.Kind),
		body:        &bytes.Buffer{},
	}
	message.SetFloat64Header("header-obj", -1.7e+308)
	got, exists := message.GetFloat64Header("header-obj")
	if exists {
		if got != -1.7e+308 {
			t.Errorf("Error BoolHeader Setters/Getters, got : %v, want : -1.7E+308", got)
		}
	}
}
