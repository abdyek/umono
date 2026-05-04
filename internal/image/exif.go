package image

import (
	"bytes"
	"encoding/binary"
)

func orientationFromBytes(data []byte, mimeType string) int {
	switch mimeType {
	case MimeJPEG:
		return jpegOrientation(data)
	case MimeWebP:
		chunks, ok := webpChunks(data)
		if !ok {
			return 1
		}
		for _, chunk := range chunks {
			if chunk.name == "EXIF" {
				return exifOrientation(chunk.data)
			}
		}
	}
	return 1
}

func jpegOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}

	for off := 2; off+4 <= len(data); {
		if data[off] != 0xff {
			return 1
		}
		marker := data[off+1]
		off += 2
		if marker == 0xda || marker == 0xd9 {
			return 1
		}
		if off+2 > len(data) {
			return 1
		}
		size := int(binary.BigEndian.Uint16(data[off : off+2]))
		if size < 2 || off+size > len(data) {
			return 1
		}
		segment := data[off+2 : off+size]
		if marker == 0xe1 && bytes.HasPrefix(segment, []byte("Exif\x00\x00")) {
			return exifOrientation(segment)
		}
		off += size
	}

	return 1
}

func exifOrientation(data []byte) int {
	data = bytes.TrimPrefix(data, []byte("Exif\x00\x00"))
	if len(data) < 8 {
		return 1
	}

	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}

	if order.Uint16(data[2:4]) != 42 {
		return 1
	}

	ifdOffset := int(order.Uint32(data[4:8]))
	if ifdOffset < 0 || ifdOffset+2 > len(data) {
		return 1
	}

	count := int(order.Uint16(data[ifdOffset : ifdOffset+2]))
	entryOffset := ifdOffset + 2
	for i := 0; i < count; i++ {
		off := entryOffset + i*12
		if off+12 > len(data) {
			return 1
		}
		tag := order.Uint16(data[off : off+2])
		if tag != 0x0112 {
			continue
		}

		fieldType := order.Uint16(data[off+2 : off+4])
		fieldCount := order.Uint32(data[off+4 : off+8])
		if fieldType != 3 || fieldCount < 1 {
			return 1
		}
		orientation := int(order.Uint16(data[off+8 : off+10]))
		if orientation < 1 || orientation > 8 {
			return 1
		}
		return orientation
	}

	return 1
}

func orientationSwapsDimensions(orientation int) bool {
	return orientation == 5 || orientation == 6 || orientation == 7 || orientation == 8
}
