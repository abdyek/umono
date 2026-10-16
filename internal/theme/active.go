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

// ActiveSet holds the loaded active themes, at most one per contract group.
// It is safe for concurrent use.
type ActiveSet struct {
	mu     sync.RWMutex
	themes map[models.ContractGroup]Loaded
}

func NewActiveSet() *ActiveSet {
	return &ActiveSet{themes: map[models.ContractGroup]Loaded{}}
}

// Set makes t the active theme of the group g.
func (a *ActiveSet) Set(g models.ContractGroup, t Loaded) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.themes[g] = t
}

// Get returns the active theme of the group g.
func (a *ActiveSet) Get(g models.ContractGroup) (Loaded, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	t, ok := a.themes[g]
	return t, ok
}
