package models

import "time"

type SitePage struct {
	ID             uint `gorm:"primaryKey"`
	Name           string
	Slug           string `gorm:"uniqueIndex"`
	Content        string
	LastModifiedAt *time.Time
	Enabled        bool
}

func (SitePage) TableName() string {
	return "pages"
}
