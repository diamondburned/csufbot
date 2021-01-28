package course

import "github.com/diamondburned/csufbot/internal/lms"

// Storer stores an internal database state of known courses.
type Storer interface {
	// Courses searches in bulk for the given list of course IDs. The returned
	// slice must be of equal length to IDs, or the error must not be nil.
	Courses(ids ...lms.CourseID) ([]lms.Course, error)
	// UpsertCourses updates or inserts the given list of courses into the
	// database.
	UpsertCourses(courses ...lms.Course) error
}
