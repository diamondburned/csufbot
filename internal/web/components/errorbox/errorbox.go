package errorbox

import (
	"net/http"

	"github.com/diamondburned/csufbot/internal/web"
)

var (
	errorbox  = web.Templater.Register("errorbox", "components/errorbox/errorbox.html")
	errorpage = web.Templater.Register("errorpage", "components/errorbox/errorpage.html")
)

type data struct {
	web.RenderConfig
	Code    int
	Status  string
	Error   error
	Referer string
}

// Render renders an error box.
func Render(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.WriteHeader(code)
	errorpage.Execute(w, data{
		RenderConfig: web.GetRenderConfig(r.Context()),
		Code:         code,
		Status:       http.StatusText(code),
		Error:        err,
		Referer:      r.Referer(),
	})
}
