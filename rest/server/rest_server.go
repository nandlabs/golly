package server

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/rest"
	"oss.nandlabs.io/golly/textutils"
	"oss.nandlabs.io/golly/turbo"
	"oss.nandlabs.io/golly/uuid"
	"oss.nandlabs.io/golly/vfs"
)

const (
	QueryParam Paramtype = iota
	PathParam
)

type HandlerFunc func(context Context)

type Paramtype int

// Server is the interface that wraps the ServeHTTP method.
type Server interface {
	// Server is a lifecytcle component
	lifecycle.Component
	// Opts returns the options of the server
	Opts() *Options
	// AddRoute adds a route to the server
	AddRoute(path string, handler HandlerFunc, method ...string) (err error)
	// AddRoute adds a route to the server
	Post(path string, handler HandlerFunc) (err error)
	// AddRoute adds a route to the server
	Get(path string, handler HandlerFunc) (err error)
	// AddRoute adds a route to the server
	Put(path string, handler HandlerFunc) (err error)
	// AddRoute adds a route to the server
	Delete(path string, handler HandlerFunc) (err error)
}
type DataTypProvider func() any

var servers = make(map[string]Server)
var mutex = &sync.RWMutex{}

type restServer struct {
	*lifecycle.SimpleComponent
	opts       *Options
	router     *turbo.Router
	httpServer *http.Server
}

// AddRoute adds a route to the server
func (rs *restServer) AddRoute(path string, handler HandlerFunc, methods ...string) (err error) {
	p := path
	if rs.opts.PathPrefix != textutils.EmptyStr {
		if !strings.HasPrefix(path, rest.PathSeparator) {
			p = "/" + path
		}
		if strings.HasSuffix(rs.opts.PathPrefix, rest.PathSeparator) {
			p = path[1:]
		}
	}
	p = rs.opts.PathPrefix + p
	_, err = rs.router.Add(p, func(w http.ResponseWriter, r *http.Request) {
		ctx := Context{
			request:  r,
			response: w,
		}
		handler(ctx)
	}, methods...)
	return
}

// Post adds a route to the server
func (rs *restServer) Post(path string, handler HandlerFunc) (err error) {
	return rs.AddRoute(path, handler, http.MethodPost)
}

// Get adds a route to the server
func (rs *restServer) Get(path string, handler HandlerFunc) (err error) {
	return rs.AddRoute(path, handler, http.MethodGet)
}

// Put adds a route to the server
func (rs *restServer) Put(path string, handler HandlerFunc) (err error) {
	return rs.AddRoute(path, handler, http.MethodPut)
}

// Delete adds a route to the server
func (rs *restServer) Delete(path string, handler HandlerFunc) (err error) {
	return rs.AddRoute(path, handler, http.MethodDelete)
}

// Opts returns the options of the server
func (rs *restServer) Opts() *Options {
	return rs.opts
}

// New creates a new Server with the given configuration file of the options.
func NewServerFrom(configPath string) (Server, error) {
	// Read from file.
	vFile, err := vfs.GetManager().OpenRaw(configPath)
	var opts *Options
	if err != nil {
		return nil, err
	}

	mimeType := ioutils.GetMimeFromExt(path.Ext(configPath))
	// Get the codec for the file.
	codec, err := codec.GetDefault(mimeType)
	if err != nil {
		return nil, err
	}

	err = codec.Read(vFile, &opts)
	if err != nil {
		return nil, err
	}
	return NewServer(opts)

}

// DefaultServer creates a new Server with the default options.
func DefaultServer() (Server, error) {
	opts := DefaultOptions()
	uid, err := uuid.V4()
	if err != nil {
		return nil, err

	}
	opts.Id = uid.String()
	return NewServer(opts)
}

// NewServer creates a new Server with the given options.
func NewServer(opts *Options) (rServer Server, err error) {
	if opts == nil {
		return nil, ErrNilOptions
	}
	err = opts.Validate()
	if err != nil {
		return
	}
	router := turbo.NewRouter()
	httpServer := &http.Server{
		Handler:      router,
		Addr:         opts.ListenHost + ":" + strconv.Itoa(int(opts.ListenPort)),
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Second,
	}
	rServer = &restServer{
		SimpleComponent: &lifecycle.SimpleComponent{
			CompId: opts.Id,
			StartFunc: func() error {

				go httpServer.ListenAndServe()
				return nil

			},
			StopFunc: func() error {
				fmt.Println("Stopping HTTP server")
				return httpServer.Shutdown(context.Background())
			},
		},
		opts:       opts,
		router:     router,
		httpServer: httpServer,
	}
	mutex.Lock()
	defer mutex.Unlock()
	servers[opts.Id] = rServer
	return
}
