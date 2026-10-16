package models

import "time"

type SitePage struct {
	ID             uint `gorm:"primaryKey"`
	Name           string
	Slug           string `gorm:"uniqueIndex"`
	Content        string
	LastModifiedAt *time.Time
	Enabled        bool
	ContractID     *uint     `gorm:"index"`
	LayoutID       *string   // id of a [[layouts]] entry in the contract record. For example "post"
	Contract       *Contract `gorm:"foreignKey:ContractID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	SystemKey      *string   `gorm:"uniqueIndex"` // NULL: an ordinary page. NotFoundSystemKey: the 404 page
}

func (SitePage) TableName() string {
	return "pages"
}

// NotFoundSystemKey is the SystemKey of the 404 page. A page with a SystemKey
// is a system page: it is not resolved by its slug, cannot be deleted and is
// always enabled.
const NotFoundSystemKey = "not_found"

// NotFoundSlug is the reserved slug of the 404 page. Slug validation keeps
// slugs starting with "_" away from ordinary pages.
const NotFoundSlug = "_not-found"
