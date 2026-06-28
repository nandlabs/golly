package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type DBConf struct {
	DSN  string        `config:"dsn"`
	Wait time.Duration `config:"wait"`
}

type ServerConf struct {
	Addr  string   `config:"addr"`
	TLS   bool     `config:"tls"`
	Hosts []string `config:"hosts"`
	DB    DBConf   `config:"db"`
}

// ---- FromMap ----

func TestLoadInto_FromMap_Basic(t *testing.T) {
	src := FromMap(map[string]any{
		"addr":  ":8080",
		"tls":   true,
		"hosts": []any{"a", "b"},
		"db": map[string]any{
			"dsn":  "postgres://x",
			"wait": "1500ms",
		},
	})

	var cfg ServerConf
	if err := LoadInto(&cfg, src); err != nil {
		t.Fatalf("LoadInto: %v", err)
	}
	if cfg.Addr != ":8080" || !cfg.TLS {
		t.Errorf("scalar fields wrong: %+v", cfg)
	}
	if len(cfg.Hosts) != 2 || cfg.Hosts[0] != "a" {
		t.Errorf("slice wrong: %v", cfg.Hosts)
	}
	if cfg.DB.DSN != "postgres://x" {
		t.Errorf("nested DSN wrong: %q", cfg.DB.DSN)
	}
	if cfg.DB.Wait != 1500*time.Millisecond {
		t.Errorf("duration parsing wrong: %v", cfg.DB.Wait)
	}
}

// ---- layering: later wins ----

func TestLoadInto_Layered_LaterWins(t *testing.T) {
	defaults := FromMap(map[string]any{
		"addr": ":80",
		"db":   map[string]any{"dsn": "default-dsn"},
	})
	override := FromMap(map[string]any{
		"addr": ":9090",
	})

	var cfg ServerConf
	_ = LoadInto(&cfg, defaults, override)
	if cfg.Addr != ":9090" {
		t.Errorf("later source should win on addr; got %q", cfg.Addr)
	}
	if cfg.DB.DSN != "default-dsn" {
		t.Errorf("untouched nested key should remain from defaults; got %q", cfg.DB.DSN)
	}
}

// ---- FromFile YAML + JSON + .properties ----

func TestFromFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	body := `
addr: ":7777"
tls: true
hosts: [x, y, z]
db:
  dsn: "yaml-dsn"
  wait: "250ms"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	var cfg ServerConf
	if err := LoadInto(&cfg, FromFile(path)); err != nil {
		t.Fatalf("LoadInto: %v", err)
	}
	if cfg.Addr != ":7777" || cfg.DB.DSN != "yaml-dsn" {
		t.Errorf("yaml decode wrong: %+v", cfg)
	}
}

func TestFromFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	body := `{"addr":":6666","db":{"dsn":"json-dsn"}}`
	_ = os.WriteFile(path, []byte(body), 0o600)
	var cfg ServerConf
	if err := LoadInto(&cfg, FromFile(path)); err != nil {
		t.Fatalf("LoadInto: %v", err)
	}
	if cfg.Addr != ":6666" || cfg.DB.DSN != "json-dsn" {
		t.Errorf("json decode wrong: %+v", cfg)
	}
}

func TestFromFile_Properties(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.properties")
	body := "# a comment\naddr=:5555\ndb.dsn=props-dsn\n"
	_ = os.WriteFile(path, []byte(body), 0o600)
	var cfg ServerConf
	if err := LoadInto(&cfg, FromFile(path)); err != nil {
		t.Fatalf("LoadInto: %v", err)
	}
	if cfg.Addr != ":5555" || cfg.DB.DSN != "props-dsn" {
		t.Errorf("properties decode wrong: %+v", cfg)
	}
}

// ---- FromEnv ----

func TestFromEnv_PrefixAndUnderscores(t *testing.T) {
	t.Setenv("MYAPP_ADDR", ":4444")
	t.Setenv("MYAPP_DB_DSN", "env-dsn") // → db.dsn
	t.Setenv("MYAPP_DB__WAIT", "200ms") // double-underscore → literal _, key db.wait? no — single _ remains separator unless doubled
	t.Setenv("OTHER_IGNORED", "leave-alone")

	var cfg ServerConf
	if err := LoadInto(&cfg, FromEnv("MYAPP_")); err != nil {
		t.Fatalf("LoadInto: %v", err)
	}
	if cfg.Addr != ":4444" {
		t.Errorf("addr from env wrong: %q", cfg.Addr)
	}
	if cfg.DB.DSN != "env-dsn" {
		t.Errorf("db.dsn from env wrong: %q", cfg.DB.DSN)
	}
}

// ---- coercions ----

func TestCoerce_StringToBool(t *testing.T) {
	type T struct{ B bool }
	var cfg T
	_ = LoadInto(&cfg, FromMap(map[string]any{"b": "true"}))
	if !cfg.B {
		t.Errorf("expected bool true from string 'true'")
	}
}

func TestCoerce_StringToInt(t *testing.T) {
	type T struct {
		N int    `config:"n"`
		U uint64 `config:"u"`
	}
	var cfg T
	_ = LoadInto(&cfg, FromMap(map[string]any{"n": "42", "u": "100"}))
	if cfg.N != 42 || cfg.U != 100 {
		t.Errorf("numeric coercion wrong: %+v", cfg)
	}
}

func TestCoerce_CSVStringToSlice(t *testing.T) {
	type T struct {
		Tags []string `config:"tags"`
	}
	var cfg T
	_ = LoadInto(&cfg, FromMap(map[string]any{"tags": "a, b ,c"}))
	if len(cfg.Tags) != 3 || cfg.Tags[1] != "b" {
		t.Errorf("csv slice wrong: %v", cfg.Tags)
	}
}

// ---- errors ----

func TestLoadInto_NilDst(t *testing.T) {
	if err := LoadInto(nil); err == nil {
		t.Error("expected error for nil dst")
	}
}

func TestLoadInto_NonPointerDst(t *testing.T) {
	var cfg ServerConf
	if err := LoadInto(cfg); err == nil {
		t.Error("expected error for non-pointer dst")
	}
}

func TestLoadInto_PointerToNonStruct(t *testing.T) {
	x := 5
	if err := LoadInto(&x); err == nil {
		t.Error("expected error for pointer to non-struct")
	}
}

func TestFromFile_Missing(t *testing.T) {
	var cfg ServerConf
	if err := LoadInto(&cfg, FromFile("/no/such/file")); err == nil {
		t.Error("expected error for missing file")
	}
}

// ---- env-name → key transformation unit test ----

func TestEnvNameToKey(t *testing.T) {
	cases := []struct{ in, want string }{
		{"DB_DSN", "db.dsn"},
		{"DB_HOST_NAME", "db.host.name"},
		{"DB__DSN", "db_dsn"},         // doubled → literal underscore
		{"PRE__NAME_X", "pre_name.x"}, // doubled then split
	}
	for _, c := range cases {
		if got := envNameToKey(c.in); got != c.want {
			t.Errorf("envNameToKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ---- mergeMaps recursion ----

func TestMergeMaps_DeepMerge(t *testing.T) {
	dst := map[string]any{
		"a": 1,
		"sub": map[string]any{
			"x": "from-dst",
			"y": "keep",
		},
	}
	src := map[string]any{
		"a": 2,
		"sub": map[string]any{
			"x": "from-src",
			"z": "new",
		},
	}
	mergeMaps(dst, src)
	if dst["a"] != 2 {
		t.Errorf("scalar should be replaced; got %v", dst["a"])
	}
	sub := dst["sub"].(map[string]any)
	if sub["x"] != "from-src" || sub["y"] != "keep" || sub["z"] != "new" {
		t.Errorf("nested merge wrong: %+v", sub)
	}
}
