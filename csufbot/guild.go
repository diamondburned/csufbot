package csufbot

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/pkg/errors"
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
}

// CourseMap constructs a backwards-lookup map to look up courses from roles.
func (g Guild) CourseMap() map[discord.RoleID]lms.CourseID {
	courseMap := make(map[discord.RoleID]lms.CourseID, len(g.RoleMap))
	for courseID, roleID := range g.RoleMap {
		courseMap[roleID] = courseID
	}
	return courseMap
}

// Courses returns a map of courses registered with the guild.
func (g Guild) Courses(store CourseStorer) (map[lms.CourseID]Course, error) {
	out := make(map[lms.CourseID]Course, len(g.RoleMap))
	for id := range g.RoleMap {
		out[id] = Course{}
	}

	return out, store.Courses(out)
}

// GuildsCourses searches for the enrolled courses of each guild. It writes
// directly to the given output map.
func GuildsCourses(store Store, out map[discord.GuildID][]Course) error {
	courseMap := map[lms.CourseID]Course{}

	for guildID := range out {
		g, err := store.Guilds.Guild(guildID)
		if err != nil {
			return errors.Wrapf(err, "failed to get guild ID %d", guildID)
		}

		coursesList := make([]Course, 0, len(g.RoleMap))
		out[guildID] = coursesList

		for id := range g.RoleMap {
			courseMap[id] = Course{}
			coursesList = append(coursesList, Course{
				Course: lms.Course{ID: id},
			})
		}
	}

	if err := store.Courses.Courses(courseMap); err != nil {
		return errors.Wrap(err, "failed to get courses")
	}

	for _, guildCourses := range out {
		for i, course := range guildCourses {
			guildCourses[i] = courseMap[course.ID]
		}
	}

	return nil
}
