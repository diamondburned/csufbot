package oauth

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	// cacheAge is the age for each cache item.
	cacheAge = 30 * time.Minute
	// globalClean is the duration to globally clean the cache.
	globalClean = 5 * time.Hour
)

type CacheKey uint32

var keyIncr uint32

// NewCacheKey generates a global cache key unique to each usage of the cache.
func NewCacheKey() CacheKey {
	return CacheKey(atomic.AddUint32(&keyIncr, 1))
}

// CacheStore contains cache values.
type CacheStore map[CacheKey]interface{}

// cacheItem is a single cache item. The underlying storage is thread-safe.
type cacheItem struct {
	mutex sync.Mutex
	store CacheStore

	added uint32 // synchronized independently
}

func (it *cacheItem) isExpired(t time.Time) bool {
	return it.added+uint32(cacheAge/time.Second) < uint32(t.Unix())
}

type cacheRepository struct {
	mu sync.Mutex
	cc map[string]*cacheItem

	lastClean time.Time
}

// newCache constructs a new cache instance.
func newCache() *cacheRepository {
	return &cacheRepository{
		cc:        map[string]*cacheItem{},
		lastClean: time.Now(),
	}
}

// acquire tries to get a cache item and acquires it. This acquisition will last
// for an entire request.
func (c *cacheRepository) acquire(token string) *cacheItem {
	now := time.Now()
	expired := false

	c.mu.Lock()

	it, exists := c.cc[token]
	if !exists {
		it = &cacheItem{
			store: make(CacheStore),
			added: uint32(now.Unix()),
		}

		c.cc[token] = it
		c.cleanup(now)
	} else {
		// We have to check if the cache is expired while we're locking the main
		// mutex. This is to prevent racing with another cleanup routine.
		expired = it.isExpired(now)
		it.added = uint32(now.Unix())
	}

	c.mu.Unlock()

	it.mutex.Lock()

	// Recreate the store if the cache is too old.
	if expired {
		it.store = make(CacheStore)
	}

	return it
}

func (c *cacheRepository) release(item *cacheItem) {
	item.mutex.Unlock()
}

func (c *cacheRepository) cleanup(now time.Time) {
	// Infrequently clean the cache.
	if !c.lastClean.Add(globalClean).Before(now) {
		return
	}

	c.lastClean = now
	nowSecs := uint32(now.Unix())

	if c.lastClean.Add(cacheAge).Before(now) {
		c.lastClean = now

		for k, item := range c.cc {
			added := atomic.LoadUint32(&item.added)
			if added+uint32(cacheAge/time.Second) < nowSecs {
				delete(c.cc, k)
			}
		}
	}
}
