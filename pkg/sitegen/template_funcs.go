package sitegen

import (
	"html/template"
	"strings"
	"time"
)

var tmplFuncs = template.FuncMap{
	"truncate": truncate,
	"time":     formatTime,
}

func truncate(s string, length int) string {
	if len(s) > length {
		s = s[:length]
	}
	i := strings.LastIndex(s, " ")
	if i > 0 {
		s = s[:i]
	}

	if len(s) > 0 {
		s += "..."
	}

	return s
}

func formatTime(t time.Time) string {
	return t.Format("Jan 2 2006 @ 3:04 PM")
}
