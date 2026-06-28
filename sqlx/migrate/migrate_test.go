package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitVersionName(t *testing.T) {
	cases := []struct {
		base    string
		version string
		name    string
	}{
		{"0001_init", "0001", "init"},
		{"20260101_add_users", "20260101", "add_users"},
		{"single", "single", ""},
		{"v1.2.3_release", "v1.2.3", "release"},
	}
	for _, c := range cases {
		ver, name := splitVersionName(c.base)
		if ver != c.version || name != c.name {
			t.Errorf("splitVersionName(%q) = (%q, %q), want (%q, %q)",
				c.base, ver, name, c.version, c.name)
		}
	}
}

func TestNewFromDir_LoadsAndSorts(t *testing.T) {
	dir := t.TempDir()
	// Write files out of order — loader should sort them.
	files := map[string]string{
		"0002_add_table.sql": "CREATE TABLE x();",
		"0001_init.sql":      "CREATE TABLE u();",
		"0003_drop.sql":      "DROP TABLE x;",
		"README.md":          "ignored",
		"sub":                "", // skipped: directory check
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		if name == "sub" {
			continue
		}
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	m, err := NewFromDir(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.mig) != 3 {
		t.Fatalf("expected 3 migrations (excluding non-sql and dirs); got %d", len(m.mig))
	}
	wantVersions := []string{"0001", "0002", "0003"}
	for i, w := range wantVersions {
		if m.mig[i].Version != w {
			t.Errorf("mig[%d].Version = %q, want %q", i, m.mig[i].Version, w)
		}
	}
	if m.mig[0].Name != "init" {
		t.Errorf("mig[0].Name = %q, want init", m.mig[0].Name)
	}
}

func TestNewFromDir_BadPath(t *testing.T) {
	if _, err := NewFromDir(nil, "/definitely/does/not/exist"); err == nil {
		t.Error("expected error for missing dir")
	}
}

func TestWithTable_OverridesDefault(t *testing.T) {
	m := New(nil, nil, WithTable("my_tracker"))
	if m.table != "my_tracker" {
		t.Errorf("WithTable not applied; got %q", m.table)
	}
}

func TestWithTable_EmptyStringKeepsDefault(t *testing.T) {
	m := New(nil, nil, WithTable(""))
	if m.table != "schema_migrations" {
		t.Errorf("empty WithTable should keep default; got %q", m.table)
	}
}
