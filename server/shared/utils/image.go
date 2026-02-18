package utils

import (
	"errors"
	"mime/multipart"
	"net/http"
)

const MaxAvatarSize = 1 * 1024 * 1024 // 2MB

// ValidateImage validates image file header for size and MIME type (JPG/PNG only)
func ValidateImage(header *multipart.FileHeader) error {
	if header.Size > MaxAvatarSize {
		return errors.New("image size must not exceed 1MB")
	}

	f, err := header.Open()
	if err != nil {
		return errors.New("failed to read image file")
	}
	defer f.Close()

	// Read first 512 bytes for MIME detection
	buf := make([]byte, 512)
	if _, err := f.Read(buf); err != nil {
		return errors.New("failed to read image content")
	}

	mimeType := http.DetectContentType(buf)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return errors.New("only JPG and PNG images are allowed")
	}

	return nil
}

// ImageContentType returns content type from first 512 bytes of a file
func ImageContentType(header *multipart.FileHeader) (string, error) {
	f, err := header.Open()
	if err != nil {
		return "", errors.New("failed to open image file")
	}
	defer f.Close()

	buf := make([]byte, 512)
	if _, err := f.Read(buf); err != nil {
		return "", errors.New("failed to read image content")
	}

	return http.DetectContentType(buf), nil
}