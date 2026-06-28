package ws

import (
	"sort"
	"sync"
	"testing"
)

// makeStubConn returns a minimally-initialized *Conn that's safe to use as a
// Hub map key. Tests that exercise Hub.Send (which writes to the conn) must
// use the full server+client pattern instead.
func makeStubConn(id string) *Conn {
	return &Conn{id: id}
}

func TestHub_AddAndRooms(t *testing.T) {
	h := NewHub()
	a := makeStubConn("a")
	b := makeStubConn("b")

	h.Add(a, "user-1", "room:42")
	h.Add(b, "user-2", "room:42")

	if got := h.Len(); got != 2 {
		t.Errorf("Len = %d, want 2", got)
	}
	if got := h.CountKey("room:42"); got != 2 {
		t.Errorf("CountKey(room:42) = %d, want 2", got)
	}
	if got := h.CountKey("user-1"); got != 1 {
		t.Errorf("CountKey(user-1) = %d, want 1", got)
	}

	rooms := h.Rooms(a)
	sort.Strings(rooms)
	want := []string{"room:42", "user-1"}
	if len(rooms) != 2 || rooms[0] != want[0] || rooms[1] != want[1] {
		t.Errorf("Rooms(a) = %v, want %v", rooms, want)
	}
}

func TestHub_AddIsIdempotent(t *testing.T) {
	h := NewHub()
	c := makeStubConn("c")
	h.Add(c, "k1", "k2")
	h.Add(c, "k2", "k3") // overlap k2; new k3

	if got := h.CountKey("k1"); got != 1 {
		t.Errorf("k1 should still have 1 conn; got %d", got)
	}
	if got := h.CountKey("k2"); got != 1 {
		t.Errorf("k2 should remain 1 (not duplicate); got %d", got)
	}
	if got := h.CountKey("k3"); got != 1 {
		t.Errorf("k3 should have 1 conn; got %d", got)
	}
	if got := h.Len(); got != 1 {
		t.Errorf("Len = %d, want 1 (same conn added twice)", got)
	}
}

func TestHub_RemoveDropsFromAllKeys(t *testing.T) {
	h := NewHub()
	a := makeStubConn("a")
	b := makeStubConn("b")
	h.Add(a, "x", "y", "z")
	h.Add(b, "y")

	h.Remove(a)

	if got := h.Len(); got != 1 {
		t.Errorf("Len after remove = %d, want 1", got)
	}
	if got := h.CountKey("x"); got != 0 {
		t.Errorf("x should be empty; got %d", got)
	}
	if got := h.CountKey("y"); got != 1 {
		t.Errorf("y should retain b; got %d", got)
	}
	if got := h.CountKey("z"); got != 0 {
		t.Errorf("z should be empty (and pruned); got %d", got)
	}
	if rooms := h.Rooms(a); rooms != nil {
		t.Errorf("Rooms(removed) = %v, want nil", rooms)
	}
}

func TestHub_RemoveUnknown(t *testing.T) {
	h := NewHub()
	h.Remove(makeStubConn("ghost")) // no panic
	h.Remove(nil)                   // no panic
	if got := h.Len(); got != 0 {
		t.Errorf("Len = %d on empty hub, want 0", got)
	}
}

func TestHub_AddNilNoOp(t *testing.T) {
	h := NewHub()
	h.Add(nil, "x")
	if got := h.Len(); got != 0 {
		t.Errorf("Add(nil) should not register; Len = %d", got)
	}
}

func TestHub_SendToMissingKeyReturnsZero(t *testing.T) {
	h := NewHub()
	got := h.Send("nobody-home", NewTextMessage([]byte("x")))
	if got != 0 {
		t.Errorf("Send to missing key = %d, want 0", got)
	}
}

func TestHub_ConcurrentAddRemove(t *testing.T) {
	h := NewHub()
	conns := make([]*Conn, 200)
	for i := range conns {
		conns[i] = makeStubConn("c")
	}

	var wg sync.WaitGroup
	// 100 adders, 100 removers, all hitting the same key, racing.
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			h.Add(conns[i], "shared", "u")
		}(i)
		go func(i int) {
			defer wg.Done()
			h.Remove(conns[100+i])
		}(i)
	}
	wg.Wait()

	// All adders' conns should be in the hub; removers' conns should not.
	if got := h.Len(); got != 100 {
		t.Errorf("Len = %d after concurrent ops, want 100", got)
	}
	if got := h.CountKey("shared"); got != 100 {
		t.Errorf("CountKey(shared) = %d, want 100", got)
	}
}
