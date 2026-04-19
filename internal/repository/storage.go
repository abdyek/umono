package repository

import (
	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type StorageRepository struct {
	db *gorm.DB
}

func NewStorageRepository(db *gorm.DB) *StorageRepository {
	return &StorageRepository{db: db}
}

func (r *StorageRepository) GetByID(id string) models.Storage {
	var storage models.Storage
	r.db.Model(&models.Storage{}).Where("id = ?", id).First(&storage)
	return storage
}

func (r *StorageRepository) Create(storage models.Storage) models.Storage {
	r.db.Create(&storage)
	return storage
}

func (r *StorageRepository) Update(storage models.Storage) models.Storage {
	r.db.Model(&storage).Select("*").Updates(storage)
	return storage
}
