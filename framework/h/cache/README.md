# Component cache stores

A cached component stores its rendered HTML. By default each component gets its own TTL store. Use `h.Cached` for output shared by all requests and `h.CachedPerKeyT` for output that varies by user, tenant, locale, or another input. Include every input that changes the output in the key.

## Bound the number of entries

This component caches up to 1,000 profiles for 15 minutes each. Construct it once at startup, reuse the returned function across requests, and close the store when the application shuts down.

```go
package profiles

import (
	"time"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/h/cache"
)

func NewProfiles() (profile func(int) *h.Element, closeStore func()) {
	store := cache.NewLRUStore[any, string](1_000)
	profile = h.CachedPerKeyT(15*time.Minute,
		func(id int) (int, h.GetElementFunc) {
			return id, func() *h.Element {
				return h.P(h.TextF("Profile %d", id))
			}
		},
		h.WithCacheStore(store),
	)
	return profile, store.Close
}
```

The high-level helpers accept a `cache.Store[any, string]` because they store rendered HTML. Give each component a separate store or an isolated backend namespace. Two components using the same key in a shared store can return each other's HTML.

## Configure the default

`h.Cached` and `h.CachedPerKeyT` capture their store when the component is constructed. A provider assigned in `init()` applies to components constructed afterward. Package-level variables in that package have already been initialized. For package-level components that require a limit, pass `h.WithCacheStore` directly.

To set an application default, assign the provider at startup before constructing components or serving requests. Return a fresh store for each component.

```go
h.DefaultCacheProvider = func() cache.Store[any, string] {
	return cache.NewLRUStore[any, string](10_000)
}
```

## Implement a store

The [Store interface](interface.go) specifies six methods:

```go
type Store[K comparable, V any] interface {
	Set(key K, value V, ttl time.Duration)
	Get(key K) (V, bool)
	GetOrCompute(key K, compute func() V, ttl time.Duration) V
	Delete(key K)
	Purge()
	Close()
}
```

`Get` reports a miss for absent or expired entries. Concurrent calls to `GetOrCompute` for the same missing key must share one computation and return its result. A separate lookup followed by an unconditional write allows concurrent callers to repeat the work. `Set` and `GetOrCompute` receive the TTL for that entry. `Close` releases resources owned by the store.

- The [TTL store](ttl_store.go) expires entries by time and periodically removes them. `NewTTLStoreWithMaxSize` also caps the entry count and evicts the earliest expiration when full.
- The [LRU store](lru_store_example.go) enforces an entry limit using recency and checks TTL on lookup.
- When adapting Redis or Memcached, namespace keys by component, preserve per-entry expiration, and define how concurrent fills coordinate across application instances. For a two-tier cache, promoting an entry to the local tier must preserve its remaining lifetime.

[Runnable examples](example_test.go) cover bounded per-key caching, changing the default, and wiring a mock distributed backend. Run them from the framework module with `go test ./h/cache -run '^Example'`.
