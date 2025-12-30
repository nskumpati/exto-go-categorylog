package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"

	"github.com/disintegration/imaging"
)

func GetThumbnailImage(filePath string) (string, error) {
	img, err := imaging.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Create thumbnail
	thumb := imaging.Thumbnail(img, 128, 128, imaging.CatmullRom)

	// Encode thumbnail to a buffer
	buf := new(bytes.Buffer)
	opts := &jpeg.Options{Quality: 60}
	err = jpeg.Encode(buf, thumb, opts)
	if err != nil {
		return "", fmt.Errorf("failed to encode thumbnail to buffer: %w", err)
	}

	// Base64 encode the buffer
	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/jpeg;base64," + base64String, nil
}

// get base64 string from file path
func GetBase64FromFilePath(filePath string) (string, error) {
	img, err := imaging.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	// Encode image to a buffer
	buf := new(bytes.Buffer)
	opts := &jpeg.Options{Quality: 80}
	err = jpeg.Encode(buf, img, opts)
	if err != nil {
		return "", fmt.Errorf("failed to encode image to buffer: %w", err)
	}

	// Base64 encode the buffer
	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/jpeg;base64," + base64String, nil
}

// get image from base64 string
func GetImageFromBase64(base64String string) ([]byte, error) {
	// Remove the data URL prefix if present
	if idx := bytes.Index([]byte(base64String), []byte(",")); idx != -1 {
		base64String = base64String[idx+1:]
	}

	// Decode the base64 string
	imageData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	return imageData, nil
}
