package themeapi

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/umono-cms/themecore"
)

func TestCheckBreak(t *testing.T) {
	old := Manifest{ThemeAPIMin: 1, Goldens: map[string]string{"a.golden": "1", "b.golden": "2"}}

	cases := []struct {
		name     string
		hashes   map[string]string
		min      int
		wantErr  bool
		contains []string
	}{
		{
			name:    "unchanged",
			hashes:  map[string]string{"a.golden": "1", "b.golden": "2"},
			min:     1,
			wantErr: false,
		},
		{
			name:    "new golden",
			hashes:  map[string]string{"a.golden": "1", "b.golden": "2", "c.golden": "3"},
			min:     1,
			wantErr: false,
		},
		{
			name:     "changed",
			hashes:   map[string]string{"a.golden": "9", "b.golden": "2"},
			min:      1,
			wantErr:  true,
			contains: []string{"a.golden", "Break"},
		},
		{
			name:     "dropped",
			hashes:   map[string]string{"a.golden": "1"},
			min:      1,
			wantErr:  true,
			contains: []string{"b.golden", "Break"},
		},
		{
			name:    "changed with a Break",
			hashes:  map[string]string{"a.golden": "9"},
			min:     2,
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckBreak(old, tc.hashes, tc.min)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				for _, s := range tc.contains {
					if !strings.Contains(err.Error(), s) {
						t.Errorf("error %q does not contain %q", err.Error(), s)
					}
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestFrozenDOM(t *testing.T) {
	const dir = "testdata/dom"

	goldens, err := RenderCorpus(dir)
	if err != nil {
		t.Fatalf("RenderCorpus: %v", err)
	}
	if len(goldens) == 0 {
		t.Fatalf("RenderCorpus returned no goldens")
	}

	paths := make([]string, 0, len(goldens))
	for p := range goldens {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		want := goldens[path]
		full := filepath.Join(dir, filepath.FromSlash(path))
		data, err := os.ReadFile(full)
		if err != nil {
			t.Errorf("%s: no golden; run go generate ./internal/themeapi", path)
			continue
		}
		if string(data) != want {
			t.Errorf("%s: the frozen DOM changed: Compono's output no longer matches the golden", path)
		}
	}

	seen := make(map[string]bool)
	for _, p := range paths {
		seen[p] = true
	}

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".golden") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !seen[rel] {
			t.Errorf("%s: golden without a .comp input", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}

	manifest, err := ReadManifest(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}
	hashes := Hashes(goldens)
	if err := CheckBreak(manifest, hashes, themecore.ThemeAPIMin()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(manifest.Goldens, hashes) || manifest.ThemeAPIMin != themecore.ThemeAPIMin() {
		t.Fatal("manifest.json is stale; run go generate ./internal/themeapi")
	}
}
