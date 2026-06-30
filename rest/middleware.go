package rest

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"oss.nandlabs.io/golly/turbo"
)

// AccessLog returns a turbo.FilterFunc that logs one line per request
// at INFO via the package's l3 logger. The line carries: remote addr,
// method, path, status, response size in bytes, and wall-clock
// duration.
//
// Register globally with Server.AddGlobalFilter(rest.AccessLog()).
func AccessLog() turbo.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			logger.InfoF("%s %s %s %d %d %s",
				r.RemoteAddr,
				r.Method,
				r.URL.RequestURI(),
				rw.status,
				rw.bytes,
				time.Since(start),
			)
		})
	}
}

// Recover returns a turbo.FilterFunc that recovers from a handler
// panic, writes a 500 response (best-effort — see below), and logs
// the panic value + stack at ERROR via l3.
//
// Register globally as the FIRST filter so it wraps everything else:
//
//	srv.AddGlobalFilter(rest.Recover())
//	srv.AddGlobalFilter(rest.AccessLog())
//
// Without recovery, a panic in any handler crashes the request's
// goroutine and surfaces as a connection reset to the client.
//
// Best-effort 500: if the handler had already started writing the
// response before panicking (WriteHeader / Write already called),
// net/http silently no-ops the second WriteHeader and the client
// sees a truncated body. That's a fundamental HTTP limitation, not
// a recovery bug — the panic itself is always logged.
func Recover() turbo.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				// http.ErrAbortHandler is the documented way for a
				// handler to give up without logging — propagate so
				// net/http closes the connection silently.
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				logger.ErrorF("rest: panic recovered in %s %s: %v\n%s",
					r.Method, r.URL.Path, rec, debug.Stack())
				http.Error(w, fmt.Sprintf("internal server error: %v", rec), http.StatusInternalServerError)
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder wraps http.ResponseWriter so AccessLog can read back
// the status code and bytes written. WriteHeader / Write are passed
// through; the recorder just captures the metadata.
type statusRecorder struct {
	http.ResponseWriter
	status  int
	bytes   int
	written bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.written {
		s.status = code
		s.written = true
		s.ResponseWriter.WriteHeader(code)
	}
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.written {
		s.status = http.StatusOK
		s.written = true
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}
