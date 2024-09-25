package server

import (
	"net/http"
	"testing"
	"time"

	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/testing/assert"
	"oss.nandlabs.io/golly/uuid"
)

// TestNewServerFrom tests the NewServerFrom function
func TestNewServerFrom(t *testing.T) {
	// This test requires a valid config file path
	// configPath := "path/to/config/file"
	// _, err := NewServerFrom(configPath)
	// if err != nil {
	// 	t.Errorf("NewServerFrom() error = %v", err)
	// }
}

// TestDefaultServer tests the DefaultServer function
func TestDefaultServer(t *testing.T) {
	server, err := Default()
	if err != nil {
		t.Errorf("DefaultServer() error = %v", err)
	}
	if server == nil {
		t.Errorf("DefaultServer() = nil, want non-nil")
	}
}

// TestNewServer tests the NewServer function
func TestNewServer(t *testing.T) {
	opts := DefaultOptions()
	uid, err := uuid.V4()
	if err != nil {
		t.Errorf("uuid.V4() error = %v", err)
	}
	opts.Id = uid.String()
	server, err := New(opts)
	if err != nil {
		t.Errorf("NewServer() error = %v", err)
	}
	if server == nil {
		t.Errorf("NewServer() = nil, want non-nil")
	}
}

// TestRestServer_AddRoute tests the AddRoute function
func TestRestServer_AddRoute(t *testing.T) {
	server, err := Default()
	if err != nil {
		t.Fatalf("DefaultServer() error = %v", err)
	}
	rs := server.(*restServer)
	handler := func(ctx Context) {}
	err = rs.AddRoute("/test", handler, http.MethodGet)
	if err != nil {
		t.Errorf("AddRoute() error = %v", err)
	}
}

// TestRestServer_Post tests the Post function
func TestRestServer_Post(t *testing.T) {
	server, err := Default()
	if err != nil {
		t.Fatalf("DefaultServer() error = %v", err)
	}
	rs := server.(*restServer)
	handler := func(ctx Context) {}
	err = rs.Post("/test", handler)
	if err != nil {
		t.Errorf("Post() error = %v", err)
	}
}

// TestRestServer_Get tests the Get function
func TestRestServer_Get(t *testing.T) {
	server, err := Default()
	if err != nil {
		t.Fatalf("DefaultServer() error = %v", err)
	}
	rs := server.(*restServer)
	handler := func(ctx Context) {}
	err = rs.Get("/test", handler)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
}

// TestRestServer_Put tests the Put function
func TestRestServer_Put(t *testing.T) {
	server, err := Default()
	if err != nil {
		t.Fatalf("DefaultServer() error = %v", err)
	}
	rs := server.(*restServer)
	handler := func(ctx Context) {}
	err = rs.Put("/test", handler)
	if err != nil {
		t.Errorf("Put() error = %v", err)
	}
}

// TestRestServer_Delete tests the Delete function
func TestRestServer_Delete(t *testing.T) {
	server, err := Default()
	if err != nil {
		t.Fatalf("DefaultServer() error = %v", err)
	}
	rs := server.(*restServer)
	handler := func(ctx Context) {}
	err = rs.Delete("/test", handler)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

// TestRestServer_Opts tests the Opts function
func TestRestServer_Opts(t *testing.T) {
	server, err := Default()
	if err != nil {
		t.Fatalf("DefaultServer() error = %v", err)
	}
	rs := server.(*restServer)
	opts := rs.Opts()
	if opts == nil {
		t.Errorf("Opts() = nil, want non-nil")
	}
}

// TestRestServer_Lifecycle tests the lifecycle functions
func TestRestServer_Lifecycle(t *testing.T) {
	server, err := Default()
	assert.NoError(t, err)
	mgr := lifecycle.NewSimpleComponentManager()
	mgr.Register(server)
	err = mgr.StartAll()
	go func() {
		time.Sleep(1000 * time.Millisecond)
		err := mgr.StopAll()
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)
}
