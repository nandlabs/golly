// Package cache provides a small, generic key/value cache abstraction plus
// an in-memory backend. The interface is provider-agnostic so downstream
// modules (redis, memcached, cloud-managed caches) can implement it without
// pulling those dependencies into the core package — matching the same
// pattern used by golly/vfs and golly/messaging.
package cache

import (
	"context"
	"errors"
	"time"
)

// NoExpiry is a TTL value that disables expiry for a given entry.
const NoExpiry time.Duration = 0

// ErrClosed is returned by operations on a cache that has been closed.
var ErrClosed = errors.New("cache: closed")

// Cache is a generic key/value cache with optional per-entry TTL.
// Implementations MUST be safe for concurrent use.
//
// The context is intended for backends that perform network I/O
// (Redis, Memcached, cloud caches); in-memory backends may ignore it.
type Cache[K comparable, V any] interface {
	// Get returns the value for the key and whether it was found. Entries
	// past their TTL return (zero value, false).
	Get(ctx context.Context, key K) (V, bool)

	// Set stores value under key with no expiry. Equivalent to
	// SetWithTTL(ctx, key, value, NoExpiry).
	Set(ctx context.Context, key K, value V) error

	// SetWithTTL stores value under key with the given TTL. A ttl of
	// NoExpiry (zero) means the entry never expires.
	SetWithTTL(ctx context.Context, key K, value V, ttl time.Duration) error

	// Delete removes the key. Returns true if it existed.
	Delete(ctx context.Context, key K) bool

	// Has reports whether key is present and not expired.
	Has(ctx context.Context, key K) bool

	// Clear empties the cache.
	Clear(ctx context.Context) error

	// Len returns the current number of stored entries. Lazy-expiry
	// backends may include entries past their TTL until next access;
	// callers wanting a precise live count should call Sweep first if
	// the backend implements Sweeper.
	Len(ctx context.Context) int

	// Close releases any resources held by the cache (background
	// goroutines, network connections). After Close, all other methods
	// return ErrClosed. Safe to call multiple times.
	Close() error
}

// Sweeper is an optional capability for caches that retain expired entries
// until accessed; calling Sweep removes them eagerly.
type Sweeper interface {
	// Sweep removes all expired entries and returns the number removed.
	Sweep() int
}

// Loader is an optional capability some caches expose for the
// "get-or-load" pattern that protects the underlying source from
// thundering-herd traffic.
type Loader[K comparable, V any] interface {
	// GetOrLoad returns the cached value for key. If absent or expired,
	// load is invoked exactly once per concurrent miss and the result is
	// cached with the supplied ttl. Other concurrent callers for the
	// same key wait for the single load and share its result.
	GetOrLoad(ctx context.Context, key K, ttl time.Duration, load func(context.Context) (V, error)) (V, error)
}
