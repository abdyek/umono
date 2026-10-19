package theme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/umono-cms/themecore"
)

// DevPath is the install path of the theme in development under themes/.
// Only the dev mode gives it the privileges of dev/dev.
const DevPath = "dev/dev/0.0.0"

// DevRelease is the release of the theme in development, a fixed value.
const DevRelease = "dev"

// ErrDevRoot is returned for a value of UMONO_THEME_DEV that is not the
// absolute path of a directory.
var ErrDevRoot = errors.New("UMONO_THEME_DEV must be the absolute path of the theme project's directory")

// ParseDevRoot checks the value of UMONO_THEME_DEV, the root of the theme
// project in development, which must be the absolute path of a directory.
func ParseDevRoot(v string) (string, error) {
	if v == "" || !filepath.IsAbs(v) {
		return "", fmt.Errorf("%w: %q", ErrDevRoot, v)
	}
	info, err := os.Stat(v)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDevRoot, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%w: %q is not a directory", ErrDevRoot, v)
	}
	return v, nil
}

// CheckDev runs themecore.CheckDev on the theme in development under
// themesDir with the privileges of dev/dev; root is the theme project, so
// the symlinks must resolve inside it. A theme-api-version of built.toml
// outside the range this Umono supports is fatal, as in production.
func CheckDev(themesDir, root string) (themecore.DevResult, error) {
	r, err := themecore.CheckDev(filepath.Join(themesDir, filepath.FromSlash(DevPath)), root)
	if err != nil {
		return themecore.DevResult{}, err
	}
	if min, current := themecore.ThemeAPIMin(), themecore.ThemeAPICurrent(); r.Built.ThemeAPIVersion < min || r.Built.ThemeAPIVersion > current {
		return themecore.DevResult{}, fmt.Errorf("%w: %d, supported %d-%d", ErrThemeAPI, r.Built.ThemeAPIVersion, min, current)
	}
	return r, nil
}
