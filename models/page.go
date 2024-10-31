package models

import (
	"time"
)

type Page struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `json:"name" validate:"required"`
	Slug           string     `json:"slug" validate:"slug"`
	Content        string     `json:"content"`
	LastModifiedAt *time.Time `json:"last_modified_at"`
	Enabled        bool       `json:"enabled"`
}
