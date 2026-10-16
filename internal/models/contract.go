package models

import "time"

// Contract is a theme contract stored on activation. A record is never
// updated: Source is the contract.toml of the package, unchanged.
type Contract struct {
	ID        uint   `gorm:"primaryKey"`
	Org       string `gorm:"not null;uniqueIndex:idx_contracts_identity"`
	Name      string `gorm:"not null;uniqueIndex:idx_contracts_identity"`
	Major     uint   `gorm:"not null;uniqueIndex:idx_contracts_identity"`
	Minor     uint   `gorm:"not null;uniqueIndex:idx_contracts_identity"`
	Source    string `gorm:"not null"`
	Checksum  string `gorm:"not null"` // sha256 of Source
	CreatedAt time.Time
}
