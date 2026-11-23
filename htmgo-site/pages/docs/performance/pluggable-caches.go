package performance

import (
	"github.com/franchb/htmgo/framework/v2/h"
	. "htmgo-site/pages/docs"
	"htmgo-site/ui"
)

func PluggableCaches(ctx *h.RequestContext) *h.Page {
	return DocPage(
		ctx,
		h.Div(
			h.Class("flex flex-col gap-3"),
			Title("Creating Custom Cache Stores"),
			Text("Each cached component stores rendered HTML. The default TTL store removes expired entries in the background. Use a bounded LRU store when many different keys can arrive during a single TTL."),
			SubTitle("Bound a component's cache"),
			Text("Construct this component once at startup. Keep the returned profile function for request handlers and call the cleanup function when the application shuts down."),
			ui.GoCodeSnippet(PerComponentExample),
			Text("Use a separate store for each component. For user-specific output, include the user and every other input that changes the rendered HTML in the key."),
			SubTitle("Set an application default"),
			Text("Assign the provider before constructing components or serving requests. Each call creates a separate store. Existing components keep their original store, including package-level components initialized before init() runs."),
			ui.GoCodeSnippet(GlobalConfigExample),
			SubTitle("Implement a store"),
			ui.GoCodeSnippet(CacheStoreInterface),
			Text("Get reports a miss for absent or expired entries. GetOrCompute must coordinate concurrent misses for a key so that its callers share one computation. Both Set and GetOrCompute take a per-entry TTL. Close releases resources owned by the store."),
			Text("The component helpers take a Store[any, string]. For a Redis or Memcached backend, give each component its own key namespace, preserve expiration, and coordinate concurrent fills across instances. A local cache in front of a remote one must preserve the remaining TTL when promoting entries."),
			Link("Cache implementations and runnable examples", "https://github.com/franchb/htmgo/tree/master/framework/h/cache"),
			NextStep(
				"mt-4",
				PrevBlock("Caching Per Key", DocPath("/performance/caching-per-key")),
				NextBlock("Server Sent Events", DocPath("/pushing-data/sse")),
			),
		),
	)
}

const PerComponentExample = `package profiles

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
}`

const GlobalConfigExample = `h.DefaultCacheProvider = func() cache.Store[any, string] {
	return cache.NewLRUStore[any, string](10_000)
}`

const CacheStoreInterface = `type Store[K comparable, V any] interface {
	Set(key K, value V, ttl time.Duration)
	Get(key K) (V, bool)
	GetOrCompute(key K, compute func() V, ttl time.Duration) V
	Delete(key K)
	Purge()
	Close()
}`
