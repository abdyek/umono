package themeapi

import "testing"

func TestParseTxtar(t *testing.T) {
	input := []byte("comment\n-- a.txt --\nA\n-- globals/CARD.comp --\nB\nC\n")
	files, err := parseTxtar(input)
	if err != nil {
		t.Fatalf("parseTxtar returned error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if got := string(files["a.txt"]); got != "A\n" {
		t.Errorf("files[\"a.txt\"] = %q, want %q", got, "A\n")
	}
	if got := string(files["globals/CARD.comp"]); got != "B\nC\n" {
		t.Errorf("files[\"globals/CARD.comp\"] = %q, want %q", got, "B\nC\n")
	}
}

func TestParseTxtarRejectsDuplicates(t *testing.T) {
	input := []byte("-- a --\nx\n-- a --\ny\n")
	_, err := parseTxtar(input)
	if err == nil {
		t.Fatalf("expected error for duplicate file names, got nil")
	}
}

func TestScenarioLayout(t *testing.T) {
	layouts := []string{"post", "post-list", "page"}
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"post-list-empty", "post-list", false},
		{"post-block", "post", false},
		{"post", "post", false},
		{"page-link", "page", false},
		{"pages", "", true},
		{"other-x", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scenarioLayout(tt.name, layouts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("scenarioLayout(%q) error = %v, wantErr = %v", tt.name, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("scenarioLayout(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
