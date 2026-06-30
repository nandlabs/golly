package rest

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Patch/Head/Options registrars ---

func TestServer_PatchHeadOptionsRegistrars(t *testing.T) {
	srv, err := DefaultServer()
	if err != nil {
		t.Fatalf("DefaultServer: %v", err)
	}
	called := map[string]bool{}
	mk := func(name string) HandlerFunc {
		return func(ctx ServerContext) {
			called[name] = true
			ctx.SetStatusCode(http.StatusOK)
		}
	}
	if _, err := srv.Patch("/r", mk("PATCH")); err != nil {
		t.Fatalf("Patch register: %v", err)
	}
	if _, err := srv.Head("/r", mk("HEAD")); err != nil {
		t.Fatalf("Head register: %v", err)
	}
	if _, err := srv.Options("/r", mk("OPTIONS")); err != nil {
		t.Fatalf("Options register: %v", err)
	}
	router := srv.Router()
	for _, m := range []string{http.MethodPatch, http.MethodHead, http.MethodOptions} {
		called = map[string]bool{}
		req := httptest.NewRequest(m, "/r", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", m, rec.Code)
		}
		if !called[m] {
			t.Errorf("%s: handler not invoked", m)
		}
	}
}

// --- TLS option validation ---

func TestValidate_TLS_RequiresCertOrConfig(t *testing.T) {
	opts := &SrvOptions{
		Id: "t", ListenHost: "localhost", ListenPort: 8080,
		EnableTLS: true, // no cert paths, no TLSConfig
	}
	if err := opts.Validate(); err == nil {
		t.Error("expected error when EnableTLS=true but no cert/key/TLSConfig")
	}
}

func TestValidate_TLS_AcceptsCertPaths(t *testing.T) {
	opts := &SrvOptions{
		Id: "t", ListenHost: "localhost", ListenPort: 8080,
		EnableTLS:      true,
		CertPath:       "/tmp/x.pem",
		PrivateKeyPath: "/tmp/x.key",
	}
	if err := opts.Validate(); err != nil {
		t.Errorf("expected nil for cert paths, got %v", err)
	}
}

func TestValidate_TLS_AcceptsTLSConfigWithCerts(t *testing.T) {
	opts := &SrvOptions{
		Id: "t", ListenHost: "localhost", ListenPort: 8080,
		EnableTLS: true,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{{}}, // synthetic; validation only checks non-empty
		},
	}
	if err := opts.Validate(); err != nil {
		t.Errorf("expected nil for TLSConfig with Certificates, got %v", err)
	}
}

func TestValidate_TLS_AcceptsTLSConfigWithGetCertificate(t *testing.T) {
	opts := &SrvOptions{
		Id: "t", ListenHost: "localhost", ListenPort: 8080,
		EnableTLS: true,
		TLSConfig: &tls.Config{
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return nil, nil },
		},
	}
	if err := opts.Validate(); err != nil {
		t.Errorf("expected nil for TLSConfig with GetCertificate, got %v", err)
	}
}

// --- StrictSlash forwarding ---

func TestStrictSlash_ForwardedToRouter(t *testing.T) {
	opts := DefaultSrvOptions()
	opts.StrictSlash = true
	srv, err := NewServer(opts)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	called := false
	if _, err := srv.Get("/users", func(ctx ServerContext) {
		called = true
		ctx.SetStatusCode(http.StatusOK)
	}); err != nil {
		t.Fatalf("Get register: %v", err)
	}
	// With StrictSlash on, "/users/" must NOT match "/users"
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	rec := httptest.NewRecorder()
	srv.Router().ServeHTTP(rec, req)
	if called && rec.Code == http.StatusOK {
		t.Error("StrictSlash=true but /users/ matched /users handler")
	}
}
