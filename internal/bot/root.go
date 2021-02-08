package bot

import (
	"strings"

	"github.com/diamondburned/arikawa/v2/bot"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/csufbot/internal/bot/colors"
	"github.com/diamondburned/csufbot/internal/config"
)

// root is the root command structure.
type root struct {
	Ctx    *bot.Context
	Config *config.Config
}

func (r *root) Setup(cmd *bot.Subcommand) {
	cmd.ChangeCommandInfo(r.Help, "", "print this help message.")
	cmd.ChangeCommandInfo(r.Sync, "", "synchronize new courses to your account.")
}

func (r *root) Help(m *gateway.MessageCreateEvent) (string, error) {
	return r.Ctx.Help(), nil
}

func (r *root) Sync(m *gateway.MessageCreateEvent) (*discord.Embed, error) {
	embedDesc := strings.Builder{}
	embedDesc.WriteString("__Quick links:__\n")

	for _, svc := range r.Config.Services {
		embedDesc.WriteString("- ")
		embedDesc.WriteByte('[')
		embedDesc.WriteString(svc.Name)
		embedDesc.WriteByte(']')

		embedDesc.WriteByte('(')
		embedDesc.WriteString(r.Config.Site.FrontURL)
		embedDesc.WriteString("/sync/")
		embedDesc.WriteString(string(svc.Host))
		embedDesc.WriteByte(')')

		embedDesc.WriteByte('\n')
	}

	return &discord.Embed{
		Title:       "Synchronize Courses",
		Color:       colors.Info,
		URL:         r.Config.Site.FrontURL + "/sync",
		Description: embedDesc.String(),
	}, nil
}
