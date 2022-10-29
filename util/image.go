package util

import (
	"bytes"
	"image/jpeg"
	"image/png"
)

const (
	ImgFormatUnknown = ""
	ImgFormatJPEG    = "jpeg"
	ImgFormatPNG     = "png"
)

// ParseRawImageFormat parse image in jpeg or png format
func ParseRawImageFormat(img []byte) string {
	format := ImgFormatJPEG
	_, err := jpeg.Decode(bytes.NewReader(img))
	if err != nil {
		format = ImgFormatPNG
		_, err = png.Decode(bytes.NewReader(img))
		if err != nil {
			return ImgFormatUnknown
		}
	}
	return format
}
