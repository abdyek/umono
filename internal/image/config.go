package image

import (
	"sort"
	"strings"
)

const (
	MimeJPEG = "image/jpeg"
	MimePNG  = "image/png"
	MimeWebP = "image/webp"
)

type VariantGenerationConfig struct {
	Version         int
	Widths          []int
	FormatsPolicy   string
	JPEGQuality     int
	WebPQuality     int
	ResizeAlgorithm string
	NormalizeExif   bool
}

var DefaultVariantGenerationConfig = VariantGenerationConfig{
	Version:         1,
	Widths:          []int{160, 320, 640, 960, 1280, 1920},
	FormatsPolicy:   "v1",
	JPEGQuality:     85,
	WebPQuality:     82,
	ResizeAlgorithm: "lanczos",
	NormalizeExif:   true,
}

type SourceInfo struct {
	MimeType    string
	Width       int
	Height      int
	Orientation int
	Animated    bool
	HasAlpha    bool
}

type VariantTarget struct {
	Width    int
	MimeType string
}

func PlanVariants(info SourceInfo, cfg VariantGenerationConfig) []VariantTarget {
	if info.Animated || info.Width <= 0 {
		return nil
	}

	formats := targetFormats(info)
	if len(formats) == 0 {
		return nil
	}

	widthSet := map[int]struct{}{}
	for _, width := range cfg.Widths {
		if width > 0 && width <= info.Width {
			widthSet[width] = struct{}{}
		}
	}
	widthSet[info.Width] = struct{}{}

	widths := make([]int, 0, len(widthSet))
	for width := range widthSet {
		widths = append(widths, width)
	}
	sort.Ints(widths)

	targets := make([]VariantTarget, 0, len(widths)*len(formats))
	for _, width := range widths {
		for _, format := range formats {
			if width == info.Width && format == info.MimeType {
				continue
			}
			targets = append(targets, VariantTarget{
				Width:    width,
				MimeType: format,
			})
		}
	}

	return targets
}

func ExtensionByMimeType(mimeType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case MimePNG:
		return "png", true
	case MimeWebP:
		return "webp", true
	case MimeJPEG:
		return "jpeg", true
	default:
		return "", false
	}
}

func targetFormats(info SourceInfo) []string {
	switch strings.ToLower(strings.TrimSpace(info.MimeType)) {
	case MimeJPEG:
		return []string{MimeWebP, MimeJPEG}
	case MimePNG:
		return []string{MimeWebP, MimePNG}
	case MimeWebP:
		if info.HasAlpha {
			return []string{MimeWebP, MimePNG}
		}
		return []string{MimeWebP, MimeJPEG}
	default:
		return nil
	}
}
