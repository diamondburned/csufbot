package bot

import (
	"github.com/diamondburned/arikawa/v2/bot"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/pkg/errors"
)

func Start(s *state.State, store csufbot.Store) (*bot.Context, error) {
	b, err := bot.New(s, &root{Store: store})
	if err != nil {
		return nil, err
	}
	// Bind a handler once forever.
	b.Start()

	if err := b.Open(); err != nil {
		return nil, errors.Wrap(err, "failed to open Discord")
	}

	return b, nil
}
