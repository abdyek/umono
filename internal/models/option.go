package models

const (
	DefaultNotFoundTitle   = "Page Not Found"
	DefaultNotFoundContent = "# Page Not Found\nThe page you're looking for may have been moved, deleted, or never existed."
)

type Option struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Key   string `gorm:"uniqueIndex;not null"`
	Value string
}

type NotFoundPageOption struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
