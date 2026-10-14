package view

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
)

func rng(sl, sc, el, ec int) compono.Range {
	return compono.Range{
		Start: compono.Position{Line: sl, Column: sc},
		End:   compono.Position{Line: el, Column: ec},
	}
}

func TestEditorDiagnostics(t *testing.T) {
	tests := []struct {
		name        string
		diagnostics []compono.Diagnostic
		root        []string
		want        []DiagnosticItem
	}{
		{
			name: "page source",
			diagnostics: []compono.Diagnostic{
				{
					Code:    "unknown-component",
					Message: "The component **FOO** is not defined.",
					Range:   rng(2, 3, 2, 12),
				},
			},
			root: nil,
			want: []DiagnosticItem{
				{
					Code:        "unknown-component",
					Message:     template.HTML("The component <strong>FOO</strong> is not defined."),
					Source:      "",
					StartLine:   2,
					StartColumn: 3,
					EndLine:     2,
					EndColumn:   12,
				},
			},
		},
		{
			name: "page through a global",
			diagnostics: []compono.Diagnostic{
				{
					Code:    "invalid-url",
					Message: "Bad",
					Source:  []string{"LAYOUT"},
					Range:   rng(3, 1, 3, 10),
					Calls: []compono.Call{
						{Name: "LAYOUT", Range: rng(1, 1, 1, 20)},
						{Name: "LINK", Source: []string{"LAYOUT"}, Range: rng(3, 1, 3, 10)},
					},
				},
			},
			root: nil,
			want: []DiagnosticItem{
				{
					Code:        "invalid-url",
					Message:     "Bad",
					Source:      "LAYOUT",
					StartLine:   1,
					StartColumn: 1,
					EndLine:     1,
					EndColumn:   20,
				},
			},
		},
		{
			name: "global source",
			diagnostics: []compono.Diagnostic{
				{
					Code:    "unknown-parameter",
					Message: "Bad",
					Source:  []string{"CARD"},
					Range:   rng(2, 1, 2, 5),
					Calls: []compono.Call{
						{Name: "CARD", Range: rng(1, 1, 1, 11)},
					},
				},
			},
			root: []string{"CARD"},
			want: []DiagnosticItem{
				{
					Code:        "unknown-parameter",
					Message:     "Bad",
					Source:      "CARD",
					StartLine:   2,
					StartColumn: 1,
					EndLine:     2,
					EndColumn:   5,
				},
			},
		},
		{
			name: "sub component of the global",
			diagnostics: []compono.Diagnostic{
				{
					Code:    "unknown-parameter",
					Message: "Bad",
					Source:  []string{"CARD", "SUB"},
					Range:   rng(1, 1, 1, 4),
					Calls: []compono.Call{
						{Name: "CARD", Range: rng(1, 1, 1, 11)},
						{Name: "SUB", Source: []string{"CARD"}, Range: rng(5, 1, 5, 10)},
					},
				},
			},
			root: []string{"CARD"},
			want: []DiagnosticItem{
				{
					Code:        "unknown-parameter",
					Message:     "Bad",
					Source:      "CARD > SUB",
					StartLine:   5,
					StartColumn: 1,
					EndLine:     5,
					EndColumn:   10,
				},
			},
		},
		{
			name: "synthetic call of a global preview",
			diagnostics: []compono.Diagnostic{
				{
					Code:    "missing-argument",
					Message: "Bad",
					Range:   rng(1, 1, 1, 11),
				},
			},
			root: []string{"CARD"},
			want: []DiagnosticItem{},
		},
		{
			name: "another global",
			diagnostics: []compono.Diagnostic{
				{
					Code:    "invalid-url",
					Message: "Bad",
					Source:  []string{"OTHER"},
					Range:   rng(1, 1, 1, 5),
					Calls: []compono.Call{
						{Name: "CARD", Range: rng(1, 1, 1, 11)},
						{Name: "OTHER", Source: []string{"CARD"}, Range: rng(4, 1, 4, 9)},
					},
				},
			},
			root: []string{"CARD"},
			want: []DiagnosticItem{},
		},
		{
			name: "message escaping",
			diagnostics: []compono.Diagnostic{
				{
					Code:    "x",
					Message: "Use **" + "<b>" + "** and **open",
					Range:   rng(1, 1, 1, 2),
				},
			},
			root: nil,
			want: []DiagnosticItem{
				{
					Code:        "x",
					Message:     template.HTML("Use <strong>&lt;b&gt;</strong> and **open"),
					Source:      "",
					StartLine:   1,
					StartColumn: 1,
					EndLine:     1,
					EndColumn:   2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EditorDiagnostics(tt.diagnostics, tt.root)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPublishedDiagnostics(t *testing.T) {
	rows := []models.PublishedDiagnostic{
		{PageID: 7, PageName: "About", Code: "unknown-component", Message: "The component **FOO** is not defined.", Source: ""},
		{PageID: 12, PageName: "Home", Code: "unknown-key", Message: "a < b", Source: "LAYOUT > CARD"},
	}
	want := []PublishedDiagnosticItem{
		{PageName: "About", PageURL: "/admin/site-pages/7", Source: "", Code: "unknown-component", Message: template.HTML("The component <strong>FOO</strong> is not defined.")},
		{PageName: "Home", PageURL: "/admin/site-pages/12", Source: "LAYOUT > CARD", Code: "unknown-key", Message: template.HTML("a &lt; b")},
	}
	got := PublishedDiagnostics(rows)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	if got := PublishedDiagnostics(nil); got == nil || len(got) != 0 {
		t.Fatalf("PublishedDiagnostics(nil) = %#v, want non-nil empty slice", got)
	}
}
