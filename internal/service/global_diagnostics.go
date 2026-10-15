package service

import (
	"sort"

	"github.com/umono-cms/compono"
)

// requiredBodyDiagnostics returns the diagnostics of the body of the global
// name when its parameter line has a required parameter. The preview calls
// the global without arguments, so such a call drops with missing-argument
// and the body is never rendered. They are the diagnostics of
// compono.Analyze and an unknown-component for every call or value reference
// whose name resolves to no local, no built-in, no loaded global and not to
// name itself. Each has name as its source. They are sorted by start offset.
func requiredBodyDiagnostics(name string, source []byte, globals map[string]string) []compono.Diagnostic {
	analysis, err := compono.Analyze(source)
	if err != nil {
		return nil
	}
	hasRequired := false
	for _, p := range analysis.Params {
		if p.Required {
			hasRequired = true
			break
		}
	}
	if !hasRequired {
		return nil
	}

	var diags []compono.Diagnostic
	for _, d := range analysis.Diagnostics {
		d.Source = []string{name}
		diags = append(diags, d)
	}

	builtins := make(map[string]struct{}, 0)
	for _, b := range compono.Builtins() {
		builtins[b] = struct{}{}
	}
	for _, r := range analysis.Refs {
		if r.Kind != "call" && r.Kind != "value" {
			continue
		}
		if r.Local {
			continue
		}
		if _, ok := builtins[r.Name]; ok {
			continue
		}
		if _, ok := globals[r.Name]; ok {
			continue
		}
		if r.Name == name {
			continue
		}
		diags = append(diags, compono.Diagnostic{
			Code:    compono.CodeUnknownComponent,
			Message: "The component **" + r.Name + "** is not defined.",
			Source:  []string{name},
			Range:   r.Range,
		})
	}

	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Range.Start.Offset < diags[j].Range.Start.Offset
	})
	return diags
}
