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

func (r *OptionRepository) SaveNotFoundPageOption(option models.NotFoundPageOption) error {
	optInDB := r.GetOptionByKey("not_found_page_option")
	alreadyExists := optInDB.ID != 0

	value, err := json.Marshal(option)
	if err != nil {
		return err
	}

	if alreadyExists && option.Title == "" && option.Content == "" {
		// Delete
		r.db.Unscoped().Delete(&optInDB)
		return nil
	} else if alreadyExists && (option.Title != "" || option.Content != "") {
		// Update
		r.db.Model(&models.Option{}).Where("ID = ?", optInDB.ID).Updates(&models.Option{
			Value: string(value),
		})
		return nil
	} else {
		// Create
		r.db.Create(&models.Option{
			Key:   "not_found_page_option",
			Value: string(value),
		})
		return nil
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
