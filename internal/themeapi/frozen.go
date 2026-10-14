package themeapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/umono-cms/compono"
)

// Manifest records the frozen DOM: the SHA-256 of each golden by its path
// under the corpus directory, and the ThemeAPIMin the corpus was frozen at.
type Manifest struct {
	ThemeAPIMin int               `json:"theme_api_min"`
	Goldens     map[string]string `json:"goldens"`
}

// RenderCorpus converts every .comp file under dir with Compono and returns
// the content of its golden, the canonical form of the output and a newline,
// by the golden's slash path relative to dir: x.comp has the golden x.golden.
// A diagnostic is an error, because the corpus holds only error-free inputs.
func RenderCorpus(dir string) (map[string]string, error) {
	c := compono.New()
	goldens := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".comp") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		diagnostics, err := c.Convert(src, &buf)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if len(diagnostics) > 0 {
			return fmt.Errorf("%s: %s: %s", rel, diagnostics[0].Code, diagnostics[0].Message)
		}
		canonical, err := Canonical(buf.Bytes())
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		goldens[strings.TrimSuffix(rel, ".comp")+".golden"] = canonical + "\n"
		return nil
	})
	if err != nil {
		return nil, err
	}
	return goldens, nil
}

// Hashes returns the SHA-256 in hex of each golden by its path.
func Hashes(goldens map[string]string) map[string]string {
	hashes := make(map[string]string, len(goldens))
	for path, content := range goldens {
		sum := sha256.Sum256([]byte(content))
		hashes[path] = hex.EncodeToString(sum[:])
	}
	return hashes
}

// CheckBreak returns an error when hashes change or drop a golden of old while
// min is not greater than old.ThemeAPIMin: the frozen DOM can only change with
// a Break line in themecore's ledger. A new golden needs no Break.
func CheckBreak(old Manifest, hashes map[string]string, min int) error {
	if min > old.ThemeAPIMin {
		return nil
	}
	var paths []string
	for path, oldHash := range old.Goldens {
		if newHash, ok := hashes[path]; !ok || newHash != oldHash {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil
	}
	return fmt.Errorf("the frozen DOM changed but themecore's ledger has no Break line: %s", strings.Join(paths, ", "))
}

// ReadManifest reads the manifest at path. A missing file is an empty
// manifest.
func ReadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Manifest{Goldens: map[string]string{}}, nil
		}
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, err
	}
	if m.Goldens == nil {
		m.Goldens = map[string]string{}
	}
	return m, nil
}

// WriteManifest writes m to path as indented JSON with a trailing newline.
func WriteManifest(path string, m Manifest) error {
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}
