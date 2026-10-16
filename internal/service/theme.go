package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"github.com/umono-cms/umono/internal/theme"
)

var (
	// ErrContractChanged is returned when a package's contract.toml differs
	// from the stored record of the same contract version.
	ErrContractChanged = errors.New("contract.toml differs from the stored contract of the same version")
	// ErrNotCovered is returned when a theme's contract does not cover the
	// base of its group.
	ErrNotCovered = errors.New("the theme's contract does not cover the base of its group")
	// ErrNotInstalled is returned for a pending install without an installed
	// theme at its path. It can only be rejected.
	ErrNotInstalled = errors.New("no theme is installed at the path of the pending install")
	// ErrGroupChanged is returned when the pending install of an active theme
	// has a contract of another group. The theme is deactivated first.
	ErrGroupChanged = errors.New("the pending install's contract is in another group; deactivate the theme first")
)

// ContractRegistry verifies a package's contract.toml against the published
// contract. theme.Registry is the registry of umono.dev.
type ContractRegistry interface {
	Verify(ctx context.Context, ref models.Contract, src []byte) error
}

// ThemeCandidate is an installed theme that passed every check of
// activation.
type ThemeCandidate struct {
	theme.Installed
	// Contract is the record of the theme's contract: the stored one, or a new
	// one with ID 0 that activation adds.
	Contract models.Contract
	// Asked tells whether the registry was asked.
	Asked bool
}

// ThemeService activates the themes installed under themesDir and applies the pending installs under pendingDir.
type ThemeService struct {
	themesDir  string
	pendingDir string
	contracts  *ContractService
	themes     *repository.ThemeRepository
	registry   ContractRegistry
	active     *theme.ActiveSet
	// activateMu serializes activations, so two of them do not add the same
	// contract record or overwrite each other's active_themes.
	activateMu         sync.Mutex
	renderInputChanged func(ctx context.Context)
}

func NewThemeService(themesDir, pendingDir string, contracts *ContractService, themes *repository.ThemeRepository, registry ContractRegistry) *ThemeService {
	return &ThemeService{
		themesDir:  themesDir,
		pendingDir: pendingDir,
		contracts:  contracts,
		themes:     themes,
		registry:   registry,
		active:     theme.NewActiveSet(),
	}
}

// Check runs the checks of activation on the installed theme at path, in
// their order: the package checks, the contract check and the coverage check.
// The registry is asked only when the database has no record of the
// contract; a stored record never changes, so its checksum is enough.
func (s *ThemeService) Check(ctx context.Context, path string) (ThemeCandidate, error) {
	installed, err := theme.CheckInstalled(s.themesDir, path)
	if err != nil {
		return ThemeCandidate{}, err
	}
	return s.checkContract(ctx, installed)
}

// checkContract runs the contract check and the coverage check on a theme
// that passed the package checks.
func (s *ThemeService) checkContract(ctx context.Context, installed theme.Installed) (ThemeCandidate, error) {
	ref, err := models.ParseContractRef(installed.Built.Contract)
	if err != nil {
		return ThemeCandidate{}, err
	}
	sum := sha256.Sum256(installed.ContractSource)
	checksum := hex.EncodeToString(sum[:])

	record, ok := s.contracts.Record(ref)
	asked := false
	if ok {
		if record.Checksum != checksum {
			return ThemeCandidate{}, fmt.Errorf("%w: %s", ErrContractChanged, ref.Ref())
		}
	} else {
		if err := s.registry.Verify(ctx, ref, installed.ContractSource); err != nil {
			return ThemeCandidate{}, err
		}
		record = models.Contract{
			Org:      ref.Org,
			Name:     ref.Name,
			Major:    ref.Major,
			Minor:    ref.Minor,
			Source:   string(installed.ContractSource),
			Checksum: checksum,
		}
		asked = true
	}

	if !s.contracts.Selectable(ref) {
		return ThemeCandidate{}, fmt.Errorf("%w: %d pages are bound to a newer minor", ErrNotCovered, len(s.contracts.UncoveredPages(ref)))
	}
	return ThemeCandidate{Installed: installed, Contract: record, Asked: asked}, nil
}

// Activate runs the checks of activation on the installed theme at path and
// activates it: the contract record, when new, and active_themes are written
// in one transaction, and the theme replaces the active theme of its group in
// memory with its release and attributes table. The release is never stored.
func (s *ThemeService) Activate(ctx context.Context, path string) (ThemeCandidate, error) {
	s.activateMu.Lock()
	defer s.activateMu.Unlock()

	c, err := s.Check(ctx, path)
	if err != nil {
		return ThemeCandidate{}, err
	}
	if err := s.themes.Activate(&c.Contract, models.ThemeOption{Path: c.Path, ContractID: c.Contract.Ref()}); err != nil {
		return ThemeCandidate{}, err
	}
	s.active.Set(c.Contract.Group(), theme.Loaded{
		Path:       c.Path,
		Release:    c.Release,
		ContractID: c.Contract.ID,
		Contract:   c.Result.Contract,
		Attributes: c.Result.Attributes,
	})
	s.notifyRenderInputChange()
	return c, nil
}

// Loaded returns the active theme of the group g loaded into memory.
func (s *ThemeService) Loaded(g models.ContractGroup) (theme.Loaded, bool) {
	return s.active.Get(g)
}

// Deactivate returns the group g to fallback render: its element is removed
// from active_themes and its theme from memory. The theme directory, the
// contract record and the bindings of the pages stay. It has no
// precondition.
func (s *ThemeService) Deactivate(g models.ContractGroup) error {
	s.activateMu.Lock()
	defer s.activateMu.Unlock()

	if err := s.themes.Deactivate(g); err != nil {
		return err
	}
	s.active.Delete(g)
	s.notifyRenderInputChange()
	return nil
}

// OnRenderInputChange sets fn to be called after an activation or a
// deactivation, which changes how the pages of a group render.
func (s *ThemeService) OnRenderInputChange(fn func(ctx context.Context)) {
	s.renderInputChanged = fn
}

func (s *ThemeService) notifyRenderInputChange() {
	if s.renderInputChanged != nil {
		s.renderInputChanged(context.Background())
	}
}

// ApplyPending uses the new content of the pending install at path. The
// checks of activation run on the pending directory; for a theme that is not
// active the contract and coverage checks are skipped, because activation
// runs them. The new contract of an active theme must stay in its group. A
// changed contract is written to active_themes first, then the pending
// directory replaces the installed one and the new release and attributes
// table replace the old ones in memory. When a step fails, the installed
// theme stays in use.
func (s *ThemeService) ApplyPending(ctx context.Context, path string) error {
	s.activateMu.Lock()
	defer s.activateMu.Unlock()

	installed, err := theme.CheckInstalled(s.pendingDir, path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(s.themesDir, filepath.FromSlash(path))); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrNotInstalled, path)
		}
		return err
	}

	g, old, active := s.active.ByPath(path)
	if !active {
		return theme.Replace(s.themesDir, s.pendingDir, path)
	}

	ref, err := models.ParseContractRef(installed.Built.Contract)
	if err != nil {
		return err
	}
	if ref.Group() != g {
		return fmt.Errorf("%w: %s", ErrGroupChanged, ref.Ref())
	}
	c, err := s.checkContract(ctx, installed)
	if err != nil {
		return err
	}

	contractChanged := c.Contract.ID != old.ContractID
	if contractChanged {
		if err := s.themes.Activate(&c.Contract, models.ThemeOption{Path: path, ContractID: c.Contract.Ref()}); err != nil {
			return err
		}
	}
	if err := theme.Replace(s.themesDir, s.pendingDir, path); err != nil {
		if contractChanged {
			previous := models.Contract{ID: old.ContractID}
			_ = s.themes.Activate(&previous, models.ThemeOption{Path: path, ContractID: old.Contract.ID + "@" + old.Contract.Version})
		}
		return err
	}

	s.active.Set(g, theme.Loaded{
		Path:       path,
		Release:    c.Release,
		ContractID: c.Contract.ID,
		Contract:   c.Result.Contract,
		Attributes: c.Result.Attributes,
	})
	if contractChanged {
		s.notifyRenderInputChange()
	}
	return nil
}
