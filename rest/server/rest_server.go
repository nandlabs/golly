package server

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/lifecycle"
	"oss.nandlabs.io/golly/turbo"
	"oss.nandlabs.io/golly/uuid"
	"oss.nandlabs.io/golly/vfs"
)

// Server is the interface that wraps the ServeHTTP method.
type Server interface {
	lifecycle.Component
	Opts() *Options
}

var servers = make(map[string]Server)
var mutex = &sync.RWMutex{}

type restServer struct {
	*lifecycle.SimpleComponent
	opts       *Options
	router     *turbo.Router
	httpServer *http.Server
}

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

func DefaultServer() (Server, error) {
	opts := DefaultOptions()
	uid, err := uuid.V4()
	if err != nil {
		return nil, err

	}
	opts.Id = uid.String()
	return NewServer(opts)
}

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
