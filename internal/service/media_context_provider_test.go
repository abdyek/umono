package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/umono-cms/umono/internal/models"
)

func TestMediaContextProviderBuildsMediaByIDContext(t *testing.T) {
	resolver := newFakeMediaContextResolver()
	resolver.mediaByID["media-1"] = models.Media{
		ID:       "media-1",
		MimeType: "image/png",
		Metadata: models.JSONMap{
			"width":  640,
			"height": 320,
			"alias":  "hero-image",
		},
		Variants: []models.MediaVariant{
			{
				ID:       "png-640",
				MediaID:  "media-1",
				MimeType: "image/png",
				Metadata: models.JSONMap{"width": 640, "height": 320, "alias": "ignored"},
			},
			{
				ID:       "webp-640",
				MediaID:  "media-1",
				MimeType: "image/webp",
				Metadata: models.JSONMap{"width": 640, "height": 320},
			},
			{
				ID:       "png-320",
				MediaID:  "media-1",
				MimeType: "image/png",
				Metadata: models.JSONMap{"width": 320, "height": 160},
			},
			{
				ID:       "webp-320",
				MediaID:  "media-1",
				MimeType: "image/webp",
				Metadata: models.JSONMap{"width": 320, "height": 160},
			},
		},
	}

	provider := NewMediaContextProvider(resolver)
	values, err := provider.BuildCompileContext(context.Background(), []string{
		"app/version",
		"media-by-id/media-1",
	})
	if err != nil {
		t.Fatalf("BuildCompileContext() error = %v", err)
	}

	value, ok := values["media-by-id/media-1"].(map[string]any)
	if !ok {
		t.Fatalf("media context value = %#v", values["media-by-id/media-1"])
	}

	want := map[string]any{
		"url":       "/media/media-1",
		"width":     640,
		"height":    320,
		"mime-type": "image/png",
		"variants": []map[string]any{
			{"url": "/variants/webp-320", "width": 320, "height": 160, "mime-type": "image/webp"},
			{"url": "/variants/webp-640", "width": 640, "height": 320, "mime-type": "image/webp"},
			{"url": "/variants/png-320", "width": 320, "height": 160, "mime-type": "image/png"},
			{"url": "/variants/png-640", "width": 640, "height": 320, "mime-type": "image/png"},
		},
	}
	if !reflect.DeepEqual(value, want) {
		t.Fatalf("media context value = %#v", value)
	}
	if _, ok := value["id"]; ok {
		t.Fatal("media context must not expose media id")
	}
	if _, ok := value["alias"]; ok {
		t.Fatal("media context must not expose media alias")
	}

	for _, variant := range value["variants"].([]map[string]any) {
		if _, ok := variant["id"]; ok {
			t.Fatal("variant context must not expose variant id")
		}
		if _, ok := variant["alias"]; ok {
			t.Fatal("variant context must not expose variant alias")
		}
	}
	if resolver.idCalls != 1 {
		t.Fatalf("resolver id calls = %d", resolver.idCalls)
	}
}

func TestMediaContextProviderBuildsMediaByAliasContext(t *testing.T) {
	resolver := newFakeMediaContextResolver()
	resolver.mediaByAlias["hero"] = models.Media{
		ID:       "media-1",
		MimeType: "image/jpeg",
		Metadata: models.JSONMap{
			"width":  float64(800),
			"height": float64(450),
		},
	}

	values, err := NewMediaContextProvider(resolver).BuildCompileContext(context.Background(), []string{
		"media-by-alias/hero",
	})
	if err != nil {
		t.Fatalf("BuildCompileContext() error = %v", err)
	}

	value := values["media-by-alias/hero"].(map[string]any)
	if value["url"] != "/media/media-1" {
		t.Fatalf("url = %q", value["url"])
	}
	if got := value["variants"]; !reflect.DeepEqual(got, []map[string]any{}) {
		t.Fatalf("variants = %#v", got)
	}
	if resolver.aliasCalls != 1 {
		t.Fatalf("resolver alias calls = %d", resolver.aliasCalls)
	}
}

func TestMediaContextProviderSortsWebPBeforeJPEGVariants(t *testing.T) {
	resolver := newFakeMediaContextResolver()
	resolver.mediaByID["media-1"] = models.Media{
		ID:       "media-1",
		MimeType: "image/jpeg",
		Metadata: models.JSONMap{"width": 640, "height": 320},
		Variants: []models.MediaVariant{
			{ID: "jpeg-640", MediaID: "media-1", MimeType: "image/jpeg", Metadata: models.JSONMap{"width": 640, "height": 320}},
			{ID: "webp-640", MediaID: "media-1", MimeType: "image/webp", Metadata: models.JSONMap{"width": 640, "height": 320}},
			{ID: "jpeg-320", MediaID: "media-1", MimeType: "image/jpeg", Metadata: models.JSONMap{"width": 320, "height": 160}},
			{ID: "webp-320", MediaID: "media-1", MimeType: "image/webp", Metadata: models.JSONMap{"width": 320, "height": 160}},
		},
	}

	values, err := NewMediaContextProvider(resolver).BuildCompileContext(context.Background(), []string{
		"media-by-id/media-1",
	})
	if err != nil {
		t.Fatalf("BuildCompileContext() error = %v", err)
	}

	variants := values["media-by-id/media-1"].(map[string]any)["variants"].([]map[string]any)
	got := make([]string, 0, len(variants))
	for _, variant := range variants {
		got = append(got, variant["url"].(string))
	}
	want := []string{
		"/variants/webp-320",
		"/variants/webp-640",
		"/variants/jpeg-320",
		"/variants/jpeg-640",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("variant order = %#v", got)
	}
}

func TestMediaContextProviderFeedsImageBuiltin(t *testing.T) {
	resolver := newFakeMediaContextResolver()
	resolver.mediaByID["media-1"] = models.Media{
		ID:       "media-1",
		MimeType: "image/png",
		Metadata: models.JSONMap{"width": 640, "height": 320},
		Variants: []models.MediaVariant{
			{ID: "png-640", MediaID: "media-1", MimeType: "image/png", Metadata: models.JSONMap{"width": 640, "height": 320}},
			{ID: "webp-320", MediaID: "media-1", MimeType: "image/webp", Metadata: models.JSONMap{"width": 320, "height": 160}},
			{ID: "png-320", MediaID: "media-1", MimeType: "image/png", Metadata: models.JSONMap{"width": 320, "height": 160}},
			{ID: "webp-640", MediaID: "media-1", MimeType: "image/webp", Metadata: models.JSONMap{"width": 640, "height": 320}},
		},
	}
	compiler, err := NewContentCompiler(nil, NewMediaContextProvider(resolver))
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.CompileWithProviderContext(context.Background(), `{{ IMAGE media = context(media-by-id/media-1) alt = "Hero" }}`)
	if err != nil {
		t.Fatalf("CompileWithProviderContext() error = %v", err)
	}

	want := `<picture><source type="image/webp" srcset="/variants/webp-320 320w, /variants/webp-640 640w"><source type="image/png" srcset="/variants/png-320 320w, /variants/png-640 640w"><img src="/media/media-1" alt="Hero" width="640" height="320"></picture>`
	if strings.TrimSpace(output) != want {
		t.Fatalf("output = %q", output)
	}
}

func TestMediaContextProviderFeedsImageBuiltinWithoutVariants(t *testing.T) {
	resolver := newFakeMediaContextResolver()
	resolver.mediaByID["media-1"] = models.Media{
		ID:       "media-1",
		MimeType: "image/jpeg",
		Metadata: models.JSONMap{"width": 800, "height": 450},
	}
	compiler, err := NewContentCompiler(nil, NewMediaContextProvider(resolver))
	if err != nil {
		t.Fatalf("NewContentCompiler() error = %v", err)
	}

	output, err := compiler.CompileWithProviderContext(context.Background(), `{{ IMAGE media = context(media-by-id/media-1) alt = "Hero" }}`)
	if err != nil {
		t.Fatalf("CompileWithProviderContext() error = %v", err)
	}

	want := `<img src="/media/media-1" alt="Hero" width="800" height="450">`
	if strings.TrimSpace(output) != want {
		t.Fatalf("output = %q", output)
	}
}

type fakeMediaContextResolver struct {
	mediaByID    map[string]models.Media
	mediaByAlias map[string]models.Media
	idCalls      int
	aliasCalls   int
}

func newFakeMediaContextResolver() *fakeMediaContextResolver {
	return &fakeMediaContextResolver{
		mediaByID:    map[string]models.Media{},
		mediaByAlias: map[string]models.Media{},
	}
}

func (r *fakeMediaContextResolver) GetByIDWithVariants(id string) (models.Media, error) {
	r.idCalls++
	item, ok := r.mediaByID[id]
	if !ok {
		return models.Media{}, ErrMediaNotFound
	}
	return item, nil
}

func (r *fakeMediaContextResolver) GetByAliasWithVariants(alias string) (models.Media, error) {
	r.aliasCalls++
	item, ok := r.mediaByAlias[alias]
	if !ok {
		return models.Media{}, ErrMediaNotFound
	}
	return item, nil
}

func (r *fakeMediaContextResolver) DirectURL(_ context.Context, item models.Media) (string, error) {
	if item.ID == "broken" {
		return "", errors.New("broken media")
	}
	return "/media/" + item.ID, nil
}

func (r *fakeMediaContextResolver) VariantDirectURL(_ context.Context, variant models.MediaVariant) (string, error) {
	if variant.ID == "broken" {
		return "", errors.New("broken variant")
	}
	return "/variants/" + variant.ID, nil
}
