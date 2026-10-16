package theme

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/umono-cms/themecore"
)

// Installed is an installed theme that passed the package checks of
// activation.
type Installed struct {
	// Path is the install path under themes/: org/id/version.
	Path string
	// Release is the first 12 hex digits of the SHA-256 of checksums.json. It
	// is computed when the theme is loaded and never stored.
	Release string
	Built   themecore.Built
	// ContractSource is the contract.toml of the package, unchanged.
	ContractSource []byte
	// Result holds the parsed contract and the compiled attributes.toml.
	Result themecore.Result
}

// ErrInvalidThemePath is returned for a path that is not org/id/version.
var ErrInvalidThemePath = errors.New("invalid theme path")

// ErrThemeAPI is returned for a theme whose theme-api-version is outside the
// range this Umono supports.
var ErrThemeAPI = errors.New("unsupported theme API version")

// themePathPattern is an install path: the org/id and the version of
// themecore's grammars.
var themePathPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*/[a-z0-9]+(-[a-z0-9]+)*/(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// CheckInstalled reads the theme at path under themesDir and runs the
// package checks of activation in their order: checksums.json, the Theme API
// range of built.toml, then themecore.CheckPackage for the static check,
// attributes.toml and the feature floor. The contract and coverage checks
// need the database and come after it.
func CheckInstalled(themesDir, path string) (Installed, error) {
	if !themePathPattern.MatchString(path) {
		return Installed{}, fmt.Errorf("%w: %q", ErrInvalidThemePath, path)
	}
	p, err := themecore.ReadDir(filepath.Join(themesDir, filepath.FromSlash(path)))
	if err != nil {
		return Installed{}, err
	}
	if err := themecore.VerifyChecksums(p); err != nil {
		return Installed{}, err
	}
	built, err := themecore.ParseBuilt(p.Files["built.toml"])
	if err != nil {
		return Installed{}, err
	}
	if min, current := themecore.ThemeAPIMin(), themecore.ThemeAPICurrent(); built.ThemeAPIVersion < min || built.ThemeAPIVersion > current {
		return Installed{}, fmt.Errorf("%w: %d, supported %d-%d", ErrThemeAPI, built.ThemeAPIVersion, min, current)
	}
	result, err := themecore.CheckPackage(p)
	if err != nil {
		return Installed{}, err
	}
	sum := sha256.Sum256(p.Files["checksums.json"])
	return Installed{
		Path:           path,
		Release:        hex.EncodeToString(sum[:])[:12],
		Built:          built,
		ContractSource: p.Files["contract.toml"],
		Result:         result,
	}, nil
}
