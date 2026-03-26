package view

import (
	"time"

	"github.com/umono-cms/umono/internal/i18n"
	"github.com/umono-cms/umono/internal/reltime"
)

func RelativeTime(t *time.Time) string {
	return RelativeTimeWithTranslator(t, nil)
}

func RelativeTimeWithTranslator(t *time.Time, translator any) string {
	if t == nil {
		return ""
	}

	return reltime.Format(*t, relativeTimeTranslator(translator))
}

func Translate(translator any, key string) string {
	if tr, ok := translator.(interface{ T(string) string }); ok && tr != nil {
		return tr.T(key)
	}
	return i18n.Missing(key)
}

func relativeTimeTranslator(translator any) reltime.Translator {
	if tr, ok := translator.(interface{ T(string) string }); ok && tr != nil {
		return reltime.TranslatorFunc(tr.T)
	}

	return reltime.TranslatorFunc(i18n.Missing)
}
