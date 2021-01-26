package web

import (
	"context"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/diamondburned/csufbot/internal/lms"
	"github.com/diamondburned/csufbot/internal/session"
	"github.com/diamondburned/tmplutil"
	"github.com/phogolabs/parcello"
)

//go:generate go run github.com/phogolabs/parcello/cmd/parcello -r -i *.go

// Templater is the global template tree.
var Templater = tmplutil.Templater{
	Includes: map[string]string{
		"css":    "components/css.html",
		"header": "components/header.html",
		"footer": "components/footer.html",
	},
	Functions: template.FuncMap{},
}

type ctxTypes uint8

const (
	renderCfgCtx ctxTypes = iota
)

// LMSService describes a LMS service.
type LMSService struct {
	lms.Service
	NameHash    string
	Instruction template.HTML
}

// NewLMSService creates a new LMS service.
func NewLMSService(svc lms.Service, instruction string) LMSService {
	h := fnv.New64a()
	h.Write([]byte(svc.Name()))

	return LMSService{
		Service:     svc,
		NameHash:    strconv.FormatUint(h.Sum64(), 36),
		Instruction: template.HTML(instruction),
	}
}

// UserRegisterer is the interface to register a LMS user from the given ticket.
type UserRegisterer interface {
	RegisterUser(session.Ticket, lms.Service, lms.User)
}

// RenderConfig is the config to render with.
type RenderConfig struct {
	Services   []LMSService
	Sessions   session.Repository
	Registerer UserRegisterer
}

// FindService finds the LMS service from a given name hash.
func (rcfg RenderConfig) FindService(nameHash string) *LMSService {
	for i, svc := range rcfg.Services {
		if svc.NameHash == nameHash {
			return &rcfg.Services[i]
		}
	}
	return nil
}

// InjectConfig injects the render config.
func InjectConfig(config RenderConfig) func(http.Handler) http.Handler {
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
