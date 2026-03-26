package models

type Option struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Key   string `gorm:"uniqueIndex;not null"`
	Value string
}

type NotFoundPageOption struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
