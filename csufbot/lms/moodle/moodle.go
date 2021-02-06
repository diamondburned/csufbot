package moodle

import (
	"strconv"

	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/pkg/errors"
	"github.com/zaddok/moodle"
)

type service struct {
	name string
	host string
	icon string
}

var _ lms.IconSetter = (*service)(nil)

// New creates a new Moodle service.
func New(name, host string) lms.Service {
	return service{
		name: name,
		host: host,
		icon: "https://moodle.com/wp-content/uploads/2019/03/cropped-FAV_icon-1-192x192.png",
	}
}

// SetIcon sets the icon. Make an IconSetter interface to access this method.
func (svc *service) SetIcon(url string) {
	svc.icon = url
}

func (svc service) Name() string {
	return svc.name
}

func (svc service) Host() lms.Host {
	return lms.Host(svc.host)
}

func (svc service) Icon() string {
	return svc.icon
}

func (svc service) Authorize() lms.AuthorizationMethods {
	return lms.AuthorizationMethods{
		Token: tokenAuth{host: svc.host},
	}
}

type tokenAuth struct {
	host string
}

func (auth tokenAuth) Authorize(token string) (lms.Session, error) {
	m := moodle.NewMoodleApi(auth.host, token)

	// I HATE MOODLE. I HATE PHP. I HATE EVERYTHING THAT PHP INFECTS.
	_, _, _, userID, err := m.GetSiteInfo()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get info")
	}

	return session{m, userID}, nil
}

type session struct {
	m      *moodle.MoodleApi
	userID int64
}

func (s session) User() (*lms.User, error) {
	person, err := s.m.GetPersonByMoodleId(s.userID)
	if err != nil {
		return nil, err
	}

	return &lms.User{
		ID: lms.UserID(strconv.FormatInt(s.userID, 10)),
		Name: lms.Name{
			First: person.FirstName,
			Last:  person.LastName,
		},
		Avatar: person.ProfileImageUrl,
	}, nil
}

func (s session) Courses() ([]lms.Course, error) {
	courses, err := s.m.GetPersonCourseList(s.userID)
	if err != nil {
		return nil, err
	}

	var lmsCourses = make([]lms.Course, len(courses))
	for i, course := range courses {
		lmsCourse := lms.Course{
			ID:   lms.CourseID(strconv.FormatInt(course.MoodleId, 10)),
			Name: course.Name,
		}

		if course.Start != nil {
			lmsCourse.Start = *course.Start
		}
		if course.End != nil {
			lmsCourse.End = *course.End
		}

		lmsCourses[i] = lmsCourse
	}

	return lmsCourses, nil
}
