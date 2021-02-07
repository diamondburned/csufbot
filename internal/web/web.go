package web

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/bot"
	"github.com/diamondburned/csufbot/internal/config"
	"github.com/diamondburned/tmplutil"
	"github.com/phogolabs/parcello"

	humanize "github.com/dustin/go-humanize"
)

//go:generate go run github.com/phogolabs/parcello/cmd/parcello -r -i *.go

// Templater is the global template tree.
var Templater = tmplutil.Templater{
	Includes: map[string]string{
		"css":      "components/css.html",
		"errorbox": "components/errorbox.html",
		"header":   "components/header.html",
		"footer":   "components/footer.html",
	},
	Functions: template.FuncMap{
		"humanizeTime": humanize.Time,
		"shortError": func(err error) string {
			parts := strings.Split(err.Error(), ": ")
			if len(parts) == 0 {
				return ""
			}

			part := parts[len(parts)-1]

			r, sz := utf8.DecodeRuneInString(part)
			if sz == 0 {
				return ""
			}

			return string(unicode.ToUpper(r)) + part[sz:] + "."
		},
	},
}

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
	d, err := parcello.Manager.Dir("static/")
	if err != nil {
		log.Fatalln("Static not found:", err)
	}

	return http.StripPrefix("/static", http.FileServer(d))
}
