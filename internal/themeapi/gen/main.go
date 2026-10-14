// Command gen writes themecore's Theme API range to the theme_api_current and
// theme_api_min keys of the umono.json at the path in its argument. It keeps
// every other key.
package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/umono-cms/themecore"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: gen <path-to-umono.json>")
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	out, err := update(data, themecore.ThemeAPICurrent(), themecore.ThemeAPIMin())
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		log.Fatal(err)
	}
}

// update sets theme_api_current and theme_api_min in the JSON object data and
// keeps every other key.
func update(data []byte, current, min int) ([]byte, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	object["theme_api_current"] = json.RawMessage(strconv.Itoa(current))
	object["theme_api_min"] = json.RawMessage(strconv.Itoa(min))
	out, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
