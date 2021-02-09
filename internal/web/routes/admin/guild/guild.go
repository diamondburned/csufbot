package guild

import (
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/components/errorbox"
	"github.com/diamondburned/csufbot/internal/web/routes/admin/guild/adminonly"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

var guild = web.Templater.Register("guild", "routes/admin/guild/guild.html")

func Mount(guildIDParam string) http.Handler {
	r := chi.NewRouter()
	r.Use(adminonly.Require(guildIDParam, true))

	r.Get("/", render)
	r.Post("/refresh", refresh)

	return r
}

func refresh(w http.ResponseWriter, r *http.Request) {
	cacher := adminonly.GetCache(r.Context())
	client := oauth.Client(r.Context())

	cacher.Invalidate(client)

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

func (d *data) MemberRoles() []discord.Role {
	roles, err := d.Discord.Roles(d.Guild.ID)
	d.Error = err
	return roles
}

func render(w http.ResponseWriter, r *http.Request) {
	// TODO: button to link more classes
	cfg := web.GetRenderConfig(r.Context())
	dat := adminonly.GetData(r.Context())

	u, err := cfg.Users.User(dat.UserID)
	if err != nil { // Invalid user ID.
		errorbox.Render(w, r, 401, errors.Wrap(err, "failed to get this user"))
		return
	}

	var courseMap = make(map[lms.CourseID]csufbot.Course, 10)
	for _, svc := range u.Services {
		for _, id := range svc.Enrolled {
			courseMap[id] = csufbot.Course{}
		}
	}

	if err := cfg.Courses.Courses(courseMap); err != nil {
		errorbox.Render(w, r, 500, errors.Wrap(err, "failed to get courses"))
		return
	}

	guild.Execute(w, &data{
		RenderConfig: cfg,
		Data:         dat,
		Client:       oauth.Client(r.Context()),
	})
}
