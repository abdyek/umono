package service

import (
	"errors"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
)

// ErrContractNotFound is returned when a contract group has no record.
var ErrContractNotFound = errors.New("contract not found")

// ContractService computes the bases, coverage and offered records of
// contract groups from the contract records and the pages bound to them.
type ContractService struct {
	repo *repository.ContractRepository
}

func NewContractService(r *repository.ContractRepository) *ContractService {
	return &ContractService{repo: r}
}

// Base returns the base of the group: the highest minor of the records its
// pages are bound to, disabled pages included. ok is false when no page of
// the group is bound.
func (s *ContractService) Base(g models.ContractGroup) (minor uint, ok bool) {
	return s.repo.GroupBase(g)
}

// Record returns the stored record of the contract ref, identified by its
// org, name, major and minor. ok is false when there is none.
func (s *ContractService) Record(ref models.Contract) (models.Contract, bool) {
	c := s.repo.GetByIdentity(ref.Org, ref.Name, ref.Major, ref.Minor)
	return c, c.ID != 0
}

// Selectable reports whether a theme with the contract can be activated in
// its group: the contract covers the group's base, or the group has none.
func (s *ContractService) Selectable(theme models.Contract) bool {
	base, ok := s.Base(theme.Group())
	return !ok || theme.Minor >= base
}

// Offered returns the record that the binding interface offers for the
// group: the record of the group's active theme in active, or the record with
// the highest minor when the group has no active theme. An element of active
// whose ContractID does not parse is skipped.
func (s *ContractService) Offered(g models.ContractGroup, active []models.ThemeOption) (models.Contract, error) {
	for _, theme := range active {
		ref, err := models.ParseContractRef(theme.ContractID)
		if err != nil || ref.Group() != g {
			continue
		}
		c := s.repo.GetByIdentity(ref.Org, ref.Name, ref.Major, ref.Minor)
		if c.ID == 0 {
			return models.Contract{}, ErrContractNotFound
		}
		return c, nil
	}
	c := s.repo.GetHighestInGroup(g)
	if c.ID == 0 {
		return models.Contract{}, ErrContractNotFound
	}
	return c, nil
}

// UncoveredPages returns the pages of the contract's group that are bound to
// a newer minor than the contract's, ordered by ID. They keep a theme with
// the contract from being selected.
func (s *ContractService) UncoveredPages(theme models.Contract) []models.SitePage {
	return s.repo.GetPagesAboveMinor(theme.Group(), theme.Minor)
}

// CountPages returns the number of pages bound to the record id, disabled
// ones included.
func (s *ContractService) CountPages(id uint) int64 {
	return s.repo.CountPages(id)
}

// Delete deletes the record id. The caller checks the conditions of
// deletion first.
func (s *ContractService) Delete(id uint) error {
	return s.repo.Delete(id)
}
