package auth

import (
	"net/http"
)

// TODO : Documentation

type Authenticator interface {
	Apply(handler http.Handler) http.Handler
}

// OAuth Filter Implementation placeholder.
