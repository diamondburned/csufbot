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
	"github.com/diamondburned/tmplutil"
	"github.com/go-chi/chi"
)

var (
	configFile = "./config*.toml"
)

func init() {
	tmplutil.Log = true

	flag.StringVar(&configFile, "c", configFile, "Path to the TOML config")
	flag.Parse()
}

func main() {
	cfg, err := configFromGlob(configFile)
	if err != nil {
		log.Fatalln("failed to get services:", err)
	}

	storer, err := cfg.Database.Open()
	if err != nil {
		log.Fatalln("failed to open the session storer:", err)
	}

	d, err := cfg.Discord.Open()
	if err != nil {
		log.Fatalln("failed to open Discord:", err)
	}
	defer d.Close()

	r := chi.NewRouter()
	r.Mount("/static", web.MountStatic())
	r.Mount("/", pages.Mount(web.RenderConfig{
		HTTPS:      cfg.Site.HTTPS,
		SiteName:   cfg.Site.SiteName,
		Disclaimer: cfg.Site.Disclaimer,
		Services:   cfg.Services.WebServices(),
		Discord:    d,
		Store:      storer,
	}))

	log.Println("Listen and serve at", cfg.Site.Address)
	log.Fatalln(http.ListenAndServe(cfg.Site.Address, r))
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
