package theme

import (
	"sync"

	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/models"
)

// Loaded is an active theme loaded into memory. Render and the asset service
// read its release and attributes table from here, never from disk.
type Loaded struct {
	// Path is the install path under themes/: org/id/version.
	Path    string
	Release string
	// ContractID is the ID of the record of the theme's contract.
	ContractID uint
	Contract   themecore.Contract
	Attributes themecore.AttributeTable
}

// Broken is an active theme that failed the checks of loading. Its group
// renders with the fallback until it is retried or deactivated.
type Broken struct {
	// Path is the install path under themes/: org/id/version.
	Path string
	Err  error
}

// ActiveSet holds the loaded active themes, at most one per contract group.
// It is safe for concurrent use.
type ActiveSet struct {
	mu     sync.RWMutex
	themes map[models.ContractGroup]Loaded
	broken map[models.ContractGroup]Broken
}

func NewActiveSet() *ActiveSet {
	return &ActiveSet{
		themes: map[models.ContractGroup]Loaded{},
		broken: map[models.ContractGroup]Broken{},
	}
}

// Set makes t the active theme of the group g.
func (a *ActiveSet) Set(g models.ContractGroup, t Loaded) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.themes[g] = t
	delete(a.broken, g)
}

// Get returns the active theme of the group g.
func (a *ActiveSet) Get(g models.ContractGroup) (Loaded, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	t, ok := a.themes[g]
	return t, ok
}

// SetBroken marks the active theme of the group g broken.
func (a *ActiveSet) SetBroken(g models.ContractGroup, b Broken) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.themes, g)
	a.broken[g] = b
}

// Broken returns the broken active theme of the group g.
func (a *ActiveSet) Broken(g models.ContractGroup) (Broken, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	b, ok := a.broken[g]
	return b, ok
}

// Delete removes the active theme of the group g, loaded or broken.
func (a *ActiveSet) Delete(g models.ContractGroup) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.themes, g)
	delete(a.broken, g)
}

// ByPath returns the active theme installed at path and its group.
func (a *ActiveSet) ByPath(path string) (models.ContractGroup, Loaded, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for g, t := range a.themes {
		if t.Path == path {
			return g, t, true
		}
	}
	return models.ContractGroup{}, Loaded{}, false
}
