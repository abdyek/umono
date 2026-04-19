package media

import "strings"

func NormalizeAlias(alias string) string {
	return strings.TrimSpace(alias)
}

func AllowedMimeType(mimeType string) bool {
	_, ok := extByMimeType(mimeType)
	return ok
}

func ExtensionByMimeType(mimeType string) (string, bool) {
	return extByMimeType(mimeType)
}

func extByMimeType(mimeType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/png":
		return "png", true
	case "image/webp":
		return "webp", true
	case "image/jpeg":
		return "jpeg", true
	default:
		return "", false
	}
}
