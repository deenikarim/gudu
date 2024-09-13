package mails

import (
	"encoding/base64"
	"errors"
	"mime"
	"os"
	"path/filepath"
)

// Attachment represents an email attachment
type Attachment struct {
	Name     string
	Data     []byte
	MimeType string
	Inline   bool
}

// NewAttachmentFromFile creates a new attachment from a file path
func NewAttachmentFromFile(filePath string, inline bool) (*Attachment, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(filePath)
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &Attachment{
		Name:     filename,
		Data:     data,
		MimeType: mimeType,
		Inline:   inline,
	}, nil
}

// NewAttachmentFromBase64 creates a new attachment from a base64 string
func NewAttachmentFromBase64(name, b64Data, mimeType string, inline bool) (*Attachment, error) {
	data, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, err
	}

	if mimeType == "" {
		return nil, errors.New("mimeType is required for base64 attachment")
	}

	return &Attachment{
		Name:     name,
		Data:     data,
		MimeType: mimeType,
		Inline:   inline,
	}, nil
}

// NewAttachmentFromBytes creates a new attachment from a byte slice
func NewAttachmentFromBytes(name string, data []byte, mimeType string, inline bool) (*Attachment, error) {
	if mimeType == "" {
		return nil, errors.New("mimeType is required for byte slice attachment")
	}

	return &Attachment{
		Name:     name,
		Data:     data,
		MimeType: mimeType,
		Inline:   inline,
	}, nil
}
