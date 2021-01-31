package service

import (
	"net/http"
	"time"

	"github.com/diamondburned/csufbot/internal/csufbot"
	"github.com/diamondburned/csufbot/internal/lms"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/pages/oauth"
	"github.com/pkg/errors"
)

var post = web.Templater.Register("post", "pages/sync/service/post.html")

type postSyncData struct {
	web.RenderConfig
	Service *web.LMSService
	done    chan error
}

// postSync is called after the user has submitted their LMS token.
func postSync(w http.ResponseWriter, r *http.Request) {
	cfg := web.GetRenderConfig(r.Context())

	svc := getService(r.Context())
	if svc == nil {
		w.WriteHeader(404)
		return
	}

	done := make(chan error, 1)
	go func() { done <- processSync(r, cfg, svc) }()

	post.Execute(w, postSyncData{
		RenderConfig: cfg,
		Service:      svc,
		done:         done,
	})
}

func (d *postSyncData) Wait() error {
	return <-d.done
}

func processSync(r *http.Request, cfg web.RenderConfig, svc *web.LMSService) error {
	token := r.FormValue("token")
	if token == "" {
		return errors.New("missing token")
	}

	discordUser := oauth.Client(r.Context())
	userID, err := discordUser.UserID()
	if err != nil {
		return errors.Wrap(err, "failed to get user ID")
	}

	auth := svc.Authorize()

	session, err := auth.Token.Authorize(token)
	if err != nil {
		return errors.Wrap(err, "invalid token")
	}

	courses, err := session.Courses()
	if err != nil {
		return errors.Wrap(err, "failed to get courses")
	}

	user, err := session.User()
	if err != nil {
		return errors.Wrap(err, "failed to get this user")
	}

	enrolledIDs := make([]lms.CourseID, len(courses))
	newCourses := make([]csufbot.Course, len(courses))

	for i, course := range courses {
		enrolledIDs[i] = course.ID
		newCourses[i] = csufbot.Course{
			Course:      course,
			ServiceHost: svc.Host(),
		}
	}

	if err := cfg.Courses.UpsertCourses(newCourses...); err != nil {
		return errors.Wrap(err, "failed to update courses")
	}

	userService := csufbot.UserInService{
		User:        *user,
		Enrolled:    enrolledIDs,
		LastSynced:  time.Now(),
		ServiceHost: svc.Host(),
	}

	if err := cfg.Users.Sync(userID, userService); err != nil {
		return errors.Wrap(err, "failed to sync")
	}

	return nil
}