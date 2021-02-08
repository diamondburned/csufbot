package courses

import (
	"net/http"

	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/diamondburned/csufbot/internal/web/routes/admin/adminonly"
)

type chooseCoursesData struct {
	web.RenderConfig
	adminonly.Data
	CourseMap map[lms.CourseID]csufbot.Course
	Services  []csufbot.UserInService
}

func chooseCourses(w http.ResponseWriter, r *http.Request) {
	// TODO: button to link more classes
	cfg := web.GetRenderConfig(r.Context())
	dat := adminonly.GetData(r.Context())

	u, err := cfg.Users.User(dat.UserID)
	if err != nil { // Invalid user ID.
		w.WriteHeader(401)
		return
	}

	var courseMap = make(map[lms.CourseID]csufbot.Course, 10)
	for _, svc := range u.Services {
		for _, id := range svc.Enrolled {
			courseMap[id] = csufbot.Course{}
		}
	}

	if err := cfg.Courses.Courses(courseMap); err != nil {
		w.WriteHeader(500)
		return
	}

	// courses.Execute(w, chooseCoursesData{
	// 	RenderConfig: cfg,
	// 	Data:         dat,
	// 	CourseMap:    courseMap,
	// 	Services:     u.Services,
	// })
}
