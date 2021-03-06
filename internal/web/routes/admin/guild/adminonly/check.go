package adminonly

import (
	"context"
	"log"
	"net/http"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/components/errorbox"
	"github.com/diamondburned/csufbot/internal/web/routes/oauth"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

type ctxKey uint8

const (
	routeDataKey ctxKey = iota
	cacheDataKey
)

var adminCacheKey = oauth.NewCacheKey()

// Data contains data that follows after the admin check and contains
// information about the current route.
//
// Routes that use the middleware is guaranteed to have routeData.
type Data struct {
	UserID discord.UserID

	Guild       *discord.Guild
	MemberCount int
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
func Require(routeParam string) web.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, ok := load(w, r, routeParam)
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

func load(w http.ResponseWriter, r *http.Request, routeParam string) (Data, bool) {
	snowflake, err := discord.ParseSnowflake(chi.URLParam(r, routeParam))
	if err != nil {
		errorbox.Render(w, r, 404, errors.Wrap(err, "invalid snowflake"))
		return Data{}, false
	}

	guildID := discord.GuildID(snowflake)

	userClient := oauth.Client(r.Context())

	cache, ok := userClient.Cache[adminCacheKey].(Data)
	if ok && cache.Guild.ID == guildID {
		return cache, true
	}

	userID, err := userClient.UserID()
	if err != nil {
		// Must be a token error that this fails.
		errorbox.Render(w, r, 400, errors.Wrap(err, "failed to get user ID"))
		return Data{}, false
	}

	// Get the guild from the user's perspective.
	guild, err := userClient.Guild(guildID)
	if err != nil {
		errorbox.Render(w, r, 401, errors.Wrap(err, "invalid guild"))
		return Data{}, false
	}

	// We can then check for the right permissions this way.
	if !guild.Permissions.Has(discord.PermissionAdministrator) {
		errorbox.Render(w, r, 401, errors.New("not an administrator"))
		return Data{}, false
	}

	cfg := web.GetRenderConfig(r.Context())

	// Check if we know this guild.
	botGuild, err := cfg.Discord.GuildWithCount(guildID)
	if err != nil {
		errorbox.Render(w, r, 401, errors.New("unknown guild is given"))
		return Data{}, false
	}

	data := Data{
		UserID:      userID,
		Guild:       guild,
		MemberCount: int(botGuild.ApproximateMembers),
	}

	userClient.Cache[adminCacheKey] = data
	return data, true
}
