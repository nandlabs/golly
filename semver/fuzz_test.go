package semver

import (
	"testing"
)

// FuzzParse asserts Parse never panics on arbitrary input. Malformed input
// must return an error rather than crash; well-formed input must round-trip
// back to a string that re-parses cleanly.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		"0.0.0",
		"1.2.3",
		"v1.2.3",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-0.3.7",
		"1.0.0-x.7.z.92",
		"1.0.0+20130313144700",
		"1.0.0-beta+exp.sha.5114f85",
		"1.0.0+21AF26D3----117B344092BD",
		"99999999999999999999999.999999999999999999.99999999999999999",
		"1",
		"1.2",
		"1.2.3.4",
		"a.b.c",
		"-1.-2.-3",
		"1.2.3-+",
		"1.2.3-",
		"1.2.3+",
		"1.0.0-α",
		"  1.2.3  ",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := Parse(s)
		if err != nil {
			return // invalid input — must not panic, but failure is OK
		}
		// Round-trip: a successfully parsed version should re-print to a
		// non-empty string that also parses successfully.
		out := v.String()
		if out == "" {
			t.Fatalf("Parse(%q) succeeded but String() empty", s)
		}
		if _, err := Parse(out); err != nil {
			t.Fatalf("round-trip failed: Parse(%q) -> %q -> Parse error: %v", s, out, err)
		}
	})
}

// FuzzCompareRaw asserts CompareRaw never panics on arbitrary pair input.
func FuzzCompareRaw(f *testing.F) {
	seeds := [][2]string{
		{"1.0.0", "1.0.0"},
		{"1.0.0", "2.0.0"},
		{"1.0.0-alpha", "1.0.0"},
		{"", ""},
		{"v1.2.3", "1.2.3"},
		{"1.2.3-alpha+build", "1.2.3-alpha"},
	}
	for _, p := range seeds {
		f.Add(p[0], p[1])
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		_, _ = CompareRaw(a, b) // success or error are both acceptable; must not panic
	})
}
