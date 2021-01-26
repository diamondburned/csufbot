package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/diamondburned/csufbot/internal/session"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/pages"
	"github.com/go-chi/chi"
)

var (
	configFile = "./config.toml"
)

func init() {
	flag.StringVar(&configFile, "c", configFile, "Path to the TOML config")
	flag.Parse()
}

func main() {
	cfg, err := configFromFile(configFile)
	if err != nil {
		log.Fatalln("failed to get services:", err)
	}

	storer, err := cfg.Session.Open()
	if err != nil {
		log.Fatalln("failed to open the session storer:", err)
	}

	sessionRepo := session.NewRepository(storer)

	r := chi.NewRouter()
	r.Mount("/static", web.MountStatic())
	r.Mount("/", pages.Mount(web.RenderConfig{
		Services: cfg.Services.WebServices(),
		Sessions: sessionRepo,
	}))

	log.Println("Listen and serve at 127.0.0.1:8081")
	log.Fatalln(http.ListenAndServe("127.0.0.1:8081", r))
}
