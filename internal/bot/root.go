package bot

import (
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/v2/bot"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/csufbot/internal/config"
)

// root is the root command structure.
type root struct {
	Ctx    *bot.Context
	Config *config.Config
}

func (r *root) Setup(cmd *bot.Subcommand) {}

func (r *root) Sync(m *gateway.MessageCreateEvent) (*discord.Embed, error) {
	return &discord.Embed{
		Description: "[hyperlink test](https://google.com)",
		Fields: []discord.EmbedField{
			{Name: "a", Value: "[a](https://google.com)"},
			{Name: "a", Value: "[a](https://google.com)"},
			{Name: "a", Value: "[a](https://google.com)"},
		},
	}, nil
}

func (r *root) Help() (string, error) {
	return r.Ctx.Help(), nil
}
