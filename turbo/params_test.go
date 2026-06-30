package turbo

import (
	"net/http"
	"testing"
)

func mkReq(t *testing.T, raw string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return r
}

// --- Query (always nil-error variant) ---

func TestQuery_PresentAndMissing(t *testing.T) {
	r := mkReq(t, "/?a=hello&empty=")
	if got := Query(r, "a"); got != "hello" {
		t.Errorf("Query(a) = %q, want hello", got)
	}
	if got := Query(r, "missing"); got != "" {
		t.Errorf("Query(missing) = %q, want empty", got)
	}
	if got := Query(r, "empty"); got != "" {
		t.Errorf("Query(empty) = %q, want empty", got)
	}
}

// --- (value, bool) typed variants ---

func TestQueryInt_DistinguishesMissingFromZero(t *testing.T) {
	r := mkReq(t, "/?n=0&bad=abc")
	if v, ok := QueryInt(r, "n"); !ok || v != 0 {
		t.Errorf("QueryInt(n) = (%d, %v), want (0, true)", v, ok)
	}
	if _, ok := QueryInt(r, "absent"); ok {
		t.Errorf("QueryInt(absent) should report ok=false")
	}
	if _, ok := QueryInt(r, "bad"); ok {
		t.Errorf("QueryInt(bad) should report ok=false on parse failure")
	}
}

func TestQueryFloat_PresentMissingInvalid(t *testing.T) {
	r := mkReq(t, "/?f=3.14&bad=NaNNaN")
	if v, ok := QueryFloat(r, "f"); !ok || v != 3.14 {
		t.Errorf("QueryFloat(f) = (%v, %v), want (3.14, true)", v, ok)
	}
	if _, ok := QueryFloat(r, "absent"); ok {
		t.Errorf("absent should be ok=false")
	}
	if _, ok := QueryFloat(r, "bad"); ok {
		t.Errorf("invalid should be ok=false")
	}
}

func TestQueryBool_PresentMissingInvalid(t *testing.T) {
	r := mkReq(t, "/?t=true&f=false&bad=maybe")
	for k, want := range map[string]bool{"t": true, "f": false} {
		v, ok := QueryBool(r, k)
		if !ok || v != want {
			t.Errorf("QueryBool(%q) = (%v, %v), want (%v, true)", k, v, ok, want)
		}
	}
	if _, ok := QueryBool(r, "absent"); ok {
		t.Errorf("absent should be ok=false")
	}
	if _, ok := QueryBool(r, "bad"); ok {
		t.Errorf("invalid should be ok=false")
	}
}

// --- Require* variants ---

func TestRequireQuery_PresentReturnsValue(t *testing.T) {
	r := mkReq(t, "/?name=alice")
	v, err := RequireQuery(r, "name")
	if err != nil || v != "alice" {
		t.Errorf("RequireQuery(name) = (%q, %v), want (alice, nil)", v, err)
	}
}

func TestRequireQuery_MissingErrors(t *testing.T) {
	r := mkReq(t, "/?")
	if _, err := RequireQuery(r, "missing"); err == nil {
		t.Error("expected error for missing required param")
	}
}

func TestRequireQueryInt_PresentInvalidMissing(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr bool
		wantVal int
	}{
		{"present-valid", "/?n=42", false, 42},
		{"present-invalid", "/?n=abc", true, 0},
		{"missing", "/?", true, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := mkReq(t, c.url)
			v, err := RequireQueryInt(r, "n")
			if c.wantErr && err == nil {
				t.Errorf("want error, got nil (v=%d)", v)
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !c.wantErr && v != c.wantVal {
				t.Errorf("v = %d, want %d", v, c.wantVal)
			}
		})
	}
}

// Sanity: the legacy GetQueryParam contract (returns nil error for
// missing) still holds — we don't want to regress the #87 fix.
func TestLegacyGetQueryParam_MissingNoError(t *testing.T) {
	r := mkReq(t, "/?")
	v, err := GetQueryParam("missing", r)
	if err != nil {
		t.Errorf("legacy GetQueryParam should not error on missing param; got %v", err)
	}
	if v != "" {
		t.Errorf("legacy GetQueryParam(missing) = %q, want empty", v)
	}
}
