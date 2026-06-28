// Package migrate is a small file-based SQL migration runner. It targets
// any database/sql driver and uses one tracking table (default name
// "schema_migrations") to record applied versions.
//
// Migration files are named "<version>_<description>.sql" where <version>
// is a string-comparable identifier (zero-padded numbers like 0001 or
// timestamps like 20260101120000 both work — comparison is lexicographic).
// Each file is applied in a transaction; a failure rolls back and stops.
//
// Usage:
//
//	m, _ := migrate.NewFromDir(db, "./migrations")
//	if err := m.Up(ctx); err != nil { /* migration failed */ }
//
// Stdlib only — database/sql + os + path/filepath + sort + strings.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migration is one ordered SQL script.
type Migration struct {
	Version string
	Name    string
	SQL     string
}

// Migrator runs migrations against db.
type Migrator struct {
	db    *sql.DB
	mig   []Migration
	table string
}

// Option configures the Migrator.
type Option func(*Migrator)

// WithTable overrides the tracking table name (default: schema_migrations).
func WithTable(name string) Option {
	return func(m *Migrator) {
		if name != "" {
			m.table = name
		}
	}
}

// New returns a Migrator over the given (already-ordered) migrations.
func New(db *sql.DB, migrations []Migration, opts ...Option) *Migrator {
	m := &Migrator{db: db, mig: migrations, table: "schema_migrations"}
	for _, o := range opts {
		o(m)
	}
	return m
}

// NewFromDir loads "*.sql" files from dir, sorts them by filename, and
// returns a Migrator. The file basename (minus extension) is split on the
// first "_" into <version>_<name>; both are recorded.
func NewFromDir(db *sql.DB, dir string, opts ...Option) (*Migrator, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("migrate: read dir %q: %w", dir, err)
	}
	var mig []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("migrate: read %q: %w", path, err)
		}
		base := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		ver, name := splitVersionName(base)
		mig = append(mig, Migration{Version: ver, Name: name, SQL: string(data)})
	}
	sort.Slice(mig, func(i, j int) bool { return mig[i].Version < mig[j].Version })
	return New(db, mig, opts...), nil
}

func splitVersionName(base string) (version, name string) {
	if i := strings.IndexByte(base, '_'); i > 0 {
		return base[:i], base[i+1:]
	}
	return base, ""
}

// Up applies every migration whose version is not yet recorded in the
// tracking table, in ascending order. Each migration runs in its own
// transaction.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	applied, err := m.applied(ctx)
	if err != nil {
		return err
	}
	for _, mig := range m.mig {
		if applied[mig.Version] {
			continue
		}
		if err := m.applyOne(ctx, mig); err != nil {
			return err
		}
	}
	return nil
}

// Applied returns the set of versions already recorded in the tracking
// table. Useful for status reports.
func (m *Migrator) Applied(ctx context.Context) ([]string, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	set, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Strings(out)
	return out, nil
}

// Pending returns the versions known to the migrator but not yet applied.
func (m *Migrator) Pending(ctx context.Context) ([]Migration, error) {
	applied, err := m.applied(ctx)
	if err != nil {
		return nil, err
	}
	var out []Migration
	for _, mig := range m.mig {
		if !applied[mig.Version] {
			out = append(out, mig)
		}
	}
	return out, nil
}

func (m *Migrator) ensureTable(ctx context.Context) error {
	// The schema is intentionally simple and portable across SQLite/Postgres/
	// MySQL. Apply-time defaults: TEXT version, TEXT name, INTEGER applied_at
	// (unix seconds). We don't use CURRENT_TIMESTAMP because dialects diverge.
	q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		version TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at INTEGER NOT NULL
	)`, m.table)
	_, err := m.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("migrate: create %s: %w", m.table, err)
	}
	return nil
}

func (m *Migrator) applied(ctx context.Context) (map[string]bool, error) {
	rows, err := m.db.QueryContext(ctx, fmt.Sprintf("SELECT version FROM %s", m.table))
	if err != nil {
		return nil, fmt.Errorf("migrate: read %s: %w", m.table, err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func (m *Migrator) applyOne(ctx context.Context, mig Migration) error {
	if mig.SQL == "" {
		return errors.New("migrate: empty migration: " + mig.Version)
	}
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migrate: begin %s: %w", mig.Version, err)
	}
	if _, err := tx.ExecContext(ctx, mig.SQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("migrate: apply %s: %w", mig.Version, err)
	}
	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (version, name, applied_at) VALUES (?, ?, ?)", m.table),
		mig.Version, mig.Name, time.Now().Unix()); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("migrate: record %s: %w", mig.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migrate: commit %s: %w", mig.Version, err)
	}
	return nil
}
