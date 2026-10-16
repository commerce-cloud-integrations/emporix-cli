package cli

import (
	"net/http"
	"path/filepath"
	"strings"
)

// uploadFlag is the flag of a write that uploads a file of any kind, such as
// the file of a media asset (uploadPart).
const uploadFlag = "upload"

// maxUploadBytes bounds the request of an upload, the whole form that sends
// the file (the bound of uploadPart), and so the file that it reads: Emporix
// answers a larger request with 413 Request Entity Too Large, whatever part
// makes it so. uploadLimitHelp states it in help.
const maxUploadBytes = 30 << 20

// uploadLimitHelp states maxUploadBytes in the help of a file to upload.
const uploadLimitHelp = "up to 30 MB including the form"

// uploadPart is the part name of a form that sends the file that --upload
// (described by usage) names, when it is given, as filePartOf sends a file:
// whatever it holds, up to maxUploadBytes, with its content type
// (contentTypeOf), in a form of up to maxUploadBytes (its bound).
func uploadPart(name, usage string) formPart {
	part := filePartOf(name, bodyFile{flag: uploadFlag, limit: maxUploadBytes}, usage, contentTypeOf)
	part.bound = maxUploadBytes
	return part
}

// contentTypeOf returns the content type of a file to upload named fileName:
// the one that its extension names (uploadTypes), else the one that its
// content shows (application/octet-stream when it shows none). It is the same
// on every system, so a dry run shows what is sent anywhere. Any file can be
// sent: it never fails.
func contentTypeOf(fileName string, content []byte) (string, error) {
	if byExtension, known := uploadTypes()[strings.ToLower(filepath.Ext(fileName))]; known {
		return byExtension, nil
	}
	return http.DetectContentType(content), nil
}

// uploadTypes returns the content types of files to upload by their
// extension, lower-cased: those of images, documents, video and audio that a
// shop stores as media.
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
