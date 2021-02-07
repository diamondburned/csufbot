package csufbot

import "io"

// Store contains database store interfaces. These store interfaces may
// optionally implement io.Closer.
type Store struct {
	Users   UserStorer
	Guilds  GuildStorer
	Courses CourseStorer
}

// Close tries to close the backing store.
func (s Store) Close() error {
	var fields = []interface{}{
		s.Users,
		s.Guilds,
		s.Courses,
	}

	for _, field := range fields {
		c, ok := field.(io.Closer)
		if ok {
			c.Close()
		}
	}

	return nil
}
