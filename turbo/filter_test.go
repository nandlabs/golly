package turbo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func filterFunction(input string) FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(input))
			next.ServeHTTP(w, r)
		})
	}
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("testHandler"))
})

func TestFilter(t *testing.T) {
	var router = NewRouter()
	route := router.Get("/api/foo", testHandler)
	path := "/api/foo"

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Test1",
			input: "v1/",
		},
		{
			name:  "Test2",
			input: "v2/",
		},
		{
			name:  "Test3",
			input: "v3/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route.AddFilter(filterFunction(tt.input))
		})
	}
	w := httptest.NewRecorder()
	r, err := http.NewRequest(GET, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(route.filters) != len(tests) {
		t.Error("All Test Filters not added")
	}
	router.ServeHTTP(w, r)
	if w.Body.String() != "v1/v2/v3/testHandler" {
		t.Error("Filter Chain not working")
	}
}

type BasicAuthFilter struct{}

func (ba *BasicAuthFilter) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("token"); token != "" {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Not Authorised"))
		}
	})
}

func CreateBasicAuthAuthenticator() *BasicAuthFilter {
	return &BasicAuthFilter{}
}

func TestAuthenticatorFilter(t *testing.T) {
	var router = NewRouter()
	route := router.Get("/api/foo", testHandler)
	path := "/api/foo"

	var authenticator = CreateBasicAuthAuthenticator()

	route.AddAuthenticator(authenticator)

	var w *httptest.ResponseRecorder
	var r *http.Request
	var err error

	w = httptest.NewRecorder()
	r, err = http.NewRequest(GET, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("token", "value")

	if route.authFilter == nil {
		t.Error("Authenticator Filters not added")
	}
	router.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Error("Auth Filter not working")
	}

	w = httptest.NewRecorder()
	r, err = http.NewRequest(GET, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if route.authFilter == nil {
		t.Error("Authenticator Filters not added")
	}
	router.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusForbidden {
		t.Error("Auth Filter not working")
	}
}
