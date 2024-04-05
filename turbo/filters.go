package turbo

import (
	"net/http"

	"go.nandlabs.io/golly/l3"
	"go.nandlabs.io/golly/turbo/auth"
)

// FilterFunc FuncHandler for with which the Filters need to be defined
type FilterFunc func(http.Handler) http.Handler

// AddFilter Making the Filter Chain in the order of filters being added
// if f1, f2, f3, finalHandler handlers are added to the filter chain then the order of execution remains
// f1 -> f2 -> f3 -> finalHandler
func (route *Route) AddFilter(filter ...FilterFunc) *Route {
	newFilters := make([]FilterFunc, 0, len(route.filters)+len(filter))
	newFilters = append(newFilters, route.filters...)
	newFilters = append(newFilters, filter...)
	route.filters = newFilters
	return route
}

// AddAuthenticator Adding the authenticator filter to the route
func (route *Route) AddAuthenticator(auth auth.Authenticator) *Route {
	route.authFilter = auth
	return route
}

// SetLogger Sets the custom logger is required at the route level
func (route *Route) SetLogger(logger *l3.BaseLogger) *Route {
	route.logger = logger
	return route
}
