package rest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestContext_GetParam tests the GetParam function
func TestContext_GetParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	ctx := &ServerContext{request: req}

	value, err := ctx.GetParam("param", QueryParam)
	if err != nil {
		t.Errorf("GetParam() error = %v", err)
	}
	if value != "value" {
		t.Errorf("GetParam() = %v, want %v", value, "value")
	}
}

// TestContext_GetBody tests the GetBody function
func TestContext_GetBody(t *testing.T) {
	body := "test body"
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	ctx := &ServerContext{request: req}

	r, err := ctx.GetBody()
	if err != nil {
		t.Errorf("GetBody() error = %v", err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	if buf.String() != body {
		t.Errorf("GetBody() = %v, want %v", buf.String(), body)
	}
}

// TestContext_GetHeader tests the GetHeader function
func TestContext_GetHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Test-Header", "test-value")
	ctx := &ServerContext{request: req}

	value := ctx.GetHeader("X-Test-Header")
	if value != "test-value" {
		t.Errorf("GetHeader() = %v, want %v", value, "test-value")
	}
}

// TestContext_InHeaders tests the InHeaders function
func TestContext_InHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Test-Header", "test-value")
	ctx := &ServerContext{request: req}

	headers := ctx.InHeaders()
	if headers.Get("X-Test-Header") != "test-value" {
		t.Errorf("InHeaders() = %v, want %v", headers.Get("X-Test-Header"), "test-value")
	}
}

// TestContext_GetMethod tests the GetMethod function
func TestContext_GetMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	ctx := &ServerContext{request: req}

	method := ctx.GetMethod()
	if method != http.MethodPost {
		t.Errorf("GetMethod() = %v, want %v", method, http.MethodPost)
	}
}

// TestContext_GetURL tests the GetURL function
func TestContext_GetURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := &ServerContext{request: req}

	url := ctx.GetURL()
	if url != "/test" {
		t.Errorf("GetURL() = %v, want %v", url, "/test")
	}
}

// TestContext_GetRequest tests the GetRequest function
func TestContext_GetRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := &ServerContext{request: req}

	request := ctx.GetRequest()
	if request != req {
		t.Errorf("GetRequest() = %v, want %v", request, req)
	}
}

// TestContext_Read tests the Read function
func TestContext_Read(t *testing.T) {
	body := `{"key":"value"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set(ContentTypeHeader, "application/json")
	ctx := &ServerContext{request: req}

	var obj map[string]string
	err := ctx.Read(&obj)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	if obj["key"] != "value" {
		t.Errorf("Read() = %v, want %v", obj["key"], "value")
	}
}

// TestContext_Write tests the Write function
func TestContext_Write(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	data := map[string]string{"key": "value"}
	err := ctx.Write(data, "application/json")
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if rec.Header().Get(ContentTypeHeader) != "application/json" {
		t.Errorf("Write() Content-Type = %v, want %v", rec.Header().Get(ContentTypeHeader), "application/json")
	}
}

// TestContext_WriteData tests the WriteData function
func TestContext_WriteData(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	data := []byte("test data")
	n, err := ctx.WriteData(data)
	if err != nil {
		t.Errorf("WriteData() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("WriteData() = %v, want %v", n, len(data))
	}
	if rec.Body.String() != string(data) {
		t.Errorf("WriteData() = %v, want %v", rec.Body.String(), string(data))
	}
}

// TestContext_WriteString tests the WriteString function
func TestContext_WriteString(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	data := "test string"
	ctx.WriteString(data)
	if rec.Body.String() != data {
		t.Errorf("WriteString() = %v, want %v", rec.Body.String(), data)
	}
}

// TestContext_SetHeader tests the SetHeader function
func TestContext_SetHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	ctx.SetHeader("X-Test-Header", "test-value")
	if rec.Header().Get("X-Test-Header") != "test-value" {
		t.Errorf("SetHeader() = %v, want %v", rec.Header().Get("X-Test-Header"), "test-value")
	}
}

// TestContext_SetContentType tests the SetContentType function
func TestContext_SetContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	ctx.SetContentType("application/json")
	if rec.Header().Get(ContentTypeHeader) != "application/json" {
		t.Errorf("SetContentType() = %v, want %v", rec.Header().Get(ContentTypeHeader), "application/json")
	}
}

// TestContext_SetStatusCode tests the SetStatusCode function
func TestContext_SetStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	ctx.SetStatusCode(http.StatusCreated)
	if rec.Code != http.StatusCreated {
		t.Errorf("SetStatusCode() = %v, want %v", rec.Code, http.StatusCreated)
	}
}

// TestContext_SetCookie tests the SetCookie function
func TestContext_SetCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	cookie := &http.Cookie{Name: "test-cookie", Value: "test-value"}
	ctx.SetCookie(cookie)
	if rec.Header().Get("Set-Cookie") != "test-cookie=test-value" {
		t.Errorf("SetCookie() = %v, want %v", rec.Header().Get("Set-Cookie"), "test-cookie=test-value")
	}
}

// TestContext_WriteFrom tests the WriteFrom function
func TestContext_WriteFrom(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	data := "test data"
	ctx.WriteFrom(strings.NewReader(data))
	if rec.Body.String() != data {
		t.Errorf("WriteFrom() = %v, want %v", rec.Body.String(), data)
	}
}

// TestContext_HttpResWriter tests the HttpResWriter function
func TestContext_HttpResWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := &ServerContext{response: rec}

	writer := ctx.HttpResWriter()
	if writer != rec {
		t.Errorf("HttpResWriter() = %v, want %v", writer, rec)
	}
}
