package theme

import "strings"

// assetTypes are the Content-Types of the files themecore allows under
// assets/: theme.css and the fonts and images by their extension.
var assetTypes = map[string]string{
	".woff2": "font/woff2",
	".woff":  "font/woff",
	".png":   "image/png",
	".jpeg":  "image/jpeg",
	".jpg":   "image/jpeg",
	".webp":  "image/webp",
	".avif":  "image/avif",
	".gif":   "image/gif",
}

// ContentType returns the Content-Type of the asset name, a path relative to
// assets/, or "" for a name themecore does not allow.
func ContentType(name string) string {
	if name == "theme.css" {
		return "text/css; charset=utf-8"
	}
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return ""
	}
	return assetTypes[name[i:]]
}
