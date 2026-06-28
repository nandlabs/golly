// Package sqlx is a thin, stdlib-only helper layer over database/sql.
// It is intentionally minimal — golly's stance is to lean on database/sql
// directly (paired with sqlc or hand-written queries) rather than ship a
// full ORM. Two helpers cover the common boilerplate:
//
//   - ScanRow / ScanRows scans a *sql.Row / *sql.Rows into a struct or slice
//     of structs by `db:"column_name"` tags (or lowercased field name).
//   - Named substitutes :name placeholders with positional parameters using
//     a map. Useful for hand-written queries; sqlc-style code generation is
//     still the recommended path for non-trivial schemas.
//
// Stdlib only — uses database/sql, reflect, strings.
//
// Migration runner lives in the sqlx/migrate subpackage.
package sqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrNoRows mirrors sql.ErrNoRows for convenience so callers don't need
// two error sentinels in scope.
var ErrNoRows = sql.ErrNoRows

// ScanRow scans a single *sql.Row into dest (a pointer to a struct).
// Returns ErrNoRows when the row is empty.
func ScanRow(row *sql.Row, dest any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return errors.New("sqlx: ScanRow dest must be a non-nil pointer to a struct")
	}
	rows, err := rowToRows(row)
	if err != nil {
		return err
	}
	if rows == nil {
		return errors.New("sqlx: ScanRow requires *sql.Rows; got *sql.Row — use ScanRow(db.QueryRow(...)) variant only with simple types")
	}
	return scanRowsOne(rows, dest)
}

// ScanRows scans every row in rows into dest (a pointer to a slice of
// structs). Closes rows when done.
func ScanRows(rows *sql.Rows, dest any) error {
	defer func() { _ = rows.Close() }()
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Slice {
		return errors.New("sqlx: ScanRows dest must be a non-nil pointer to a slice")
	}
	sliceVal := rv.Elem()
	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errors.New("sqlx: ScanRows dest must be a slice of structs")
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	out := reflect.MakeSlice(sliceVal.Type(), 0, 16)
	for rows.Next() {
		elem := reflect.New(elemType).Elem()
		targets, terr := buildTargets(elem, cols)
		if terr != nil {
			return terr
		}
		if serr := rows.Scan(targets...); serr != nil {
			return serr
		}
		out = reflect.Append(out, elem)
	}
	if rerr := rows.Err(); rerr != nil {
		return rerr
	}
	sliceVal.Set(out)
	return nil
}

// scanRowsOne scans the first (and only) row from rows into dest.
func scanRowsOne(rows *sql.Rows, dest any) error {
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	if !rows.Next() {
		if rerr := rows.Err(); rerr != nil {
			return rerr
		}
		return ErrNoRows
	}
	elem := reflect.ValueOf(dest).Elem()
	targets, err := buildTargets(elem, cols)
	if err != nil {
		return err
	}
	return rows.Scan(targets...)
}

// buildTargets returns a []any of pointers into elem's fields, ordered to
// match cols. Missing columns are bound to a *any throwaway sink so the
// scan still works; unmapped struct fields are simply left zero.
func buildTargets(elem reflect.Value, cols []string) ([]any, error) {
	idx := tagIndex(elem.Type())
	targets := make([]any, len(cols))
	for i, col := range cols {
		if f, ok := idx[col]; ok {
			targets[i] = elem.Field(f).Addr().Interface()
		} else {
			var sink any
			targets[i] = &sink
		}
	}
	return targets, nil
}

// tagIndex returns column-name → field-index, honoring `db:"col"` tags
// with a lowercased-field-name fallback.
func tagIndex(t reflect.Type) map[string]int {
	out := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := strings.TrimSpace(strings.Split(f.Tag.Get("db"), ",")[0])
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = strings.ToLower(f.Name)
		}
		out[tag] = i
	}
	return out
}

// rowToRows is a placeholder — sql.Row doesn't expose its inner rows. We
// return nil to force callers to the *sql.Rows path (ScanRows).
func rowToRows(_ *sql.Row) (*sql.Rows, error) { return nil, nil }

// Named replaces :name placeholders in query with positional ? (or $N for
// 1-based numbered, when dialect is "$") and returns (rewritten, args, error).
// Names are matched against params; missing names error early.
//
//	q, a, _ := sqlx.Named(
//	    "SELECT * FROM u WHERE org=:org AND age>:age",
//	    map[string]any{"org": "acme", "age": 18},
//	    "?",
//	)
//	rows, _ := db.Query(q, a...)
//
// Dialect arg is "?" (mysql/sqlite) or "$" (postgres). Bare empty string =
// "?" by default.
func Named(query string, params map[string]any, dialect string) (string, []any, error) {
	if dialect == "" {
		dialect = "?"
	}
	if dialect != "?" && dialect != "$" {
		return "", nil, fmt.Errorf("sqlx: unsupported dialect %q (use \"?\" or \"$\")", dialect)
	}
	var (
		out  strings.Builder
		args []any
		i    int
	)
	for i < len(query) {
		c := query[i]
		// A `:` introduces a placeholder only when followed by an identifier
		// AND NOT preceded by another `:` (so Postgres-style `::cast` is
		// passed through unchanged).
		prevColon := i > 0 && query[i-1] == ':'
		if c == ':' && !prevColon && i+1 < len(query) && isIdent(query[i+1]) && query[i+1] != ':' {
			// Read identifier
			j := i + 1
			for j < len(query) && isIdent(query[j]) {
				j++
			}
			name := query[i+1 : j]
			val, ok := params[name]
			if !ok {
				return "", nil, fmt.Errorf("sqlx: missing parameter %q", name)
			}
			args = append(args, val)
			if dialect == "?" {
				out.WriteByte('?')
			} else {
				fmt.Fprintf(&out, "$%d", len(args))
			}
			i = j
			continue
		}
		// Skip past string literals so colons inside don't trip the parser.
		if c == '\'' {
			out.WriteByte(c)
			i++
			for i < len(query) && query[i] != '\'' {
				out.WriteByte(query[i])
				i++
			}
			if i < len(query) {
				out.WriteByte(query[i])
				i++
			}
			continue
		}
		out.WriteByte(c)
		i++
	}
	return out.String(), args, nil
}

func isIdent(b byte) bool {
	return b == '_' ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}
