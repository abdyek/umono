package models

import "time"

type Page struct {
	ID             uint `gorm:"primaryKey"`
	Name           string
	Slug           string `gorm:"uniqueIndex"`
	Content        string
	LastModifiedAt *time.Time
	Enabled        bool
}
