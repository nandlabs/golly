package authz

import (
	"testing"
)

// ---- RBAC ----

func TestRBAC_Grants(t *testing.T) {
	r := NewRBAC()
	r.AddRole("admin", "read", "write", "delete")
	r.AddRole("viewer", "read")

	admin := &BasicPrincipal{RoleList: []string{"admin"}}
	viewer := &BasicPrincipal{RoleList: []string{"viewer"}}
	multi := &BasicPrincipal{RoleList: []string{"viewer", "admin"}}

	cases := []struct {
		name   string
		p      Principal
		action string
		want   bool
	}{
		{"admin read", admin, "read", true},
		{"admin write", admin, "write", true},
		{"admin unknown", admin, "explode", false},
		{"viewer read", viewer, "read", true},
		{"viewer write", viewer, "write", false},
		{"multi-role admin grant", multi, "delete", true},
		{"nil principal", nil, "read", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := r.Can(c.p, c.action, nil); got != c.want {
				t.Errorf("Can = %v, want %v", got, c.want)
			}
		})
	}
}

func TestRBAC_RemoveRole(t *testing.T) {
	r := NewRBAC()
	r.AddRole("temp", "x")
	if !r.Can(&BasicPrincipal{RoleList: []string{"temp"}}, "x", nil) {
		t.Fatal("initial grant should work")
	}
	r.RemoveRole("temp")
	if r.Can(&BasicPrincipal{RoleList: []string{"temp"}}, "x", nil) {
		t.Error("removed role should no longer grant")
	}
}

// ---- CapabilityList ----

func TestCapabilityList(t *testing.T) {
	cl := CapabilityList{}
	p := &BasicPrincipal{CapList: []string{"posts:read", "posts:create"}}
	if !cl.Can(p, "posts:read", nil) {
		t.Error("posts:read should be granted")
	}
	if cl.Can(p, "posts:delete", nil) {
		t.Error("posts:delete should NOT be granted")
	}
	if cl.Can(nil, "anything", nil) {
		t.Error("nil principal should be denied")
	}
}

// ---- composers ----

func TestAll_RequiresEvery(t *testing.T) {
	yes := PolicyFunc(func(_ Principal, _ string, _ any) bool { return true })
	no := PolicyFunc(func(_ Principal, _ string, _ any) bool { return false })

	if !All(yes, yes, yes).Can(nil, "x", nil) {
		t.Error("All(yes...) should grant")
	}
	if All(yes, no, yes).Can(nil, "x", nil) {
		t.Error("All with any no should deny")
	}
	if All().Can(nil, "x", nil) {
		t.Error("All() (empty) should deny")
	}
}

func TestAny_OnlyNeedsOne(t *testing.T) {
	yes := PolicyFunc(func(_ Principal, _ string, _ any) bool { return true })
	no := PolicyFunc(func(_ Principal, _ string, _ any) bool { return false })

	if !Any(no, no, yes).Can(nil, "x", nil) {
		t.Error("Any with any yes should grant")
	}
	if Any(no, no).Can(nil, "x", nil) {
		t.Error("Any with all no should deny")
	}
	if Any().Can(nil, "x", nil) {
		t.Error("Any() (empty) should deny")
	}
}

func TestDeny_Inverts(t *testing.T) {
	yes := PolicyFunc(func(_ Principal, _ string, _ any) bool { return true })
	if Deny(yes).Can(nil, "x", nil) {
		t.Error("Deny(yes) should deny")
	}
	no := PolicyFunc(func(_ Principal, _ string, _ any) bool { return false })
	if !Deny(no).Can(nil, "x", nil) {
		t.Error("Deny(no) should grant")
	}
	// nil policy treated as "denies" → Deny flips it → grants.
	if !Deny(nil).Can(nil, "x", nil) {
		t.Error("Deny(nil) should grant (vacuous)")
	}
}

// ---- composition example: RBAC AND not-suspended ----

func TestComposition_RBACAndNotSuspended(t *testing.T) {
	rbac := NewRBAC()
	rbac.AddRole("editor", "post:create")

	// "suspended" policy returns true for principals with role "suspended".
	suspended := PolicyFunc(func(p Principal, _ string, _ any) bool {
		for _, r := range p.Roles() {
			if r == "suspended" {
				return true
			}
		}
		return false
	})

	policy := All(rbac, Deny(suspended))

	ok := &BasicPrincipal{RoleList: []string{"editor"}}
	if !policy.Can(ok, "post:create", nil) {
		t.Error("active editor should be allowed to post")
	}
	susp := &BasicPrincipal{RoleList: []string{"editor", "suspended"}}
	if policy.Can(susp, "post:create", nil) {
		t.Error("suspended editor should be denied")
	}
}

// ---- PolicyFunc.Can on nil receiver ----

func TestPolicyFunc_NilSafe(t *testing.T) {
	var f PolicyFunc
	if f.Can(nil, "x", nil) {
		t.Error("nil PolicyFunc should deny")
	}
}
