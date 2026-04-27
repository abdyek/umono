package repository

import (
	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type SecretRepository struct {
	db *gorm.DB
}

func NewSecretRepository(db *gorm.DB) *SecretRepository {
	return &SecretRepository{db: db}
}

func (r *SecretRepository) GetByID(id string) (models.Secret, error) {
	var secret models.Secret
	err := r.db.Model(&models.Secret{}).Where("id = ?", id).First(&secret).Error
	return secret, err
}

func (r *SecretRepository) Create(secret models.Secret) (models.Secret, error) {
	err := r.db.Create(&secret).Error
	return secret, err
}

func (r *SecretRepository) Update(secret models.Secret) (models.Secret, error) {
	err := r.db.Model(&secret).Select("*").Updates(secret).Error
	return secret, err
}

func (r *SecretRepository) Delete(id string) error {
	return r.db.Delete(&models.Secret{}, "id = ?", id).Error
}
