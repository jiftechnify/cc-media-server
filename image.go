package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"image"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
	webpenc "github.com/kolesa-team/go-webp/encoder"
)

var (
	supportedImageFormats = map[string]struct{}{
		"jpeg": {},
		"png":  {},
		"webp": {},
		"bmp":  {},
	}
	errUnsupportedImageFormat = fmt.Errorf("unsupported image format")
)

func validateImage(data io.Reader) (string, error) {
	_, format, err := image.DecodeConfig(data)
	if err != nil {
		log.Printf("failed to decode image header: %v", err)
		return "", errUnsupportedImageFormat
	}
	if _, ok := supportedImageFormats[format]; !ok {
		return "", errUnsupportedImageFormat
	}
	return format, nil
}

func processImage(data io.Reader) (*bytes.Reader, error) {
	imgFmt, err := validateImage(data)
	if err != nil {
		return nil, fmt.Errorf("image validation failed: %w", err)
	}

	// this strips EXIF data away from the image if it's a JPEG image
	img, err := imaging.Decode(data, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)

	}
	resized, err := resizeImage(img)
	if err != nil {
		return nil, fmt.Errorf("failed to resize image: %w", err)
	}
	return toLossyWebp(resized, imgFmt)
}

const (
	resizeNumPixelsThreshold = 1024 * 576
	resizeTargetSize         = 1024
)

func resizeImage(img image.Image) (image.Image, error) {
	var (
		targWidth  int
		targHeight int
	)
	size := img.Bounds().Size()
	log.Printf("image size: %dx%d", size.X, size.Y)

	if size.X*size.Y <= resizeNumPixelsThreshold {
		return img, nil
	}
	if size.X >= size.Y {
		if size.X <= resizeTargetSize {
			return img, nil
		}
		targWidth = resizeTargetSize
		targHeight = 0
	} else {
		if size.Y <= resizeTargetSize {
			return img, nil
		}
		targWidth = 0
		targHeight = resizeTargetSize
	}

	log.Println("start resizeimage")
	resized := imaging.Resize(img, targWidth, targHeight, imaging.Lanczos)
	log.Println("finish resizeImage")
	return resized, nil
}

func toLossyWebp(img image.Image, _ string) (*bytes.Reader, error) {
	opts, _ := webpenc.NewLossyEncoderOptions(webpenc.PresetDefault, 75)
	e, err := webpenc.NewEncoder(img, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize webp encoder: %w", err)
	}

	var buf bytes.Buffer
	log.Println("start encode to lossy webp")
	if err := e.Encode(&buf); err != nil {
		return nil, fmt.Errorf("failed to encode image to webp: %w", err)
	}
	log.Println("finish encode to lossy webp")
	return bytes.NewReader(buf.Bytes()), nil
}
