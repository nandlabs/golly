// Package authz provides a small, datastore-agnostic authorization model.
// Three primitives:
//
//   - Principal — who is acting (with roles and capabilities)
//   - Policy    — pure-function authorization check Can(principal, action, resource)
//   - composers — All / Any / Deny to combine multiple policies
//
// Two built-in Policy implementations: RBAC (roles → capabilities) and
// CapabilityList (direct capability allowlist on the principal).
//
// Stdlib only. Persistence and principal sourcing are intentionally out of
// scope — this package is purely the decision model.
package authz

import "sync"

// Principal identifies the actor making a request. Roles and Capabilities
// are sets of opaque strings — the authz package never interprets them.
type Principal interface {
	Roles() []string
	Capabilities() []string
}

// Policy is a single authorization decision function.
//
//	resource is whatever the application passes — a user id, a struct, a
//	"resource:type/id" string, etc. Policies decide what to make of it.
type Policy interface {
	Can(p Principal, action string, resource any) bool
}

// PolicyFunc adapts a plain function to the Policy interface.
type PolicyFunc func(p Principal, action string, resource any) bool

// Can satisfies the Policy interface.
func (f PolicyFunc) Can(p Principal, action string, resource any) bool {
	if f == nil {
		return false
	}
	return f(p, action, resource)
}

// --- BasicPrincipal: a trivial implementation for tests / simple apps ---

// BasicPrincipal is a Principal whose roles and capabilities are static
// slices. The slices are returned as-is — callers must not mutate them
// concurrently with calls into Roles() / Capabilities().
type BasicPrincipal struct {
	RoleList []string
	CapList  []string
}

// Roles returns the static role list.
func (b *BasicPrincipal) Roles() []string {
	if b == nil {
		return nil
	}
	return b.RoleList
}

// Capabilities returns the static capability list.
func (b *BasicPrincipal) Capabilities() []string {
	if b == nil {
		return nil
	}
	return b.CapList
}

// --- RBAC: role → capabilities ---

// RBAC implements Policy by mapping a Principal's Roles() through a
// role→capability table and checking whether the requested action is
// present in any role's capability set. Resource is ignored — this is a
// coarse role-based check; pair with a CapabilityList or custom Policy
// composed via All for resource-level scoping.
//
// Safe for concurrent reads after construction. Use AddRole to mutate.
type RBAC struct {
	mu    sync.RWMutex
	roles map[string]map[string]struct{} // role -> capabilities set
}

// NewRBAC returns an empty RBAC policy.
func NewRBAC() *RBAC {
	return &RBAC{roles: map[string]map[string]struct{}{}}
}

// AddRole grants the given capabilities to role (idempotent).
func (r *RBAC) AddRole(role string, capabilities ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	set, ok := r.roles[role]
	if !ok {
		set = make(map[string]struct{}, len(capabilities))
		r.roles[role] = set
	}
	for _, c := range capabilities {
		set[c] = struct{}{}
	}
}

// RemoveRole drops the role and all its grants.
func (r *RBAC) RemoveRole(role string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.roles, role)
}

// Can returns true iff p has at least one role that grants action.
func (r *RBAC) Can(p Principal, action string, _ any) bool {
	if p == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, role := range p.Roles() {
		if caps, ok := r.roles[role]; ok {
			if _, granted := caps[action]; granted {
				return true
			}
		}
	}
	return false
}

// --- CapabilityList: direct capability allowlist ---

// CapabilityList implements Policy by checking the Principal's own
// Capabilities() set for action. Useful when capabilities are issued per-user
// (e.g. carried in a JWT scope claim) without a role table.
type CapabilityList struct{}

// Can returns true iff action appears in p.Capabilities().
func (c CapabilityList) Can(p Principal, action string, _ any) bool {
	if p == nil {
		return false
	}
	for _, cap := range p.Capabilities() {
		if cap == action {
			return true
		}
	}
	return false
}

// --- composers ---

// All returns a Policy that grants iff every supplied policy grants.
// Empty input returns a Policy that always denies (vacuously, to avoid
// surprises — explicit caller intent is safer than vacuous truth here).
func All(policies ...Policy) Policy {
	if len(policies) == 0 {
		return PolicyFunc(func(_ Principal, _ string, _ any) bool { return false })
	}
	return PolicyFunc(func(p Principal, action string, resource any) bool {
		for _, pol := range policies {
			if !pol.Can(p, action, resource) {
				return false
			}
		}
		return true
	})
}

// Any returns a Policy that grants iff at least one supplied policy grants.
// Empty input returns a Policy that always denies.
func Any(policies ...Policy) Policy {
	if len(policies) == 0 {
		return PolicyFunc(func(_ Principal, _ string, _ any) bool { return false })
	}
	return PolicyFunc(func(p Principal, action string, resource any) bool {
		for _, pol := range policies {
			if pol.Can(p, action, resource) {
				return true
			}
		}
		return false
	})
}

// Deny inverts a policy: it grants iff the wrapped policy denies. Useful for
// "everyone except these" patterns.
//
//	authz.All(rbac, authz.Deny(suspended))
//
// — suspended is a custom Policy that says "yes" for suspended users.
func Deny(policy Policy) Policy {
	return PolicyFunc(func(p Principal, action string, resource any) bool {
		if policy == nil {
			return true
		}
		return !policy.Can(p, action, resource)
	})
}
