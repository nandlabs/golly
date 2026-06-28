package l3

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"
)

// Field is one key-value pair attached to a structured log record.
type Field struct {
	Key   string
	Value any
}

// F is a shorthand constructor for a Field, useful in With() and *KV calls:
//
//	logger.With(l3.F("user_id", "42")).InfoKV("login ok")
func F(k string, v any) Field {
	return Field{Key: k, Value: v}
}

// FieldLogger extends Logger with structured (key-value) methods and a
// builder pattern. The classic Info / Warn / Error / etc. positional methods
// remain available via embedding so existing call sites are untouched.
type FieldLogger interface {
	Logger

	// With returns a derived logger that always includes the given fields.
	// The receiver is unchanged — FieldLoggers are immutable values, safe
	// to share across goroutines.
	With(fields ...Field) FieldLogger

	// Ctx returns a derived logger that, on each emit, pulls any keys
	// registered via RegisterContextKey from ctx into the field bag.
	Ctx(ctx context.Context) FieldLogger

	// Structured emit methods. kv must be an even-length sequence of
	// alternating string keys and any values; an odd trailing entry is
	// recorded as the value of an "MISSING_KEY" key so nothing is dropped.
	ErrorKV(msg string, kv ...any)
	WarnKV(msg string, kv ...any)
	InfoKV(msg string, kv ...any)
	DebugKV(msg string, kv ...any)
	TraceKV(msg string, kv ...any)

	// Fields returns a snapshot of the bag currently attached to this logger.
	Fields() []Field
}

// fieldLogger is the immutable FieldLogger implementation.
type fieldLogger struct {
	base   Logger
	fields []Field
	ctx    context.Context
}

// NewFieldLogger wraps an existing Logger with optional starting fields.
// The base Logger handles actual emission; this wrapper just bundles fields
// into the message string (so any existing writer works) and, when paired
// with NewJSONHandler, produces structured JSON output.
func NewFieldLogger(base Logger, fields ...Field) FieldLogger {
	if base == nil {
		base = Get()
	}
	// Defensive copy so callers can mutate their slice afterwards.
	cp := make([]Field, len(fields))
	copy(cp, fields)
	return &fieldLogger{base: base, fields: cp}
}

func (f *fieldLogger) With(fields ...Field) FieldLogger {
	merged := make([]Field, 0, len(f.fields)+len(fields))
	merged = append(merged, f.fields...)
	merged = append(merged, fields...)
	return &fieldLogger{base: f.base, fields: merged, ctx: f.ctx}
}

func (f *fieldLogger) Ctx(ctx context.Context) FieldLogger {
	return &fieldLogger{base: f.base, fields: f.fields, ctx: ctx}
}

func (f *fieldLogger) Fields() []Field {
	out := make([]Field, len(f.fields))
	copy(out, f.fields)
	return out
}

// --- structured emit methods ---

func (f *fieldLogger) ErrorKV(msg string, kv ...any) { f.emit(Err, msg, kv) }
func (f *fieldLogger) WarnKV(msg string, kv ...any)  { f.emit(Warn, msg, kv) }
func (f *fieldLogger) InfoKV(msg string, kv ...any)  { f.emit(Info, msg, kv) }
func (f *fieldLogger) DebugKV(msg string, kv ...any) { f.emit(Debug, msg, kv) }
func (f *fieldLogger) TraceKV(msg string, kv ...any) { f.emit(Trace, msg, kv) }

func (f *fieldLogger) emit(level Level, msg string, kv []any) {
	all := f.assemble(kv)
	formatted := formatLogfmt(msg, all)
	switch level {
	case Err:
		f.base.Error(formatted)
	case Warn:
		f.base.Warn(formatted)
	case Info:
		f.base.Info(formatted)
	case Debug:
		f.base.Debug(formatted)
	case Trace:
		f.base.Trace(formatted)
	}
}

// assemble merges the field bag + ctx-derived fields + per-call kv pairs into
// a single ordered slice. Per-call kv wins on duplicate keys.
func (f *fieldLogger) assemble(kv []any) []Field {
	all := make([]Field, 0, len(f.fields)+len(kv)/2)
	all = append(all, f.fields...)
	if f.ctx != nil {
		all = append(all, ctxKeyRegistry.extract(f.ctx)...)
	}
	all = append(all, kvToFields(kv)...)
	return all
}

// kvToFields converts an even-length alternating key/value pair list into a
// []Field. An odd trailing value is recorded under the "MISSING_KEY" key so
// callers never silently drop data.
func kvToFields(kv []any) []Field {
	if len(kv) == 0 {
		return nil
	}
	out := make([]Field, 0, (len(kv)+1)/2)
	for i := 0; i < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", kv[i])
		}
		var val any
		if i+1 < len(kv) {
			val = kv[i+1]
		} else {
			key = "MISSING_KEY"
			val = kv[i]
		}
		out = append(out, Field{Key: key, Value: val})
	}
	return out
}

// formatLogfmt renders msg with appended logfmt-ish fields. Used by the
// default (printf-style) writers so structured logs are still readable when
// no JSON handler is configured.
func formatLogfmt(msg string, fields []Field) string {
	if len(fields) == 0 {
		return msg
	}
	// Stable order for deterministic console output.
	sorted := make([]Field, len(fields))
	copy(sorted, fields)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Key < sorted[j].Key })

	var b []byte
	b = append(b, msg...)
	for _, f := range sorted {
		b = append(b, ' ')
		b = append(b, f.Key...)
		b = append(b, '=')
		b = append(b, fmt.Sprintf("%v", f.Value)...)
	}
	return string(b)
}

// --- Logger pass-throughs (positional / printf API) ---

func (f *fieldLogger) Error(a ...interface{})         { f.base.Error(a...) }
func (f *fieldLogger) ErrorF(fmtStr string, a ...any) { f.base.ErrorF(fmtStr, a...) }
func (f *fieldLogger) Warn(a ...interface{})          { f.base.Warn(a...) }
func (f *fieldLogger) WarnF(fmtStr string, a ...any)  { f.base.WarnF(fmtStr, a...) }
func (f *fieldLogger) Info(a ...interface{})          { f.base.Info(a...) }
func (f *fieldLogger) InfoF(fmtStr string, a ...any)  { f.base.InfoF(fmtStr, a...) }
func (f *fieldLogger) Debug(a ...interface{})         { f.base.Debug(a...) }
func (f *fieldLogger) DebugF(fmtStr string, a ...any) { f.base.DebugF(fmtStr, a...) }
func (f *fieldLogger) Trace(a ...interface{})         { f.base.Trace(a...) }
func (f *fieldLogger) TraceF(fmtStr string, a ...any) { f.base.TraceF(fmtStr, a...) }

// --- context key registry ---

type contextKeyEntry struct {
	key       any
	fieldName string
}

type contextKeyRegistryT struct {
	mu      sync.RWMutex
	entries []contextKeyEntry
}

var ctxKeyRegistry = &contextKeyRegistryT{}

// RegisterContextKey records that ctx values keyed by ctxKey should be
// extracted into a Field named fieldName whenever a FieldLogger emits with
// a context.
//
// Typical usage at startup:
//
//	type reqIDKey struct{}
//	l3.RegisterContextKey(reqIDKey{}, "request_id")
//	type userIDKey struct{}
//	l3.RegisterContextKey(userIDKey{}, "user_id")
//
// Later, handlers do:
//
//	logger.Ctx(r.Context()).InfoKV("processed", "duration_ms", elapsed)
//
// and both request_id and user_id (if present in ctx) appear in the output.
func RegisterContextKey(ctxKey any, fieldName string) {
	if ctxKey == nil || fieldName == "" {
		return
	}
	ctxKeyRegistry.mu.Lock()
	defer ctxKeyRegistry.mu.Unlock()
	// Replace if already registered with same key.
	for i, e := range ctxKeyRegistry.entries {
		if e.key == ctxKey {
			ctxKeyRegistry.entries[i].fieldName = fieldName
			return
		}
	}
	ctxKeyRegistry.entries = append(ctxKeyRegistry.entries, contextKeyEntry{
		key:       ctxKey,
		fieldName: fieldName,
	})
}

// unregisterAllContextKeys is provided for tests to reset registry state.
func unregisterAllContextKeys() {
	ctxKeyRegistry.mu.Lock()
	ctxKeyRegistry.entries = nil
	ctxKeyRegistry.mu.Unlock()
}

func (r *contextKeyRegistryT) extract(ctx context.Context) []Field {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.entries) == 0 {
		return nil
	}
	out := make([]Field, 0, len(r.entries))
	for _, e := range r.entries {
		if v := ctx.Value(e.key); v != nil {
			out = append(out, Field{Key: e.fieldName, Value: v})
		}
	}
	return out
}

// --- JSON handler ---

// JSONRecord is the single-line JSON shape emitted by NewJSONHandler.
type JSONRecord struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"msg"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// JSONHandler is a Logger implementation that emits line-delimited JSON
// records (one per line) to an io.Writer. It's a leaf logger — you typically
// wrap it with NewFieldLogger so With/Ctx/KV methods work:
//
//	jl := l3.NewJSONHandler(os.Stdout, l3.Info)
//	log := l3.NewFieldLogger(jl, l3.F("service", "api"))
//	log.InfoKV("started", "addr", ":8080")
//	// → {"time":"...","level":"INFO","msg":"started","fields":{"addr":":8080","service":"api"}}
type JSONHandler struct {
	mu    sync.Mutex
	out   io.Writer
	level Level
}

// NewJSONHandler returns a JSONHandler that writes JSON records at or above
// the given level. Below the level is silently dropped.
func NewJSONHandler(w io.Writer, level Level) *JSONHandler {
	return &JSONHandler{out: w, level: level}
}

// SetLevel updates the minimum level recorded by this handler.
func (h *JSONHandler) SetLevel(level Level) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.level = level
}

func (h *JSONHandler) enabled(l Level) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return l <= h.level
}

// emitJSON parses a logfmt-formatted msg from formatLogfmt back into
// {message, fields} so the JSON output preserves structure.
//
// We split on the first space; everything before is the message, the
// remainder is parsed as space-separated key=value tokens. This is exactly
// what formatLogfmt produces.
func emitJSON(h *JSONHandler, level Level, full string) {
	if !h.enabled(level) {
		return
	}
	msg, fields := parseLogfmt(full)
	rec := JSONRecord{
		Time:    time.Now().UTC(),
		Level:   levelName(level),
		Message: msg,
		Fields:  fields,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return
	}
	b = append(b, '\n')
	h.mu.Lock()
	_, _ = h.out.Write(b)
	h.mu.Unlock()
}

func levelName(l Level) string {
	if int(l) >= 0 && int(l) < len(Levels) {
		return Levels[l]
	}
	return "UNKNOWN"
}

func parseLogfmt(s string) (msg string, fields map[string]any) {
	// Find first space; if none, whole string is the message.
	idx := -1
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return s, nil
	}
	msg = s[:idx]
	rest := s[idx+1:]
	// Walk through key=value tokens (no quoted-value support: values are
	// %v-formatted and may contain spaces, so we accept the rest as a
	// best-effort split).
	fields = map[string]any{}
	cursor := 0
	for cursor < len(rest) {
		// Find the next '=' to delimit key
		eq := -1
		for i := cursor; i < len(rest); i++ {
			if rest[i] == '=' {
				eq = i
				break
			}
			if rest[i] == ' ' {
				// Unexpected space inside what should be a key — drop.
				break
			}
		}
		if eq < 0 {
			break
		}
		key := rest[cursor:eq]
		// Value runs until next space at the start of another key (i.e. ' ' followed by [^=]*'=').
		valStart := eq + 1
		valEnd := len(rest)
		for i := valStart + 1; i < len(rest); i++ {
			if rest[i-1] == ' ' && nextIsKey(rest, i) {
				valEnd = i - 1
				break
			}
		}
		fields[key] = rest[valStart:valEnd]
		cursor = valEnd
		// Skip the separating space.
		if cursor < len(rest) && rest[cursor] == ' ' {
			cursor++
		}
	}
	if len(fields) == 0 {
		// Couldn't parse anything — preserve the whole string as the message.
		return s, nil
	}
	return msg, fields
}

// nextIsKey reports whether the substring starting at i looks like the start
// of a new key=value pair (a non-space, non-= run followed by '=').
func nextIsKey(s string, i int) bool {
	for j := i; j < len(s); j++ {
		if s[j] == '=' {
			return j > i
		}
		if s[j] == ' ' {
			return false
		}
	}
	return false
}

// Logger interface implementation: each method dispatches to emitJSON.
func (h *JSONHandler) Error(a ...any)            { emitJSON(h, Err, fmt.Sprint(a...)) }
func (h *JSONHandler) ErrorF(f string, a ...any) { emitJSON(h, Err, fmt.Sprintf(f, a...)) }
func (h *JSONHandler) Warn(a ...any)             { emitJSON(h, Warn, fmt.Sprint(a...)) }
func (h *JSONHandler) WarnF(f string, a ...any)  { emitJSON(h, Warn, fmt.Sprintf(f, a...)) }
func (h *JSONHandler) Info(a ...any)             { emitJSON(h, Info, fmt.Sprint(a...)) }
func (h *JSONHandler) InfoF(f string, a ...any)  { emitJSON(h, Info, fmt.Sprintf(f, a...)) }
func (h *JSONHandler) Debug(a ...any)            { emitJSON(h, Debug, fmt.Sprint(a...)) }
func (h *JSONHandler) DebugF(f string, a ...any) { emitJSON(h, Debug, fmt.Sprintf(f, a...)) }
func (h *JSONHandler) Trace(a ...any)            { emitJSON(h, Trace, fmt.Sprint(a...)) }
func (h *JSONHandler) TraceF(f string, a ...any) { emitJSON(h, Trace, fmt.Sprintf(f, a...)) }
