package service

type SettingsService struct{}

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

type SettingsMenuItem struct {
	Title   string
	Slug    string
	Partial string
}

func (*SettingsService) MenuItems() []SettingsMenuItem {
	return []SettingsMenuItem{
		{
			Title:   "404 Page",
			Slug:    "404-page",
			Partial: "settings-404-page",
		},
		{
			Title:   "About",
			Slug:    "about",
			Partial: "settings-about",
		},
	}
}
