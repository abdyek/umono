package media

import (
	"encoding/binary"
	"errors"
	"image"
	"io"

	_ "image/jpeg"
	_ "image/png"
)

var ErrInvalidImageDimensions = errors.New("invalid image dimensions")

func DimensionsFromReader(mimeType string, r io.Reader) (int, int, error) {
	switch mimeType {
	case "image/jpeg", "image/png":
		cfg, _, err := image.DecodeConfig(r)
		if err != nil {
			return 0, 0, err
		}
		return cfg.Width, cfg.Height, nil
	case "image/webp":
		return decodeWEBPDimensions(r)
	default:
		return 0, 0, ErrInvalidImageDimensions
	}
}

func decodeWEBPDimensions(r io.Reader) (int, int, error) {
	header := make([]byte, 12)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, 0, err
	}

	if string(header[:4]) != "RIFF" || string(header[8:12]) != "WEBP" {
		return 0, 0, ErrInvalidImageDimensions
	}

	for {
		chunkHeader := make([]byte, 8)
		if _, err := io.ReadFull(r, chunkHeader); err != nil {
			return 0, 0, err
		}

		chunkType := string(chunkHeader[:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])
		chunkData := make([]byte, chunkSize)
		if _, err := io.ReadFull(r, chunkData); err != nil {
			return 0, 0, err
		}

		if chunkSize%2 == 1 {
			if _, err := io.CopyN(io.Discard, r, 1); err != nil {
				return 0, 0, err
			}
		}

		switch chunkType {
		case "VP8X":
			if len(chunkData) < 10 {
				return 0, 0, ErrInvalidImageDimensions
			}
			width := 1 + int(uint32(chunkData[4])|uint32(chunkData[5])<<8|uint32(chunkData[6])<<16)
			height := 1 + int(uint32(chunkData[7])|uint32(chunkData[8])<<8|uint32(chunkData[9])<<16)
			return width, height, nil
		case "VP8 ":
			if len(chunkData) < 10 {
				return 0, 0, ErrInvalidImageDimensions
			}
			if chunkData[3] != 0x9d || chunkData[4] != 0x01 || chunkData[5] != 0x2a {
				return 0, 0, ErrInvalidImageDimensions
			}
			width := int(binary.LittleEndian.Uint16(chunkData[6:8]) & 0x3fff)
			height := int(binary.LittleEndian.Uint16(chunkData[8:10]) & 0x3fff)
			return width, height, nil
		case "VP8L":
			if len(chunkData) < 5 || chunkData[0] != 0x2f {
				return 0, 0, ErrInvalidImageDimensions
			}
			bits := uint32(chunkData[1]) |
				uint32(chunkData[2])<<8 |
				uint32(chunkData[3])<<16 |
				uint32(chunkData[4])<<24
			width := int(bits&0x3FFF) + 1
			height := int((bits>>14)&0x3FFF) + 1
			return width, height, nil
		}
	}
}
