package rest

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oss.nandlabs.io/golly/clients"
)

func TestClient_Execute(t *testing.T) {
	tests := []struct {
		name           string
		clientOptions  *ClientOpts
		requestURL     string
		requestMethod  string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Successful GET request",
			clientOptions: &ClientOpts{
				ClientOptions:  clients.NewOptionsBuilder().Build(),
				requestTimeout: 5 * time.Second,
			},
			requestURL:     "/test",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Invalid URL",
			clientOptions: &ClientOpts{
				ClientOptions:  clients.NewOptionsBuilder().Build(),
				requestTimeout: 5 * time.Second,
			},
			requestURL:    "://invalid-url",
			requestMethod: http.MethodGet,
			expectError:   true,
		},
		{
			name: "Error on status",
			clientOptions: &ClientOpts{
				ClientOptions:  clients.NewOptionsBuilder().Build(),
				requestTimeout: 5 * time.Second,
				errorOnMap: map[int]int{
					http.StatusInternalServerError: http.StatusInternalServerError,
				},
			},
			requestURL:     "/error",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/error" {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			// Update request URL to use the test server URL
			if tt.requestURL != "://invalid-url" {
				tt.requestURL = ts.URL + tt.requestURL
			}

			client := NewClientWithOptions(tt.clientOptions)
			req, err := client.NewRequest(tt.requestURL, tt.requestMethod)
			if err != nil {
				if !tt.expectError {
					t.Errorf("unexpected error creating request: %v", err)
				}
				return
			}

			res, err := client.Execute(req)
			if tt.expectError && res.StatusCode() != 500 {
				t.Errorf("unexpected error executing request: %v", err)
				return
			}

			if res != nil && res.raw.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.raw.StatusCode)
			}
		})
	}
}
func TestClientOptsBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder *ClientOptsBuilder
		verify  func(opts *ClientOpts) bool
	}{
		{
			name:    "Default options",
			builder: RestCliOptBuilder(),
			verify: func(opts *ClientOpts) bool {
				return opts.ClientOptions != nil &&
					opts.tlsConfig != nil &&
					opts.maxIdlePerHost == 0 &&
					opts.useCustomTLSConfig == false
			},
		},
		{
			name:    "Set MaxIdlePerHost",
			builder: RestCliOptBuilder().MaxIdlePerHost(10),
			verify: func(opts *ClientOpts) bool {
				return opts.maxIdlePerHost == 10
			},
		},
		{
			name:    "Set ProxyAuth",
			builder: RestCliOptBuilder().ProxyAuth("user", "pass", ""),
			verify: func(opts *ClientOpts) bool {
				expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
				return opts.proxyBasicAuth == expected
			},
		},
		{
			name: "Set BaseUrl",
			builder: func() *ClientOptsBuilder {
				builder := RestCliOptBuilder()
				err := builder.BaseUrl("http://example.com")
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return builder
			}(),
			verify: func(opts *ClientOpts) bool {
				return opts.baseUrl.String() == "http://example.com/"
			},
		},
		{
			name:    "Set SSLVerify",
			builder: RestCliOptBuilder().SSLVerify(true),
			verify: func(opts *ClientOpts) bool {
				return opts.tlsConfig.InsecureSkipVerify == true
			},
		},
		{
			name:    "Set RequestTimeoutMs",
			builder: RestCliOptBuilder().RequestTimeoutMs(5000),
			verify: func(opts *ClientOpts) bool {
				return opts.requestTimeout == 5*time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.builder.Build()
			if !tt.verify(opts) {
				t.Errorf("verification failed for test: %s", tt.name)
			}
		})
	}
}
func TestClient_isError(t *testing.T) {
	tests := []struct {
		name       string
		clientOpts *ClientOpts
		err        error
		httpRes    *http.Response
		wantErr    bool
	}{
		{
			name: "No error and no errorOnMap",
			clientOpts: &ClientOpts{
				ClientOptions: clients.NewOptionsBuilder().Build(),
			},
			err:     nil,
			httpRes: &http.Response{StatusCode: http.StatusOK},
			wantErr: false,
		},
		{
			name: "Error present",
			clientOpts: &ClientOpts{
				ClientOptions: clients.NewOptionsBuilder().Build(),
			},
			err:     fmt.Errorf("some error"),
			httpRes: &http.Response{StatusCode: http.StatusOK},
			wantErr: true,
		},
		{
			name: "Error on status code",
			clientOpts: &ClientOpts{
				ClientOptions: clients.NewOptionsBuilder().Build(),
				errorOnMap: map[int]int{
					http.StatusInternalServerError: http.StatusInternalServerError,
				},
			},
			err:     nil,
			httpRes: &http.Response{StatusCode: http.StatusInternalServerError},
			wantErr: true,
		},
		{
			name: "No error on status code",
			clientOpts: &ClientOpts{
				ClientOptions: clients.NewOptionsBuilder().Build(),
				errorOnMap: map[int]int{
					http.StatusInternalServerError: http.StatusInternalServerError,
				},
			},
			err:     nil,
			httpRes: &http.Response{StatusCode: http.StatusOK},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				options: tt.clientOpts,
			}
			if gotErr := client.isError(tt.err, tt.httpRes); gotErr != tt.wantErr {
				t.Errorf("Client.isError() = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}
func TestClient_ExecuteWithCircuitBreaker(t *testing.T) {
	coptsBuilder1 := RestCliOptBuilder().ErrOnStatus(http.StatusInternalServerError)
	coptsBuilder1.CircuitBreaker(3, 3, 3, 5000)

	tests := []struct {
		name           string
		clientOptions  *ClientOpts
		requestURL     string
		requestMethod  string
		expectedStatus int
		expectError    bool
	}{
		{
			name:          "Circuit breaker open",
			clientOptions: coptsBuilder1.Build(),
			requestURL:    "/test",
			requestMethod: http.MethodGet,
			expectError:   true,
		},
		{
			name:           "Circuit breaker closed",
			clientOptions:  coptsBuilder1.Build(),
			requestURL:     "/test",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectError {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer ts.Close()

			tt.requestURL = ts.URL + tt.requestURL
			client := NewClientWithOptions(tt.clientOptions)
			req, err := client.NewRequest(tt.requestURL, tt.requestMethod)
			if err != nil {
				if !tt.expectError {
					t.Errorf("unexpected error creating request: %v", err)
				}
				return
			}
			var res *Response

			for i := 0; i < 5; i++ {
				res, err = client.Execute(req)
			}
			if tt.expectError {
				if !client.isError(err, res.raw) {
					t.Errorf("expected error, got none")
				}
				return
			}

			if res.raw.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.raw.StatusCode)
			}
		})
	}
}

func TestClient_ExecuteWithRetries(t *testing.T) {
	coptsBuilder2 := RestCliOptBuilder().ErrOnStatus(http.StatusInternalServerError)
	coptsBuilder2.RetryPolicy(3, 1000, true, 5000)
	tests := []struct {
		name           string
		clientOptions  *ClientOpts
		requestURL     string
		requestMethod  string
		expectedStatus int
		expectError    bool
	}{
		{
			name:          "Retry on failure",
			clientOptions: coptsBuilder2.Build(),
			requestURL:    "/error",
			requestMethod: http.MethodGet,
			expectError:   true,
		},
		{
			name:           "No retry on success",
			clientOptions:  coptsBuilder2.Build(),
			requestURL:     "/test",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/error" {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			tt.requestURL = ts.URL + tt.requestURL

			client := NewClientWithOptions(tt.clientOptions)
			req, err := client.NewRequest(tt.requestURL, tt.requestMethod)
			if err != nil {
				if !tt.expectError {
					t.Errorf("unexpected error creating request: %v", err)
				}
				return
			}

			var res *Response

			res, err = client.Execute(req)

			if tt.expectError {
				if !client.isError(err, res.raw) {
					t.Errorf("expected error, got none")
				}
				return
			}

			if res.raw.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.raw.StatusCode)
			}
		})
	}
}

func TestClient_ExecuteWithBasicAuth(t *testing.T) {
	coptsBuilder1 := RestCliOptBuilder().ErrOnStatus(http.StatusUnauthorized)
	coptsBuilder1.BasicAuth("user", "pass")

	coptsBuilder2 := RestCliOptBuilder().ErrOnStatus(http.StatusUnauthorized)
	coptsBuilder2.BasicAuth("user1", "pass1")
	tests := []struct {
		name           string
		clientOptions  *ClientOpts
		requestURL     string
		requestMethod  string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Successful Basic Auth",
			clientOptions:  coptsBuilder1.Build(),
			requestURL:     "/test",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:          "Invalid Basic Auth",
			clientOptions: coptsBuilder2.Build(),
			requestURL:    "/test",
			requestMethod: http.MethodGet,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				if !ok || user != "user" || pass != "pass" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			tt.requestURL = ts.URL + tt.requestURL

			client := NewClientWithOptions(tt.clientOptions)
			req, err := client.NewRequest(tt.requestURL, tt.requestMethod)
			if err != nil {
				if !tt.expectError {
					t.Errorf("unexpected error creating request: %v", err)
				}
				return
			}

			res, err := client.Execute(req)
			if tt.expectError {
				if !client.isError(err, res.raw) {
					t.Errorf("expected error, got none")
				}
				return
			}

			if res.raw.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.raw.StatusCode)
			}
		})
	}
}

func TestClient_ExecuteWithBearerAuth(t *testing.T) {
	coptsBuilder1 := RestCliOptBuilder().ErrOnStatus(http.StatusInternalServerError)
	coptsBuilder1.TokenBearerAuth("valid-token")
	coptsBuilder2 := RestCliOptBuilder().ErrOnStatus(http.StatusInternalServerError)
	coptsBuilder2.TokenBearerAuth("invalid-token")
	tests := []struct {
		name           string
		clientOptions  *ClientOpts
		requestURL     string
		requestMethod  string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Successful Bearer Auth",
			clientOptions:  coptsBuilder1.Build(),
			requestURL:     "/test",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:          "Invalid Bearer Auth",
			clientOptions: coptsBuilder2.Build(),
			requestURL:    "/test",
			requestMethod: http.MethodGet,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token := r.Header.Get("Authorization")
				if token != "Bearer valid-token" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			tt.requestURL = ts.URL + tt.requestURL

			client := NewClientWithOptions(tt.clientOptions)
			req, err := client.NewRequest(tt.requestURL, tt.requestMethod)
			if err != nil {
				if !tt.expectError {
					t.Errorf("unexpected error creating request: %v", err)
				}
				return
			}

			res, err := client.Execute(req)
			if tt.expectError {
				if !client.isError(err, res.raw) {
					t.Errorf("expected error, got none")
				}
				return
			}
		})
	}
}
