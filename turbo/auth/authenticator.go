package auth

import (
	"net/http"
)

// TODO : Documentation

type Authenticator interface {
	Apply(handler http.Handler) http.Handler
}

//OAuth Filter Implementation

/*type OAuth struct {}
func (oAuth *OAuth) Apply(next http.Handler) http.Handler {
}*/
