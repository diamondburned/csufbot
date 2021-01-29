package pages

import (
	"net/http"

	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/pages/admin"
	"github.com/diamondburned/csufbot/internal/web/pages/oauth"
	"github.com/diamondburned/csufbot/internal/web/pages/sync"
	"github.com/go-chi/chi"
)

func Mount(cfg web.RenderConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(web.InjectConfig(cfg))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Only write 404 so we can do the same on other handlers as well.
		w.WriteHeader(404)
	})

	r.Mount("/sync", sync.Mount())
	r.Mount("/admin", admin.Mount())
	r.Mount("/oauth", oauth.Mount())

	return r
}
