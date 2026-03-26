package reltime

import (
	"testing"
	"time"
)

func TestFormatAt(t *testing.T) {
	now := time.Date(2026, time.March, 26, 12, 0, 0, 0, time.UTC)
	translator := TranslatorFunc(func(key string) string {
		values := map[string]string{
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
		}
		return values[key]
	})

	tests := []struct {
		name string
		when time.Time
		want string
	}{
		{name: "future becomes just now", when: now.Add(time.Minute), want: "just now"},
		{name: "seconds becomes just now", when: now.Add(-59 * time.Second), want: "just now"},
		{name: "one minute", when: now.Add(-1 * time.Minute), want: "1 minute ago"},
		{name: "many minutes", when: now.Add(-59 * time.Minute), want: "59 minutes ago"},
		{name: "one hour", when: now.Add(-1 * time.Hour), want: "1 hour ago"},
		{name: "many hours", when: now.Add(-23 * time.Hour), want: "23 hours ago"},
		{name: "one day", when: now.Add(-24 * time.Hour), want: "1 day ago"},
		{name: "many days", when: now.Add(-6 * 24 * time.Hour), want: "6 days ago"},
		{name: "one week", when: now.Add(-7 * 24 * time.Hour), want: "1 week ago"},
		{name: "many weeks", when: now.Add(-29 * 24 * time.Hour), want: "4 weeks ago"},
		{name: "one month", when: now.Add(-30 * 24 * time.Hour), want: "1 month ago"},
		{name: "many months", when: now.Add(-364 * 24 * time.Hour), want: "12 months ago"},
		{name: "one year", when: now.Add(-365 * 24 * time.Hour), want: "1 year ago"},
		{name: "many years", when: now.Add(-800 * 24 * time.Hour), want: "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAt(tt.when, now, translator)
			if got != tt.want {
				t.Fatalf("FormatAt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatAtWithoutTranslatorFallsBackToKey(t *testing.T) {
	now := time.Date(2026, time.March, 26, 12, 0, 0, 0, time.UTC)

	got := FormatAt(now.Add(-2*time.Hour), now, nil)
	want := "relative_time.hour_ago.other"

	if got != want {
		t.Fatalf("FormatAt() = %q, want %q", got, want)
	}
}
