package service

type SettingsService struct{}

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

type SettingsMenuItem struct {
	TitleKey string
	Slug     string
	Partial  string
}

func (*SettingsService) MenuItems() []SettingsMenuItem {
	return []SettingsMenuItem{
		{
			TitleKey: "settings.menu.general",
			Slug:     "general",
			Partial:  "settings-general",
		},
		{
			TitleKey: "settings.menu.not_found_page",
			Slug:     "404-page",
			Partial:  "settings-404-page",
		},
		{
			TitleKey: "settings.menu.storage",
			Slug:     "storage",
			Partial:  "settings-storage",
		},
		{
			TitleKey: "settings.menu.about",
			Slug:     "about",
			Partial:  "settings-about",
		},
	}
}
