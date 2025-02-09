package models

import (
	"time"

	"gorm.io/gorm"
)

type Component struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `json:"name" validate:"required,numeric-screaming-snake-case"`
	Content        string     `json:"content"`
	LastModifiedAt *time.Time `json:"last_modified_at"`
}

func (c *Component) FillByName(db *gorm.DB) {
	db.Model(&Component{}).Where("name = ?", c.Name).First(&c)
}
