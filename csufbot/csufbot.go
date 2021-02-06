package csufbot

// Store contains database store interfaces.
type Store struct {
	Users   UserStorer
	Guilds  GuildStorer
	Courses CourseStorer
}
