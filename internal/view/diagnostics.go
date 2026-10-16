package view

import (
	"html/template"
	"slices"
	"strconv"
	"strings"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/umono/internal/models"
)

// DiagnosticItem is a Compono diagnostic as an editor lists it. The range is
// where the editor selects it in the edited source.
type DiagnosticItem struct {
	Code        string
	Message     template.HTML
	Source      string
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

// EditorDiagnostics returns the diagnostics an editor shows for the source at
// root: nil for a page, []string{NAME} for the global NAME. A diagnostic is
// shown when its source is root or below it. It is placed at its own range
// when its source is root, otherwise at the first call written in root; one
// without such a call is left out.
func EditorDiagnostics(diagnostics []compono.Diagnostic, root []string) []DiagnosticItem {
	items := []DiagnosticItem{}
	for _, d := range diagnostics {
		if len(d.Source) < len(root) {
			continue
		}
		if !slices.Equal(d.Source[:len(root)], root) {
			continue
		}
		r := d.Range
		if !slices.Equal(d.Source, root) {
			found := false
			for _, c := range d.Calls {
				if slices.Equal(c.Source, root) {
					r = c.Range
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		items = append(items, DiagnosticItem{
			Code:        string(d.Code),
			Message:     diagnosticMessage(d.Message),
			Source:      strings.Join(d.Source, " > "),
			StartLine:   r.Start.Line,
			StartColumn: r.Start.Column,
			EndLine:     r.End.Line,
			EndColumn:   r.End.Column,
		})
	}
	return items
}

// diagnosticMessage escapes message and turns its ** emphasis into <strong>.
// A ** without a closing one stays as it is.
func diagnosticMessage(message string) template.HTML {
	escaped := template.HTMLEscapeString(message)
	parts := strings.Split(escaped, "**")
	var b strings.Builder
	b.WriteString(parts[0])
	last := len(parts) - 1
	for i := 1; i < len(parts); i++ {
		switch {
		case i%2 == 1 && i == last:
			b.WriteString("**")
			b.WriteString(parts[i])
		case i%2 == 1:
			b.WriteString("<strong>")
			b.WriteString(parts[i])
		default:
			b.WriteString("</strong>")
			b.WriteString(parts[i])
		}
	}
	return template.HTML(b.String())
}

// PublishedDiagnosticItem is a stored diagnostic as the list of published
// errors shows it.
type PublishedDiagnosticItem struct {
	PageName string
	PageURL  string
	Source   string
	Code     string
	Message  template.HTML
}

// PublishedDiagnostics returns the items of the list of published errors. A
// diagnostic of the 404 page links to its settings screen and shows
// notFoundTitle when the page has no name.
func PublishedDiagnostics(rows []models.PublishedDiagnostic, notFoundTitle string) []PublishedDiagnosticItem {
	items := []PublishedDiagnosticItem{}
	for _, row := range rows {
		item := PublishedDiagnosticItem{
			PageName: row.PageName,
			PageURL:  "/admin/site-pages/" + strconv.FormatUint(uint64(row.PageID), 10),
			Source:   row.Source,
			Code:     row.Code,
			Message:  diagnosticMessage(row.Message),
		}
		if row.SystemKey != nil && *row.SystemKey == models.NotFoundSystemKey {
			item.PageURL = "/admin/settings/404-page"
			if item.PageName == "" {
				item.PageName = notFoundTitle
			}
		}
		items = append(items, item)
	}
	return items
}
