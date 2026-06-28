package sqlx

import (
	"reflect"
	"testing"
)

type User struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string // no tag → falls back to lowercased field name "email"
	Skip  string `db:"-"`
}

func TestTagIndex(t *testing.T) {
	idx := tagIndex(reflect.TypeOf(User{}))
	wantKeys := []string{"id", "name", "email"}
	for _, k := range wantKeys {
		if _, ok := idx[k]; !ok {
			t.Errorf("missing key %q in index: %+v", k, idx)
		}
	}
	if _, ok := idx["Skip"]; ok {
		t.Error("db:'-' field should be excluded")
	}
	if _, ok := idx["-"]; ok {
		t.Error("'-' should not appear as a key")
	}
}

func TestNamed_QuestionMarkDialect(t *testing.T) {
	q, args, err := Named(
		"SELECT * FROM u WHERE org=:org AND age>:age",
		map[string]any{"org": "acme", "age": 18},
		"?",
	)
	if err != nil {
		t.Fatal(err)
	}
	if q != "SELECT * FROM u WHERE org=? AND age>?" {
		t.Errorf("rewritten = %q", q)
	}
	if len(args) != 2 || args[0] != "acme" || args[1] != 18 {
		t.Errorf("args = %v", args)
	}
}

func TestNamed_DollarDialect(t *testing.T) {
	q, args, err := Named(
		"INSERT INTO t (a,b) VALUES (:a, :b)",
		map[string]any{"a": 1, "b": 2},
		"$",
	)
	if err != nil {
		t.Fatal(err)
	}
	if q != "INSERT INTO t (a,b) VALUES ($1, $2)" {
		t.Errorf("rewritten = %q", q)
	}
	if len(args) != 2 || args[0] != 1 || args[1] != 2 {
		t.Errorf("args = %v", args)
	}
}

func TestNamed_DefaultDialect(t *testing.T) {
	q, _, _ := Named("WHERE x=:x", map[string]any{"x": 1}, "")
	if q != "WHERE x=?" {
		t.Errorf("default dialect should be ?; got %q", q)
	}
}

func TestNamed_MissingParam(t *testing.T) {
	if _, _, err := Named("WHERE a=:missing", map[string]any{}, "?"); err == nil {
		t.Error("expected error for missing param")
	}
}

func TestNamed_UnsupportedDialect(t *testing.T) {
	if _, _, err := Named("a=:x", map[string]any{"x": 1}, "@"); err == nil {
		t.Error("expected error for unsupported dialect")
	}
}

func TestNamed_IgnoresColonsInsideStringLiterals(t *testing.T) {
	q, args, err := Named(
		"SELECT 'time::stamp' FROM t WHERE x=:x",
		map[string]any{"x": 1},
		"?",
	)
	if err != nil {
		t.Fatal(err)
	}
	if q != "SELECT 'time::stamp' FROM t WHERE x=?" {
		t.Errorf("string literal protection failed: %q", q)
	}
	if len(args) != 1 {
		t.Errorf("args wrong: %v", args)
	}
}

func TestNamed_ColonNotFollowedByIdent_LeftAlone(t *testing.T) {
	// Naked colon (e.g. type cast in postgres) shouldn't be treated as a placeholder.
	q, _, err := Named("WHERE x::int = 5", nil, "?")
	if err != nil {
		t.Fatal(err)
	}
	if q != "WHERE x::int = 5" {
		t.Errorf("naked colons should pass through; got %q", q)
	}
}
