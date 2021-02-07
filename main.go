package main

import (
	"flag"
	"log"
	"path/filepath"
	"sort"
	"time"

	"github.com/diamondburned/csufbot/internal/bot"
	"github.com/diamondburned/csufbot/internal/config"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes"
	"github.com/diamondburned/gracehttp"
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

	if err := cfg.Database.Open(); err != nil {
		log.Fatalln("failed to open the session storer:", err)
	}
	defer cfg.Database.Close()

	d, err := bot.Open(cfg)
	if err != nil {
		log.Fatalln("failed to open Discord:", err)
	}
	defer d.Close()

	r := chi.NewRouter()
	r.Mount("/static", web.MountStatic())
	r.Mount("/", routes.Mount(web.RenderConfig{
		Store:    cfg.Database.Store,
		Site:     cfg.Site,
		Discord:  d,
		Services: cfg.Services,
	}))

	log.Println("Listen and serve at", cfg.Site.Address)

	s := gracehttp.ListenAndServeAsync(cfg.Site.Address, r)
	defer s.ShutdownTimeout(5 * time.Second)

	gracehttp.WaitForInterrupt()
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
