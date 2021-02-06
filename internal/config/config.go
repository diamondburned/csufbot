package config

import (
	"fmt"
	"os"

	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms/canvas"
	"github.com/diamondburned/csufbot/csufbot/lms/moodle"
	"github.com/diamondburned/csufbot/internal/db/badger"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/pkg/errors"

	toml "github.com/pelletier/go-toml"
)

// Config is the application configuration.
type Config struct {
	Site     SiteConfig     `toml:"site"`
	Discord  DiscordConfig  `toml:"discord"`
	Database DatabaseConfig `toml:"database"`
	Services Services       `toml:"services"`
}

// FromFile parses the file and overrides the config.
func (cfg *Config) FromFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(cfg); err != nil {
		return errors.Wrap(err, "failed to parse TOML")
	}

	return nil
}

// SiteConfig describes the configuration for the HTTP server.
type SiteConfig struct {
	Address  string `toml:"address"`
	FrontURL string `toml:"fronturl"`

	SiteName   string `toml:"site_name"`
	Disclaimer string `toml:"disclaimer"`
}

// DiscordConfig describes the configuration for the Discord bot.
type DiscordConfig struct {
	Token  string `toml:"token"`
	Secret string `toml:"secret"`
}

// Open opens a new Discord state connection.
func (dcfg DiscordConfig) Open() (web.DiscordState, error) {
	s, err := state.New("Bot " + dcfg.Token)
	if err != nil {
		return web.DiscordState{}, errors.Wrap(err, "failed to create state")
	}

	s.Gateway.AddIntents(gateway.IntentGuilds)
	s.Gateway.AddIntents(gateway.IntentGuildMembers)

	return web.DiscordState{
		State:  s,
		Secret: dcfg.Secret,
	}, nil
}

// DatabaseConfig describes the configuration for the underlying database
// storage.
type DatabaseConfig struct {
	Name    string `toml:"name"`
	Address string `toml:"address"`
}

// Open opens a new session storer from the given SessionConfig.
func (dbcfg DatabaseConfig) Open() (store csufbot.Store, err error) {
	switch dbcfg.Name {
	case "badger":
		store, err = badger.New(dbcfg.Address)
	case "sqlite":
		panic("TODO")
	case "pgx":
		panic("TODO")
	default:
		err = fmt.Errorf("unknown database %q", dbcfg.Name)
	}

	return
}

// Services describes the types of LMS services.
type Services struct {
	Canvas []Service `toml:"canvas"`
	Moodle []Service `toml:"moodle"`
}

// WebServices converts Services to a list of package web's LMSService.
func (svcs Services) WebServices() []web.LMSService {
	var services []web.LMSService

	for _, svc := range svcs.Canvas {
		services = append(services, web.NewLMSService(
			canvas.New(svc.Name, svc.Host), svc.Instruction,
		))
	}

	for _, svc := range svcs.Moodle {
		services = append(services, web.NewLMSService(
			moodle.New(svc.Name, svc.Host), svc.Instruction,
		))
	}

	return services
}

// Service describes a single hosted LMS service.
type Service struct {
	Name        string `toml:"name"`
	Host        string `toml:"host"`
	Instruction string `toml:"instruction"` // HTML
}
