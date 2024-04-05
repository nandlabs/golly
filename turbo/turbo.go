package turbo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"go.nandlabs.io/golly/l3"
	"go.nandlabs.io/golly/textutils"
	"go.nandlabs.io/golly/turbo/auth"
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
}

// Param to hold key value
type Param struct {
	key   string
	value string
}

// Route base struct to hold the route information
type Route struct {
	//name of the route fragment if this is a path variable the name of the variable will be used here.
	path      string
	paramType string
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
	logger.InfoF("Initiating Turbo")
	return &Router{
		lock:                     sync.RWMutex{},
		unManagedRouteHandler:    endpointNotFoundHandler(),
		unsupportedMethodHandler: methodNotAllowedHandler(),
		topLevelRoutes:           make(map[string]*Route),
	}
}

// Get to Add a turbo handler for GET method
func (router *Router) Get(path string, f func(w http.ResponseWriter, r *http.Request)) *Route {
	return router.Add(path, f, GET)
}

// Post to Add a turbo handler for POST method
func (router *Router) Post(path string, f func(w http.ResponseWriter, r *http.Request)) *Route {
	return router.Add(path, f, POST)
}

// Put to Add a turbo handler for PUT method
func (router *Router) Put(path string, f func(w http.ResponseWriter, r *http.Request)) *Route {
	return router.Add(path, f, PUT)
}

// Delete to Add a turbo handler for DELETE method
func (router *Router) Delete(path string, f func(w http.ResponseWriter, r *http.Request)) *Route {
	return router.Add(path, f, DELETE)
}

// Add a turbo handler for one or more HTTP methods.
func (router *Router) Add(path string, f func(w http.ResponseWriter, r *http.Request), methods ...string) *Route {
	router.lock.Lock()
	defer router.lock.Unlock()
	var route *Route = nil
	//Check if the methods provided are valid if not return error straight away
	for _, method := range methods {
		if _, contains := Methods[method]; !contains {
			panic(fmt.Sprintf("Invalid/Unsupported Http method  %s provided", method))
		}
	}
	logger.InfoF("Registering New Route: %s", path)
	//TODO add path check for any query variables specified.
	pathValue := strings.TrimSpace(path)

	//Adds support to path with variables in {} format instead of : prefix
	var sb strings.Builder
	for _, c := range pathValue {
		if c == textutils.OpenBraceChar {
			sb.WriteRune(textutils.ColonChar)
		} else if c == textutils.CloseBraceChar {
			logger.Debug("Ignoring char ", textutils.CloseBraceStr)
		} else {
			sb.WriteRune(c)
		}
	}
	pathValue = sb.String()

	pathValues := strings.Split(pathValue, PathSeparator)[1:]
	length := len(pathValues)
	if length > 0 && pathValues[0] != textutils.EmptyStr {
		isPathVar := false
		name := textutils.EmptyStr
		for i, pathValue := range pathValues {
			isPathVar = pathValue[0] == textutils.ColonChar
			if isPathVar {
				name = pathValue[1:]
			} else {
				name = pathValue
			}
			currentRoute := &Route{
				path:         name,
				isPathVar:    isPathVar,
				childVarName: textutils.EmptyStr,
				hasChildVar:  false,
				authFilter:   nil,
				logger:       logger,
				handlers:     make(map[string]http.Handler),
				subRoutes:    make(map[string]*Route),
				queryParams:  make(map[string]*QueryParam),
			}
			if route == nil {
				if v, ok := router.topLevelRoutes[name]; ok {
					route = v
				} else {
					//No Parent present add the current route as route and continue
					if currentRoute.isPathVar {
						panic("the framework does not support path variables at root context")
					}
					router.topLevelRoutes[name] = currentRoute
					route = currentRoute
				}
			} else {
				if v, ok := route.subRoutes[name]; ok {
					if v.isPathVar && isPathVar && v.path != name {
						panic("one path cannot have multiple names")
					}
					route = v
				} else {
					route.subRoutes[name] = currentRoute
					if isPathVar {
						route.childVarName = name
						route.hasChildVar = true
					}
					route = currentRoute
				}
			}
			//At Last index add the method(s) to the map.
			if i == len(pathValues)-1 {
				for _, method := range methods {
					currentRoute.handlers[method] = http.HandlerFunc(f)
				}
			}
		}
	} else {
		//TODO Handle the Root context path
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
			currentRoute.handlers[method] = prepareHandler(method, http.HandlerFunc(f))
		}
		//Root route will not have any path value
		router.topLevelRoutes[textutils.EmptyStr] = currentRoute
	}
	return route
}

// prepareHandler to add any default features like logging, auth... will be injected here
func prepareHandler(method string, handler http.Handler) http.Handler {
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
		_, err := w.Write([]byte("Path Moved : " + p + "\n"))
		if err != nil {
			logger.Error(err)
		}
		return
	}
	// start by checking where the method of the Request is same as that of the registered method
	match, params := router.findRoute(r)
	if match != nil {
		handler = match.handlers[r.Method]
		if len(match.filters) > 0 {
			//Middlewares added
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

// GetPathParams fetches the path parameters
func (router *Router) GetPathParams(id string, r *http.Request) (string, error) {
	params, ok := r.Context().Value("params").([]Param)
	if !ok {
		logger.ErrorF("Error Fetching Path Param %s", id)
		return "err", errors.New(fmt.Sprintf("error fetching path param %s", id))
	}
	for _, p := range params {
		if p.key == id {
			return p.value, nil
		}
	}
	return "", errors.New(fmt.Sprintf("No Such parameter %s", id))
}

// GetIntPathParams fetches the int path parameters
func (router *Router) GetIntPathParams(id string, r *http.Request) (int, error) {
	val, err := router.GetPathParams(id, r)
	if err != nil {
		return -1, err
	}
	valInt, err := strconv.Atoi(val)
	if err != nil {
		return -1, err
	}
	return valInt, nil
}

// GetFloatPathParams fetches the float path parameters
func (router *Router) GetFloatPathParams(id string, r *http.Request) (float64, error) {
	val, err := router.GetPathParams(id, r)
	if err != nil {
		return -1, err
	}
	valFloat, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return -1, err
	}
	return valFloat, nil
}

// GetBoolPathParams fetches the bool path parameters
func (router *Router) GetBoolPathParams(id string, r *http.Request) (bool, error) {
	val, err := router.GetPathParams(id, r)
	if err != nil {
		return false, err
	}
	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return false, err
	}
	return valBool, nil
}

// GetQueryParams fetches the query parameters
func (router *Router) GetQueryParams(id string, r *http.Request) (string, error) {
	val := r.URL.Query().Get(id)
	if val == "" {
		logger.ErrorF("Error Fetching Query Param %s", id)
		return "err", errors.New(fmt.Sprintf("error fetching query param %s", id))
	}
	return val, nil
}

// GetIntQueryParams fetches the int query parameters
func (router *Router) GetIntQueryParams(id string, r *http.Request) (int, error) {
	val, ok := strconv.Atoi(r.URL.Query().Get(id))
	if ok != nil {
		logger.ErrorF("Error Fetching Query Parameter %s", id)
		return -1, errors.New(fmt.Sprintf("error fetching query param %s", id))
	}
	return val, nil
}

// GetFloatQueryParams fetches the float query parameters
func (router *Router) GetFloatQueryParams(id string, r *http.Request) (float64, error) {
	val, ok := strconv.ParseFloat(r.URL.Query().Get(id), 64)
	if ok != nil {
		logger.ErrorF("Error Fetching Query Parameter %s", id)
		return -1, errors.New(fmt.Sprintf("error fetching query param %s", id))
	}
	return val, nil
}

// GetBoolQueryParams fetches the boolean query parameters
func (router *Router) GetBoolQueryParams(id string, r *http.Request) (bool, error) {
	val, ok := strconv.ParseBool(r.URL.Query().Get(id))
	if ok != nil {
		logger.ErrorF("Error Fetching Query Parameter %s", id)
		return false, errors.New(fmt.Sprintf("error fetching query param %s", id))
	}
	return val, nil
}
