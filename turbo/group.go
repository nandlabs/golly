package turbo

import (
	"net/http"
	"strings"

	"oss.nandlabs.io/golly/turbo/auth"
)

// Group is a routing prefix with its own stack of filters / authenticator.
// Groups make it easy to attach middleware to a whole subtree of routes
// (e.g. "/api/v1" with JWT auth) without repeating boilerplate on every
// route registration.
//
// Construct via Router.Group; nested groups are allowed and compose
// prefixes + filters in declaration order.
type Group struct {
	router  *Router
	parent  *Group
	prefix  string
	filters []FilterFunc
	auth    auth.Authenticator
}

// Group returns a child Group rooted at prefix. The prefix is concatenated
// with any parent prefix; filters and authenticators added on the parent
// are inherited.
//
//	api := router.Group("/api/v1")
//	api.Use(myLogger)
//	api.AddAuthenticator(jwtAuth)
//	api.Get("/users/:id", listUser)         // → GET /api/v1/users/:id
//	admin := api.Group("/admin")
//	admin.Get("/audit", listAudit)          // → GET /api/v1/admin/audit (still authed)
func (router *Router) Group(prefix string) *Group {
	return &Group{router: router, prefix: normalizeGroupPrefix(prefix)}
}

// Group on a Group returns a nested Group whose prefix and filters compose
// with the parent's.
func (g *Group) Group(prefix string) *Group {
	return &Group{router: g.router, parent: g, prefix: normalizeGroupPrefix(prefix)}
}

// Use appends one or more filters that will be applied to every route
// registered *after* this call on this Group (and any nested children).
// Returns the Group for chaining.
func (g *Group) Use(filters ...FilterFunc) *Group {
	g.filters = append(g.filters, filters...)
	return g
}

// AddAuthenticator sets the authenticator for every route registered on
// this Group (and nested children that don't override it).
func (g *Group) AddAuthenticator(a auth.Authenticator) *Group {
	g.auth = a
	return g
}

// Get / Post / Put / Delete register handlers with the group's prefix
// applied; the group's filters + authenticator are attached automatically.
func (g *Group) Get(path string, h func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return g.add(path, h, GET)
}
func (g *Group) Post(path string, h func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return g.add(path, h, POST)
}
func (g *Group) Put(path string, h func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return g.add(path, h, PUT)
}
func (g *Group) Delete(path string, h func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return g.add(path, h, DELETE)
}

// Patch registers a PATCH handler under the group's prefix.
func (g *Group) Patch(path string, h func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return g.add(path, h, PATCH)
}

// Head registers a HEAD handler under the group's prefix.
func (g *Group) Head(path string, h func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return g.add(path, h, HEAD)
}

// Options registers an OPTIONS handler under the group's prefix.
func (g *Group) Options(path string, h func(w http.ResponseWriter, r *http.Request)) (*Route, error) {
	return g.add(path, h, OPTIONS)
}

// Add registers an arbitrary method handler (or methods) on this Group.
func (g *Group) Add(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) (*Route, error) {
	return g.add(path, h, methods...)
}

func (g *Group) add(path string, h func(w http.ResponseWriter, r *http.Request), methods ...string) (*Route, error) {
	full := g.fullPath(path)
	route, err := g.router.Add(full, h, methods...)
	if err != nil {
		return nil, err
	}
	// Apply collected filters in chain order (root → leaf).
	for _, f := range g.collectFilters() {
		route.AddFilter(f)
	}
	if a := g.effectiveAuth(); a != nil {
		route.AddAuthenticator(a)
	}
	return route, nil
}

// fullPath concatenates all parent group prefixes with this group's prefix
// and the per-route path.
func (g *Group) fullPath(path string) string {
	parts := []string{}
	for cur := g; cur != nil; cur = cur.parent {
		if cur.prefix != "" {
			parts = append([]string{cur.prefix}, parts...)
		}
	}
	parts = append(parts, normalizeRoutePath(path))
	joined := strings.Join(parts, "")
	if joined == "" {
		joined = "/"
	}
	return joined
}

// collectFilters returns the parent-then-self filter list.
func (g *Group) collectFilters() []FilterFunc {
	if g.parent == nil {
		return g.filters
	}
	pf := g.parent.collectFilters()
	out := make([]FilterFunc, 0, len(pf)+len(g.filters))
	out = append(out, pf...)
	out = append(out, g.filters...)
	return out
}

// effectiveAuth returns this Group's authenticator if set, else the
// nearest ancestor's.
func (g *Group) effectiveAuth() auth.Authenticator {
	if g.auth != nil {
		return g.auth
	}
	if g.parent != nil {
		return g.parent.effectiveAuth()
	}
	return nil
}

// normalizeGroupPrefix ensures the prefix starts with "/" and does not end
// with "/" (so concatenation with "/path" produces "/prefix/path").
func normalizeGroupPrefix(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.TrimRight(p, "/")
	return p
}

// normalizeRoutePath ensures the path starts with "/".
func normalizeRoutePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
