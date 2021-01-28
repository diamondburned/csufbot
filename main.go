package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/diamondburned/csufbot/internal/csufbot/session"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/pages"
	"github.com/go-chi/chi"
)

var (
	configFile = "./config*.toml"
)

func init() {
	flag.StringVar(&configFile, "c", configFile, "Path to the TOML config")
	flag.Parse()
}

func main() {
	cfg, err := configFromGlob(configFile)
	if err != nil {
		log.Fatalln("failed to get services:", err)
	}

	// 	storer, err := cfg.Session.Open()
	// 	if err != nil {
	// 		log.Fatalln("failed to open the session storer:", err)
	// 	}

	state, err := cfg.Discord.Open()
	if err != nil {
		log.Fatalln("failed to open Discord:", err)
	}
	defer state.Close()

	storer := newMockStorer([]session.Ticket{
		{
			Token:   "a",
			GuildID: 803771406197325844,
			UserID:  170132746042081280,
		},
	})

	sessionRepo := session.NewRepository(storer)

	r := chi.NewRouter()
	r.Mount("/static", web.MountStatic())
	r.Mount("/", pages.Mount(web.RenderConfig{
		Services: cfg.Services.WebServices(),
		Discord:  state,
		Sessions: sessionRepo,
	}))

	log.Println("Listen and serve at 127.0.0.1:8081")
	log.Fatalln(http.ListenAndServe("127.0.0.1:8081", r))
}

type mockStorer struct {
	mutex   sync.Mutex
	tickets map[string]session.Ticket
}

func newMockStorer(tickets []session.Ticket) session.Storer {
	var ticketMap = make(map[string]session.Ticket, len(tickets))

	for _, ticket := range tickets {
		ticketMap[ticket.Token] = ticket
	}

	return &mockStorer{tickets: ticketMap}
}

func (s *mockStorer) InsertTicket(t *session.Ticket) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.tickets[t.Token]; ok {
		return session.ErrCollidingToken
	}

	s.tickets[t.Token] = *t
	return nil
}

func (s *mockStorer) FindTicket(token string) (*session.Ticket, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	tk, ok := s.tickets[token]
	if !ok {
		return nil, session.ErrTicketNotFound
	}

	return &tk, nil
}

func (s *mockStorer) InvalidateTicket(token string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.tickets, token)
}
