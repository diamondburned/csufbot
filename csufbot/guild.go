package csufbot

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot/lms"
)

// Guild contains per-guild relationships.
type Guild struct {
	ID discord.GuildID
	// RoleMap maps each course to its appropriate role.
	RoleMap map[lms.CourseID]discord.RoleID
}

// GuildStorer stores guild relationships and states.
type GuildStorer interface {
	// Guild gets a guild.
	Guild(id discord.GuildID) (*Guild, error)
	// AddCourses adds courses into the guild wherein each course must be mapped
	// to a role. The courses must already be added into the database through
	// CourseStorer.
	SetCourses(guild discord.GuildID, roleMap map[lms.CourseID]discord.RoleID) error
	// GuildCourses searches for the enrolled courses of each guild. It writes
	// directly to the given output map.
	GuildCourses(courses CourseStorer, out map[discord.GuildID][]Course) error
}

// CourseMap constructs a backwards-lookup map to look up courses from roles.
func (g Guild) CourseMap() map[discord.RoleID]lms.CourseID {
	courseMap := make(map[discord.RoleID]lms.CourseID, len(g.RoleMap))
	for courseID, roleID := range g.RoleMap {
		courseMap[roleID] = courseID
	}
	return courseMap
}
