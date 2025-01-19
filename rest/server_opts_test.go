package rest

import (
	"testing"
)

// TestOptions_Validate tests the Validate function
func TestOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options SrvOptions
		wantErr bool
	}{
		{"Valid options", SrvOptions{Id: "123", ListenHost: "localhost", ListenPort: 8080}, false},
		{"Invalid ID", SrvOptions{ListenHost: "localhost", ListenPort: 8080}, true},
		{"Invalid ListenHost", SrvOptions{Id: "123", ListenPort: 8080}, true},
		{"Invalid ListenPort", SrvOptions{Id: "123", ListenHost: "localhost"}, true},
		{"Invalid TLS options", SrvOptions{Id: "123", ListenHost: "localhost", ListenPort: 8080, EnableTLS: true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.options.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOptions_Getters tests the getter functions
func TestOptions_Getters(t *testing.T) {
	opts := SrvOptions{
		ListenHost:     "localhost",
		ListenPort:     8080,
		EnableTLS:      true,
		PrivateKeyPath: "/path/to/private.key",
		CertPath:       "/path/to/cert.crt",
	}
	if got := opts.GetListenHost(); got != opts.ListenHost {
		t.Errorf("GetListenHost() = %v, want %v", got, opts.ListenHost)
	}
	if got := opts.GetListenPort(); got != opts.ListenPort {
		t.Errorf("GetListenPort() = %v, want %v", got, opts.ListenPort)
	}
	if got := opts.GetEnableTLS(); got != opts.EnableTLS {
		t.Errorf("GetEnableTLS() = %v, want %v", got, opts.EnableTLS)
	}
	if got := opts.GetPrivateKeyPath(); got != opts.PrivateKeyPath {
		t.Errorf("GetPrivateKeyPath() = %v, want %v", got, opts.PrivateKeyPath)
	}
	if got := opts.GetCertPath(); got != opts.CertPath {
		t.Errorf("GetCertPath() = %v, want %v", got, opts.CertPath)
	}
}

// TestOptions_Setters tests the setter functions
func TestOptions_Setters(t *testing.T) {
	opts := &SrvOptions{}
	opts = opts.SetListenHost("localhost")
	if opts.ListenHost != "localhost" {
		t.Errorf("SetListenHost() = %v, want %v", opts.ListenHost, "localhost")
	}
	opts = opts.SetListenPort(8080)
	if opts.ListenPort != 8080 {
		t.Errorf("SetListenPort() = %v, want %v", opts.ListenPort, 8080)
	}
	opts = opts.SetEnableTLS(true)
	if !opts.EnableTLS {
		t.Errorf("SetEnableTLS() = %v, want %v", opts.EnableTLS, true)
	}
	opts = opts.SetPrivateKeyPath("/path/to/private.key")
	if opts.PrivateKeyPath != "/path/to/private.key" {
		t.Errorf("SetPrivateKeyPath() = %v, want %v", opts.PrivateKeyPath, "/path/to/private.key")
	}
	opts = opts.SetCertPath("/path/to/cert.crt")
	if opts.CertPath != "/path/to/cert.crt" {
		t.Errorf("SetCertPath() = %v, want %v", opts.CertPath, "/path/to/cert.crt")
	}
}

// TestNewOptions tests the NewOptions function
func TestNewOptions(t *testing.T) {
	opts := EmptySrvOptions()
	if opts.ListenHost != "" || opts.ListenPort != 0 {
		t.Errorf("NewOptions() = %v, want default values", opts)
	}
}

// TestDefaultOptions tests the DefaultOptions function
func TestDefaultOptions(t *testing.T) {
	opts := DefaultSrvOptions()
	if opts.ListenHost != "localhost" || opts.ListenPort != 8080 || opts.ReadTimeout != 20000 || opts.WriteTimeout != 20000 {
		t.Errorf("DefaultOptions() = %v, want default values", opts)
	}
}
