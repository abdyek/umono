package service

import (
	"fmt"
	"sort"

	"github.com/umono-cms/compono"
	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/theme"
)

// BindingOption is a layout the Layout field offers.
type BindingOption struct {
	// Value is the form value: the contract record's ID and the layout id,
	// joined by a colon.
	Value string
	// Group is the group key, for example "umono/blog v2".
	Group string
	// Label is "org/id vM.N · layout-id", followed by " — description" when
	// the layout has one.
	Label      string
	ContractID uint
	LayoutID   string
	// Old marks the page's own binding to an older record than the offered
	// one. Choosing the layout again moves the page to the offered record.
	Old bool
	// Params are the layout's parameters, from its source's parameter line.
	Params []compono.Param
	// RaisesBase tells whether binding the layout raises the group's base,
	// which only a group without an active theme allows.
	RaisesBase bool
	// NewBase is the version the base rises to, for example "v1.3".
	NewBase string
	// Unselectable are the installed themes, "id version", that the raised
	// base makes unselectable.
	Unselectable []string
}

// BindingOptions returns the options of the Layout field for page: the
// layouts of the offered record of every group with a stored record, sorted
// by group and label, and the page's own binding when it is to an older
// record than the offered one.
func (s *ThemeService) BindingOptions(page models.SitePage) ([]BindingOption, error) {
	active, err := s.themes.GetActiveThemes()
	if err != nil {
		return nil, err
	}
	activeGroups := map[models.ContractGroup]bool{}
	for _, a := range active {
		if ref, err := models.ParseContractRef(a.ContractID); err == nil {
			activeGroups[ref.Group()] = true
		}
	}
	installed, err := theme.ListInstalled(s.themesDir)
	if err != nil {
		return nil, err
	}

	groups := map[models.ContractGroup]bool{}
	for _, r := range s.contracts.All() {
		groups[r.Group()] = true
	}

	var options []BindingOption
	for g := range groups {
		offered, err := s.contracts.Offered(g, active)
		if err != nil {
			return nil, err
		}
		raises, newBase, unselectable := false, "", []string(nil)
		if !activeGroups[g] {
			base, has := s.contracts.Base(g)
			if !has || offered.Minor > base {
				raises = true
				newBase = fmt.Sprintf("v%d.%d", offered.Major, offered.Minor)
				for _, t := range installed {
					if t.Err != nil {
						continue
					}
					ref, err := models.ParseContractRef(t.Built.Contract)
					if err != nil || ref.Group() != g {
						continue
					}
					if ref.Minor < offered.Minor && (!has || ref.Minor >= base) {
						unselectable = append(unselectable, t.Built.ID+" "+t.Built.Version)
					}
				}
			}
		}
		layouts, err := s.bindingOptions(offered, false)
		if err != nil {
			return nil, err
		}
		for _, o := range layouts {
			o.RaisesBase, o.NewBase, o.Unselectable = raises, newBase, unselectable
			options = append(options, o)
		}

		if page.ContractID != nil && page.LayoutID != nil && *page.ContractID != offered.ID {
			if current, ok := s.contracts.ByID(*page.ContractID); ok && current.Group() == g {
				old, err := s.bindingOptions(current, true)
				if err != nil {
					return nil, err
				}
				for _, o := range old {
					if o.LayoutID == *page.LayoutID {
						options = append(options, o)
					}
				}
			}
		}
	}

	sort.SliceStable(options, func(i, j int) bool {
		if options[i].Group != options[j].Group {
			return options[i].Group < options[j].Group
		}
		return options[i].Label < options[j].Label
	})
	return options, nil
}

// bindingOptions returns an option for every layout of record.
func (s *ThemeService) bindingOptions(record models.Contract, old bool) ([]BindingOption, error) {
	contract, err := s.parsedContract(record)
	if err != nil {
		return nil, err
	}
	var options []BindingOption
	for _, l := range contract.Layouts {
		analysis, err := compono.Analyze([]byte(l.Source))
		if err != nil {
			return nil, err
		}
		options = append(options, BindingOption{
			Value:      fmt.Sprintf("%d:%s", record.ID, l.ID),
			Group:      record.Group().String(),
			Label:      bindingLabel(record, l),
			ContractID: record.ID,
			LayoutID:   l.ID,
			Old:        old,
			Params:     analysis.Params,
		})
	}
	return options, nil
}

func bindingLabel(record models.Contract, l themecore.Layout) string {
	label := fmt.Sprintf("%s/%s v%d.%d · %s", record.Org, record.Name, record.Major, record.Minor, l.ID)
	if l.Description != nil && *l.Description != "" {
		label += " — " + *l.Description
	}
	return label
}
