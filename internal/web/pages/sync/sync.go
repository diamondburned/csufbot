// Package sync contains routes for users to sync courses from LMS services into
// their account.
package sync

import (
	"log"
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/csufbot"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/pages/oauth"
	"github.com/go-chi/chi"
)

var sync = web.Templater.Register("sync", "pages/sync/sync.html")

func Mount() http.Handler {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(oauth.Require)

		r.Get("/", render)
	})

	return r
}

type syncData struct {
	web.RenderConfig
	Client *oauth.UserClient
}

func (data syncData) Me() *discord.User {
	u, _ := data.Client.Me()
	return u
}

func (data syncData) User(id discord.UserID) *csufbot.User {
	user, _ := data.Users.User(id)
	return user
}

func render(w http.ResponseWriter, r *http.Request) {
	err := sync.Execute(w, syncData{
		RenderConfig: web.GetRenderConfig(r.Context()),
		Client:       oauth.Client(r.Context()),
	})

	if err != nil {
		log.Println("failed to render:", err)
	}
}
