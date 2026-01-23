package repository

import (
	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type ComponentRepository struct {
	db *gorm.DB
}

func NewComponentRepository(db *gorm.DB) *ComponentRepository {
	return &ComponentRepository{
		db: db,
	}
}

func (r *ComponentRepository) GetByID(ID uint) models.Component {
	var c models.Component
	r.db.Model(&models.Component{}).First(&c, ID)
	return c
}

func (r *ComponentRepository) GetAll() []models.Component {
	var all []models.Component
	r.db.Model(&models.Component{}).Order("last_modified_at DESC").Find(&all)
	return all
}

func (r *ComponentRepository) GetByName(name string) models.Component {
	var c models.Component
	r.db.Model(&models.Component{}).Where("name = ?", name).First(&c)
	return c
}

func (r *ComponentRepository) Create(c models.Component) models.Component {
	r.db.Create(&c)
	return c
}

func (r *ComponentRepository) Update(c models.Component) models.Component {
	r.db.Model(&c).Select("*").Updates(c)
	return c
}

func (r *ComponentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Component{}, id).Error
}
