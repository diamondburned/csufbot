package csufbot

// Store contains database store implementations.
type Store struct {
	Users   UserStorer
	Guilds  GuildStorer
	Courses CourseStorer
}
