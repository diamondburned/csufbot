package user

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/lms"
)

// User describes per-user data and relationships. Note that a user does not
// have to be guild-specific.
type User struct {
	ID discord.UserID
	// Enrolled contains a list of course IDs that the user is enrolled in.
	Enrolled []lms.CourseID
}

// Storer stores an internal database of users and their relationships.
type Storer interface {
	// User gets a user.
	User(id discord.UserID) (*User, error)
	// Register registers a new user with the given user ID and optionally a
	// list of currently enrolled courses.
	Register(id discord.UserID, courseIDs ...lms.CourseID) error
}
