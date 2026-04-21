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

func (r *StorageRepository) GetAll() []models.Storage {
	var storages []models.Storage
	r.db.Model(&models.Storage{}).Order("created_at ASC").Find(&storages)
	return storages
}

func (r *StorageRepository) Create(storage models.Storage) models.Storage {
	r.db.Create(&storage)
	return storage
}

func (r *StorageRepository) Update(storage models.Storage) models.Storage {
	r.db.Model(&storage).Select("*").Updates(storage)
	return storage
}

func (r *StorageRepository) Delete(id string) error {
	return r.db.Delete(&models.Storage{}, "id = ?", id).Error
}
