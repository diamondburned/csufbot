package main

import (
	"fmt"
	"os"

	"github.com/diamondburned/csufbot/internal/lms/canvas"
	"github.com/diamondburned/csufbot/internal/lms/moodle"
	"github.com/diamondburned/csufbot/internal/session"
	"github.com/diamondburned/csufbot/internal/session/badger"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/pkg/errors"

	toml "github.com/pelletier/go-toml"
)

// Config is the application configuration.
type Config struct {
	Session  SessionConfig `toml:"session"`
	Services Services      `toml:"services"`
}

func configFromFile(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer f.Close()

	var cfg Config

	if err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to parse TOML")
	}

	return &cfg, nil
}

// SessionConfig describes the configuration for the session storage.
type SessionConfig struct {
	Database string `toml:"database"`
	Address  string `toml:"address"`
}

// Open opens a new session storer from the given SessionConfig.
func (scfg SessionConfig) Open() (storer session.Storer, err error) {
	switch scfg.Database {
	case "badger":
		storer, err = badger.New(scfg.Address)
	default:
		err = fmt.Errorf("unknown database %q", scfg.Database)
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
