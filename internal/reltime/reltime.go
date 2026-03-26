package reltime

import (
	"strconv"
	"strings"
	"time"
)

type Translator interface {
	Translate(string) string
}

type TranslatorFunc func(string) string

func (f TranslatorFunc) Translate(key string) string {
	return f(key)
}

func Format(t time.Time, translator Translator) string {
	return FormatAt(t, time.Now(), translator)
}

func FormatAt(t time.Time, now time.Time, translator Translator) string {
	if t.After(now) {
		t = now
	}

	diff := now.Sub(t)
	if diff < time.Minute {
		return translate(translator, "relative_time.just_now")
	}

	type unit struct {
		duration time.Duration
		singular string
		plural   string
	}

	units := []unit{
		{duration: time.Hour * 24 * 365, singular: "relative_time.year_ago.one", plural: "relative_time.year_ago.other"},
		{duration: time.Hour * 24 * 30, singular: "relative_time.month_ago.one", plural: "relative_time.month_ago.other"},
		{duration: time.Hour * 24 * 7, singular: "relative_time.week_ago.one", plural: "relative_time.week_ago.other"},
		{duration: time.Hour * 24, singular: "relative_time.day_ago.one", plural: "relative_time.day_ago.other"},
		{duration: time.Hour, singular: "relative_time.hour_ago.one", plural: "relative_time.hour_ago.other"},
		{duration: time.Minute, singular: "relative_time.minute_ago.one", plural: "relative_time.minute_ago.other"},
	}

	for _, unit := range units {
		count := int(diff / unit.duration)
		if count < 1 {
			continue
		}

		key := unit.plural
		if count == 1 {
			key = unit.singular
		}

		return replaceCount(translate(translator, key), count)
	}

	return translate(translator, "relative_time.just_now")
}

func translate(translator Translator, key string) string {
	if translator == nil {
		return key
	}

	return translator.Translate(key)
}

func replaceCount(value string, count int) string {
	return strings.ReplaceAll(value, "{{count}}", strconv.Itoa(count))
}
