package database

import (
	"github.com/umono-cms/umono/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(dsn string) error {
	err := connect(dsn)

	if err != nil {
		return err
	}

	autoMigrate(DB)
	return nil
}

func connect(dsn string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
	})

	return err
}

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(&models.Page{})
}
