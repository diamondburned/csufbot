package service

import (
	"context"
	"log"
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/diamondburned/tmplutil"
	"github.com/go-chi/chi"
)

type ctxKey uint8

const (
	serviceCtxKey ctxKey = iota
)

var service = web.Templater.Register("service", "pages/sync/service/service.html")

func Mount(paramName string) http.Handler {
	r := chi.NewRouter()
	r.Use(tmplutil.AlwaysFlush)
	r.Use(needService(paramName))
	r.Get("/", renderServiceSync)
	r.Post("/", postSync)

	return r
}

func needService(name string) web.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serviceHost := lms.Host(chi.URLParam(r, name))

			cfg := web.GetRenderConfig(r.Context())
			svc := cfg.FindService(serviceHost)
			if svc == nil {
				w.WriteHeader(404)
				return
			}

			ctx := context.WithValue(r.Context(), serviceCtxKey, svc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getService(ctx context.Context) *web.LMSService {
	sv, ok := ctx.Value(serviceCtxKey).(*web.LMSService)
	if !ok {
		log.Panicln("missing *web.LMSService in request")
	}
	return sv
}

type serviceSyncData struct {
	web.RenderConfig
	Client  *oauth.UserClient
	Service *web.LMSService
}

func (data serviceSyncData) Me() *discord.User {
	u, _ := data.Client.Me()
	return u
}

func renderServiceSync(w http.ResponseWriter, r *http.Request) {
	cfg := web.GetRenderConfig(r.Context())
	cli := oauth.Client(r.Context())
	svc := getService(r.Context())

	service.Execute(w, serviceSyncData{
		RenderConfig: cfg,
		Client:       cli,
		Service:      svc,
	})
}
