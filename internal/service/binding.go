package service

import (
	"fmt"
	"strings"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/models"
)

// BindingErrorKind is a reason the layout binding of a page cannot be saved.
type BindingErrorKind string

const (
	// BindingUnboundCall: a page without a binding calls UMONO_LAYOUT.
	BindingUnboundCall BindingErrorKind = "unbound-call"
	// BindingRoot: the root of a bound page is not a single block
	// UMONO_LAYOUT call.
	BindingRoot BindingErrorKind = "root"
	// BindingLocalCall: a local component of the page calls UMONO_LAYOUT.
	BindingLocalCall BindingErrorKind = "local-call"
	// BindingUnknownParameter: an argument of the call is no parameter of
	// the bound layout.
	BindingUnknownParameter BindingErrorKind = "unknown-parameter"
	// BindingMissingParameter: a required parameter of the layout is not
	// given.
	BindingMissingParameter BindingErrorKind = "missing-parameter"
	// BindingNotOffered: a new binding is not to the group's offered record.
	BindingNotOffered BindingErrorKind = "not-offered"
	// BindingLayoutNotFound: the bound record has no such layout.
	BindingLayoutNotFound BindingErrorKind = "layout-not-found"
)

// BindingError is a reason the layout binding of a page cannot be saved.
type BindingError struct {
	Kind BindingErrorKind
	// Names are the parameters the error is about.
	Names []string
	// Offered are, of an unknown parameter's Names, those that the layout of
	// the group's offered record has. Binding the layout again moves the page
	// to that record.
	Offered []string
	// Skeleton is a call of the layout with every parameter, for the root
	// error, to copy into the content.
	Skeleton string
}

func (e *BindingError) Error() string {
	if len(e.Names) == 0 {
		return "layout binding: " + string(e.Kind)
	}
	return "layout binding: " + string(e.Kind) + ": " + strings.Join(e.Names, ", ")
}

// CheckBinding runs the save checks of page's layout binding. previous is the
// stored page, nil for a new one: a page whose ContractID does not change
// stays on its record, a new binding must be to the group's offered record.
// Types and values are not checked. The checks reject saving only; the
// render and the previews convert any content.
func (s *ThemeService) CheckBinding(page models.SitePage, previous *models.SitePage) error {
	analysis, err := compono.Analyze([]byte(page.Content))
	if err != nil {
		return err
	}
	var layoutRefs []compono.Ref
	for _, ref := range analysis.Refs {
		if ref.Name == layoutGlobal && ref.Kind != "default" {
			layoutRefs = append(layoutRefs, ref)
		}
	}

	if page.ContractID == nil || page.LayoutID == nil {
		if len(layoutRefs) > 0 {
			return &BindingError{Kind: BindingUnboundCall}
		}
		return nil
	}

	record, ok := s.contracts.ByID(*page.ContractID)
	if !ok {
		return fmt.Errorf("%w: %d", ErrContractNotFound, *page.ContractID)
	}
	if previous == nil || previous.ContractID == nil || *previous.ContractID != *page.ContractID {
		active, err := s.themes.GetActiveThemes()
		if err != nil {
			return err
		}
		offered, err := s.contracts.Offered(record.Group(), active)
		if err != nil {
			return err
		}
		if offered.ID != record.ID {
			return &BindingError{Kind: BindingNotOffered}
		}
	}

	params, err := s.layoutParams(record, *page.LayoutID)
	if err != nil {
		return err
	}

	for _, ref := range layoutRefs {
		if ref.In != "" {
			return &BindingError{Kind: BindingLocalCall}
		}
	}

	if len(analysis.Root) != 1 || analysis.Root[0].Kind != "call" {
		return &BindingError{Kind: BindingRoot, Skeleton: layoutSkeleton(params)}
	}
	call := analysis.Refs[analysis.Root[0].Ref]
	if call.Name != layoutGlobal || !call.Block {
		return &BindingError{Kind: BindingRoot, Skeleton: layoutSkeleton(params)}
	}

	defined := map[string]bool{}
	for _, p := range params {
		defined[p.Name] = true
	}
	var unknown []string
	for _, name := range call.Args {
		if !defined[name] {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		e := &BindingError{Kind: BindingUnknownParameter, Names: unknown}
		e.Offered = s.offeredParams(record, *page.LayoutID, unknown)
		return e
	}

	given := map[string]bool{}
	for _, name := range call.Args {
		given[name] = true
	}
	var missing []string
	for _, p := range params {
		if p.Required && !given[p.Name] {
			missing = append(missing, p.Name)
		}
	}
	if len(missing) > 0 {
		return &BindingError{Kind: BindingMissingParameter, Names: missing}
	}
	return nil
}

// layoutParams returns the parameters of the layout id of record, from its
// source's parameter line.
func (s *ThemeService) layoutParams(record models.Contract, id string) ([]compono.Param, error) {
	contract, err := s.parsedContract(record)
	if err != nil {
		return nil, err
	}
	layout, ok := findLayout(contract, id)
	if !ok {
		return nil, &BindingError{Kind: BindingLayoutNotFound, Names: []string{id}}
	}
	analysis, err := compono.Analyze([]byte(layout.Source))
	if err != nil {
		return nil, err
	}
	return analysis.Params, nil
}

// offeredParams returns those of names that the layout id has in the offered
// record of record's group.
func (s *ThemeService) offeredParams(record models.Contract, id string, names []string) []string {
	active, err := s.themes.GetActiveThemes()
	if err != nil {
		return nil
	}
	offered, err := s.contracts.Offered(record.Group(), active)
	if err != nil || offered.ID == record.ID {
		return nil
	}
	params, err := s.layoutParams(offered, id)
	if err != nil {
		return nil
	}
	has := map[string]bool{}
	for _, p := range params {
		has[p.Name] = true
	}
	var found []string
	for _, name := range names {
		if has[name] {
			found = append(found, name)
		}
	}
	return found
}

func findLayout(c themecore.Contract, id string) (themecore.Layout, bool) {
	for _, l := range c.Layouts {
		if l.ID == id {
			return l, true
		}
	}
	return themecore.Layout{}, false
}

// layoutSkeleton returns a call of UMONO_LAYOUT with every parameter and an
// empty value of its type; a Component parameter gets a component named
// after it, for example content = CONTENT.
func layoutSkeleton(params []compono.Param) string {
	var b strings.Builder
	b.WriteString("{{ " + layoutGlobal)
	for _, p := range params {
		b.WriteString(" " + p.Name + " = " + emptyValue(p))
	}
	b.WriteString(" }}")
	return b.String()
}

// quotes is an empty string literal of Compono.
const quotes = quote + quote

func emptyValue(p compono.Param) string {
	switch p.Type {
	case "integer":
		return "0"
	case "bool":
		return "false"
	case "component":
		return strings.ToUpper(strings.ReplaceAll(p.Name, "-", "_"))
	case "array":
		return "[]"
	case "record":
		return "{}"
	default:
		return quotes
	}
}
