package themeapi

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/umono-cms/themecore"
)

type umonoJSON struct {
	ThemeAPICurrent int `json:"theme_api_current"`
	ThemeAPIMin     int `json:"theme_api_min"`
}

func readUmonoJSON(t *testing.T) umonoJSON {
	t.Helper()
	data, err := os.ReadFile("../../umono.json")
	if err != nil {
		t.Fatalf("os.ReadFile(../../umono.json): %v", err)
	}
	var f umonoJSON
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("json.Unmarshal(umono.json): %v", err)
	}
	return f
}

func TestUmonoJSONMatchesThemecore(t *testing.T) {
	f := readUmonoJSON(t)
	wantCurrent := themecore.ThemeAPICurrent()
	wantMin := themecore.ThemeAPIMin()
	if f.ThemeAPICurrent != wantCurrent || f.ThemeAPIMin != wantMin {
		t.Fatalf("umono.json is stale: theme_api_current = %d, theme_api_min = %d, want %d and %d; run go generate ./internal/themeapi",
			f.ThemeAPICurrent, f.ThemeAPIMin, wantCurrent, wantMin)
	}
}
