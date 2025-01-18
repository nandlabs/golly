package rest

import (
	"context"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/textutils"
	"oss.nandlabs.io/golly/turbo"
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
	AddRoute(path string, handler HandlerFunc, method ...string) (route *turbo.Route, err error)
	// AddRoute adds a route to the server
	Post(path string, handler HandlerFunc) (route *turbo.Route, err error)
	// AddRoute adds a route to the server
	Get(path string, handler HandlerFunc) (route *turbo.Route, err error)
	// AddRoute adds a route to the server
	Put(path string, handler HandlerFunc) (route *turbo.Route, err error)
	// AddRoute adds a route to the server
	Delete(path string, handler HandlerFunc) (route *turbo.Route, err error)
	// Unhandled adds a handler for unhandled routes
	Unhandled(handler HandlerFunc) (err error)
	// Unsupported adds a handler for unsupported methods
	Unsupported(handler HandlerFunc) (err error)
	// AddGlobalFilter adds a global filter to the server
	AddGlobalFilter(filter turbo.FilterFunc) (err error)
	//Turbo returns the turbo router
	Router() *turbo.Router
}
type DataTypProvider func() any

type restServer struct {
	*lifecycle.SimpleComponent
	opts       *Options
	router     *turbo.Router
	httpServer *http.Server
}

// AddRoute adds a route to the server
func (rs *restServer) AddRoute(path string, handler HandlerFunc, methods ...string) (route *turbo.Route, err error) {
	p := path
	if rs.opts.PathPrefix != textutils.EmptyStr {
		if !strings.HasPrefix(path, PathSeparator) {
			p = "/" + path
		}
		if strings.HasSuffix(rs.opts.PathPrefix, PathSeparator) {
			p = path[1:]
		}
	}
	p = rs.opts.PathPrefix + p
	route, err = rs.router.Add(p, func(w http.ResponseWriter, r *http.Request) {
		ctx := Context{
			request:  r,
			response: w,
		}
		handler(ctx)
	}, methods...)
	return
}

// Post adds a route to the server
func (rs *restServer) Post(path string, handler HandlerFunc) (route *turbo.Route, err error) {
	return rs.AddRoute(path, handler, http.MethodPost)
}

// Get adds a route to the server
func (rs *restServer) Get(path string, handler HandlerFunc) (route *turbo.Route, err error) {
	return rs.AddRoute(path, handler, http.MethodGet)
}

// Put adds a route to the server
func (rs *restServer) Put(path string, handler HandlerFunc) (route *turbo.Route, err error) {
	return rs.AddRoute(path, handler, http.MethodPut)
}

// Delete adds a route to the server
func (rs *restServer) Delete(path string, handler HandlerFunc) (route *turbo.Route, err error) {
	return rs.AddRoute(path, handler, http.MethodDelete)
}

// Unhandled adds a handler for unhandled routes
func (rs *restServer) Unhandled(handler HandlerFunc) (err error) {
	rs.router.SetUnmanaged(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := Context{
			request:  r,
			response: w,
		}
		handler(ctx)
	}))
	return
}

// Unsupported adds a handler for unsupported methods
func (rs *restServer) Unsupported(handler HandlerFunc) (err error) {
	rs.router.SetUnsupportedMethod(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := Context{
			request:  r,
			response: w,
		}
		handler(ctx)
	}))
	return
}

// AddFilter adds a filter to the server
func (rs *restServer) AddGlobalFilter(filter turbo.FilterFunc) (err error) {
	rs.router.AddGlobalFilter(filter)
	return
}

// Router returns the turbo router
func (rs *restServer) Router() *turbo.Router {
	return rs.router
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
	opts := DefaultSrvOptions()
	// uid, err := uuid.V4()
	// if err != nil {
	// 	return nil, err

	// }
	// opts.Id = uid.String()
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
	router.AddCorsFilter(opts.Cors)

	httpServer := &http.Server{
		Handler:      router,
		Addr:         opts.ListenHost + ":" + strconv.Itoa(int(opts.ListenPort)),
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Second,
	}
	var listener net.Listener
	rServer = &restServer{
		SimpleComponent: &lifecycle.SimpleComponent{
			CompId: opts.Id,
			StartFunc: func() error {

				listener, err = net.Listen("tcp", httpServer.Addr)
				if err != nil {
					logger.ErrorF("Error starting server: %v", err)
				}
				return err
			},
			AfterStart: func(err error) {

				if err == nil {
					go func() {

						if opts.EnableTLS && opts.CertPath != textutils.EmptyStr && opts.PrivateKeyPath != textutils.EmptyStr {
							logger.InfoF("Starting to accept rest(https) requests on %s", httpServer.Addr)
							err = httpServer.ServeTLS(listener, opts.CertPath, opts.PrivateKeyPath)
							if err != nil {
								// if the server was closed intentionally, do not log the error
								if err != http.ErrServerClosed {
									logger.ErrorF("Error starting https server: %v", err)
								}
							}
							ioutils.CloserFunc(listener)

						} else {
							logger.InfoF("Starting to accept rest (http) requests on %s", httpServer.Addr)
							err = httpServer.Serve(listener)
							if err != nil {
								// if the server was closed intentionally, do not log the error
								if err != http.ErrServerClosed {
									logger.ErrorF("Error starting https server: %v", err)
								}
							}
							ioutils.CloserFunc(listener)
						}
					}()
				}
			},

			StopFunc: func() error {
				logger.Info("Stopping rest server at ", httpServer.Addr)
				return httpServer.Shutdown(context.Background())
			},
		},
		opts:       opts,
		router:     router,
		httpServer: httpServer,
	}

	return
}
