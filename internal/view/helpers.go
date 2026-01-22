package view

import (
	"time"

	"github.com/dustin/go-humanize"
)

func RelativeTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return humanize.Time(*t)
}
