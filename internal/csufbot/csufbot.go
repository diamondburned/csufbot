package csufbot

import (
	"github.com/diamondburned/csufbot/internal/csufbot/course"
	"github.com/diamondburned/csufbot/internal/csufbot/guild"
	"github.com/diamondburned/csufbot/internal/csufbot/session"
	"github.com/diamondburned/csufbot/internal/csufbot/user"
)

// Store contains database store implementations.
type Store struct {
	Users    user.Storer
	Guilds   guild.Storer
	Courses  course.Storer
	Sessions session.Storer
}
