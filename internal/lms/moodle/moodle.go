package moodle

import (
	"strconv"

	"github.com/diamondburned/csufbot/internal/lms"
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
	return session{m}, nil
}

type session struct {
	m *moodle.MoodleApi
}

func (s session) User() (*lms.User, error) {
	// I HATE MOODLE. I HATE PHP. I HATE EVERYTHING THAT PHP INFECTS.
	_, first, last, userID, err := s.m.GetSiteInfo()
	if err != nil {
		return nil, err
	}

	person, err := s.m.GetPersonByMoodleId(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get person")
	}

	return &lms.User{
		ID: strconv.FormatInt(userID, 10),
		Name: lms.Name{
			First: first,
			Last:  last,
		},
		Avatar: person.ProfileImageUrl,
	}, nil
}
