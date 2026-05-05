package service

import (
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func TestContentCompilerLoadsGlobalComponents(t *testing.T) {
	compiler, err := NewContentCompiler([]models.Component{
		{Name: "GREETING", Content: "# Hello"},
	})
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.Compile("{{ GREETING }}")
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if strings.TrimSpace(output) != "<h1>Hello</h1>" {
		t.Fatalf("Compile() = %q", output)
	}
}

func TestContentCompilerPreviewComponent(t *testing.T) {
	compiler, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.PreviewComponent("GREETING", "# Hello")
	if err != nil {
		t.Fatalf("PreviewComponent() error = %v", err)
	}

	if strings.TrimSpace(output) != "<h1>Hello</h1>" {
		t.Fatalf("PreviewComponent() = %q", output)
	}
}

func TestContentCompilerBuildsComponoContext(t *testing.T) {
	compiler, err := NewContentCompiler(nil)
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.CompileWithContext("Version: {{ context(app/version) }}", map[string]any{
		"app/version": "1.2.0",
	})
	if err != nil {
		t.Fatalf("CompileWithContext() error = %v", err)
	}

	if strings.TrimSpace(output) != "<p>Version: 1.2.0</p>" {
		t.Fatalf("CompileWithContext() = %q", output)
	}
}
