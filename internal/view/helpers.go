package view

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/umono-cms/umono/internal/i18n"
)

func RelativeTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return humanize.Time(*t)
}

func Translate(translator any, key string) string {
	if tr, ok := translator.(interface{ T(string) string }); ok && tr != nil {
		return tr.T(key)
	}
	return i18n.Missing(key)
}
