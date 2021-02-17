package web

import (
	"context"
	"embed"
	"net/http"

	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/bot"
	"github.com/diamondburned/csufbot/internal/config"
	"github.com/diamondburned/tmplutil"
)

//go:embed *
var webFS embed.FS

// Templater is the global template tree.
var Templater = tmplutil.Preregister(&tmplutil.Templater{
	FileSystem: webFS,
	Includes: map[string]string{
		"css":    "components/css.html",
		"header": "components/header.html",
		"footer": "components/footer.html",
	},
	Functions: funcs,
})

type ctxTypes uint8

const (
	renderCfgCtx ctxTypes = iota
)

// RenderConfig is the config to render with.
type RenderConfig struct {
	csufbot.Store
	config.Site

	Discord  *bot.Discord
	Services []config.Service
}

// Service gets a service from the provided host. It returns nil if none is
// found.
func (rcfg RenderConfig) Service(host lms.Host) *config.Service {
	return rcfg.FindService(string(host))
}

// FindService finds the LMS service from a given name hash.
func (rcfg RenderConfig) FindService(host string) *config.Service {
	for i, svc := range rcfg.Services {
		if svc.Host == lms.Host(host) {
			return &rcfg.Services[i]
		}
	}
	return nil
}

// Middleware is the type for a middleware.
type Middleware = func(http.Handler) http.Handler

// InjectConfig injects the render config.
func InjectConfig(config RenderConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), renderCfgCtx, config)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRenderConfig gets the render config from the given context. It panics if
// the config is not available.
func GetRenderConfig(ctx context.Context) RenderConfig {
	config, ok := ctx.Value(renderCfgCtx).(RenderConfig)
	if !ok {
		panic("no RenderConfig in context")
	}
	return config
}

// MountStatic mounts the /static folder.
func MountStatic() http.Handler {
	d := tmplutil.MustSub(webFS, "static")
	return http.StripPrefix("/static", http.FileServer(http.FS(d)))
}
