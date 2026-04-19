package media

import (
	"regexp"
	"strings"
)

var kebabCasePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func NormalizeAlias(alias string) string {
	return strings.TrimSpace(alias)
}

func IsKebabCase(alias string) bool {
	return kebabCasePattern.MatchString(NormalizeAlias(alias))
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
