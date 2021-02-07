package config

import (
	"fmt"
	"html/template"
	"os"

	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/csufbot/lms/canvas"
	"github.com/diamondburned/csufbot/csufbot/lms/moodle"
	"github.com/diamondburned/csufbot/internal/db/badger"
	"github.com/pkg/errors"

	toml "github.com/pelletier/go-toml"
)

// Site describes the configuration for the HTTP server.
type Site struct {
	Address  string `toml:"address"`
	FrontURL string `toml:"fronturl"`

	SiteName   string `toml:"site_name"`
	Disclaimer string `toml:"disclaimer"`
}

// Config is the application configuration.
type Config struct {
	Site     Site      `toml:"site"`
	Discord  Discord   `toml:"discord"`
	Database Database  `toml:"database"`
	Services []Service `toml:"services"`
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

	// verify
	for _, s := range cfg.Services {
		if s.LMS != nil {
			continue
		}

		switch s.Type {
		case "canvas":
			s.LMS = canvas.New(s.Host)
		case "moodle":
			s.LMS = moodle.New(s.Host)
		default:
			return fmt.Errorf("unknown type %q for service %s", s.Type, s.Name)
		}
	}

	return nil
}

// Discord describes the configuration for the Discord bot.
type Discord struct {
	Token  string `toml:"token"`
	Secret string `toml:"secret"`
}

// Database describes the configuration for the underlying database storage.
type Database struct {
	Name    string `toml:"name"`
	Address string `toml:"address"`

	csufbot.Store `toml:"-"`
}

// Open opens a new session storer from the given SessionConfig.
func (dbcfg *Database) Open() (err error) {
	switch dbcfg.Name {
	case "badger":
		dbcfg.Store, err = badger.New(dbcfg.Address)
	default:
		err = fmt.Errorf("unknown database %q", dbcfg.Name)
	}

	return
}

// Service describes a single hosted LMS service.
type Service struct {
	Name string `toml:"name"`
	Type string `toml:"type"`
	Icon string `toml:"icon"` // URL

	Host        lms.Host      `toml:"host"`
	Instruction template.HTML `toml:"instruction"` // HTML

	LMS lms.Service `toml:"-"`
}
