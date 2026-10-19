package repository

import (
	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type ContractRepository struct {
	db *gorm.DB
}

func NewContractRepository(db *gorm.DB) *ContractRepository {
	return &ContractRepository{db: db}
}

// inGroup limits a query on contracts to the group g. table is the name or
// alias of the contracts table in the query.
func inGroup(q *gorm.DB, table string, g models.ContractGroup) *gorm.DB {
	q = q.Where(table+".org = ? AND "+table+".name = ? AND "+table+".major = ?", g.Org, g.Name, g.Major)
	if g.Major == 0 {
		q = q.Where(table+".minor = ?", g.Minor)
	}
	return q
}

// GetByIdentity returns the record of org/name@vmajor.minor, or a zero
// Contract when there is none.
func (r *ContractRepository) GetByIdentity(org, name string, major, minor uint) models.Contract {
	var c models.Contract
	r.db.Where("org = ? AND name = ? AND major = ? AND minor = ?", org, name, major, minor).Limit(1).Find(&c)
	return c
}

// GetByID returns the record id, or a zero Contract when there is none.
func (r *ContractRepository) GetByID(id uint) models.Contract {
	var c models.Contract
	r.db.Where("id = ?", id).Limit(1).Find(&c)
	return c
}

// GetHighestInGroup returns the record of the group with the highest minor,
// or a zero Contract when the group has none.
func (r *ContractRepository) GetHighestInGroup(g models.ContractGroup) models.Contract {
	var c models.Contract
	inGroup(r.db.Model(&models.Contract{}), "contracts", g).Order("minor DESC").Limit(1).Find(&c)
	return c
}

// GroupBase returns the highest minor of the records that pages of the group
// are bound to, disabled pages included. ok is false when no page is bound.
func (r *ContractRepository) GroupBase(g models.ContractGroup) (minor uint, ok bool) {
	var base *uint
	inGroup(r.db.Table("pages p").Joins("JOIN contracts c ON c.id = p.contract_id"), "c", g).
		Select("MAX(c.minor)").Scan(&base)
	if base == nil {
		return 0, false
	}
	return *base, true
}

// GetPagesAboveMinor returns the pages of the group bound to a record with a
// minor above minor, ordered by ID.
func (r *ContractRepository) GetPagesAboveMinor(g models.ContractGroup, minor uint) []models.SitePage {
	var pages []models.SitePage
	inGroup(r.db.Model(&models.SitePage{}).Joins("JOIN contracts c ON c.id = pages.contract_id"), "c", g).
		Preload("Contract").Where("c.minor > ?", minor).Order("pages.id ASC").Find(&pages)
	return pages
}

// CountPages returns the number of pages bound to the record id, disabled
// ones included.
func (r *ContractRepository) CountPages(id uint) int64 {
	var n int64
	r.db.Model(&models.SitePage{}).Where("contract_id = ?", id).Count(&n)
	return n
}

// Delete deletes the record id.
func (r *ContractRepository) Delete(id uint) error {
	return r.db.Delete(&models.Contract{}, id).Error
}

// Create adds the contract record c. On success c has its ID.
func (r *ContractRepository) Create(c *models.Contract) error {
	return r.db.Create(c).Error
}

// GetAll returns every record, ordered by org, name, major and minor.
func (r *ContractRepository) GetAll() []models.Contract {
	var all []models.Contract
	r.db.Order("org ASC, name ASC, major ASC, minor ASC").Find(&all)
	return all
}

// CountGroupPages returns the number of pages bound to the records of the
// group g, disabled ones included.
func (r *ContractRepository) CountGroupPages(g models.ContractGroup) int64 {
	var n int64
	inGroup(r.db.Model(&models.SitePage{}).Joins("JOIN contracts c ON c.id = pages.contract_id"), "c", g).Count(&n)
	return n
}
