package models

type Option struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Key   string `gorm:"uniqueIndex;not null"`
	Value string
}

// ThemeOption is an element of the active_themes option, a JSON array with at
// most one element per contract group.
type ThemeOption struct {
	Path       string `json:"path"`        // install path under themes/. For example "acme/minimal/0.1.0"
	ContractID string `json:"contract_id"` // org/id@vMAJOR.MINOR
}
