package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUpdateKeepsOtherKeys(t *testing.T) {
	in, err := json.Marshal(map[string]any{"min_cli_version": "0.2.2", "theme_api_min": 7})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	out, err := update(in, 3, 2)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		t.Fatalf("update output does not end with newline: got %q", out)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("json.Unmarshal(out): %v", err)
	}
	want := map[string]any{
		"min_cli_version":   "0.2.2",
		"theme_api_current": float64(3),
		"theme_api_min":     float64(2),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("update result mismatch: got %v, want %v", got, want)
	}
}

func TestUpdateRejectsNonObject(t *testing.T) {
	if _, err := update([]byte("[]"), 1, 1); err == nil {
		t.Fatalf("update([]byte(\"[]\"), 1, 1) returned no error; want an error")
	}
}
