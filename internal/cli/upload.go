package cli

import (
	"net/http"
	"path/filepath"
	"strings"
)

const uploadFlag = "upload"

// Emporix answers a larger request with 413 Request Entity Too Large,
// whatever part makes it so.
const maxUploadBytes = 30 << 20

const uploadLimitHelp = "up to 30 MB including the form"

func uploadPart(name, usage string) formPart {
	part := filePartOf(name, bodyFile{flag: uploadFlag, limit: maxUploadBytes}, usage, contentTypeOf)
	part.bound = maxUploadBytes
	return part
}

func contentTypeOf(fileName string, content []byte) (string, error) {
	if byExtension, known := uploadTypes()[strings.ToLower(filepath.Ext(fileName))]; known {
		return byExtension, nil
	}
	return http.DetectContentType(content), nil
}

func uploadTypes() map[string]string {
	return map[string]string{
		".avif": "image/avif", ".gif": "image/gif", ".jpeg": "image/jpeg", ".jpg": "image/jpeg", ".png": "image/png",
		".svg": "image/svg+xml", ".webp": "image/webp",
		".csv": "text/csv", ".json": "application/json", ".pdf": "application/pdf", ".text": "text/plain",
		".txt": "text/plain", ".xml": "application/xml", ".zip": "application/zip",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".mp3":  "audio/mpeg", ".mp4": "video/mp4", ".webm": "video/webm",
	}
}
