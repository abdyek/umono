package models

import "time"

type Component struct {
	ID             uint   `gorm:"primaryKey"`
	Name           string `gorm:"uniqueIndex"`
	Content        string
	LastModifiedAt *time.Time
}
