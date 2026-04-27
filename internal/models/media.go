package models

import "time"

type Media struct {
	ID           string         `gorm:"primaryKey;type:text" json:"id" db:"id"`
	StorageID    string         `gorm:"not null;index" json:"storage_id" db:"storage_id"`
	OriginalName string         `gorm:"not null" json:"original_name" db:"original_name"`
	PathKey      string         `gorm:"not null" json:"path_key" db:"path_key"`
	MimeType     string         `gorm:"not null" json:"mime_type" db:"mime_type"`
	Size         int64          `gorm:"not null" json:"size" db:"size"`
	Metadata     JSONMap        `gorm:"type:json" json:"metadata" db:"metadata"`
	Hash         string         `gorm:"not null;index" json:"hash" db:"hash"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
	Storage      Storage        `gorm:"foreignKey:StorageID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Variants     []MediaVariant `gorm:"foreignKey:MediaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type MediaVariant struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id" db:"id"`
	MediaID   string    `gorm:"not null;index" json:"media_id" db:"media_id"`
	PathKey   string    `gorm:"not null" json:"path_key" db:"path_key"`
	Size      int64     `gorm:"not null" json:"size" db:"size"`
	MimeType  string    `gorm:"not null" json:"mime_type" db:"mime_type"`
	Metadata  JSONMap   `gorm:"type:json" json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Media     Media     `gorm:"foreignKey:MediaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
