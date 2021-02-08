package adminonly

import (
	"net/http"
	"sync"
	"time"

	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
)

// CacheAge is the age for each cache item.
const CacheAge = 15 * time.Minute

// cache implements a small data cache.
type cache struct {
	mu sync.RWMutex
	cc map[string]cacheItem

	lastClean time.Time
}

func newCache() *cache {
	return &cache{
		lastClean: time.Now(),
	}
}

type cacheItem struct {
	data  Data
	added uint64
}

func (c *cache) get(token string) (Data, bool) {
	now := uint64(time.Now().UnixNano())

	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.cc[token]
	if !ok || item.added+uint64(CacheAge) < now {
		return Data{}, false
	}

	return item.data, ok
}

func (c *cache) set(token string, data Data) {
	now := time.Now()
	u64 := uint64(now.UnixNano())

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lastClean.Add(CacheAge).Before(now) {
		c.lastClean = now

		for k, item := range c.cc {
			if item.added+uint64(CacheAge) < u64 {
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
	var cache cache

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

			next.ServeHTTP(w, setData(r, data))
		})
	}
}
