package filters

import (
	"net/http"
	"strconv"
	"strings"

	"oss.nandlabs.io/golly/assertion"
	"oss.nandlabs.io/golly/textutils"
)

const (
	AllowCredentials             = "Access-Control-Allow-Credentials"
	AllowExposeHeaders           = "Access-Control-Expose-Headers"
	AllowHeadersHeader           = "Access-Control-Allow-Headers"
	AllowMethodsHeader           = "Access-Control-Allow-Methods"
	AllowOriginHeader            = "Access-Control-Allow-Origin"
	MaxAgeHeader                 = "Access-Control-Max-Age"
	ExposeHeaders                = "Access-Control-Expose-Headers"
	AccessControlReqHeaders      = "Access-Control-Request-Headers"
	AccessControlReqMethodHdr    = "Access-Control-Request-Method"
	AccessControlPvtNetworkHdr   = "Access-Control-Allow-Private-Network"
	OriginHeader                 = "Origin"
	VaryHeader                   = "Vary"
	AccessControlAllowAllOrigins = "*"
	// DefaultAccessControlMaxAge is the default value for the MaxAge field
	DefaultAccessControlMaxAge = 0
	trueStr                    = "true"
)

// CorsOptions represents the options for the CORS filter
type CorsOptions struct {
	AllowCredentials     bool     `json:"allowCredentials" yaml:"allowCredentials"`
	AllowedHeaders       []string `json:"allowedHeaders" yaml:"allowedHeaders"`
	AllowedMethods       []string `json:"allowedMethods" yaml:"allowedMethods"`
	AllowedOrigins       []string `json:"allowedOrigins" yaml:"allowedOrigins"`
	ExposeHeaders        []string `json:"exposeHeaders" yaml:"exposeHeaders"`
	MaxAge               int      `json:"maxAge" yaml:"maxAge"`
	ResponseStatus       int      `json:"responseStatus" yaml:"responseStatus"`
	PreFlightPassThrough bool     `json:"preFlightPassThrough" yaml:"preFlightPassThrough"`
	AllowPvtNetwork      bool     `json:"allowPvtNetwork" yaml:"allowPvtNetwork"`
}

func (co *CorsOptions) NewFilter() *CorsFilter {

	cf := &CorsFilter{
		CorsOptions: co,
	}
	// Lowercase all the allowed origins
	for i, o := range co.AllowedOrigins {
		co.AllowedOrigins[i] = strings.ToLower(o)
	}
	// Uppercase all the allowed methods
	for i, m := range co.AllowedMethods {
		co.AllowedMethods[i] = strings.ToUpper(m)
	}

	cf.SetAllowPvtNetwork(false)
	for _, origin := range co.AllowedOrigins {
		if origin == AccessControlAllowAllOrigins {
			cf.AllowAllOrigins = true
			break
		}
	}

	for _, method := range co.AllowedMethods {
		if method == "*" {
			cf.AllowAllMethods = true
			break
		}
	}

	return cf
}

// CorsFilter represents the CORS filter
type CorsFilter struct {
	*CorsOptions
	accessControlReqHdrsStr string   `json:"-"`
	PreFlightVary           []string `json:"-"`
	AllowAllOrigins         bool     `json:"-" yaml:"-"`
	AllowAllMethods         bool     `json:"-" yaml:"-"`
}

// NewCorsFilter creates a new CorsFilter
func NewCorsFilter(allowedOrigins ...string) *CorsFilter {
	co := &CorsOptions{
		MaxAge:         DefaultAccessControlMaxAge,
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		ResponseStatus: http.StatusNoContent,
	}
	return co.NewFilter()
}

// SetAllowedHeaders sets the allowed headers
func (cf *CorsFilter) SetAllowedHeaders(headers ...string) {
	hdrs := make([]string, len(headers))
	for i, h := range headers {
		hdrs[i] = strings.ToLower(h)
	}
	cf.accessControlReqHdrsStr = strings.Join(headers, ",")
	cf.AllowedHeaders = hdrs
}

// SetAllowedMethods sets the allowed methods
func (cf *CorsFilter) SetAllowedMethods(methods ...string) {
	cf.AllowedMethods = methods
}

// SetExposeHeaders sets the headers to expose
func (cf *CorsFilter) SetExposeHeaders(headers ...string) {
	cf.ExposeHeaders = headers
}

// SetMaxAge sets the max age
func (cf *CorsFilter) SetMaxAge(maxAge int) {
	cf.MaxAge = maxAge
}

// SetAllowCredentials sets the allow credentials
func (cf *CorsFilter) SetAllowCredentials(allow bool) {
	cf.AllowCredentials = allow
}

// SetAllowedOrigins sets the allowed origins
func (cf *CorsFilter) SetAllowedOrigins(origins ...string) {
	cf.AllowedOrigins = origins
}

// SetPreFlightPassThrough sets the preflight pass through
func (cf *CorsFilter) SetPreFlightPassThrough(passThrough bool) {
	cf.PreFlightPassThrough = passThrough
}

// SetResponseStatus sets the response status
func (cf *CorsFilter) SetResponseStatus(status int) {
	cf.ResponseStatus = status
}

// SetAllowPvtNetwork sets the allow private network
func (cf *CorsFilter) SetAllowPvtNetwork(allow bool) {
	cf.AllowPvtNetwork = allow
	cf.PreFlightVary = []string{OriginHeader, AccessControlReqMethodHdr, AccessControlReqHeaders}
	if allow {
		cf.PreFlightVary = append(cf.PreFlightVary, AccessControlPvtNetworkHdr)
	}
}

// isOriginAllowed checks if the origin is allowed
func (cf *CorsFilter) isOriginAllowed(origin string) (bool, string) {
	if cf.AllowAllOrigins {
		return true, AccessControlAllowAllOrigins
	}
	return assertion.ListHas(strings.ToLower(origin), cf.AllowedOrigins), origin
}

// isMethodAllowed checks if the method is allowed
func (cf *CorsFilter) isMethodAllowed(method string) bool {
	return cf.AllowAllMethods || assertion.ListHas(method, cf.AllowedMethods)
}

// handlePreflight handles the preflight request
// It checks if the method and origin are allowed
// and sets the appropriate headers

func (cf *CorsFilter) handlePreflight(w http.ResponseWriter, r *http.Request) {
	reqOrigin := r.Header.Get(OriginHeader)

	// Add Vary Headers
	for _, vh := range cf.PreFlightVary {
		w.Header().Add(VaryHeader, vh)
	}

	// Check if the origin is originAllowed
	originAllowed, origin := cf.isOriginAllowed(reqOrigin)
	if !originAllowed {
		return
	}

	// Check if the method is allowed.
	if methodAllowed := cf.isMethodAllowed(r.Header.Get(AccessControlReqMethodHdr)); !methodAllowed {

		return
	}
	w.Header().Set(AllowOriginHeader, origin)
	w.Header().Set(AllowMethodsHeader, r.Header.Get(AccessControlReqMethodHdr))

	if len(cf.AllowedHeaders) > 0 {
		// Set the allowed headers
		w.Header().Set(AllowHeadersHeader, cf.accessControlReqHdrsStr)
	} else {
		// Set the allowed headers
		w.Header().Set(AllowHeadersHeader, r.Header.Get(AccessControlReqHeaders))
	}

	// Allow credentials
	if cf.AllowCredentials {
		w.Header().Set(AllowCredentials, trueStr)
	}

	// Set the max age
	if cf.MaxAge > 0 {
		w.Header().Set(MaxAgeHeader, strconv.Itoa(cf.MaxAge))
	}
	// Set the Private Network Header
	if cf.AllowPvtNetwork && r.Header.Get(AccessControlPvtNetworkHdr) == trueStr {
		w.Header().Set(AccessControlPvtNetworkHdr, trueStr)
	}

}

// HandleActualRequest handles the actual request
func (cf *CorsFilter) HandleActualRequest(w http.ResponseWriter, r *http.Request) {
	reqOrigin := r.Header.Get(OriginHeader)
	// Check if the origin is originAllowed
	originAllowed, origin := cf.isOriginAllowed(reqOrigin)
	if !originAllowed {
		return
	}
	// check if the method is allowed
	if methodAllowed := cf.isMethodAllowed(r.Method); !methodAllowed {
		return
	}
	w.Header().Set(AllowOriginHeader, origin)
	if cf.AllowCredentials {
		w.Header().Set(AllowCredentials, trueStr)
	}
	if len(cf.ExposeHeaders) > 0 {
		w.Header().Set(ExposeHeaders, strings.Join(cf.ExposeHeaders, ","))
	}

}

// CorsFilter handles the CORS headers
func (cf *CorsFilter) HandleCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(AccessControlReqHeaders) != textutils.EmptyStr && r.Method == http.MethodOptions {
			cf.handlePreflight(w, r)
			if cf.PreFlightPassThrough {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(cf.ResponseStatus)
			}
		} else {
			cf.HandleActualRequest(w, r)
			next.ServeHTTP(w, r)
		}
	})
}
