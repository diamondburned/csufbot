package oauth

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/components/errorbox"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

func Mount() http.Handler {
	r := chi.NewRouter()
	r.HandleFunc("/redirect", redirect)
	return r
}

// redirect is the endpoint that Discord redirects the user to. It expects a
// "code" URL parameter.
func redirect(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		errorbox.Render(w, r, 401, errors.New("?code not found"))
		return
	}

	cfg := web.GetRenderConfig(r.Context())
	oa := cfg.Discord.OAuth(cfg.FrontURL)

	t, err := oa.Exchange(r.Context(), code, oauth2.AccessTypeOnline)
	if err != nil {
		errorbox.Render(w, r, 401, errors.Wrap(err, "error xchg OAuth"))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "discord",
		Path:     "/",
		Value:    t.AccessToken,
		Expires:  t.Expiry,
		HttpOnly: true,
	})

	redirectTo := "/"

	// Try and get the redirection URL set during Require. If not, redirect to
	// root.
	redirCookie, err := r.Cookie("redirect-after")
	if err == nil {
		redirectTo = redirCookie.Value
		// Wipe the cookie.
		http.SetCookie(w, &http.Cookie{
			Name:    "redirect-after",
			Expires: time.Unix(0, 0),
		})
	}

	http.Redirect(w, r, redirectTo, http.StatusFound)
}

type ctxKey uint8

const (
	oauthClientKey ctxKey = iota
)

// Require marks handlers as requiring Discord authentication. If not, the
// client will be promptly redirected to Discord.
func Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("discord")
		// If we still have the token cookie, then serve as usual.
		if err == nil {
			cli := api.NewClient("Bearer " + c.Value)
			ctx := context.WithValue(r.Context(), oauthClientKey, cli)
			next.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		// Somehow save the current URL into cookies so we can get it after
		// OAuth redirection.
		if r.URL.Path != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "redirect-after",
				Path:     "/oauth/redirect",
				Value:    r.URL.Path,
				HttpOnly: true,
			})
		}

		cfg := web.GetRenderConfig(r.Context())
		oa := cfg.Discord.OAuth(cfg.FrontURL)
		rd := oa.AuthCodeURL("state", oauth2.AccessTypeOnline)
		http.Redirect(w, r, rd, http.StatusFound)
	})
}

// UserClient provides a small cache on top of common self-identifying
// endpoints.
type UserClient struct {
	*api.Client
	userID discord.UserID
	guilds []discord.Guild
}

// UserID gets the current user's ID. This method is not thread-safe, as it
// relies on a stateful cache.
func (c *UserClient) UserID() (discord.UserID, error) {
	if c.userID.IsValid() {
		return c.userID, nil
	}

	u, err := c.Me()
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}

// Me gets the current user. It saves the user ID, so it is not thread-safe.
func (c *UserClient) Me() (*discord.User, error) {
	u, err := c.Client.Me()
	if err != nil {
		return nil, err
	}

	c.userID = u.ID
	return u, nil
}

// Guilds fetches a list of guilds. It is cached, and is therefore not
// thread-safe. It limits itself to 100 guilds.
func (c *UserClient) Guilds() ([]discord.Guild, error) {
	if c.guilds != nil {
		return c.guilds, nil
	}
	var err error
	c.guilds, err = c.Client.Guilds(100)
	return c.guilds, err
}

// Guild overrides the other method to fetch the list and manually search
// through it. This has to be done because Discord does not allow single guild
// ID lookup.
//
// The method limits itself to 100 guilds. It also caches, making it thread
// unsafe.
func (c *UserClient) Guild(guildID discord.GuildID) (*discord.Guild, error) {
	guilds, err := c.Guilds()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get guilds")
	}

	for _, guild := range guilds {
		if guild.ID == guildID {
			return &guild, nil
		}
	}

	return nil, errors.New("guild not found")
}

// Client gets the Discord API client with the OAuth Bearer token from the
// cookies set after a redirection. The function requires the RequireOAuth
// middleware; it panics if the middleware is not used.
func Client(ctx context.Context) *UserClient {
	dc, ok := ctx.Value(oauthClientKey).(*api.Client)
	if !ok {
		log.Panicln("missing api.Client")
	}
	return &UserClient{Client: dc}
}
