package service

import (
	"sort"

	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/theme"
)

// ThemeList is the list of the theme management page.
type ThemeList struct {
	// Groups are sorted by their key.
	Groups []ThemeGroup
	// Unreadable are the installed themes whose built.toml cannot be read or
	// names no contract reference.
	Unreadable []theme.InstalledTheme
	Pending    []theme.PendingInstall
}

// ThemeGroup is a contract group of the list: the union of the stored
// records, the installed themes and the active themes.
type ThemeGroup struct {
	Group models.ContractGroup
	// PageCount is the number of pages bound to the group's records,
	// disabled ones included.
	PageCount int64
	// Base is the base minor, meaningful when HasBase.
	Base    uint
	HasBase bool
	// Active is the group's active theme, nil without one.
	Active *ActiveTheme
	// Themes are the installed themes of the group, sorted by path.
	Themes []ListedTheme
	// Records are the stored records of the group, sorted by minor.
	Records []ListedRecord
}

// ActiveTheme is the active theme of a group as active_themes names it.
type ActiveTheme struct {
	Path       string
	ContractID string
	// Broken is set when the theme failed the checks of loading.
	Broken *theme.Broken
}

// ListedTheme is an installed theme of a group. Only coverage and the Theme
// API range make it unselectable in the list; the other checks run on
// activation.
type ListedTheme struct {
	Path  string
	Built themecore.Built
	// Contract is the reference of built.toml's contract.
	Contract models.Contract
	Active   bool
	// UncoveredPages is the number of pages bound to a newer minor.
	UncoveredPages int
	// ThemeAPI is true when theme-api-version is outside the range this
	// Umono supports.
	ThemeAPI bool
}

// ListedRecord is a stored record of a group.
type ListedRecord struct {
	Contract  models.Contract
	PageCount int64
	// CanDelete tells whether the conditions of deletion hold: no installed
	// theme has the contract and no page is bound to the record.
	CanDelete bool
}

// List returns the list of the theme management page.
func (s *ThemeService) List() (ThemeList, error) {
	installed, err := theme.ListInstalled(s.themesDir)
	if err != nil {
		return ThemeList{}, err
	}
	pending, err := theme.ListPending(s.themesDir, s.pendingDir)
	if err != nil {
		return ThemeList{}, err
	}
	active, err := s.themes.GetActiveThemes()
	if err != nil {
		return ThemeList{}, err
	}

	groups := map[models.ContractGroup]*ThemeGroup{}
	group := func(g models.ContractGroup) *ThemeGroup {
		tg, ok := groups[g]
		if !ok {
			tg = &ThemeGroup{Group: g}
			groups[g] = tg
		}
		return tg
	}

	for _, r := range s.contracts.All() {
		tg := group(r.Group())
		tg.Records = append(tg.Records, ListedRecord{Contract: r, PageCount: s.contracts.CountPages(r.ID)})
	}
	for _, a := range active {
		ref, err := models.ParseContractRef(a.ContractID)
		if err != nil {
			continue
		}
		at := &ActiveTheme{Path: a.Path, ContractID: a.ContractID}
		if b, ok := s.active.Broken(ref.Group()); ok {
			at.Broken = &b
		}
		group(ref.Group()).Active = at
	}

	list := ThemeList{Pending: pending}
	installedRefs := map[string]bool{}
	min, current := themecore.ThemeAPIMin(), themecore.ThemeAPICurrent()
	for _, t := range installed {
		if t.Err != nil {
			list.Unreadable = append(list.Unreadable, t)
			continue
		}
		ref, err := models.ParseContractRef(t.Built.Contract)
		if err != nil {
			t.Err = err
			list.Unreadable = append(list.Unreadable, t)
			continue
		}
		installedRefs[ref.Ref()] = true
		tg := group(ref.Group())
		tg.Themes = append(tg.Themes, ListedTheme{
			Path:           t.Path,
			Built:          t.Built,
			Contract:       ref,
			Active:         tg.Active != nil && tg.Active.Path == t.Path,
			UncoveredPages: len(s.contracts.UncoveredPages(ref)),
			ThemeAPI:       t.Built.ThemeAPIVersion < min || t.Built.ThemeAPIVersion > current,
		})
	}

	for g, tg := range groups {
		tg.PageCount = s.contracts.CountGroupPages(g)
		tg.Base, tg.HasBase = s.contracts.Base(g)
		for i, r := range tg.Records {
			tg.Records[i].CanDelete = r.PageCount == 0 && !installedRefs[r.Contract.Ref()] && len(list.Unreadable) == 0
		}
		sort.Slice(tg.Themes, func(i, j int) bool { return tg.Themes[i].Path < tg.Themes[j].Path })
		sort.Slice(tg.Records, func(i, j int) bool { return tg.Records[i].Contract.Minor < tg.Records[j].Contract.Minor })
		list.Groups = append(list.Groups, *tg)
	}
	sort.Slice(list.Groups, func(i, j int) bool { return list.Groups[i].Group.String() < list.Groups[j].Group.String() })
	return list, nil
}
