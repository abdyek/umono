package models

import "time"

const (
	StorageTypeLocal = "local"
	StorageTypeS3    = "s3"
)

type Storage struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id" db:"id"`
	Name      string    `gorm:"not null" json:"name" db:"name"`
	Type      string    `gorm:"not null;index" json:"type" db:"type"`
	Config    JSONMap   `gorm:"not null;type:json" json:"config" db:"config"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
