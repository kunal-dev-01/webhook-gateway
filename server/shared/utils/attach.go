package utils

import (
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// ============================================================================
// Allowed MIME types for support attachments
// ============================================================================

var allowedAttachmentMIMEs = map[string]bool{
	// Images
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	// Documents
	"application/pdf":                                                          true,
	"application/msword":                                                       true, // .doc
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
	"application/vnd.ms-excel":                                                true, // .xls
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true, // .xlsx
	"text/plain": true,
	"text/csv":   true,
	// Archives
	"application/zip":  true,
	"application/gzip": true,
}

// allowedAttachmentExtensions — fallback when magic bytes detection is insufficient.
// Needed for zip-based formats (.docx, .xlsx) which http.DetectContentType misidentifies.
var allowedAttachmentExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".zip":  "application/zip",
	".gz":   "application/gzip",
}

// ============================================================================
// AttachmentContentType
// ============================================================================

// AttachmentContentType detects and validates the content type of a support ticket attachment.
// Uses magic bytes (first 512 bytes) as primary detection, with file extension as fallback
// for zip-based formats like .docx and .xlsx that http.DetectContentType misidentifies.
// Returns a validated MIME type string, or an error if the type is not permitted.
func AttachmentContentType(header *multipart.FileHeader) (string, error) {
	f, err := header.Open()
	if err != nil {
		return "", errors.New("failed to open file")
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return "", errors.New("failed to read file content")
	}

	detected := http.DetectContentType(buf[:n])

	// http.DetectContentType returns "application/octet-stream" or "application/zip"
	// for many binary/Office formats — fall back to extension lookup
	if detected == "application/octet-stream" || detected == "application/zip" {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if mimeType, ok := allowedAttachmentExtensions[ext]; ok {
			return mimeType, nil
		}
	}

	// Strip optional charset (e.g. "text/plain; charset=utf-8" → "text/plain")
	mimeType := strings.TrimSpace(strings.Split(detected, ";")[0])

	if !allowedAttachmentMIMEs[mimeType] {
		return "", errors.New("file type not allowed — accepted: images (JPG/PNG/GIF/WebP), PDF, Word, Excel, CSV, text, ZIP")
	}

	return mimeType, nil
}