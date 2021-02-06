package bot

import (
	"github.com/diamondburned/arikawa/v2/bot"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/csufbot/csufbot"
)

// root is the root command structure.
type root struct {
	Ctx   *bot.Context
	Store csufbot.Store
}

func (r *root) Setup(cmd *bot.Subcommand) {}

func (r *root) Sync(m *gateway.MessageCreateEvent) (string, error) {
	return m.Author.Mention() + ", ", nil
}
