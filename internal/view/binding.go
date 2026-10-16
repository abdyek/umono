package view

import "github.com/umono-cms/umono/internal/service"

// LayoutField is the Layout field of the page editor and the 404 screen.
type LayoutField struct {
	// Value is the chosen option's value, "" for no layout.
	Value string
	// Current is the page's saved binding, the value its own option has.
	Current string
	Groups  []LayoutGroup
	Info    LayoutInfo
	// Error is the translated binding error and Skeleton the call skeleton
	// of the root rule's error.
	Error    string
	Skeleton string
}

// LayoutGroup is the options of one contract group.
type LayoutGroup struct {
	Name    string
	Options []LayoutOption
}

// LayoutOption is an option of the Layout field.
type LayoutOption struct {
	Value    string
	Label    string
	Old      bool
	Selected bool
}

// LayoutInfo is what the Layout field shows about the chosen layout.
type LayoutInfo struct {
	Selected     bool
	Old          bool
	Params       []LayoutParam
	RaisesBase   bool
	NewBase      string
	Unselectable []string
	I18n         any
}

// LayoutParam is a parameter of a layout.
type LayoutParam struct {
	Name     string
	Type     string
	Required bool
}

// NewLayoutField builds the Layout field from the binding options, with value
// chosen and current saved.
func NewLayoutField(options []service.BindingOption, value, current string, translator any) LayoutField {
	field := LayoutField{Value: value, Current: current, Info: NewLayoutInfo(options, value, translator)}
	index := map[string]int{}
	for _, o := range options {
		i, ok := index[o.Group]
		if !ok {
			i = len(field.Groups)
			index[o.Group] = i
			field.Groups = append(field.Groups, LayoutGroup{Name: o.Group})
		}
		field.Groups[i].Options = append(field.Groups[i].Options, LayoutOption{
			Value:    o.Value,
			Label:    o.Label,
			Old:      o.Old,
			Selected: o.Value == value,
		})
	}
	return field
}

// NewLayoutInfo returns what the Layout field shows about the option with
// value: its parameters, whether it is the page's old record and whether it
// raises the base.
func NewLayoutInfo(options []service.BindingOption, value string, translator any) LayoutInfo {
	info := LayoutInfo{I18n: translator}
	for _, o := range options {
		if o.Value != value {
			continue
		}
		info.Selected = true
		info.Old = o.Old
		info.RaisesBase = o.RaisesBase
		info.NewBase = o.NewBase
		info.Unselectable = o.Unselectable
		for _, p := range o.Params {
			info.Params = append(info.Params, LayoutParam{Name: p.Name, Type: string(p.Type), Required: p.Required})
		}
	}
	return info
}
