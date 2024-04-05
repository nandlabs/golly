package turbo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// BenchmarkFindRouteStatic: Static Path Test
func BenchmarkFindRouteStatic(b *testing.B) {
	var router = NewRouter()
	router.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]byte("hello from turbo"))
	})
	testUrl, _ := url.Parse("/api/v1/health")
	req := &http.Request{
		Method:           "",
		URL:              testUrl,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             nil,
		GetBody:          nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Host:             "",
		Form:             nil,
		PostForm:         nil,
		MultipartForm:    nil,
		Trailer:          nil,
		RemoteAddr:       "",
		RequestURI:       "",
		TLS:              nil,
		Cancel:           nil,
		Response:         nil,
	}
	for i := 0; i < b.N; i++ {
		router.findRoute(req)
	}
}

// BenchmarkFindRoutePathParam: Path Param Test
func BenchmarkFindRoutePathParam(b *testing.B) {
	var router = NewRouter()
	router.Get("/api/v1/health/:id", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]byte("hello from turbo"))
	})
	testUrl, _ := url.Parse("/api/v1/health/123")
	req := &http.Request{
		Method:           "",
		URL:              testUrl,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             nil,
		GetBody:          nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Host:             "",
		Form:             nil,
		PostForm:         nil,
		MultipartForm:    nil,
		Trailer:          nil,
		RemoteAddr:       "",
		RequestURI:       "",
		TLS:              nil,
		Cancel:           nil,
		Response:         nil,
	}
	for i := 0; i < b.N; i++ {
		router.findRoute(req)
	}
}

func BenchmarkRouter_ServeHTTPStatic(b *testing.B) {
	var router = NewRouter()
	router.Get("/api/fooTest", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Benchmarking!"))
	})
	w := httptest.NewRecorder()
	r, err := http.NewRequest(GET, "/api/fooTest", nil)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, r)
	}
}

func BenchmarkRouter_ServeHTTPParams(b *testing.B) {
	var router = NewRouter()
	router.Get("/api/fooTest/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Benchmarking!"))
	})
	w := httptest.NewRecorder()
	r, err := http.NewRequest(GET, "/api/fooTest/123", nil)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, r)
	}
}
