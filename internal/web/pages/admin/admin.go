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

type ctxKey uint8

const (
	routeDataKey ctxKey = iota
)

func Mount() http.Handler {
	r := chi.NewRouter()

	r.Route("/{guildID}", func(r chi.Router) {
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

type chooseCoursesData struct {
	web.RenderConfig
	routeData
	CourseMap map[lms.CourseID]csufbot.Course
	Services  []csufbot.UserInService
}

func chooseCourses(w http.ResponseWriter, r *http.Request) {
	// TODO: button to link more classes
	cfg := web.GetRenderConfig(r.Context())
	routeData := getRouteData(r.Context())

	u, err := cfg.Users.User(routeData.UserID)
	if err != nil { // Invalid user ID.
		w.WriteHeader(401)
		return
	}

	var courseIDs = make([]lms.CourseID, 0, 10)
	for _, svc := range u.Services {
		courseIDs = append(courseIDs, svc.Enrolled...)
	}

	c, err := cfg.Courses.Courses(courseIDs...)
	if err != nil {
		// Database contains invalid courses for some reason.
		w.WriteHeader(500)
		return
	}

	var courseMap = make(map[lms.CourseID]csufbot.Course, len(c))
	for _, course := range c {
		courseMap[course.ID] = course
	}

	courses.Execute(w, chooseCoursesData{
		RenderConfig: cfg,
		routeData:    routeData,
		CourseMap:    courseMap,
		Services:     u.Services,
	})
}
