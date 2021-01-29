// Package admin contains routes for guild owners to set up their guilds.
package admin

import (
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/csufbot"
	"github.com/diamondburned/csufbot/internal/lms"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/pages/oauth"
	"github.com/go-chi/chi"
)

var (
	courses = web.Templater.Register("courses", "pages/admin/courses.html")
)

func Mount() http.Handler {
	r := chi.NewRouter()

	r.Route("{guildID}", func(r chi.Router) {
		r.Use(oauth.Require)
		r.Use(adminOnly)

		r.Get("/courses", chooseCourses)
	})

	return r
}

func guildID(r *http.Request) discord.GuildID {
	s, _ := discord.ParseSnowflake(chi.URLParam(r, "guildID"))
	return discord.GuildID(s)
}

func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		guildID := guildID(r)
		if !guildID.IsValid() {
			w.WriteHeader(404)
			return
		}

		userClient := oauth.Client(r.Context())
		userID, err := userClient.UserID()
		if err != nil {
			// Must be a token error that this fails.
			w.WriteHeader(400)
			return
		}

		cfg := web.GetRenderConfig(r.Context())

		if !csufbot.UserIsAdmin(cfg.Discord.State, userID, guildID) {
			w.WriteHeader(401)
			return
		}
	})
}

type chooseCoursesData struct {
	web.RenderConfig
	Courses []lms.Course
}

func chooseCourses(w http.ResponseWriter, r *http.Request) {
	// TODO: button to link more classes
	cfg := web.GetRenderConfig(r.Context())
	cli := oauth.Client(r.Context())

	userID, err := cli.UserID()
	if err != nil {
		w.WriteHeader(400)
		return
	}

	u, err := cfg.Users.User(userID)
	if err != nil { // Invalid user ID.
		w.WriteHeader(401)
		return
	}

	c, err := cfg.Courses.Courses(u.Enrolled...)
	if err != nil {
		// Database contains invalid courses for some reason.
		w.WriteHeader(500)
		return
	}

	courses.Execute(w, chooseCoursesData{
		RenderConfig: cfg,
		Courses:      c,
	})
}
