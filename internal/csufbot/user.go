package csufbot

import (
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/csufbot/internal/lms"
)

// User describes per-user data and relationships. Note that a user does not
// have to be guild-specific.
type User struct {
	ID discord.UserID
	// Services is the list of services that the user has previously synced
	// with.
	Services []UserInService
	// Enrolled contains a list of course IDs that the user is enrolled in.
	Enrolled []lms.CourseID
}

// LastSynced returns the last synced time of the user's service. It returns a
// zero-value time if the user has never synced with the given service.
func (u User) LastSynced(host lms.Host) time.Time {
	for _, service := range u.Services {
		if service.ServiceHost == host {
			return service.LastSynced
		}
	}
	return time.Time{}
}

// UserInService describes the user in a service.
type UserInService struct {
	lms.User
	// ServiceHost is the service host that this information is from.
	ServiceHost lms.Host
	// LastSynced is the last time this information was synced.
	LastSynced time.Time
}

// UserStorer stores an internal database of users and their relationships.
type UserStorer interface {
	// User gets a user.
	User(id discord.UserID) (*User, error)
	// Sync synchronizes the given list of courses and the user information from
	// the service with the database.
	Sync(id discord.UserID, svc UserInService, courses []lms.CourseID) error
}

// UserIsAdmin returns true if the given user ID is the owner or administrator
// of the given guild.
func UserIsAdmin(s *state.State, uID discord.UserID, gID discord.GuildID) bool {
	guild, err := s.Guild(gID)
	if err != nil {
		return false
	}

	member, err := s.Member(gID, uID)
	if err != nil {
		return false
	}

	perms := discord.CalcOverwrites(*guild, discord.Channel{}, *member)
	return perms.Has(discord.PermissionAdministrator)
}