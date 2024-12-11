package turbo

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"oss.nandlabs.io/golly/l3"
	"oss.nandlabs.io/golly/textutils"
	"oss.nandlabs.io/golly/turbo/auth"
)

// Router struct that holds the router configuration
type Router struct {
	lock sync.RWMutex
	//Handler for any route that is not defined
	unManagedRouteHandler http.Handler
	//Handler for any methods that are not supported
	unsupportedMethodHandler http.Handler
	//Routes Managed by this router
	topLevelRoutes map[string]*Route
	//global filters
	globalFilters []FilterFunc
}

// Param to hold key value
type Param struct {
	key   string
	value string
}

// Route base struct to hold the route information
type Route struct {
	//name of the route fragment if this is a path variable the name of the variable will be used here.
	path string
	//Checks if this is a variable. only one path variable at this level will be supported.
	isPathVar bool
	//childVarName varName
	childVarName string
	//hasChildVar
	hasChildVar bool
	//isAuthenticated keeps a check whether the route is authenticated or not
	authFilter auth.Authenticator
	//filters array to store the ...http.handler being registered for middleware in the router
	filters []FilterFunc
	//handlers for HTTP Methods <method>|<Handler>
	handlers map[string]http.Handler
	//Sub Routes from this path
	subRoutes map[string]*Route
	//Query Parameters that may be used.
	queryParams map[string]*QueryParam
	//logger to set the external logger if required using SetLogger()
	logger l3.Logger
}

// QueryParam for the Route configuration
type QueryParam struct {
	//required flag : fail upfront if a required query param not present
	required bool
	//name of the query parameter
	name string
	// TODO add mechanism for creating a typed query parameter to do auto type conversion in the framework.
}

// NewRouter registers the new instance of the Turbo Framework
func NewRouter() *Router {
	logger.DebugF("Initiating Turbo")
	return &Router{
		lock:                     sync.RWMutex{},
		unManagedRouteHandler:    endpointNotFoundHandler(),
		unsupportedMethodHandler: methodNotAllowedHandler(),
		topLevelRoutes:           make(map[string]*Route),
	}
}

// AddGlobalFilter to add a global filter to the router
func (router *Router) AddGlobalFilter(filter ...FilterFunc) *Router {
	router.lock.Lock()
	defer router.lock.Unlock()
	router.globalFilters = append(router.globalFilters, filter...)
	return router
}

// Get to Add a turbo handler for GET method
func (router *Router) Get(path string, f func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return router.Add(path, f, GET)
}

// Post to Add a turbo handler for POST method
func (router *Router) Post(path string, f func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return router.Add(path, f, POST)
}

// Put to Add a turbo handler for PUT method
func (router *Router) Put(path string, f func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return router.Add(path, f, PUT)
}

// Delete to Add a turbo handler for DELETE method
func (router *Router) Delete(path string, f func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return router.Add(path, f, DELETE)
}

func sanitizePath(p string) (string, error) {
	path := strings.TrimSpace(p)
	if path == textutils.EmptyStr {
		return textutils.EmptyStr, ErrInvalidPath
	}
	if !strings.HasPrefix(path, textutils.ForwardSlashStr) {
		path = textutils.ForwardSlashStr + path
	}
	var sb strings.Builder
	for _, c := range path {
		// Path Variable can be defined using {<name>} syntax or :<name> syntax
		// Allowed characters in the path are A-Z, a-z, 0-9, -, _, ., ~, :, /, {, }
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || c == 45 || c == 95 || c == 46 || c == 126 || c == 58 || c == 47 || c == 123 || c == 125 {
			if c == textutils.OpenBraceChar {
				sb.WriteRune(textutils.ColonChar)
			} else if c == textutils.CloseBraceChar {
				logger.Debug("Ignoring char ", textutils.CloseBraceStr)
			} else {
				sb.WriteRune(c)
			}
		} else {
			return textutils.EmptyStr, ErrInvalidPath
		}

	}
	return sb.String(), nil
}

func (router *Router) AddHandler(path string, h http.Handler, methods ...string) (route *Route, err error) {

	router.lock.Lock()
	defer router.lock.Unlock()
	var pathValue string
	var pathValues []string
	var length int
	//Check if the methods provided are valid if not return error straight away
	for _, method := range methods {
		if _, contains := Methods[strings.ToUpper(method)]; !contains {
			return nil, ErrInvalidMethod
		}
	}

	pathValue, err = sanitizePath(path)
	if err != nil {
		return
	}
	pathValues = strings.Split(pathValue, PathSeparator)
	// check for the leading empty path value and remove it
	if len(pathValues) > 1 && pathValues[0] == textutils.EmptyStr {
		pathValues = pathValues[1:]
	}
	length = len(pathValues)

	if length > 0 && pathValues[0] != textutils.EmptyStr {
		isPathVar := false
		currentPathName := textutils.EmptyStr
		for i, pathValue := range pathValues {
			isPathVar = pathValue[0] == textutils.ColonChar
			if isPathVar {
				currentPathName = pathValue[1:]
			} else {
				currentPathName = pathValue
			}
			currentRoute := &Route{
				path:         currentPathName,
				isPathVar:    isPathVar,
				childVarName: textutils.EmptyStr,
				hasChildVar:  false,
				authFilter:   nil,
				logger:       logger,
				handlers:     make(map[string]http.Handler),
				subRoutes:    make(map[string]*Route),
				queryParams:  make(map[string]*QueryParam),
			}
			if i == 0 {
				// the route will be nil only on the first iteration
				if v, ok := router.topLevelRoutes[currentPathName]; ok {
					route = v
				} else {
					//No Parent present add the current route as route and continue
					if currentRoute.isPathVar {
						return nil, ErrInvalidPath
					}
					router.topLevelRoutes[currentPathName] = currentRoute
					route = currentRoute

				}
			} else {
				// current route is not nil, it means that we are already in the middle of the path somewhere
				if v, ok := route.subRoutes[currentPathName]; ok {
					// if the path is already present in the subroutes then we will just move to the next path
					if v.isPathVar && isPathVar && v.path != currentPathName {
						return nil, ErrInvalidPath
					}
					route = v

				} else {
					// if the path is not present in the subroutes then we will add the path to the subroutes and move to the next path
					route.subRoutes[currentPathName] = currentRoute
					if isPathVar {
						route.childVarName = currentPathName
						route.hasChildVar = true
					}
					route = currentRoute
				}

			}
			if i == length-1 {
				for _, method := range methods {
					// if the handler is already present then we will overwrite it
					m := strings.ToUpper(method)
					logger.InfoF("Registering New Route: %s:%s", m, path)

					route.handlers[m] = h
				}
			}

		}
	} else {
		currentRoute := &Route{
			path:         textutils.EmptyStr,
			isPathVar:    false,
			childVarName: textutils.EmptyStr,
			handlers:     make(map[string]http.Handler),
			subRoutes:    make(map[string]*Route),
			queryParams:  make(map[string]*QueryParam),
			authFilter:   nil,
			logger:       logger,
		}
		for _, method := range methods {
			currentRoute.handlers[method] = prepareHandler(method, h)
		}
		//Root route will not have any path value
		router.topLevelRoutes[textutils.EmptyStr] = currentRoute
	}
	return route, nil

}

// Add a turbo handler for one or more HTTP methods.
func (router *Router) Add(path string, f func(w http.ResponseWriter, r *http.Request), methods ...string) (route *Route, err error) {

	return router.AddHandler(path, http.HandlerFunc(f), methods...)
}

// prepareHandler to add any default features like logging, auth... will be injected here
func prepareHandler(method string, handler http.Handler) http.Handler {
	//TODO add features later
	return handler
}

// addQueryVar to add query params to the route
func (route *Route) addQueryVar(name string, required bool) *Route {
	//TODO add name validation.
	queryParams := &QueryParam{
		required: required,
		name:     name,
	}
	//TODO Check if this name can be url encoded and save decoding per request,
	route.queryParams[name] = queryParams
	return route
}

// ServeHTTP
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var handler http.Handler
	// perform the path checks before, set the 301 status even before further computation
	// these checks need not be performed once the PreWork is refined and up to the mark
	if p := refinePath(path); p != path {
		url := *r.URL
		url.Path = p
		p = url.String()
		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently)
		fmt.Fprintf(w, "Path Moved : %q \n", html.EscapeString(p))
		return
	}
	// start by checking where the method of the Request is same as that of the registered method
	match, params := router.findRoute(r)
	if match != nil {
		handler = match.handlers[r.Method]
		//Global Middlewares added
		if router.globalFilters != nil {
			for i := range router.globalFilters {
				handler = router.globalFilters[len(router.globalFilters)-1-i](handler)
			}
		}
		//Route specific Middlewares added
		if len(match.filters) > 0 {
			for i := range match.filters {
				handler = match.filters[len(match.filters)-1-i](handler)
			}
		}
		// check for authenticated filter explicitly at the top
		// we add all the filters added by the user in its order and if the user has added an Authenticator Filter then it will always be executed first
		if match.authFilter != nil {
			handler = match.authFilter.Apply(handler)
		}
	} else {
		handler = router.unManagedRouteHandler
	}
	if handler == nil {
		handler = router.unsupportedMethodHandler
	}
	if params != nil {
		r = r.WithContext(context.WithValue(r.Context(), "params", params))
	}
	handler.ServeHTTP(w, r)
}

func (r *Router) SetUnmanaged(handler http.Handler) *Router {
	r.unManagedRouteHandler = handler
	return r
}

func (r *Router) SetUnsupportedMethod(handler http.Handler) *Router {
	r.unsupportedMethodHandler = handler
	return r
}

// findRoute performs the function checks for the incoming request path whether it matches with any registered route's path
func (router *Router) findRoute(req *http.Request) (*Route, []Param) {
	var route *Route
	var params []Param = nil
	pathLen := len(req.URL.Path)
	prevIdx := 1
	lastIdx := false
	for idx := 1; idx < pathLen; idx++ {
		lastIdx = idx == pathLen-1
		if req.URL.Path[idx] == textutils.ForwardSlashChar || lastIdx {
			if lastIdx {
				idx++
			}
			val := req.URL.Path[prevIdx:idx]
			prevIdx = idx + 1
			if route == nil {
				route = router.topLevelRoutes[val]
				continue
			} else {
				if route.hasChildVar {
					route = route.subRoutes[route.childVarName]
				} else {
					if r, ok := route.subRoutes[val]; ok {
						route = r
					} else {
						return nil, nil
					}
				}
				if route.isPathVar {
					if params == nil {
						params = []Param{}
					}
					params = append(params, Param{
						key:   route.path,
						value: val,
					})
				}
			}
		}
	}
	return route, params
}

// GetPathParam fetches the path parameters
func GetPathParam(id string, r *http.Request) (string, error) {
	params, ok := r.Context().Value("params").([]Param)
	if !ok {
		logger.ErrorF("Error Fetching Path Param %s", id)

		return textutils.EmptyStr, fmt.Errorf("error fetching path param %s", id)
	}
	for _, p := range params {
		if p.key == id {
			return p.value, nil
		}
	}
	return textutils.EmptyStr, fmt.Errorf("no such parameter %s", id)
}

// GetPathParamAsInt fetches the int path parameters
func GetPathParamAsInt(id string, r *http.Request) (int, error) {
	val, err := GetPathParam(id, r)
	if err != nil {
		return -1, err
	}
	valInt, err := strconv.Atoi(val)
	if err != nil {
		return -1, err
	}
	return valInt, nil
}

// GetPathParamAsFloat fetches the float path parameters
func GetPathParamAsFloat(id string, r *http.Request) (float64, error) {
	val, err := GetPathParam(id, r)
	if err != nil {
		return -1, err
	}
	valFloat, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, err
	}
	return valFloat, nil
}

// GetPathParamAsBool fetches the bool path parameters
func GetPathParamAsBool(id string, r *http.Request) (bool, error) {
	val, err := GetPathParam(id, r)
	if err != nil {
		return false, err
	}
	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return false, err
	}
	return valBool, nil
}

// GetQueryParam fetches the query parameters
func GetQueryParam(id string, r *http.Request) (string, error) {
	val := r.URL.Query().Get(id)
	if val == "" {
		logger.ErrorF("Error Fetching Query Param %s", id)
		return "err", fmt.Errorf("error fetching query param %s", id)
	}
	return val, nil
}

// GetQueryParamAsInt fetches the int query parameters
func GetQueryParamAsInt(id string, r *http.Request) (int, error) {
	val, ok := strconv.Atoi(r.URL.Query().Get(id))
	if ok != nil {
		logger.ErrorF("Error Fetching Query Parameter %s", id)
		return -1, fmt.Errorf("error fetching query param %s", id)
	}
	return val, nil
}

// GetQueryParamAsFloat fetches the float query parameters
func GetQueryParamAsFloat(id string, r *http.Request) (float64, error) {
	val, ok := strconv.ParseFloat(r.URL.Query().Get(id), 64)
	if ok != nil {
		logger.ErrorF("Error Fetching Query Parameter %s", id)
		return -1, fmt.Errorf("error fetching query param %s", id)
	}
	return val, nil
}

// GetQueryParamAsBool fetches the boolean query parameters
func GetQueryParamAsBool(id string, r *http.Request) (bool, error) {
	val, ok := strconv.ParseBool(r.URL.Query().Get(id))
	if ok != nil {
		logger.ErrorF("Error Fetching Query Parameter %s", id)
		return false, fmt.Errorf("error fetching query param %s", id)
	}
	return val, nil
}
