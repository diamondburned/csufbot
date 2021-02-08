package csufbot

import (
	"errors"
	"io"
)

// ErrNotFound can be returned if storers can't find the requested resource.
var ErrNotFound = errors.New("not found")

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
