package cache

import (
	"log/slog"
	"sync"
	"time"
)

// call represents an in-flight or completed computation for a single key.
type call[V any] struct {
	wg  sync.WaitGroup
	val V
}

// TTLStore is a time-to-live based cache implementation that mimics
// the original htmgo caching behavior. It stores values with expiration
// times and periodically cleans up expired entries.
type TTLStore[K comparable, V any] struct {
	cache     map[K]*entry[V]
	mutex     sync.RWMutex
	inflight  map[K]*call[V]
	imu       sync.Mutex
	maxSize   int // 0 means unlimited
	closeOnce sync.Once
	closeChan chan struct{}
}

type entry[V any] struct {
	value      V
	expiration time.Time
}

// NewTTLStore creates a new TTL-based cache store with a default 1-minute cleaner interval.
func NewTTLStore[K comparable, V any]() Store[K, V] {
	return NewTTLStoreWithInterval[K, V](time.Minute)
}

// NewTTLStoreWithInterval creates a new TTL-based cache store with a configurable cleaner interval.
func NewTTLStoreWithInterval[K comparable, V any](cleanInterval time.Duration) Store[K, V] {
	s := &TTLStore[K, V]{
		cache:     make(map[K]*entry[V]),
		inflight:  make(map[K]*call[V]),
		closeChan: make(chan struct{}),
	}
	s.startCleaner(cleanInterval)
	return s
}

// NewTTLStoreWithMaxSize creates a new TTL-based cache store with a maximum number of entries.
// When the cache exceeds maxSize during Set or GetOrCompute, the oldest entries are evicted.
// Note: eviction is O(n) per insertion that exceeds maxSize. For large caches where eviction
// performance matters, use LRUStore instead. A maxSize of 0 or less means unlimited.
func NewTTLStoreWithMaxSize[K comparable, V any](maxSize int) Store[K, V] {
	s := &TTLStore[K, V]{
		cache:     make(map[K]*entry[V]),
		inflight:  make(map[K]*call[V]),
		maxSize:   maxSize,
		closeChan: make(chan struct{}),
	}
	s.startCleaner(time.Minute)
	return s
}

// Set adds or updates an entry in the cache with the given TTL.
func (s *TTLStore[K, V]) Set(key K, value V, ttl time.Duration) {
	s.mutex.Lock()
	s.cache[key] = &entry[V]{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	if s.maxSize > 0 && len(s.cache) > s.maxSize {
		s.evictOldest()
	}
	s.mutex.Unlock()
}

// Get retrieves a value from the cache. Returns the value and true if found and not expired,
// or the zero value and false otherwise.
func (s *TTLStore[K, V]) Get(key K) (V, bool) {
	s.mutex.RLock()
	e, ok := s.cache[key]
	if ok && time.Now().Before(e.expiration) {
		val := e.value
		s.mutex.RUnlock()
		return val, true
	}
	s.mutex.RUnlock()
	var zero V
	return zero, false
}

// GetOrCompute gets an existing value or computes and stores a new value.
// Uses per-key deduplication so that concurrent requests for the same key
// only trigger a single computation, without blocking operations on other keys.
func (s *TTLStore[K, V]) GetOrCompute(key K, compute func() V, ttl time.Duration) V {
	// Fast path: read lock, check cache
	s.mutex.RLock()
	if e, ok := s.cache[key]; ok && time.Now().Before(e.expiration) {
		val := e.value
		s.mutex.RUnlock()
		return val
	}
	s.mutex.RUnlock()

	// Slow path: check/create inflight entry
	s.imu.Lock()
	if c, ok := s.inflight[key]; ok {
		// Another goroutine is already computing this key
		s.imu.Unlock()
		c.wg.Wait()
		return c.val
	}

	// Create a new inflight entry
	c := &call[V]{}
	c.wg.Add(1)
	s.inflight[key] = c
	s.imu.Unlock()

	// Double-check cache in case another goroutine just filled it
	s.mutex.RLock()
	if e, ok := s.cache[key]; ok && time.Now().Before(e.expiration) {
		val := e.value
		s.mutex.RUnlock()
		c.val = val
		c.wg.Done()
		s.imu.Lock()
		delete(s.inflight, key)
		s.imu.Unlock()
		return val
	}
	s.mutex.RUnlock()

	// Compute value with no locks held.
	// Wrap in a closure so that a panicking compute() still unblocks waiters
	// and cleans up the inflight slot (otherwise they deadlock on c.wg.Wait()).
	var value V
	panicked := true
	func() {
		defer func() {
			if panicked {
				c.wg.Done()
				s.imu.Lock()
				delete(s.inflight, key)
				s.imu.Unlock()
			}
		}()
		value = compute()
		panicked = false
	}()
	if panicked {
		// re-panic is handled by the runtime; this line is unreachable,
		// but kept for clarity.
		var zero V
		return zero
	}

	// Store the result
	s.mutex.Lock()
	s.cache[key] = &entry[V]{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	if s.maxSize > 0 && len(s.cache) > s.maxSize {
		s.evictOldest()
	}
	s.mutex.Unlock()

	// Signal waiters
	c.val = value
	c.wg.Done()

	// Clean up inflight
	s.imu.Lock()
	delete(s.inflight, key)
	s.imu.Unlock()

	return value
}

// Delete removes an entry from the cache.
func (s *TTLStore[K, V]) Delete(key K) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.cache, key)
}

// Purge removes all items from the cache.
func (s *TTLStore[K, V]) Purge() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cache = make(map[K]*entry[V])
}

// Close stops the background cleaner goroutine.
func (s *TTLStore[K, V]) Close() {
	s.closeOnce.Do(func() {
		close(s.closeChan)
	})
}

// startCleaner starts a background goroutine that periodically removes expired entries.
func (s *TTLStore[K, V]) startCleaner(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.clearExpired()
			case <-s.closeChan:
				return
			}
		}
	}()
}

// clearExpired removes expired entries from the cache in batches to limit write-lock duration.
func (s *TTLStore[K, V]) clearExpired() {
	const batchSize = 1000

	// Phase 1: RLock, collect up to batchSize expired keys
	s.mutex.RLock()
	now := time.Now()
	expired := make([]K, 0, batchSize)
	for key, e := range s.cache {
		if now.After(e.expiration) {
			expired = append(expired, key)
			if len(expired) >= batchSize {
				break
			}
		}
	}
	s.mutex.RUnlock()

	if len(expired) == 0 {
		return
	}

	// Phase 2: Lock, delete collected keys (re-verify expiration)
	s.mutex.Lock()
	now = time.Now()
	deletedCount := 0
	for _, key := range expired {
		if e, ok := s.cache[key]; ok && now.After(e.expiration) {
			delete(s.cache, key)
			deletedCount++
		}
	}
	s.mutex.Unlock()

	if deletedCount > 0 {
		slog.Debug("Deleted expired cache entries", slog.Int("count", deletedCount))
	}
}

// evictOldest removes the entry with the earliest expiration time.
// This is O(n) in the number of cache entries because it scans the entire map.
// For large caches, prefer LRUStore which has O(1) eviction.
// Must be called with the write lock held.
func (s *TTLStore[K, V]) evictOldest() {
	var oldestKey K
	var oldestTime time.Time
	first := true
	for k, e := range s.cache {
		if first || e.expiration.Before(oldestTime) {
			oldestKey = k
			oldestTime = e.expiration
			first = false
		}
	}
	if !first {
		delete(s.cache, oldestKey)
	}
}
