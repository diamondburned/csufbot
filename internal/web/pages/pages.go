package pages

import (
	"net/http"

	"github.com/diamondburned/csufbot/internal/web"
	"github.com/go-chi/chi"
)

var (
	index   = web.Templater.Register("index", "pages/index.html")
	service = web.Templater.Register("service", "pages/service.html")
)

func Mount(cfg web.RenderConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(web.InjectConfig(cfg))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Only write 404 so we can do the same on other handlers as well.
		w.WriteHeader(404)
	})

	r.Route("/{serviceHash}", func(r chi.Router) {
		r.Get("/", renderService)
		r.Post("/", serviceAuthorizer)
	})

	return r
}

func getService(r *http.Request) *web.LMSService {
	var (
		nameHash = chi.URLParam(r, "serviceHash")
		config   = web.GetRenderConfig(r.Context())
		service  = config.FindService(nameHash)
	)

	return service
}

func renderService(w http.ResponseWriter, r *http.Request) {
	service := getService(r)
	if service == nil {
		w.WriteHeader(404)
		return
	}

}

func serviceAuthorizer(w http.ResponseWriter, r *http.Request) {
	service := getService(r)
	if service == nil {
		w.WriteHeader(404)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(401)
		return
	}

}
