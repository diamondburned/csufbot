package adminonly

import (
	"context"
	"log"
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/go-chi/chi"
)

type ctxKey uint8

const (
	routeDataKey ctxKey = iota
)

// Data contains data that follows after the admin check and contains
// information about the current route.
//
// Routes that use the middleware is guaranteed to have routeData.
type Data struct {
	UserID discord.UserID
	Guild  *discord.Guild
	Member *discord.Member
}

func GetData(ctx context.Context) Data {
	rd, ok := ctx.Value(routeDataKey).(Data)
	if !ok {
		log.Panicln("missing routeData after adminOnly")
	}
	return rd
}

// Require requires that the current OAuth user has the administrator permission
// for the current guild. It requires oauth.Require.
func Require(routeParam string, cached bool) web.Middleware {
	if cached {
		return cachedRequire(routeParam)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, ok := fetchData(w, r, routeParam)
			if !ok {
				return
			}

			next.ServeHTTP(w, setData(r, data))
		})
	}
}

func setData(r *http.Request, data Data) *http.Request {
	ctx := context.WithValue(r.Context(), routeDataKey, data)
	return r.WithContext(ctx)
}

func fetchData(w http.ResponseWriter, r *http.Request, routeParam string) (Data, bool) {
	snowflake, err := discord.ParseSnowflake(chi.URLParam(r, routeParam))
	if err != nil {
		w.WriteHeader(404)
		return Data{}, false
	}

	guildID := discord.GuildID(snowflake)

	userClient := oauth.Client(r.Context())
	userID, err := userClient.UserID()
	if err != nil {
		// Must be a token error that this fails.
		w.WriteHeader(400)
		return Data{}, false
	}

	cfg := web.GetRenderConfig(r.Context())

	guild, err := cfg.Discord.Guild(guildID)
	if err != nil {
		w.WriteHeader(401)
		return Data{}, false
	}

	if !guild.Permissions.Has(discord.PermissionAdministrator) {
		w.WriteHeader(401)
		return Data{}, false
	}

	member, err := cfg.Discord.Member(guildID, userID)
	if err != nil {
		w.WriteHeader(401)
		return Data{}, false
	}

	return Data{
		UserID: userID,
		Guild:  guild,
		Member: member,
	}, true
}
