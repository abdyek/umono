package themeapi

import (
	"testing"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/themecore"
)

func TestBuiltinSetInCompono(t *testing.T) {
	cases := []struct {
		kind      string
		themecore []string
		compono   []string
	}{
		{"built-in", themecore.Builtins(), compono.Builtins()},
		{"markdown element", themecore.MarkdownElements(), compono.MarkdownElements()},
	}

	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			if len(tc.themecore) == 0 {
				t.Fatalf("%s: themecore list is empty", tc.kind)
			}
			set := make(map[string]bool, len(tc.compono))
			for _, name := range tc.compono {
				set[name] = true
			}
			for _, name := range tc.themecore {
				if !set[name] {
					t.Errorf("%s %s is in themecore's set but not in Compono: a removal without a Break line", tc.kind, name)
				}
			}
		})
	}
}
