package guild

import (
	"net/http"

	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/config"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/admin/guild/adminonly"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

var guild = web.Templater.Register("guild", "routes/admin/guild/guild.html")

func Mount(guildIDParam string) http.Handler {
	r := chi.NewRouter()
	r.Use(adminonly.Require(guildIDParam))

	r.Get("/", render)
	r.Post("/refresh", refresh)

	return r
}

func refresh(w http.ResponseWriter, r *http.Request) {
	client := oauth.Client(r.Context())
	client.InvalidateCache()

	if referer := r.Referer(); referer != "" {
		http.Redirect(w, r, referer, http.StatusFound)
	}
}

type data struct {
	web.RenderConfig
	adminonly.Data
	Client *oauth.UserClient

	CourseMap map[lms.CourseID]csufbot.Course
	Services  []csufbot.UserInService

	Error error
}

type enrolled struct {
	Service *config.Service
	Courses []csufbot.Course
}

func (d *data) EnrolledCourses() []enrolled {
	u, err := d.Users.User(d.Data.UserID)
	if err != nil { // Invalid user ID.
		d.Error = errors.Wrap(err, "failed to get this user")
		return nil
	}

	var courseMap = make(map[lms.CourseID]csufbot.Course, 10)
	for _, svc := range u.Services {
		for _, id := range svc.Enrolled {
			courseMap[id] = csufbot.Course{}
		}
	}

	if err := d.Courses.Courses(courseMap); err != nil {
		d.Error = errors.Wrap(err, "failed to get courses")
		return nil
	}

	var enrolleds = make([]enrolled, 0, len(u.Services))

	for _, svc := range u.Services {
		service := d.Service(svc.ServiceHost)
		if service == nil {
			// Unknown service, probably removed. Ignore.
			continue
		}

		courses := make([]csufbot.Course, 0, len(svc.Enrolled))
		for _, id := range svc.Enrolled {
			// A course may have been removed; we don't show them.
			course, ok := courseMap[id]
			if ok {
				courses = append(courses, course)
			}
		}

		enrolleds = append(enrolleds, enrolled{
			Service: service,
			Courses: courses,
		})
	}

	return enrolleds
}

func render(w http.ResponseWriter, r *http.Request) {
	// TODO: button to link more classes
	cfg := web.GetRenderConfig(r.Context())
	dat := adminonly.GetData(r.Context())

	guild.Execute(w, &data{
		RenderConfig: cfg,
		Data:         dat,
		Client:       oauth.Client(r.Context()),
	})
}
