package service

import (
	"context"
	"reflect"
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

func TestContentCompilerBuildsContextFromProviders(t *testing.T) {
	provider := &recordingContextProvider{
		values: map[string]any{
			"app/version": "1.2.0",
			"unused/key":  "unused",
		},
	}
	compiler, err := NewContentCompiler(nil, provider)
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.CompileWithProviderContext(context.Background(), "Version: {{ context(app/version) }}")
	if err != nil {
		t.Fatalf("CompileWithProviderContext() error = %v", err)
	}

	if strings.TrimSpace(output) != "<p>Version: 1.2.0</p>" {
		t.Fatalf("CompileWithProviderContext() = %q", output)
	}
	if !reflect.DeepEqual(provider.calls, [][]string{{"app/version"}}) {
		t.Fatalf("provider keys = %#v", provider.calls)
	}
}

func TestContentCompilerPassesEmptyKeysWhenSourceHasNoContextReferences(t *testing.T) {
	provider := &recordingContextProvider{
		values: map[string]any{
			"app/version": "1.2.0",
		},
	}
	compiler, err := NewContentCompiler(nil, provider)
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.CompileWithProviderContext(context.Background(), "No context here.")
	if err != nil {
		t.Fatalf("CompileWithProviderContext() error = %v", err)
	}

	if strings.TrimSpace(output) != "<p>No context here.</p>" {
		t.Fatalf("CompileWithProviderContext() = %q", output)
	}
	if !reflect.DeepEqual(provider.calls, [][]string{{}}) {
		t.Fatalf("provider keys = %#v", provider.calls)
	}
}

func TestContentCompilerBuildsProviderContextForGlobalComponents(t *testing.T) {
	provider := &recordingContextProvider{
		values: map[string]any{
			"user/name": "Jane",
		},
	}
	compiler, err := NewContentCompiler([]models.Component{
		{Name: "GREETING", Content: "name = context(user/name)\nHello {{ name }}"},
	}, provider)
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.CompileWithProviderContext(context.Background(), "{{ GREETING }}")
	if err != nil {
		t.Fatalf("CompileWithProviderContext() error = %v", err)
	}

	if strings.TrimSpace(output) != "<p>Hello Jane</p>" {
		t.Fatalf("CompileWithProviderContext() = %q", output)
	}
	if !reflect.DeepEqual(provider.calls, [][]string{{"user/name"}}) {
		t.Fatalf("provider keys = %#v", provider.calls)
	}
}

type recordingContextProvider struct {
	values map[string]any
	calls  [][]string
}

func (p *recordingContextProvider) BuildCompileContext(_ context.Context, keys []string) (map[string]any, error) {
	if keys == nil {
		p.calls = append(p.calls, nil)

		values := map[string]any{}
		for key, value := range p.values {
			values[key] = value
		}
		return values, nil
	}

	copiedKeys := make([]string, len(keys))
	copy(copiedKeys, keys)
	p.calls = append(p.calls, copiedKeys)

	values := map[string]any{}
	for _, key := range keys {
		if value, ok := p.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}
