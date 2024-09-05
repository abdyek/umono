package reqbodies

import "github.com/umono-cms/umono/models"

type CreatePage struct {
	Page models.Page `json:"page"`
}
