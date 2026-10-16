package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/theme"
)

var (
	// ErrContractChanged is returned when a package's contract.toml differs
	// from the stored record of the same contract version.
	ErrContractChanged = errors.New("contract.toml differs from the stored contract of the same version")
	// ErrNotCovered is returned when a theme's contract does not cover the
	// base of its group.
	ErrNotCovered = errors.New("the theme's contract does not cover the base of its group")
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

// ThemeService activates the themes installed under themesDir.
type ThemeService struct {
	themesDir string
	contracts *ContractService
	registry  ContractRegistry
}

func NewThemeService(themesDir string, contracts *ContractService, registry ContractRegistry) *ThemeService {
	return &ThemeService{themesDir: themesDir, contracts: contracts, registry: registry}
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
