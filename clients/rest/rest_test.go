package rest

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"

	"oss.nandlabs.io/golly/codec"
)

var client = NewClient()

func TestNewClient(t *testing.T) {
	if reflect.TypeOf(client) != reflect.TypeOf(NewClient()) {
		t.Errorf("NewClient() = %v, want %v", client, NewClient())
	}
}

func TestClientOptions(t *testing.T) {
	gotReq := client.ReqTimeout(10)
	if reflect.TypeOf(client) != reflect.TypeOf(gotReq) {
		t.Errorf("NewClient() = %v, want %v", gotReq, client)
	}

	gotCodec := client.AddCodecOption(codec.PrettyPrint, true)
	if reflect.TypeOf(client) != reflect.TypeOf(gotCodec) {
		t.Errorf("NewClient() = %v, want %v", gotCodec, client)
	}

	gotIdle := client.IdleTimeout(2)
	if reflect.TypeOf(client) != reflect.TypeOf(gotIdle) {
		t.Errorf("NewClient() = %v, want %v", gotIdle, client)
	}

	gotHttpEmpty := client.ErrorOnHttpStatus()
	if reflect.TypeOf(client) != reflect.TypeOf(gotHttpEmpty) {
		t.Errorf("NewClient() = %v, want %v", gotHttpEmpty, client)
	}

	gotHttp := client.ErrorOnHttpStatus(200, 300, 404)
	if reflect.TypeOf(client) != reflect.TypeOf(gotHttp) {
		t.Errorf("NewClient() = %v, want %v", gotHttp, client)
	}

	gotMaxIdle := client.MaxIdle(3)
	if reflect.TypeOf(client) != reflect.TypeOf(gotMaxIdle) {
		t.Errorf("NewClient() = %v, want %v", gotMaxIdle, client)
	}

	gotMaxIdlePerHost := client.MaxIdlePerHost(4)
	if reflect.TypeOf(client) != reflect.TypeOf(gotMaxIdlePerHost) {
		t.Errorf("NewClient() = %v, want %v", gotMaxIdlePerHost, client)
	}

	gotEndProxy := client.UseEnvProxy("test.com", "test", "test")
	if gotEndProxy != nil {
		t.Errorf("NewClient() = %v, want %v", gotEndProxy, client)
	}

	gotRetry := client.Retry(3, 5)
	if reflect.TypeOf(client) != reflect.TypeOf(gotRetry) {
		t.Errorf("NewClient() = %v, want %v", gotRetry, client)
	}

	gotCircuitBreaker := client.UseCircuitBreaker(1, 2, 1, 3)
	if reflect.TypeOf(client) != reflect.TypeOf(gotCircuitBreaker) {
		t.Errorf("NewClient() = %v, want %v", gotCircuitBreaker, client)
	}

	gotTlsCerts, err := client.SetTLSCerts(tls.Certificate{})
	if err != nil {
		t.Errorf("unable to add tls certs")
	}
	if reflect.TypeOf(client) != reflect.TypeOf(gotTlsCerts) {
		t.Errorf("NewClient() = %v, want %v", gotTlsCerts, client)
	}
}

func TestClient_NewRequest(t *testing.T) {
	req := client.NewRequest("http://localhost:8080", http.MethodGet)
	want := &Request{
		url:    "http://localhost:8080",
		method: http.MethodGet,
	}
	if reflect.TypeOf(req) != reflect.TypeOf(want) {
		t.Errorf("NewRequest() = %v, want %v", req, want)
	}
}

func TestClient_SetCACerts(t *testing.T) {
	tests := []struct {
		name     string
		certPath string
		want     string
	}{
		{
			name:     "TestClient_SetCACerts_1",
			certPath: "./testdata/test-key.pem",
			want:     "",
		},
		{
			name:     "TestClient_SetCACerts_2",
			certPath: "./testdata/test-key-temp.pem",
			want:     "open ./testdata/test-key-temp.pem: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.SetCACerts(tt.certPath)
			if err != nil {
				if tt.want != err.Error() {
					t.Errorf("Got: %s, want: %s", err.Error(), tt.want)
				}
			}
		})
	}
}

func TestClient_SetCACerts2(t *testing.T) {
	want := ""
	_, err := client.SetCACerts("./testdata/test-key.pem", "./testdata/test-key-2.pem")
	if err != nil {
		t.Errorf("Got: %s, want: %s", err.Error(), want)
	}
}

func TestClient_SSlVerify(t *testing.T) {
	clientSSLVerify, err := client.SSlVerify(true)
	if reflect.TypeOf(clientSSLVerify) != reflect.TypeOf(client) {
		t.Errorf("Got: %s, want: %s", reflect.TypeOf(clientSSLVerify), reflect.TypeOf(client))
	}

	want := ""
	_, err = client.SetCACerts("./testdata/test-key.pem", "./testdata/test-key-2.pem")
	if err != nil {
		t.Errorf("Got: %s, want: %s", err.Error(), want)
	}
	if client.tlsConfig.InsecureSkipVerify != true {
		t.Error("client SSL setup incorrect")
	}
}

func TestClient_Execute(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
		input  interface{}
		want   interface{}
	}{
		{
			name:   "TestClient_1",
			url:    "localhost",
			method: "",
			input:  "",
			want:   "Get \"localhost\": unsupported protocol scheme \"\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := client.NewRequest(tt.url, tt.method)
			_, err := client.Execute(req)
			if tt.want != err.Error() {
				t.Errorf("Got: %s, want: %s", err, tt.want)
			}
		})
	}
}
