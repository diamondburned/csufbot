package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/csufbot"
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

	d, err := cfg.Discord.Open()
	if err != nil {
		log.Fatalln("failed to open Discord:", err)
	}
	defer d.Close()

	storer := newMockStorer()
	storer.addUsers([]csufbot.User{
		{
			ID:       170132746042081280,
			Services: []csufbot.UserInService{},
		},
	})

	r := chi.NewRouter()
	r.Mount("/static", web.MountStatic())
	r.Mount("/", pages.Mount(web.RenderConfig{
		HTTPS:    cfg.HTTP.HTTPS,
		Services: cfg.Services.WebServices(),
		Discord:  d,
		Store:    csufbot.Store{},
	}))

	log.Println("Listen and serve at", cfg.HTTP.Address)
	log.Fatalln(http.ListenAndServe(cfg.HTTP.Address, r))
}

type mockStorer struct {
	mutex sync.Mutex
	users map[discord.UserID]csufbot.User
}

func newMockStorer() *mockStorer {
	return &mockStorer{
		users: map[discord.UserID]csufbot.User{},
	}
}

func (s *mockStorer) addUsers(users []csufbot.User) {
	for _, user := range users {
		s.users[user.ID] = user
	}
}
