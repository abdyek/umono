package view

import (
	"testing"
	"time"

	"github.com/umono-cms/umono/internal/i18n"
)

type testTranslator struct {
	values map[string]string
}

func (t testTranslator) T(key string) string {
	if value, ok := t.values[key]; ok {
		return value
	}
	return ""
}

func TestRelativeTimeNil(t *testing.T) {
	if got := RelativeTime(nil); got != "" {
		t.Fatalf("RelativeTime(nil) = %q, want empty string", got)
	}
}

func TestRelativeTimeWithTranslator(t *testing.T) {
	ts := time.Now().Add(-2 * time.Hour)
	translator := testTranslator{
		values: map[string]string{
			"relative_time.just_now":         "just now",
			"relative_time.minute_ago.one":   "1 minute ago",
			"relative_time.minute_ago.other": "{{count}} minutes ago",
			"relative_time.hour_ago.one":     "1 hour ago",
			"relative_time.hour_ago.other":   "{{count}} hours ago",
			"relative_time.day_ago.one":      "1 day ago",
			"relative_time.day_ago.other":    "{{count}} days ago",
			"relative_time.week_ago.one":     "1 week ago",
			"relative_time.week_ago.other":   "{{count}} weeks ago",
			"relative_time.month_ago.one":    "1 month ago",
			"relative_time.month_ago.other":  "{{count}} months ago",
			"relative_time.year_ago.one":     "1 year ago",
			"relative_time.year_ago.other":   "{{count}} years ago",
		},
	}

	got := RelativeTimeWithTranslator(&ts, translator)
	if got != "2 hours ago" {
		t.Fatalf("RelativeTimeWithTranslator() = %q, want %q", got, "2 hours ago")
	}
}

func TestRelativeTimeFallsBackToMissingTranslations(t *testing.T) {
	ts := time.Now().Add(-2 * time.Hour)

	got := RelativeTime(&ts)
	want := i18n.Missing("relative_time.hour_ago.other")

	if got != want {
		t.Fatalf("RelativeTime() = %q, want %q", got, want)
	}
}
