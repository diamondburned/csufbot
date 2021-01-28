package pages

import (
	"log"
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/csufbot/session"
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

	r.Get("/redirect", redirect)

	r.Group(func(r chi.Router) {
		r.Use(web.RequireTicket(cfg, ""))
		r.Get("/", render)

		r.Route("/{serviceHash}", func(r chi.Router) {
			r.Get("/", renderService)
			r.Post("/", serviceAuthorizer)
		})
	})

	return r
}

type indexData struct {
	web.RenderConfig
	Ticket *session.Ticket
}

func (data indexData) Member() *discord.Member {
	m, _ := data.Discord.Member(data.Ticket.GuildID, data.Ticket.UserID)
	return m
}

func (data indexData) Guild() *discord.Guild {
	g, _ := data.Discord.Guild(data.Ticket.GuildID)
	return g
}

func render(w http.ResponseWriter, r *http.Request) {
	ticket := web.GetTicket(r.Context())
	config := web.GetRenderConfig(r.Context())

	err := index.Execute(w, indexData{
		RenderConfig: config,
		Ticket:       ticket,
	})

	if err != nil {
		log.Println("failed to render:", err)
	}
}

// redirect is used to register the ticket token to the browser cookies. The
// redirectee should craft a URL to /redirect with the "token" URL form.
func redirect(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(400)
		return
	}

	web.SetToken(w, token)
	http.Redirect(w, r, "/", http.StatusFound)
}

func getService(r *http.Request, config web.RenderConfig) *web.LMSService {
	var (
		nameHash = chi.URLParam(r, "serviceHash")
		service  = config.FindService(nameHash)
	)

	return service
}

func renderService(w http.ResponseWriter, r *http.Request) {
	// var (
	// 	ticket  = web.GetTicket(r.Context())
	// 	config  = web.GetRenderConfig(r.Context())
	// 	service = getService(r, config)
	// )

	// if service == nil {
	// 	w.WriteHeader(404)
	// 	return
	// }

}

func serviceAuthorizer(w http.ResponseWriter, r *http.Request) {
	var (
		ticket  = web.GetTicket(r.Context())
		config  = web.GetRenderConfig(r.Context())
		service = getService(r, config)
	)

	if service == nil {
		w.WriteHeader(404)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(400)
		return
	}

	methods := service.Authorize()

	session, err := methods.Token.Authorize(token)
	if err != nil {
		// TODO: proper form error
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	user, err := session.User()
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	config.Registerer.RegisterUser(ticket, service, user)
}
