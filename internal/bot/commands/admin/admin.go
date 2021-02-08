package admin

import (
	"strings"

	"github.com/diamondburned/arikawa/v2/bot"
	"github.com/diamondburned/arikawa/v2/bot/extras/middlewares"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/internal/bot/colors"
	"github.com/diamondburned/csufbot/internal/config"
	"github.com/pkg/errors"
)

type Admin struct {
	Ctx    *bot.Context
	Config *config.Config

	store csufbot.Store
}

func (r *Admin) Setup(cmd *bot.Subcommand) {
	// shortcut
	r.store = r.Config.Database.Store

	cmd.SetPlumb(r.Info)
	cmd.AddMiddleware("*", middlewares.AdminOnly(r.Ctx))
}

func (r *Admin) Info(m *gateway.MessageCreateEvent) (*discord.Embed, error) {
	g, err := r.Ctx.Guild(m.GuildID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid guild ID")
	}

	var adminURL = r.Config.Site.FrontURL + "/admin/" + g.ID.String()

	guild, err := r.store.Guilds.Guild(g.ID)
	if err != nil {
		if !errors.Is(err, csufbot.ErrNotFound) {
			return nil, errors.Wrap(err, "failed to get guild from database")
		}

		return &discord.Embed{
			Title: "Registration Required",
			Color: colors.Error,
			URL:   adminURL,
			Description: `It seems like this guild has never been registered before. ` +
				`Click the "Registration Required" link to do so.`,
		}, nil
	}

	courseMap, err := guild.Courses(r.store.Courses)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get courses")
	}

	courses := csufbot.MapCoursesToList(courseMap)
	csufbot.SortCoursesByName(courses)

	var description strings.Builder

	description.WriteString("__Courses__\n")
	for _, course := range courses {
		description.WriteString(course.Name)
		description.WriteByte(' ')
		description.WriteString(guild.RoleMap[course.ID].Mention())
		description.WriteByte('\n')
	}
	description.WriteByte('\n')

	return &discord.Embed{
		Title:       "Info for " + g.Name,
		Color:       colors.OK,
		URL:         adminURL,
		Description: description.String(),
	}, nil
}
