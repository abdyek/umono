// Package migration holds the one-off data migrations Umono runs at startup.
package migration

import (
	"regexp"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

// globalParamLineKey is the option that records that GlobalParamLine has run.
const globalParamLineKey = "migration_global_param_line"

// oldGlobalParamLine is the heuristic Compono v0.5 used to recognize the
// unmarked parameter line at the start of a global.
var oldGlobalParamLine = regexp.MustCompile(`^([a-z][a-z0-9-]*)[ \t\r\n]*=[ \t\r\n]*(".*?"|\d+(?:\.\d+)?|true|false|\[|\{|[A-Z0-9]+(?:_[A-Z0-9]+)*|context\s*\()`)

// GlobalParamLine moves the parameter lines of the saved site globals to the
// format of Compono v0.7: a global whose source matches oldGlobalParamLine gets
// "~ " in front of it. It runs once; later calls change nothing.
func GlobalParamLine(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Option{}).Where("key = ?", globalParamLineKey).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		var components []models.Component
		if err := tx.Find(&components).Error; err != nil {
			return err
		}

		for _, component := range components {
			if oldGlobalParamLine.MatchString(component.Content) {
				if err := tx.Model(&models.Component{}).Where("id = ?", component.ID).UpdateColumn("content", "~ "+component.Content).Error; err != nil {
					return err
				}
			}
		}

		return tx.Create(&models.Option{Key: globalParamLineKey, Value: "done"}).Error
	})
}
