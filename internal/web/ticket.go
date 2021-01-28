package web

import (
	"context"
	"net/http"
	"time"

	"github.com/diamondburned/csufbot/internal/csufbot/session"
)

// RequireTicket requires the token cookie. If the token cookie is not present,
// then a redirection will be done to the given URL; if the URL is empty, then
// 401 is written.
func RequireTicket(cfg RenderConfig, elseRedirectURL string) Middleware {
	notFound := func(w http.ResponseWriter, r *http.Request) {
		if elseRedirectURL == "" {
			w.WriteHeader(401)
			return
		}

		http.Redirect(w, r, elseRedirectURL, http.StatusFound)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := tokenCookie(r)
			if err != nil {
				notFound(w, r)
				return
			}

			ticket, err := cfg.Sessions.FindTicket(token)
			if err != nil {
				notFound(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ticketCtx, ticket)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTicket gets the ticket from the given context. The context must be ran
// through RequireTicket; a ticket not found will panic.
func GetTicket(r context.Context) *session.Ticket {
	ticket, ok := r.Value(ticketCtx).(*session.Ticket)
	if !ok {
		panic("no Ticket in context")
	}
	return ticket
}

// tokenCookie gets the token string from the request's cookies.
func tokenCookie(r *http.Request) (string, error) {
	c, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// SetToken sets the token cookie into the request. If the token is empty, then
// the cookie will be cleared from the client.
func SetToken(w http.ResponseWriter, token string) {
	c := &http.Cookie{
		Name:  "token",
		Path:  "/",
		Value: token,
	}

	if token == "" {
		c.Expires = time.Unix(0, 0)
	}

	http.SetCookie(w, c)
}
