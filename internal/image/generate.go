package image

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"

	deepwebp "github.com/deepteams/webp"
	xdraw "golang.org/x/image/draw"
)

type GenerateResult struct {
	Data   []byte
	Width  int
	Height int
}

func GenerateVariant(r io.Reader, sourceMimeType string, target VariantTarget, cfg VariantGenerationConfig) (GenerateResult, error) {
	info, data, err := Inspect(r, sourceMimeType)
	if err != nil {
		return GenerateResult{}, err
	}
	if info.Animated {
		return GenerateResult{}, ErrUnsupportedImage
	}

	img, _, err := decode(data, info.MimeType)
	if err != nil {
		return GenerateResult{}, err
	}
	if cfg.NormalizeExif {
		img = normalizeOrientation(img, info.Orientation)
	}

	bounds := img.Bounds()
	sourceWidth := bounds.Dx()
	sourceHeight := bounds.Dy()
	if target.Width <= 0 || target.Width > sourceWidth {
		target.Width = sourceWidth
	}

	targetHeight := sourceHeight
	if target.Width != sourceWidth {
		targetHeight = int(math.Round(float64(sourceHeight) * float64(target.Width) / float64(sourceWidth)))
		if targetHeight < 1 {
			targetHeight = 1
		}
		img = resize(img, target.Width, targetHeight)
	} else {
		img = copyImage(img)
	}

	var out bytes.Buffer
	switch target.MimeType {
	case MimeJPEG:
		if err := jpeg.Encode(&out, flattenToOpaque(img), &jpeg.Options{Quality: cfg.JPEGQuality}); err != nil {
			return GenerateResult{}, err
		}
	case MimePNG:
		encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
		if err := encoder.Encode(&out, img); err != nil {
			return GenerateResult{}, err
		}
	case MimeWebP:
		opts := deepwebp.DefaultOptions()
		opts.Quality = float32(cfg.WebPQuality)
		opts.Exact = true
		opts.UseSharpYUV = true
		if err := deepwebp.Encode(&out, img, opts); err != nil {
			return GenerateResult{}, err
		}
	default:
		return GenerateResult{}, ErrUnsupportedImage
	}

	return GenerateResult{
		Data:   out.Bytes(),
		Width:  target.Width,
		Height: targetHeight,
	}, nil
}

func resize(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Src, nil)
	return dst
}

func copyImage(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Src)
	return dst
}

func flattenToOpaque(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Over)
	return dst
}

func normalizeOrientation(src image.Image, orientation int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var dst *image.RGBA
	if orientationSwapsDimensions(orientation) {
		dst = image.NewRGBA(image.Rect(0, 0, height, width))
	} else {
		dst = image.NewRGBA(image.Rect(0, 0, width, height))
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := src.At(bounds.Min.X+x, bounds.Min.Y+y)
			var dx, dy int
			switch orientation {
			case 2:
				dx, dy = width-1-x, y
			case 3:
				dx, dy = width-1-x, height-1-y
			case 4:
				dx, dy = x, height-1-y
			case 5:
				dx, dy = y, x
			case 6:
				dx, dy = height-1-y, x
			case 7:
				dx, dy = height-1-y, width-1-x
			case 8:
				dx, dy = y, width-1-x
			default:
				dx, dy = x, y
			}
			dst.Set(dx, dy, c)
		}
	}

	return dst
}
