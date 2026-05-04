package image

import "testing"

func TestPlanVariantsIncludesOriginalWidthAndExpectedFormats(t *testing.T) {
	info := SourceInfo{
		MimeType: MimeJPEG,
		Width:    1400,
		Height:   900,
	}

	targets := PlanVariants(info, DefaultVariantGenerationConfig)

	want := []VariantTarget{
		{Width: 160, MimeType: MimeWebP},
		{Width: 160, MimeType: MimeJPEG},
		{Width: 320, MimeType: MimeWebP},
		{Width: 320, MimeType: MimeJPEG},
		{Width: 640, MimeType: MimeWebP},
		{Width: 640, MimeType: MimeJPEG},
		{Width: 960, MimeType: MimeWebP},
		{Width: 960, MimeType: MimeJPEG},
		{Width: 1280, MimeType: MimeWebP},
		{Width: 1280, MimeType: MimeJPEG},
		{Width: 1400, MimeType: MimeWebP},
	}

	if len(targets) != len(want) {
		t.Fatalf("unexpected target count: got %d want %d", len(targets), len(want))
	}
	for i := range want {
		if targets[i] != want[i] {
			t.Fatalf("target %d: got %#v want %#v", i, targets[i], want[i])
		}
	}
}

func TestPlanVariantsSkipsAnimatedImages(t *testing.T) {
	targets := PlanVariants(SourceInfo{
		MimeType: MimePNG,
		Width:    640,
		Height:   480,
		Animated: true,
	}, DefaultVariantGenerationConfig)

	if len(targets) != 0 {
		t.Fatalf("animated images should not get variants: %#v", targets)
	}
}

func TestPlanVariantsUsesPNGFallbackForWebPAlpha(t *testing.T) {
	targets := PlanVariants(SourceInfo{
		MimeType: MimeWebP,
		Width:    100,
		Height:   100,
		HasAlpha: true,
	}, DefaultVariantGenerationConfig)

	want := []VariantTarget{
		{Width: 100, MimeType: MimePNG},
	}
	if len(targets) != len(want) {
		t.Fatalf("unexpected target count: got %d want %d", len(targets), len(want))
	}
	for i := range want {
		if targets[i] != want[i] {
			t.Fatalf("target %d: got %#v want %#v", i, targets[i], want[i])
		}
	}
}
