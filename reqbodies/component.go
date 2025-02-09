package reqbodies

import "github.com/umono-cms/umono/models"

type CreateComponent struct {
	Component models.Component `json:"component"`
}
