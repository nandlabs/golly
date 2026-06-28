package l3

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

// captureLogger is a tiny Logger that records every emit for inspection.
type captureLogger struct {
	mu      sync.Mutex
	lastLvl Level
	lastMsg string
}

func (c *captureLogger) write(lvl Level, msg string) {
	c.mu.Lock()
	c.lastLvl = lvl
	c.lastMsg = msg
	c.mu.Unlock()
}
func (c *captureLogger) Error(a ...any)            { c.write(Err, joinArgs(a...)) }
func (c *captureLogger) ErrorF(f string, a ...any) { c.write(Err, sprintf(f, a...)) }
func (c *captureLogger) Warn(a ...any)             { c.write(Warn, joinArgs(a...)) }
func (c *captureLogger) WarnF(f string, a ...any)  { c.write(Warn, sprintf(f, a...)) }
func (c *captureLogger) Info(a ...any)             { c.write(Info, joinArgs(a...)) }
func (c *captureLogger) InfoF(f string, a ...any)  { c.write(Info, sprintf(f, a...)) }
func (c *captureLogger) Debug(a ...any)            { c.write(Debug, joinArgs(a...)) }
func (c *captureLogger) DebugF(f string, a ...any) { c.write(Debug, sprintf(f, a...)) }
func (c *captureLogger) Trace(a ...any)            { c.write(Trace, joinArgs(a...)) }
func (c *captureLogger) TraceF(f string, a ...any) { c.write(Trace, sprintf(f, a...)) }

func joinArgs(a ...any) string {
	parts := make([]string, len(a))
	for i, v := range a {
		parts[i] = anyToString(v)
	}
	return strings.Join(parts, "")
}
func anyToString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func sprintf(f string, a ...any) string { return f } // not used in these tests

// ---- KV bag, With, Ctx ----

func TestKV_BasicEmit(t *testing.T) {
	cap := &captureLogger{}
	l := NewFieldLogger(cap, F("service", "api"))
	l.InfoKV("started", "addr", ":8080", "tls", true)

	cap.mu.Lock()
	defer cap.mu.Unlock()
	if cap.lastLvl != Info {
		t.Errorf("level = %v, want Info", cap.lastLvl)
	}
	for _, want := range []string{"started", "addr=:8080", "tls=true", "service=api"} {
		if !strings.Contains(cap.lastMsg, want) {
			t.Errorf("missing %q in %q", want, cap.lastMsg)
		}
	}
}

func TestWith_AccumulatesImmutably(t *testing.T) {
	cap := &captureLogger{}
	base := NewFieldLogger(cap, F("a", 1))
	derived := base.With(F("b", 2))

	if len(base.Fields()) != 1 {
		t.Errorf("With() must not mutate the receiver; base fields=%v", base.Fields())
	}
	if len(derived.Fields()) != 2 {
		t.Errorf("derived should have 2 fields; got %v", derived.Fields())
	}
	derived.InfoKV("hi", "c", 3)
	if !strings.Contains(cap.lastMsg, "a=1") || !strings.Contains(cap.lastMsg, "b=2") || !strings.Contains(cap.lastMsg, "c=3") {
		t.Errorf("expected a,b,c in output; got %q", cap.lastMsg)
	}
}

func TestKV_OddTrailingValueRecordedNotDropped(t *testing.T) {
	cap := &captureLogger{}
	l := NewFieldLogger(cap)
	l.InfoKV("oops", "key1", "v1", "orphan")
	if !strings.Contains(cap.lastMsg, "MISSING_KEY=orphan") {
		t.Errorf("expected MISSING_KEY=orphan in %q", cap.lastMsg)
	}
}

// ---- context key registry ----

type reqIDKey struct{}
type userIDKey struct{}

func TestCtx_ExtractsRegisteredKeys(t *testing.T) {
	defer unregisterAllContextKeys()
	RegisterContextKey(reqIDKey{}, "request_id")
	RegisterContextKey(userIDKey{}, "user_id")

	cap := &captureLogger{}
	logger := NewFieldLogger(cap, F("service", "api"))

	ctx := context.Background()
	ctx = context.WithValue(ctx, reqIDKey{}, "abc-123")
	ctx = context.WithValue(ctx, userIDKey{}, "u-42")

	logger.Ctx(ctx).InfoKV("hello")

	for _, want := range []string{"request_id=abc-123", "user_id=u-42", "service=api"} {
		if !strings.Contains(cap.lastMsg, want) {
			t.Errorf("missing %q in %q", want, cap.lastMsg)
		}
	}
}

func TestCtx_OnlyEmitsKeysPresentInContext(t *testing.T) {
	defer unregisterAllContextKeys()
	RegisterContextKey(reqIDKey{}, "request_id")
	RegisterContextKey(userIDKey{}, "user_id")

	cap := &captureLogger{}
	logger := NewFieldLogger(cap)
	logger.Ctx(context.WithValue(context.Background(), reqIDKey{}, "only-req")).InfoKV("hi")

	if !strings.Contains(cap.lastMsg, "request_id=only-req") {
		t.Errorf("expected request_id in %q", cap.lastMsg)
	}
	if strings.Contains(cap.lastMsg, "user_id=") {
		t.Errorf("user_id should be absent; got %q", cap.lastMsg)
	}
}

func TestRegisterContextKey_Replaces(t *testing.T) {
	defer unregisterAllContextKeys()
	RegisterContextKey(reqIDKey{}, "request_id")
	RegisterContextKey(reqIDKey{}, "rid") // replaces

	cap := &captureLogger{}
	NewFieldLogger(cap).
		Ctx(context.WithValue(context.Background(), reqIDKey{}, "X")).
		InfoKV("msg")

	if !strings.Contains(cap.lastMsg, "rid=X") {
		t.Errorf("expected rid=X; got %q", cap.lastMsg)
	}
}

// ---- JSON handler ----

func TestJSONHandler_EmitsStructuredRecord(t *testing.T) {
	var buf bytes.Buffer
	jh := NewJSONHandler(&buf, Trace)
	l := NewFieldLogger(jh, F("service", "api"))
	l.InfoKV("started", "addr", ":8080")

	var rec JSONRecord
	if err := json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &rec); err != nil {
		t.Fatalf("decode: %v\nout=%q", err, buf.String())
	}
	if rec.Level != "INFO" {
		t.Errorf("level = %q, want INFO", rec.Level)
	}
	if rec.Message != "started" {
		t.Errorf("msg = %q, want started", rec.Message)
	}
	if rec.Fields["addr"] != ":8080" || rec.Fields["service"] != "api" {
		t.Errorf("fields wrong: %+v", rec.Fields)
	}
}

func TestJSONHandler_RespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	jh := NewJSONHandler(&buf, Warn) // only warn & error
	l := NewFieldLogger(jh)
	l.InfoKV("dropped")
	l.WarnKV("kept")

	out := buf.String()
	if strings.Contains(out, "dropped") {
		t.Errorf("info-level record leaked through warn filter: %q", out)
	}
	if !strings.Contains(out, "kept") {
		t.Errorf("warn-level record was dropped: %q", out)
	}
}

func TestJSONHandler_NoFields(t *testing.T) {
	var buf bytes.Buffer
	jh := NewJSONHandler(&buf, Trace)
	jh.Info("plain message no fields")

	var rec JSONRecord
	if err := json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &rec); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if rec.Message != "plain message no fields" {
		t.Errorf("msg = %q, want plain message no fields", rec.Message)
	}
	if len(rec.Fields) != 0 {
		t.Errorf("fields should be empty; got %+v", rec.Fields)
	}
}

// ---- concurrency ----

func TestFieldLogger_ConcurrentEmits(t *testing.T) {
	cap := &captureLogger{}
	l := NewFieldLogger(cap, F("s", "x"))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			l.With(F("i", i)).InfoKV("hit", "n", i)
		}(i)
	}
	wg.Wait()
}
