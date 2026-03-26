package repository

import (
	"encoding/json"
	"errors"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

var ErrOptionNotFound = errors.New("option not found")

type OptionRepository struct {
	db *gorm.DB
}

func NewOptionRepository(db *gorm.DB) *OptionRepository {
	return &OptionRepository{
		db: db,
	}
}

func (r *OptionRepository) GetOptionByKey(key string) models.Option {
	var opt models.Option
	r.db.Model(&models.Option{}).Where("key = ?", key).First(&opt)
	return opt
}

func (r *OptionRepository) SaveOption(key, value string) error {
	optInDB := r.GetOptionByKey(key)
	if optInDB.ID == 0 {
		return r.db.Create(&models.Option{
			Key:   key,
			Value: value,
		}).Error
	}

	return r.db.Model(&models.Option{}).
		Where("id = ?", optInDB.ID).
		Update("value", value).
		Error
}

func (r *OptionRepository) SaveNotFoundPageOption(option models.NotFoundPageOption) error {
	optInDB := r.GetOptionByKey("not_found_page_option")
	alreadyExists := optInDB.ID != 0

	value, err := json.Marshal(option)
	if err != nil {
		return err
	}

	if alreadyExists && option.Title == "" && option.Content == "" {
		// Delete
		return r.db.Unscoped().Delete(&optInDB).Error
	} else if alreadyExists && (option.Title != "" || option.Content != "") {
		// Update
		return r.db.Model(&models.Option{}).Where("ID = ?", optInDB.ID).Updates(&models.Option{
			Value: string(value),
		}).Error
	} else {
		// Create
		return r.db.Create(&models.Option{
			Key:   "not_found_page_option",
			Value: string(value),
		}).Error
	}
}

func GetOption[T any](repo OptionRepository, key string) (T, error) {
	var result T
	opt := repo.GetOptionByKey(key)
	if opt.ID == 0 {
		return result, ErrOptionNotFound
	}
	if err := json.Unmarshal([]byte(opt.Value), &result); err != nil {
		return result, err
	}
	return result, nil
}
