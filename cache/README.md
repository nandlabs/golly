# cache

A small, generic key/value cache abstraction for golly, plus an in-memory
backend. The interface is provider-agnostic so downstream modules (Redis,
Memcached, cloud-managed caches) can implement it without pulling those
dependencies into core — matching the pattern used by `golly/vfs` and
`golly/messaging`.

## Status

| Backend | Where it lives | Status |
| --- | --- | --- |
| `InMemory` (this package) | `oss.nandlabs.io/golly/cache` | ✅ shipped |
| Redis | proposed `oss.nandlabs.io/golly-cache-redis` | follow-up |
| Memcached | proposed `oss.nandlabs.io/golly-cache-memcached` | follow-up |
| Cloud caches (ElastiCache, MemoryDB, Memorystore, …) | `golly-aws` / `golly-gcp` | follow-up |

## Interface

```go
type Cache[K comparable, V any] interface {
    Get(ctx context.Context, key K) (V, bool)
    Set(ctx context.Context, key K, value V) error
    SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error
    Delete(ctx context.Context, key K) bool
    Has(ctx context.Context, key K) bool
    Clear(ctx context.Context) error
    Len(ctx context.Context) int
    Close() error
}
```

Two optional capabilities are surfaced as separate interfaces — implementations
opt in by satisfying them:

- `Sweeper` — eagerly evict expired entries (`InMemory` implements it).
- `Loader[K, V]` — `GetOrLoad` for thundering-herd protection (not yet
  implemented on `InMemory`; reserved interface for future backends).

## Quick start

```go
import (
    "context"
    "time"

    "oss.nandlabs.io/golly/cache"
)

c := cache.NewInMemory[string, []byte]()
defer c.Close()

ctx := context.Background()
_ = c.Set(ctx, "session:abc", payload)
_ = c.SetWithTTL(ctx, "rate:user42", token, 30*time.Second)

if v, ok := c.Get(ctx, "session:abc"); ok {
    use(v)
}
```

### Background expiry

The in-memory backend expires entries **lazily** on access. To bound memory
under churn, enable a janitor goroutine that sweeps periodically:

```go
c := cache.NewInMemory[string, int](
    cache.WithJanitor[string, int](1 * time.Minute),
)
defer c.Close() // stops the janitor
```

Or sweep manually when you know it's a quiet moment:

```go
n := c.Sweep()       // returns number of expired entries removed
```

### Zero TTL means never expires

```go
_ = c.SetWithTTL(ctx, "config", v, cache.NoExpiry) // same as Set
```

## Design notes

- **Generic over keys and values** — no `interface{}` boxing, no reflection.
- **Context-aware** — the in-memory backend ignores `ctx` but the signature
  is in place for network backends.
- **Close semantics** — after `Close`, writes return `ErrClosed`; reads
  continue against the existing snapshot for graceful drain.
- **No maximum size / eviction policy yet** — the in-memory backend has
  unbounded capacity. LRU / TinyLFU / size-based eviction are planned
  follow-ups; track via #51 follow-on issues.
- **Zero external deps** — stdlib only.

## License

Dual-licensed under [Apache 2.0](../LICENSE-APACHE) or [MIT](../LICENSE-MIT)
at your option (same as the rest of golly).
