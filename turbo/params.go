package turbo

import (
	"fmt"
	"net/http"
	"strconv"
)

// This file provides the recommended param-accessor API.
//
// Background: the original Get*Param[As*] functions return (value, error)
// for both path and query parameters. For query parameters that's the
// wrong shape — a missing optional query param is the common case, not
// an error condition (see issue #87, which was patched in-place to
// stop GetQueryParam erroring on missing).
//
// These new accessors give a cleaner, signature-correct API that callers
// should prefer in new code:
//
//   - Query(r, id) string                — empty when absent
//   - QueryInt/Float/Bool(r, id) (T, ok) — ok reports presence AND
//                                          successful parse, so callers
//                                          can distinguish "not supplied"
//                                          from a zero value
//   - RequireQuery(r, id) (string, error)  — error iff absent or empty
//   - RequireQueryInt/Float/Bool           — error iff absent / invalid
//
// PathParam accessors are deliberately unchanged: a missing path
// parameter is a router bug (the route pattern requires it), not user
// input, so an error return is appropriate.

// Query returns the named query parameter, or "" if absent. Never
// returns an error.
func Query(r *http.Request, id string) string {
	return r.URL.Query().Get(id)
}

// QueryInt returns the named query parameter parsed as int. The bool is
// true iff the parameter was present AND parsed successfully; callers
// can distinguish "not supplied" from "0" by inspecting it.
func QueryInt(r *http.Request, id string) (int, bool) {
	s := r.URL.Query().Get(id)
	if s == "" {
		return 0, false
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

// QueryFloat returns the named query parameter parsed as float64. The
// bool follows the same semantics as QueryInt.
func QueryFloat(r *http.Request, id string) (float64, bool) {
	s := r.URL.Query().Get(id)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// QueryBool returns the named query parameter parsed as bool. Uses
// strconv.ParseBool semantics ("1"/"t"/"T"/"true"/"TRUE"/"True" are
// true; "0"/"f"/"F"/"false"/"FALSE"/"False" are false). The bool
// follows the same semantics as QueryInt.
func QueryBool(r *http.Request, id string) (bool, bool) {
	s := r.URL.Query().Get(id)
	if s == "" {
		return false, false
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, false
	}
	return v, true
}

// RequireQuery returns the named query parameter or an error if it is
// missing / empty. Use for parameters the handler genuinely cannot
// proceed without; pair with a 400 response in the caller.
func RequireQuery(r *http.Request, id string) (string, error) {
	v := r.URL.Query().Get(id)
	if v == "" {
		return "", fmt.Errorf("required query parameter %q is missing", id)
	}
	return v, nil
}

// RequireQueryInt returns the int-typed query parameter or an error if
// it is missing or not a valid int.
func RequireQueryInt(r *http.Request, id string) (int, error) {
	s, err := RequireQuery(r, id)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("query parameter %q is not a valid int: %w", id, err)
	}
	return v, nil
}

// RequireQueryFloat returns the float64-typed query parameter or an
// error if it is missing or not a valid float.
func RequireQueryFloat(r *http.Request, id string) (float64, error) {
	s, err := RequireQuery(r, id)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("query parameter %q is not a valid float: %w", id, err)
	}
	return v, nil
}

// RequireQueryBool returns the bool-typed query parameter or an error
// if it is missing or not a valid bool literal.
func RequireQueryBool(r *http.Request, id string) (bool, error) {
	s, err := RequireQuery(r, id)
	if err != nil {
		return false, err
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("query parameter %q is not a valid bool: %w", id, err)
	}
	return v, nil
}
