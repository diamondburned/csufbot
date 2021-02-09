package adminonly

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
)

// cacheAge is the age for each cache item.
const cacheAge = 30 * time.Minute

// Cache implements a small data Cache.
type Cache struct {
	mu sync.RWMutex
	cc map[string]cacheItem

	lastClean time.Time
}

// newCache constructs a new cache instance.
func newCache() *Cache {
	return &Cache{
		cc:        map[string]cacheItem{},
		lastClean: time.Now(),
	}
}

type cacheItem struct {
	data  Data
	added uint64
}

// Invalidate invalidates the cache item with the given user client.
func (c *Cache) Invalidate(client *oauth.UserClient) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cc, client.Token)
}

func (c *Cache) get(token string) (Data, bool) {
	now := uint64(time.Now().UnixNano())

	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.cc[token]
	if !ok || item.added+uint64(cacheAge) < now {
		return Data{}, false
	}

	return item.data, ok
}

func (c *Cache) set(token string, data Data) {
	now := time.Now()
	u64 := uint64(now.UnixNano())

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lastClean.Add(cacheAge).Before(now) {
		c.lastClean = now

		for k, item := range c.cc {
			if item.added+uint64(cacheAge) < u64 {
				delete(c.cc, k)
			}
		}
	}

	c.cc[token] = cacheItem{
		data:  data,
		added: u64,
	}
}

func cachedRequire(routeParam string) web.Middleware {
	cache := newCache()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := oauth.Client(r.Context())

			data, ok := cache.get(c.Token)
			if !ok {
				data, ok = fetchData(w, r, routeParam)
				if !ok {
					return
				}
				cache.set(c.Token, data)
			}

			r = setData(r, data)
			r = r.WithContext(context.WithValue(r.Context(), cacheDataKey, c))
			next.ServeHTTP(w, r)
		})
	}
}

// GetCache gets the cache instance. This will only return non-nil if
// Require is called with cached being true.
func GetCache(ctx context.Context) *Cache {
	c, _ := ctx.Value(cacheDataKey).(*Cache)
	return c
}
