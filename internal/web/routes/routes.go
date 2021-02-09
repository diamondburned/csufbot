package routes

import (
	"errors"
	"net/http"

	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/components/errorbox"
	"github.com/diamondburned/csufbot/internal/web/routes/admin"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/diamondburned/csufbot/internal/web/routes/sync"
	"github.com/go-chi/chi"
)

func Mount(cfg web.RenderConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(noSniff)
	r.Use(web.InjectConfig(cfg))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Only write 404 so we can do the same on other handlers as well.
		errorbox.Render(w, r, 404, errors.New("path not found"))
	})

	r.Mount("/sync", sync.Mount())
	r.Mount("/admin", admin.Mount())
	r.Mount("/oauth", oauth.Mount())

	return r
}

func noSniff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
