package web

import (
	"html/template"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
)

func mapInitials() func(rune) rune {
	var last rune

	return func(r rune) rune {
		is := last == 0 || unicode.IsSpace(last)
		last = r

		if is {
			return unicode.ToUpper(r)
		}
		return -1
	}
}

var funcs = template.FuncMap{
	"humanizeTime": humanize.Time,
	"shortError": func(err error) string {
		parts := strings.Split(err.Error(), ": ")
		if len(parts) == 0 {
			return ""
		}

		part := parts[len(parts)-1]

		r, sz := utf8.DecodeRuneInString(part)
		if sz == 0 {
			return ""
		}

		return string(unicode.ToUpper(r)) + part[sz:] + "."
	},
	"initials": func(name string) string {
		upper := strings.Map(mapInitials(), name)
		if upper == "" {
			r, _ := utf8.DecodeRuneInString(name)
			return string(unicode.ToUpper(r))
		}
		if len(upper) > 2 {
			return upper[:2]
		}

		return upper
	},
}
