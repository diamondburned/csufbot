package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/diamondburned/csufbot/internal/bot"
	"github.com/diamondburned/csufbot/internal/config"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes"
	"github.com/diamondburned/tmplutil"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
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

	b, err := bot.Start(d.State, storer)
	if err != nil {
		log.Fatalln("failed to create bot:", err)
	}
	defer b.CloseGracefully()

	r := chi.NewRouter()
	r.Mount("/static", web.MountStatic())
	r.Mount("/", routes.Mount(web.RenderConfig{
		FrontURL:   cfg.Site.FrontURL,
		SiteName:   cfg.Site.SiteName,
		Disclaimer: cfg.Site.Disclaimer,
		Services:   cfg.Services.WebServices(),
		Discord:    d,
		Store:      storer,
	}))

	log.Println("Listen and serve at", cfg.Site.Address)
	log.Fatalln(http.ListenAndServe(cfg.Site.Address, r))
}

func configFromGlob(glob string) (*config.Config, error) {
	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, errors.Wrap(err, "glob failed")
	}

	sort.Slice(files, func(i, j int) bool {
		// Parse files with shorter names first.
		if len(files[i]) != len(files[j]) {
			return len(files[i]) < len(files[j])
		}
		// Otherwise, sort alphanumerically.
		return files[i] < files[j]
	})

	var cfg config.Config
	for _, file := range files {
		if err := cfg.FromFile(file); err != nil {
			return nil, errors.Wrapf(err, "failed to parse file %q", file)
		}
	}

	return &cfg, nil
}
