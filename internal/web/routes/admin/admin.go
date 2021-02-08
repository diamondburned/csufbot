// Package admin contains routes for guild owners to set up their guilds.
package admin

import (
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/admin/adminonly"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/go-chi/chi"
)

var (
	admin = web.Templater.Register("admin", "routes/admin/admin.html")
)

func Mount() http.Handler {
	r := chi.NewRouter()
	r.Use(oauth.Require)

	r.Route("/{guildID}", func(r chi.Router) {
		r.Use(adminonly.Require("guildID", true))

		// r.Get("/courses", chooseCourses)
	})

	r.Get("/", render)

	return r
}

type data struct {
	web.RenderConfig
	Client *oauth.UserClient
}

func (d data) Me() *discord.User {
	u, _ := d.Client.Me()
	return u
}

type guildCourse struct {
	*discord.Guild
	Courses []csufbot.Course
}

func (d data) AdminGuilds() []guildCourse {
	guilds, err := d.Client.Guilds(100)
	if err != nil {
		return nil
	}

	admins := guilds[:0]

	for _, guild := range guilds {
		if guild.Permissions.Has(discord.PermissionAdministrator) {
			admins = append(admins, guild)
		}
	}

	courseOut := make(map[discord.GuildID][]csufbot.Course, len(admins))
	for _, guild := range admins {
		courseOut[guild.ID] = nil
	}

	if err := csufbot.GuildsCourses(d.Store, courseOut); err != nil {
		return nil
	}

	guildCourses := make([]guildCourse, len(admins))
	for i, guild := range admins {
		guildCourses[i] = guildCourse{
			Guild:   &admins[i],
			Courses: courseOut[guild.ID],
		}
	}

	return guildCourses
}

func render(w http.ResponseWriter, r *http.Request) {
	admin.Execute(w, data{
		RenderConfig: web.GetRenderConfig(r.Context()),
		Client:       oauth.Client(r.Context()),
	})
}
