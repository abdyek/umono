package image

import (
	"bytes"
	"encoding/binary"
	"errors"
	stdimage "image"
	"io"
	"strings"

	deepwebp "github.com/deepteams/webp"
)

var ErrUnsupportedImage = errors.New("unsupported image")

func Inspect(r io.Reader, mimeType string) (SourceInfo, []byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return SourceInfo{}, nil, err
	}

	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if _, ok := ExtensionByMimeType(mimeType); !ok {
		return SourceInfo{}, nil, ErrUnsupportedImage
	}

	orientation := orientationFromBytes(data, mimeType)
	if orientation == 0 {
		orientation = 1
	}

	animated := isAnimated(data, mimeType)
	width, height, err := configDimensions(data, mimeType)
	if err != nil {
		return SourceInfo{}, nil, err
	}
	if orientationSwapsDimensions(orientation) {
		width, height = height, width
	}

	if animated {
		return SourceInfo{
			MimeType:    mimeType,
			Width:       width,
			Height:      height,
			Orientation: orientation,
			Animated:    true,
			HasAlpha:    hasAlphaMetadata(data, mimeType),
		}, data, nil
	}

	img, _, err := decode(data, mimeType)
	if err != nil {
		return SourceInfo{}, nil, err
	}

	info := SourceInfo{
		MimeType:    mimeType,
		Width:       width,
		Height:      height,
		Orientation: orientation,
		Animated:    false,
		HasAlpha:    hasAlpha(data, mimeType, img),
	}

	return info, data, nil
}

func configDimensions(data []byte, mimeType string) (int, int, error) {
	switch mimeType {
	case MimeWebP:
		cfg, err := deepwebp.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			return cfg.Width, cfg.Height, nil
		}
		return webpDimensions(data)
	case MimeJPEG, MimePNG:
		cfg, _, err := stdimage.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return 0, 0, err
		}
		return cfg.Width, cfg.Height, nil
	default:
		return 0, 0, ErrUnsupportedImage
	}
}

func decode(data []byte, mimeType string) (stdimage.Image, string, error) {
	switch mimeType {
	case MimeWebP:
		img, err := deepwebp.Decode(bytes.NewReader(data))
		return img, "webp", err
	case MimeJPEG, MimePNG:
		return stdimage.Decode(bytes.NewReader(data))
	default:
		return nil, "", ErrUnsupportedImage
	}
}

func hasAlpha(data []byte, mimeType string, img stdimage.Image) bool {
	if hasAlphaMetadata(data, mimeType) {
		return true
	}
	return imageHasTransparency(img)
}

func hasAlphaMetadata(data []byte, mimeType string) bool {
	switch mimeType {
	case MimeJPEG:
		return false
	case MimePNG:
		return pngHasAlphaMetadata(data)
	case MimeWebP:
		return webpHasAlphaMetadata(data)
	default:
		return false
	}
}

func imageHasTransparency(img stdimage.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0xffff {
				return true
			}
		}
	}
	return false
}

func isAnimated(data []byte, mimeType string) bool {
	switch mimeType {
	case MimePNG:
		return pngHasChunk(data, "acTL")
	case MimeWebP:
		return webpHasAnimation(data)
	default:
		return false
	}
}

func pngHasAlphaMetadata(data []byte) bool {
	if len(data) < 33 || !bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		return false
	}
	colorType := data[25]
	return colorType == 4 || colorType == 6 || pngHasChunk(data, "tRNS")
}

func pngHasChunk(data []byte, chunkName string) bool {
	if len(data) < 8 || !bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		return false
	}

	for off := 8; off+8 <= len(data); {
		size := int(binary.BigEndian.Uint32(data[off : off+4]))
		name := string(data[off+4 : off+8])
		off += 8
		if size < 0 || off+size+4 > len(data) {
			return false
		}
		if name == chunkName {
			return true
		}
		off += size + 4
	}

	return false
}

func webpHasAlphaMetadata(data []byte) bool {
	chunks, ok := webpChunks(data)
	if !ok {
		return false
	}
	for _, chunk := range chunks {
		if chunk.name == "VP8X" && len(chunk.data) >= 1 && chunk.data[0]&0x10 != 0 {
			return true
		}
		if chunk.name == "ALPH" {
			return true
		}
	}
	return false
}

func webpHasAnimation(data []byte) bool {
	chunks, ok := webpChunks(data)
	if !ok {
		return false
	}
	for _, chunk := range chunks {
		if chunk.name == "VP8X" && len(chunk.data) >= 1 && chunk.data[0]&0x02 != 0 {
			return true
		}
		if chunk.name == "ANIM" || chunk.name == "ANMF" {
			return true
		}
	}
	return false
}

type webpChunk struct {
	name string
	data []byte
}

func webpChunks(data []byte) ([]webpChunk, bool) {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return nil, false
	}

	var chunks []webpChunk
	for off := 12; off+8 <= len(data); {
		name := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		off += 8
		if size < 0 || off+size > len(data) {
			return chunks, false
		}
		chunks = append(chunks, webpChunk{name: name, data: data[off : off+size]})
		off += size
		if size%2 == 1 {
			off++
		}
	}

	return chunks, true
}

func webpDimensions(data []byte) (int, int, error) {
	chunks, ok := webpChunks(data)
	if !ok {
		return 0, 0, ErrUnsupportedImage
	}

	for _, chunk := range chunks {
		switch chunk.name {
		case "VP8X":
			if len(chunk.data) < 10 {
				return 0, 0, ErrUnsupportedImage
			}
			width := 1 + int(uint32(chunk.data[4])|uint32(chunk.data[5])<<8|uint32(chunk.data[6])<<16)
			height := 1 + int(uint32(chunk.data[7])|uint32(chunk.data[8])<<8|uint32(chunk.data[9])<<16)
			return width, height, nil
		case "VP8 ":
			if len(chunk.data) < 10 || chunk.data[3] != 0x9d || chunk.data[4] != 0x01 || chunk.data[5] != 0x2a {
				return 0, 0, ErrUnsupportedImage
			}
			width := int(binary.LittleEndian.Uint16(chunk.data[6:8]) & 0x3fff)
			height := int(binary.LittleEndian.Uint16(chunk.data[8:10]) & 0x3fff)
			return width, height, nil
		case "VP8L":
			if len(chunk.data) < 5 || chunk.data[0] != 0x2f {
				return 0, 0, ErrUnsupportedImage
			}
			bits := uint32(chunk.data[1]) |
				uint32(chunk.data[2])<<8 |
				uint32(chunk.data[3])<<16 |
				uint32(chunk.data[4])<<24
			width := int(bits&0x3fff) + 1
			height := int((bits>>14)&0x3fff) + 1
			return width, height, nil
		}
	}

	return 0, 0, ErrUnsupportedImage
}
