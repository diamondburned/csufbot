package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/csufbot/internal/csufbot"
	"github.com/diamondburned/csufbot/internal/db/badger"
	"github.com/diamondburned/csufbot/internal/lms/canvas"
	"github.com/diamondburned/csufbot/internal/lms/moodle"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/pkg/errors"

	toml "github.com/pelletier/go-toml"
)

// Config is the application configuration.
type Config struct {
	Discord  DiscordConfig  `toml:"discord"`
	Database DatabaseConfig `toml:"database"`
	Services Services       `toml:"services"`
}

func configFromGlob(glob string) (*Config, error) {
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

	var cfg Config
	for _, file := range files {
		if err := cfg.FromFile(file); err != nil {
			return nil, errors.Wrapf(err, "failed to parse file %q", file)
		}
	}

	return &cfg, nil
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

// DiscordConfig describes the configuration for the Discord bot.
type DiscordConfig struct {
	Token string `toml:"token"`
}

// Open opens a new Discord state connection.
func (dcfg DiscordConfig) Open() (*state.State, error) {
	s, err := state.New("Bot " + dcfg.Token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create state")
	}

	s.Gateway.AddIntents(gateway.IntentGuilds)
	s.Gateway.AddIntents(gateway.IntentGuildMembers)

	if err := s.Open(); err != nil {
		return nil, errors.Wrap(err, "failed to open")
	}

	return s, nil
}

// DatabaseConfig describes the configuration for the underlying database
// storage.
type DatabaseConfig struct {
	Database string `toml:"database"`
	Address  string `toml:"address"`
}

// Open opens a new session storer from the given SessionConfig.
func (dbcfg DatabaseConfig) Open() (store csufbot.Store, err error) {
	switch dbcfg.Database {
	case "badger":
		db, err := badger.Open(dbcfg.Address)
		if err != nil {
			return store, errors.Wrap(err, "failed to open badger")
		}

		store = csufbot.Store{
			Sessions: badger.NewSessionStore(db),
		}

	case "sqlite":
		panic("TODO")
	case "pgx":
		panic("TODO")
	default:
		err = fmt.Errorf("unknown database %q", dbcfg.Database)
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
