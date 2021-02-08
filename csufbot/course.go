package csufbot

import (
	"sort"

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

// MapCoursesToList converts a map of courses to its list. The order of the
// returned list is undefined.
func MapCoursesToList(courseMap map[lms.CourseID]Course) []Course {
	courses := make([]Course, 0, len(courseMap))
	for _, course := range courseMap {
		courses = append(courses, course)
	}
	return courses
}

// SortCourses sorts the given list of courses alphabetically.
func SortCoursesByName(courses []Course) {
	sort.Slice(courses, func(i, j int) bool {
		return courses[i].Name < courses[j].Name
	})
}
