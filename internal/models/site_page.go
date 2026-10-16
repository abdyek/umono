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
}

func (SitePage) TableName() string {
	return "pages"
}
