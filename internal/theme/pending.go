package theme

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/umono-cms/themecore"
)

// PendingInstall is a package that umono theme install --force unpacked into
// themes-pending/ for an installed theme. It waits for the site owner to use
// or reject it.
type PendingInstall struct {
	// Path is the install path: org/id/version.
	Path string
	// Installed tells whether themes/ has a theme at Path. A pending install
	// without one can only be rejected.
	Installed bool
	// OldBuilt is the built.toml of the installed theme, zero without one.
	OldBuilt themecore.Built
	NewBuilt themecore.Built
	// Changed, Added and Removed are the files whose entries differ between
	// the two checksums.json, sorted. They are empty without an installed
	// theme.
	Changed []string
	Added   []string
	Removed []string
	// Err tells why a built.toml or checksums.json of either side could not
	// be read. Such a pending install can only be rejected.
	Err error
}

// checksumsFile is the part of checksums.json that the listing reads.
type checksumsFile struct {
	Files map[string]string `json:"files"`
}

// ListPending returns the pending installs under pendingDir, sorted by path,
// with what differs from the installed theme under themesDir. Directories
// that are not org/id/version are skipped. Without pendingDir there are none.
func ListPending(themesDir, pendingDir string) ([]PendingInstall, error) {
	orgs, err := os.ReadDir(pendingDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []PendingInstall
	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}
		ids, err := os.ReadDir(filepath.Join(pendingDir, org.Name()))
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if !id.IsDir() {
				continue
			}
			versions, err := os.ReadDir(filepath.Join(pendingDir, org.Name(), id.Name()))
			if err != nil {
				return nil, err
			}
			for _, version := range versions {
				path := org.Name() + "/" + id.Name() + "/" + version.Name()
				if !version.IsDir() || !themePathPattern.MatchString(path) {
					continue
				}
				list = append(list, pendingInstall(themesDir, pendingDir, path))
			}
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Path < list[j].Path })
	return list, nil
}

// pendingInstall reads both sides of the pending install at path.
func pendingInstall(themesDir, pendingDir, path string) PendingInstall {
	p := PendingInstall{Path: path}
	newBuilt, newSums, err := readSide(filepath.Join(pendingDir, filepath.FromSlash(path)))
	if err != nil {
		p.Err = err
		return p
	}
	p.NewBuilt = newBuilt

	oldDir := filepath.Join(themesDir, filepath.FromSlash(path))
	if _, err := os.Stat(oldDir); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			p.Err = err
		}
		return p
	}
	p.Installed = true
	oldBuilt, oldSums, err := readSide(oldDir)
	if err != nil {
		p.Err = err
		return p
	}
	p.OldBuilt = oldBuilt

	for name, sum := range newSums {
		old, ok := oldSums[name]
		switch {
		case !ok:
			p.Added = append(p.Added, name)
		case old != sum:
			p.Changed = append(p.Changed, name)
		}
	}
	for name := range oldSums {
		if _, ok := newSums[name]; !ok {
			p.Removed = append(p.Removed, name)
		}
	}
	sort.Strings(p.Changed)
	sort.Strings(p.Added)
	sort.Strings(p.Removed)
	return p
}

// readSide reads the built.toml and the checksums.json of the package at dir.
func readSide(dir string) (themecore.Built, map[string]string, error) {
	b, err := readManifest(dir, "built.toml")
	if err != nil {
		return themecore.Built{}, nil, err
	}
	built, err := themecore.ParseBuilt(b)
	if err != nil {
		return themecore.Built{}, nil, err
	}
	b, err = readManifest(dir, "checksums.json")
	if err != nil {
		return themecore.Built{}, nil, err
	}
	var sums checksumsFile
	if err := json.Unmarshal(b, &sums); err != nil {
		return themecore.Built{}, nil, fmt.Errorf("checksums.json: %w", err)
	}
	return built, sums.Files, nil
}

// readManifest reads the file name of the package at dir, up to themecore's
// size limit of the manifest files.
func readManifest(dir, name string) ([]byte, error) {
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxContractSize+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxContractSize {
		return nil, fmt.Errorf("%s: larger than %d bytes", name, maxContractSize)
	}
	return b, nil
}

// RejectPending deletes the pending install at path and the parent
// directories it leaves empty under pendingDir. The installed theme and the
// database are not touched.
func RejectPending(pendingDir, path string) error {
	if !themePathPattern.MatchString(path) {
		return fmt.Errorf("%w: %q", ErrInvalidThemePath, path)
	}
	dir := filepath.Join(pendingDir, filepath.FromSlash(path))
	if err := removeReadOnly(dir); err != nil {
		return err
	}
	return removeEmptyParents(pendingDir, dir)
}

// removeEmptyParents removes the parent directories of dir under root that
// are empty, from the innermost.
func removeEmptyParents(root, dir string) error {
	root = filepath.Clean(root)
	for parent := filepath.Dir(dir); parent != root; parent = filepath.Dir(parent) {
		entries, err := os.ReadDir(parent)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return nil
		}
		if err := os.Remove(parent); err != nil {
			return err
		}
	}
	return nil
}

// removeReadOnly deletes dir, whose files and directories are read-only
// (0444 and 0555), by making its directories writable first.
func removeReadOnly(dir string) error {
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.Chmod(p, 0o755)
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.RemoveAll(dir)
}

// Replace moves the pending install at path over the installed theme: the
// installed directory is renamed aside, the pending one is renamed into its
// place and the old one is deleted. themesDir and pendingDir must be on the
// same file system. Moving a directory to another parent needs write
// permission on it, so the read-only pending directory is made writable for
// the move and gets its mode back after it.
func Replace(themesDir, pendingDir, path string) error {
	if !themePathPattern.MatchString(path) {
		return fmt.Errorf("%w: %q", ErrInvalidThemePath, path)
	}
	dst := filepath.Join(themesDir, filepath.FromSlash(path))
	src := filepath.Join(pendingDir, filepath.FromSlash(path))
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	mode := info.Mode().Perm()
	aside := dst + ".replaced-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	if err := os.Chmod(src, mode|0o200); err != nil {
		return err
	}
	if err := os.Rename(dst, aside); err != nil {
		_ = os.Chmod(src, mode)
		return err
	}
	if err := os.Rename(src, dst); err != nil {
		_ = os.Rename(aside, dst)
		_ = os.Chmod(src, mode)
		return err
	}
	if err := os.Chmod(dst, mode); err != nil {
		return err
	}
	if err := removeReadOnly(aside); err != nil {
		return err
	}
	return removeEmptyParents(pendingDir, src)
}
