package canvas

import (
	"strconv"

	"github.com/diamondburned/csufbot/internal/lms"
	"github.com/harrybrwn/go-canvas"
)

type service struct {
	name string
	host string
	icon string
}

var _ lms.IconSetter = (*service)(nil)

// New creates a new Canvas service.
func New(name, host string) lms.Service {
	return service{
		name: name,
		host: host,
		icon: "https://du11hjcvx0uqb.cloudfront.net/br/dist/images/apple-touch-icon-585e5d997d.png",
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
	c := canvas.WithHost(token, auth.host)
	return session{c}, nil
}

type session struct {
	c *canvas.Canvas
}

func (s session) User() (*lms.User, error) {
	user, err := s.c.CurrentUser()
	if err != nil {
		return nil, err
	}

	return &lms.User{
		ID: lms.UserID(strconv.Itoa(user.ID)),
		Name: lms.Name{
			First:     user.Name,
			Preferred: user.ShortName,
		},
		Avatar: user.AvatarURL,
	}, nil
}

func (s session) Courses() ([]lms.Course, error) {
	courses, err := s.c.Courses()
	if err != nil {
		return nil, err
	}

	var lmsCourses = make([]lms.Course, len(courses))
	for i, course := range courses {
		lmsCourses[i] = lms.Course{
			ID:    lms.CourseID(strconv.Itoa(course.ID)),
			Name:  course.Name,
			Start: course.StartAt,
			End:   course.EndAt,
		}
	}

	return lmsCourses, nil
}
