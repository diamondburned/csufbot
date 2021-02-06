package csufbot

import (
	"github.com/diamondburned/csufbot/csufbot/lms"
)

// Course represents a course with a service host string attached to it to
// identify its source.
type Course struct {
	lms.Course
	ServiceHost lms.Host
}

// CourseStorer stores an internal database state of known courses.
type CourseStorer interface {
	// Course gets a single course.
	Course(id lms.CourseID) (*Course, error)
	// Courses searches in bulk for the given output map of course IDs to
	// courses.
	Courses(out map[lms.CourseID]Course) error
	// UpsertCourses updates or inserts the given list of courses into the
	// database.
	UpsertCourses(courses ...Course) error
}
