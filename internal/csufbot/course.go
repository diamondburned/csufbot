package csufbot

import "github.com/diamondburned/csufbot/internal/lms"

type Course struct {
	lms.Course
	Service string
}

// CourseStorer stores an internal database state of known courses.
type CourseStorer interface {
	// Courses searches in bulk for the given list of course IDs. The returned
	// slice must be of equal length to IDs, or the error must not be nil.
	Courses(ids ...lms.CourseID) ([]lms.Course, error)
	// UpsertCourses updates or inserts the given list of courses into the
	// database.
	UpsertCourses(courses ...lms.Course) error
}
