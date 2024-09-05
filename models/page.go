package models

import "time"

type Page struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Content        string     `json:"content"`
	LastModifiedAt *time.Time `json:"last_modified_at"`
	Published      bool       `json:"published"`
}
