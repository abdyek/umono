package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/umono-cms/umono/internal/models"
)

const (
	mediaByIDContextPrefix    = "media-by-id/"
	mediaByAliasContextPrefix = "media-by-alias/"
)

type MediaContextResolver interface {
	GetByIDWithVariants(id string) (models.Media, error)
	GetByAliasWithVariants(alias string) (models.Media, error)
	DirectURL(ctx context.Context, item models.Media) (string, error)
	VariantDirectURL(ctx context.Context, variant models.MediaVariant) (string, error)
}

type MediaContextProvider struct {
	resolver MediaContextResolver
}

type mediaContextVariant struct {
	url      string
	width    int
	height   int
	mimeType string
	pathKey  string
	original bool
}

func NewMediaContextProvider(resolver MediaContextResolver) *MediaContextProvider {
	return &MediaContextProvider{resolver: resolver}
}

func (p *MediaContextProvider) BuildCompileContext(ctx context.Context, keys []string) (map[string]any, error) {
	values := map[string]any{}
	if p == nil || p.resolver == nil {
		return values, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for _, key := range keys {
		item, ok, err := p.resolveMedia(key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		value, err := p.mediaContextValue(ctx, item)
		if err != nil {
			return nil, err
		}
		values[key] = value
	}

	return values, nil
}

func (p *MediaContextProvider) resolveMedia(key string) (models.Media, bool, error) {
	switch {
	case strings.HasPrefix(key, mediaByIDContextPrefix):
		id := strings.TrimSpace(strings.TrimPrefix(key, mediaByIDContextPrefix))
		if id == "" {
			return models.Media{}, false, nil
		}

		item, err := p.resolver.GetByIDWithVariants(id)
		if errors.Is(err, ErrMediaNotFound) {
			return models.Media{}, false, nil
		}
		return item, err == nil, err
	case strings.HasPrefix(key, mediaByAliasContextPrefix):
		alias := strings.TrimSpace(strings.TrimPrefix(key, mediaByAliasContextPrefix))
		if alias == "" {
			return models.Media{}, false, nil
		}

		item, err := p.resolver.GetByAliasWithVariants(alias)
		if errors.Is(err, ErrMediaNotFound) {
			return models.Media{}, false, nil
		}
		return item, err == nil, err
	default:
		return models.Media{}, false, nil
	}
}

func (p *MediaContextProvider) mediaContextValue(ctx context.Context, item models.Media) (map[string]any, error) {
	url, err := p.resolver.DirectURL(ctx, item)
	if err != nil {
		return nil, err
	}

	variants, err := p.variantContextValues(ctx, item, url)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"url":       url,
		"width":     mediaContextJSONMapInt(item.Metadata, "width"),
		"height":    mediaContextJSONMapInt(item.Metadata, "height"),
		"mime-type": item.MimeType,
		"variants":  variants,
	}, nil
}

func (p *MediaContextProvider) variantContextValues(ctx context.Context, item models.Media, itemURL string) ([]map[string]any, error) {
	sorted := []mediaContextVariant{
		{
			url:      itemURL,
			width:    mediaContextJSONMapInt(item.Metadata, "width"),
			height:   mediaContextJSONMapInt(item.Metadata, "height"),
			mimeType: item.MimeType,
			pathKey:  item.PathKey,
			original: true,
		},
	}

	for _, variant := range item.Variants {
		url, err := p.resolver.VariantDirectURL(ctx, variant)
		if err != nil {
			return nil, err
		}

		sorted = append(sorted, mediaContextVariant{
			url:      url,
			width:    mediaContextJSONMapInt(variant.Metadata, "width"),
			height:   mediaContextJSONMapInt(variant.Metadata, "height"),
			mimeType: variant.MimeType,
			pathKey:  variant.PathKey,
		})
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		iRank := mediaContextMimeTypeRank(sorted[i].mimeType)
		jRank := mediaContextMimeTypeRank(sorted[j].mimeType)
		if iRank != jRank {
			return iRank < jRank
		}

		if sorted[i].width != sorted[j].width {
			return sorted[i].width < sorted[j].width
		}

		if sorted[i].original != sorted[j].original {
			return !sorted[i].original
		}

		return sorted[i].pathKey < sorted[j].pathKey
	})

	values := make([]map[string]any, 0, len(sorted))
	for _, variant := range sorted {
		values = append(values, map[string]any{
			"url":       variant.url,
			"width":     variant.width,
			"height":    variant.height,
			"mime-type": variant.mimeType,
		})
	}

	return values, nil
}

func mediaContextMimeTypeRank(mimeType string) int {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/webp":
		return 0
	case "image/png":
		return 1
	case "image/jpeg", "image/jpg":
		return 2
	default:
		return 3
	}
}

func mediaContextJSONMapInt(metadata models.JSONMap, key string) int {
	if metadata == nil {
		return 0
	}

	switch value := metadata[key].(type) {
	case int:
		return value
	case int8:
		return int(value)
	case int16:
		return int(value)
	case int32:
		return int(value)
	case int64:
		return int(value)
	case uint:
		return int(value)
	case uint8:
		return int(value)
	case uint16:
		return int(value)
	case uint32:
		return int(value)
	case uint64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}
