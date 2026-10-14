// Command domgen freezes the DOM corpus in the directory of its argument: it
// writes the golden of every .comp input, removes the goldens without one and
// writes manifest.json. It refuses to change or drop a frozen golden unless
// themecore's ledger has a Break line since the manifest was written.
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/umono-cms/themecore"
	"github.com/umono-cms/umono/internal/themeapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: domgen <corpus dir>")
	}
	dir := os.Args[1]

	goldens, err := themeapi.RenderCorpus(dir)
	if err != nil {
		log.Fatal(err)
	}

	manifestPath := filepath.Join(dir, "manifest.json")
	old, err := themeapi.ReadManifest(manifestPath)
	if err != nil {
		log.Fatal(err)
	}

	hashes := themeapi.Hashes(goldens)
	min := themecore.ThemeAPIMin()
	if err := themeapi.CheckBreak(old, hashes, min); err != nil {
		log.Fatal(err)
	}

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".golden") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if _, ok := goldens[rel]; !ok {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	for rel, content := range goldens {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			log.Fatal(err)
		}
	}

	if err := themeapi.WriteManifest(manifestPath, themeapi.Manifest{ThemeAPIMin: min, Goldens: hashes}); err != nil {
		log.Fatal(err)
	}
}
