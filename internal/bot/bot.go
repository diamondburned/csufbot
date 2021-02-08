package bot

import (
	"github.com/diamondburned/arikawa/v2/bot"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/csufbot/internal/config"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/diamondburned/csufbot/internal/bot/commands/admin"
)

var endpoint = oauth2.Endpoint{
	AuthURL:   "https://discordapp.com/api/oauth2/authorize",
	TokenURL:  "https://discordapp.com/api/oauth2/token",
	AuthStyle: oauth2.AuthStyleInParams,
}

var oauthScopes = []string{
	"identify", // for user ID
	"guilds",   // for list of guilds
}

type Discord struct {
	*bot.Context
	secret string
}

// Open opens a new Discord state connection.
func Open(cfg *config.Config) (*Discord, error) {
	s, err := state.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create state")
	}

	s.Gateway.Identifier.IdentifyData.Presence = &gateway.UpdateStatusData{
		Activities: []discord.Activity{{
			Name: "csuf!help | " + cfg.Site.FrontURL,
			Type: discord.GameActivity,
		}},
	}

	b, err := bot.New(s, &root{Config: cfg})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bot")
	}

	b.HasPrefix = bot.NewPrefix("csuf!")

	b.AddIntents(b.DeriveIntents())
	b.AddIntents(gateway.IntentGuilds)
	b.AddIntents(gateway.IntentGuildMembers)

	b.RegisterSubcommand(&admin.Admin{Config: cfg})

	b.Name = "CSUFBot"
	b.Description = "A Discord bot to help managing class servers."

	// Bind a handler once forever.
	b.Start()

	if err := b.Open(); err != nil {
		return nil, errors.Wrap(err, "failed to open Discord")
	}

	return &Discord{
		Context: b,
		secret:  cfg.Discord.Secret,
	}, nil
}

// OAuth returns a new OAuth config with the given redirection domain.
func (d *Discord) OAuth(frontURL string) oauth2.Config {
	m, err := d.Me()
	if err != nil {
		return oauth2.Config{}
	}

	return oauth2.Config{
		ClientID:     m.ID.String(),
		ClientSecret: d.secret,
		Endpoint:     endpoint,
		RedirectURL:  frontURL + "/oauth/redirect",
		Scopes:       oauthScopes,
	}
}

func (d *Discord) Close() error { return d.CloseGracefully() }
