package auth

type BasicAuthFilter struct {
}

/*func (ba *BasicAuthFilter) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO : Basic Auth Implementation
	})
}*/

func CreateBasicAuthAuthenticator() *BasicAuthFilter {
	return &BasicAuthFilter{}
}
