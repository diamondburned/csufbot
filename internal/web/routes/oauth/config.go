package oauth

import (
	"log"
	"net/http"

	"github.com/diamondburned/csufbot/internal/web"
	"golang.org/x/oauth2"
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

func config(r *http.Request) oauth2.Config {
	cfg := web.GetRenderConfig(r.Context())

	me, err := cfg.Discord.Me()
	if err != nil {
		// If this errors out, then there's something seriously wrong with the
		// state.
		log.Panicln("failed to bot information:", err)
	}

	return oauth2.Config{
		ClientID:     me.ID.String(),
		ClientSecret: cfg.Discord.Secret,
		Endpoint:     endpoint,
		RedirectURL:  cfg.FrontURL + "/oauth/redirect",
		Scopes:       oauthScopes,
	}
}
