package turbo

import (
	"net/http"
	"path"
)

//Common constants used throughout
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

//refinePath Borrowed from the golang's net/turbo package
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

//endpointNotFound to check for the request endpoint
func endpointNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, err := w.Write([]byte("Endpoint Not Found : " + r.URL.Path + "\n"))
	if err != nil {
		return
	}
}

//endpointNotFoundHandler when a requested endpoint is not found in the registered route's this handler is invoked
func endpointNotFoundHandler() http.Handler {
	return http.HandlerFunc(endpointNotFound)
}

//methodNotAllowed to check for the supported method for the incoming request
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, err := w.Write([]byte("Requested Method : " + r.Method + " not supported for Endpoint : " + r.URL.Path + "\n"))
	if err != nil {
		return
	}
}

//methodNotAllowedHandler when a requested method is not allowed in the registered route's method list this handler is invoked
func methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(methodNotAllowed)
}
