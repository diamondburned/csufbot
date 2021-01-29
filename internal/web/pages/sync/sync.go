// Package sync contains routes for users to sync courses from LMS services into
// their account.
package sync

import (
	"log"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/csufbot"
	"github.com/diamondburned/csufbot/internal/lms"
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

		r.Route("/{serviceHost}", func(r chi.Router) {
			r.Get("/", renderLink)
			r.Post("/", postLink)
		})
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

func getService(r *http.Request, cfg web.RenderConfig) *web.LMSService {
	serviceHost := lms.Host(chi.URLParam(r, "serviceHost"))
	return cfg.FindService(serviceHost)
}

func renderLink(w http.ResponseWriter, r *http.Request) {
	cfg := web.GetRenderConfig(r.Context())
	svc := getService(r, cfg)
	if svc == nil {
		w.WriteHeader(404)
		return
	}
}

// postLink is called after the user has submitted their LMS token.
func postLink(w http.ResponseWriter, r *http.Request) {
	cfg := web.GetRenderConfig(r.Context())
	svc := getService(r, cfg)
	if svc == nil {
		w.WriteHeader(404)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(400)
		return
	}

	discordUser := oauth.Client(r.Context())
	userID, err := discordUser.UserID()
	if err != nil {
		w.WriteHeader(400)
		return
	}

	auth := svc.Authorize()

	session, err := auth.Token.Authorize(token)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	courses, err := session.Courses()
	if err != nil {
		w.WriteHeader(400)
		return
	}

	user, err := session.User()
	if err != nil {
		w.WriteHeader(400)
		return
	}

	if err := cfg.Courses.UpsertCourses(courses...); err != nil {
		w.WriteHeader(500)
		return
	}

	userService := csufbot.UserInService{
		User:        *user,
		Enrolled:    make([]lms.CourseID, len(courses)),
		LastSynced:  time.Now(),
		ServiceHost: svc.Host(),
	}

	for i, course := range courses {
		userService.Enrolled[i] = course.ID
	}

	if err := cfg.Users.Sync(userID, userService); err != nil {
		w.WriteHeader(500)
		return
	}
}
