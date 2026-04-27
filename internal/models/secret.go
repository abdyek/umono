package models

type Secret struct {
	ID         string `gorm:"column:id;primaryKey;type:text" json:"id" db:"id"`
	Ciphertext []byte `gorm:"column:ciphertext;not null;type:blob" json:"-" db:"ciphertext"`
}

func (Secret) TableName() string {
	return "secrets"
}
