package service

type SystemService struct{}

func NewSystemService() *SystemService {
	return &SystemService{}
}

type SystemMenuItem struct {
	TitleKey string
	Slug     string
	Partial  string
}

func (*SystemService) MenuItems() []SystemMenuItem {
	return []SystemMenuItem{
		{
			TitleKey: "system.menu.jobs",
			Slug:     "jobs",
			Partial:  "system-jobs",
		},
	}
}
