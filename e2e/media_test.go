//go:build e2e

package e2e

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

func TestUploadMedia(t *testing.T) {
	p := newPage(t)
	p.login()

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 60, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	if _, err := p.Goto("/admin/media/new"); err != nil {
		t.Fatalf("goto /admin/media/new: %v", err)
	}
	if err := p.Locator("#media-file").SetInputFiles([]playwright.InputFile{{Name: "e2e-pixel.png", MimeType: "image/png", Buffer: buf.Bytes()}}); err != nil {
		t.Fatalf("set media file: %v", err)
	}
	if err := p.Locator("#media-alias").Fill("e2e-pixel"); err != nil {
		t.Fatalf("fill media alias: %v", err)
	}
	if err := p.Locator(`form[data-media-upload-form] button[type="submit"]`).Click(); err != nil {
		t.Fatalf("click media submit: %v", err)
	}
	if err := p.WaitForURL(baseURL + "/admin/media"); err != nil {
		t.Fatalf("wait for media list url: %v", err)
	}
	if err := expect.Locator(p.Locator("#media-content")).ToContainText("e2e-pixel"); err != nil {
		t.Fatalf("assert media list contains alias: %v", err)
	}
}
