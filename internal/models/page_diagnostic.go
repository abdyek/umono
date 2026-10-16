package models

// PageDiagnostic is a Compono diagnostic of the last render of a page.
type PageDiagnostic struct {
	ID      uint `gorm:"primaryKey"`
	PageID  uint `gorm:"index"`
	Code    string
	Message string
	// Source is the diagnostic's source joined with " > ". It is empty for the
	// page's own source.
	Source string
}

// PublishedDiagnostic is a stored diagnostic of an enabled page together with
// the page's name.
type PublishedDiagnostic struct {
	PageID   uint
	PageName string
	Code     string
	Message  string
	Source   string
	// SystemKey is the SystemKey of the page.
	SystemKey *string
}
