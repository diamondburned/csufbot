package admin

import (
	"context"
	"log"
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/pages/oauth"
)

// routeData contains data that follows after the admin check and contains
// information about the current route.
//
// Routes that use the adminOnly middleware is guaranteed to have routeData.
type routeData struct {
	UserID discord.UserID
	Guild  *discord.Guild
	Member *discord.Member
}

func getRouteData(ctx context.Context) routeData {
	rd, ok := ctx.Value(routeDataKey).(routeData)
	if !ok {
		log.Panicln("missing routeData after adminOnly")
	}
	return rd
}

func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		guildID := guildID(r)
		if !guildID.IsValid() {
			w.WriteHeader(404)
			return
		}

		userClient := oauth.Client(r.Context())
		userID, err := userClient.UserID()
		if err != nil {
			// Must be a token error that this fails.
			w.WriteHeader(400)
			return
		}

		cfg := web.GetRenderConfig(r.Context())

		guild, err := cfg.Discord.Guild(guildID)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		member, err := cfg.Discord.Member(guildID, userID)
		if err != nil {
			w.WriteHeader(401)
			return
		}

		perms := discord.CalcOverwrites(*guild, discord.Channel{}, *member)
		if !perms.Has(discord.PermissionAdministrator) {
			w.WriteHeader(401)
			return
		}

		ctx := context.WithValue(r.Context(), routeDataKey, routeData{
			UserID: userID,
			Guild:  guild,
			Member: member,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
