package rest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type person struct {
	XMLName xml.Name `xml:"person" json:"-"`
	Name    string   `json:"name" xml:"name" yaml:"name"`
	Age     int      `json:"age"  xml:"age"  yaml:"age"`
}

func newCtx(t *testing.T, req *http.Request) (*ServerContext, *httptest.ResponseRecorder) {
	t.Helper()
	rec := httptest.NewRecorder()
	return &ServerContext{request: req, response: rec}, rec
}

// --- Bind (Content-Type drives decode) ---

func TestBind_JSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"alice","age":30}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(ContentTypeHeader, "application/json")
	ctx, _ := newCtx(t, req)
	var p person
	if err := ctx.Bind(&p); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if p.Name != "alice" || p.Age != 30 {
		t.Errorf("got %+v, want {alice 30}", p)
	}
}

func TestBind_XML(t *testing.T) {
	body := bytes.NewBufferString(`<person><name>bob</name><age>40</age></person>`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(ContentTypeHeader, "application/xml")
	ctx, _ := newCtx(t, req)
	var p person
	if err := ctx.Bind(&p); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if p.Name != "bob" || p.Age != 40 {
		t.Errorf("got %+v, want {bob 40}", p)
	}
}

func TestBind_UnknownContentType_Errors(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set(ContentTypeHeader, "application/octet-stream")
	ctx, _ := newCtx(t, req)
	var p person
	if err := ctx.Bind(&p); err == nil {
		t.Error("expected error for unknown content type, got nil")
	}
}

// --- Respond (Accept drives encode) ---

func TestRespond_JSONByDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil) // no Accept
	ctx, rec := newCtx(t, req)
	if err := ctx.Respond(http.StatusOK, person{Name: "alice", Age: 30}); err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if got := rec.Header().Get(ContentTypeHeader); got != "application/json" {
		t.Errorf("content-type = %q, want application/json", got)
	}
	var p person
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("body is not JSON: %v\n%s", err, rec.Body.String())
	}
	if p.Name != "alice" {
		t.Errorf("decoded = %+v", p)
	}
}

func TestRespond_ExplicitJSONAccept(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AcceptHeader, "application/json")
	ctx, rec := newCtx(t, req)
	if err := ctx.Respond(http.StatusCreated, person{Name: "alice", Age: 30}); err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if got := rec.Header().Get(ContentTypeHeader); got != "application/json" {
		t.Errorf("content-type = %q, want application/json", got)
	}
}

func TestRespond_XMLAccept(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AcceptHeader, "application/xml")
	ctx, rec := newCtx(t, req)
	if err := ctx.Respond(http.StatusOK, person{Name: "bob", Age: 40}); err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if got := rec.Header().Get(ContentTypeHeader); got != "application/xml" {
		t.Errorf("content-type = %q, want application/xml", got)
	}
	if !strings.Contains(rec.Body.String(), "<name>bob</name>") {
		t.Errorf("body not XML-encoded: %s", rec.Body.String())
	}
}

func TestRespond_TextXMLAccept(t *testing.T) {
	// text/xml should still negotiate to the XML codec.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AcceptHeader, "text/xml")
	ctx, rec := newCtx(t, req)
	_ = ctx.Respond(http.StatusOK, person{Name: "bob", Age: 40})
	if got := rec.Header().Get(ContentTypeHeader); got != "application/xml" {
		t.Errorf("content-type = %q, want application/xml", got)
	}
}

func TestRespond_YAMLAccept(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AcceptHeader, "text/yaml")
	ctx, rec := newCtx(t, req)
	if err := ctx.Respond(http.StatusOK, person{Name: "carol", Age: 50}); err != nil {
		t.Fatalf("Respond: %v", err)
	}
	if got := rec.Header().Get(ContentTypeHeader); got != "text/yaml" {
		t.Errorf("content-type = %q, want text/yaml", got)
	}
	if !strings.Contains(rec.Body.String(), "carol") {
		t.Errorf("body missing payload: %s", rec.Body.String())
	}
}

func TestRespond_StarStarAccept(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AcceptHeader, "*/*")
	ctx, rec := newCtx(t, req)
	_ = ctx.Respond(http.StatusOK, person{Name: "dave"})
	if got := rec.Header().Get(ContentTypeHeader); got != "application/json" {
		t.Errorf("*/* should fall back to JSON; got %q", got)
	}
}

func TestRespond_UnknownAccept_Yields406(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AcceptHeader, "image/png")
	ctx, rec := newCtx(t, req)
	if err := ctx.Respond(http.StatusOK, "ignored"); err != nil {
		t.Fatalf("Respond should not error on 406; got %v", err)
	}
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("status = %d, want 406", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("406 should have empty body; got %q", rec.Body.String())
	}
}
