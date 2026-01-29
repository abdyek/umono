package umono

import (
	"embed"
	"io/fs"
)

//go:embed all:views
var viewsFS embed.FS

//go:embed all:public
var publicFS embed.FS

// Views returns the views filesystem without the "views" prefix
func Views() fs.FS {
	sub, _ := fs.Sub(viewsFS, "views")
	return sub
}

// Public returns the public filesystem without the "public" prefix
func Public() fs.FS {
	sub, _ := fs.Sub(publicFS, "public")
	return sub
}
