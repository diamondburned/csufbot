// Package sync contains routes for users to sync courses from LMS services into
// their account.
package sync

import (
	"net/http"
	"sort"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/diamondburned/csufbot/internal/web/routes/sync/service"
	"github.com/diamondburned/tmplutil"
	"github.com/go-chi/chi"
)

var sync = web.Templater.Register("sync", "routes//sync/sync.html")

func Mount() http.Handler {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(oauth.Require)
		r.Use(tmplutil.AlwaysFlush)

		r.Get("/", render)
		r.Mount("/{serviceHost}", service.Mount("serviceHost"))
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

// guildCourses extends a Discord guild to add courses as well.
type guildCourses struct {
	*discord.Guild
	Courses []csufbot.Course
}

// GuildsInServices searches for the current user's guilds and returns a list of
// guilds that they're in.
func (data syncData) GuildsInServices() map[lms.Host][]guildCourses {
	guilds, _ := data.Client.Guilds(100)
	if len(guilds) == 0 {
		return nil
	}

	return guildsInServices(guilds, data.RenderConfig)
}

func guildsInServices(guilds []discord.Guild, cfg web.RenderConfig) map[lms.Host][]guildCourses {
	var guildIDCourses = make(map[discord.GuildID][]csufbot.Course, len(guilds))
	for _, guild := range guilds {
		guildIDCourses[guild.ID] = nil
	}

	if err := csufbot.GuildsCourses(cfg.Store, guildIDCourses); err != nil {
		return nil
	}

	var guildMap = make(map[discord.GuildID]int, len(guilds))
	for i, guild := range guilds {
		guildMap[guild.ID] = i
	}

	var hostGuilds = make(map[lms.Host][]guildCourses, len(cfg.Services))

	for _, svc := range cfg.Services {
		guildCoursesList := make([]guildCourses, 0, len(guildIDCourses))

		for guildID, courses := range guildIDCourses {
			// Get a pointer to the guild inside the slice allocated from Guilds.
			guild := &guilds[guildMap[guildID]]

			var filterCourses = make([]csufbot.Course, 0, len(courses))
			for _, course := range courses {
				if course.ServiceHost == svc.Host {
					filterCourses = append(filterCourses, course)
				}
			}

			if len(filterCourses) == 0 {
				continue
			}

			guildCoursesList = append(guildCoursesList, guildCourses{
				Guild:   guild,
				Courses: filterCourses,
			})
		}

		// Sort the guilds alphabetically.
		sort.Slice(guildCoursesList, func(i, j int) bool {
			return guildCoursesList[i].Name < guildCoursesList[j].Name
		})

		hostGuilds[svc.Host] = guildCoursesList
	}

	return hostGuilds
}

func render(w http.ResponseWriter, r *http.Request) {
	sync.Execute(w, syncData{
		RenderConfig: web.GetRenderConfig(r.Context()),
		Client:       oauth.Client(r.Context()),
	})
}
