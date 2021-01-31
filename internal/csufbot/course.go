package csufbot

import (
	"github.com/diamondburned/csufbot/internal/lms"
)

// Course represents a course with a service host string attached to it to
// identify its source.
type Course struct {
	lms.Course
	ServiceHost lms.Host
}

// CourseStorer stores an internal database state of known courses.
type CourseStorer interface {
	// Courses searches in bulk for the given list of course IDs. The returned
	// slice must be of equal length to IDs, or the error must not be nil.
	Courses(ids ...lms.CourseID) ([]Course, error)
	// UpsertCourses updates or inserts the given list of courses into the
	// database.
	UpsertCourses(courses ...Course) error
}
