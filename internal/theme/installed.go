package theme

import (
	"path/filepath"

	"github.com/umono-cms/themecore"
)

// InstalledTheme is a theme under themes/ with its built.toml.
type InstalledTheme struct {
	// Path is the install path: org/id/version.
	Path  string
	Built themecore.Built
	// Err tells why built.toml could not be read.
	Err error
}

// ListInstalled returns the themes under themesDir, sorted by path, with
// their built.toml. Directories that are not org/id/version are skipped.
// Without themesDir there are none. It reads nothing but built.toml; the
// checks of activation run on activation.
func ListInstalled(themesDir string) ([]InstalledTheme, error) {
	paths, err := listThemePaths(themesDir)
	if err != nil {
		return nil, err
	}
	var list []InstalledTheme
	for _, path := range paths {
		t := InstalledTheme{Path: path}
		b, err := readManifest(filepath.Join(themesDir, filepath.FromSlash(path)), "built.toml")
		if err == nil {
			t.Built, err = themecore.ParseBuilt(b)
		}
		t.Err = err
		list = append(list, t)
	}
	return list, nil
}
