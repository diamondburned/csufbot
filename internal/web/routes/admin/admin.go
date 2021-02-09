// Package admin contains routes for guild owners to set up their guilds.
package admin

import (
	"net/http"
	"sort"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/admin/guild"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/diamondburned/tmplutil"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

var (
	admin = web.Templater.Register("admin", "routes/admin/admin.html")
)

func Mount() http.Handler {
	r := chi.NewRouter()
	r.Use(oauth.Require)
	r.Use(tmplutil.AlwaysFlush)

	r.Mount("/{guildID}", guild.Mount("guildID"))
	r.Get("/", render)

	return r
}

type data struct {
	web.RenderConfig

	Client *oauth.UserClient

	Error error

	HideUnregistered bool
}

func (d *data) Me() *discord.User {
	u, err := d.Client.Me()
	d.Error = err
	return u
}

type guildCourse struct {
	*discord.Guild
	Courses []csufbot.Course
}

func (d *data) AdminGuilds() []guildCourse {
	guilds, err := d.Client.Guilds()
	if err != nil {
		d.Error = errors.Wrap(err, "failed to get guilds")
		return nil
	}

	admins := guilds[:0]

	for _, guild := range guilds {
		if guild.Permissions.Has(discord.PermissionAdministrator) {
			admins = append(admins, guild)
		}
	}

	if len(admins) == 0 {
		return nil
	}

	courseOut := make(map[discord.GuildID][]csufbot.Course, len(admins))
	for _, guild := range admins {
		courseOut[guild.ID] = nil
	}

	if err := csufbot.GuildsCourses(d.Store, courseOut); err != nil {
		d.Error = errors.Wrap(err, "failed to get courses")
		return nil
	}

	// Sort newest guilds first.
	sort.Slice(admins, func(i, j int) bool {
		return admins[i].ID > admins[j].ID
	})

	guildCourses := make([]guildCourse, 0, len(admins))
	for i, guild := range admins {
		coursesL := courseOut[guild.ID]
		if d.HideUnregistered && len(coursesL) == 0 {
			continue
		}

		guildCourses = append(guildCourses, guildCourse{
			Guild:   &admins[i],
			Courses: coursesL,
		})
	}

	return guildCourses
}

func render(w http.ResponseWriter, r *http.Request) {
	admin.Execute(w, &data{
		RenderConfig: web.GetRenderConfig(r.Context()),
		Client:       oauth.Client(r.Context()),

		HideUnregistered: r.FormValue("hide_unregistered") == "1",
	})
}
