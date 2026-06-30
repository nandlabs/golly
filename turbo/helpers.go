package turbo

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"path"
	"sort"
	"strings"
)

// Common constants used throughout
const (
	PathSeparator = "/"
	GET           = "GET"
	HEAD          = "HEAD"
	POST          = "POST"
	PUT           = "PUT"
	DELETE        = "DELETE"
	OPTIONS       = "OPTIONS"
	TRACE         = "TRACE"
	PATCH         = "PATCH"
)

var Methods = map[string]string{
	GET:     GET,
	HEAD:    HEAD,
	POST:    POST,
	PUT:     PUT,
	DELETE:  DELETE,
	OPTIONS: OPTIONS,
	TRACE:   TRACE,
	PATCH:   PATCH,
}

var ErrInvalidMethod = errors.New("invalid method provided")
var ErrInvalidPath = errors.New("invalid path provided")
var ErrInvalidHandler = errors.New("invalid handler provided")

// refinePath Borrowed from the golang's net/turbo package
func refinePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	rp := path.Clean(p)
	if p[len(p)-1] == '/' && rp != "/" {
		rp += "/"
	}
	return rp
}

// endpointNotFound to check for the request endpoint
func endpointNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, "Endpoint not found :%q \n", html.EscapeString(r.URL.Path))
}

// endpointNotFoundHandler when a requested endpoint is not found in the registered route's this handler is invoked
func endpointNotFoundHandler() http.Handler {
	return http.HandlerFunc(endpointNotFound)
}

// methodNotAllowed to check for the supported method for the incoming request
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = fmt.Fprintf(w, "Method %q Not Supported for %q \n", html.EscapeString(r.Method), html.EscapeString(r.URL.Path))
}

// methodNotAllowedHandler when a requested method is not allowed in the registered route's method list this handler is invoked
func methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(methodNotAllowed)
}

// toggleTrailingSlash returns the path with a single trailing slash
// added or removed. Returns the empty string for the root "/" (no
// useful toggle exists). Used by ServeHTTP when StrictSlash is off to
// retry routing with the inverse trailing-slash form.
func toggleTrailingSlash(p string) string {
	if p == "" || p == "/" {
		return ""
	}
	if strings.HasSuffix(p, "/") {
		return strings.TrimRight(p, "/")
	}
	return p + "/"
}

// allowedMethods returns a comma-separated list of HTTP methods that
// the route has handlers registered for, sorted deterministically (so
// the Allow header value is stable across requests and easy to test).
// Returns the empty string when the route has no handlers.
func allowedMethods(route *Route) string {
	if route == nil || len(route.handlers) == 0 {
		return ""
	}
	methods := make([]string, 0, len(route.handlers))
	for m := range route.handlers {
		methods = append(methods, m)
	}
	sort.Strings(methods)
	return strings.Join(methods, ", ")
}
